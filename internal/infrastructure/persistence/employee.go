package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/domain/vos"
	"rhapp/internal/utils"

	"github.com/google/uuid"
)

type employeePgRepository struct {
	db *sql.DB
}

func NewEmployeePgRepository(db *sql.DB) *employeePgRepository {
	return &employeePgRepository{db: db}
}

func (r *employeePgRepository) Create(ctx context.Context, e entities.Employee) error {
	query := `
		INSERT INTO employees (
			id, full_name, gender, date_of_birth, nationality, marital_status,
			phone_number, email, bi, id_date, iban, department_id, position_id, address, district_id,
			hiring_date, contract_type, salary, social_security,
			is_active, created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5, $6, $7,
			$8, $9, $10, $11, $12, $13, $14, $15,
			$16, $17, $18, $19, $20, $21, $22
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		e.ID,
		e.FullName.String(),
		e.Gender.String(),
		e.DateOfBirth,
		e.Nationality.String(),
		e.MaritalStatus.String(),
		e.PhoneNumber.String(),
		e.Email.String(),
		e.BI.String(),
		e.IDValidationDate,
		e.IBAN.String(),
		e.DepartmentID,
		e.PositionID,
		e.Address.String(),
		e.DistrictID,
		e.HiringDate,
		e.ContractType.String(),
		e.Salary.Float64(),
		e.SocialSecurity.String(),
		e.IsActive,
		e.CreatedAt,
		e.UpdatedAt,
	)
	return err
}

func (r *employeePgRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.Employee, error) {
	query := `SELECT 
		id, employee_number, full_name, gender, date_of_birth, nationality, marital_status,
		phone_number, email, bi, id_date, iban, department_id, position_id, address, district_id,
		hiring_date, contract_type, salary, social_security,
		is_active, created_at, updated_at
		FROM employees WHERE id = $1`

	row := r.db.QueryRowContext(ctx, query, id)
	return scanEmployeeFromRow(row)
}

func (r *employeePgRepository) GetMaxHeadCount(ctx context.Context, id, positionID uuid.UUID) (int, error) {

	query := `SELECT p.max_headcount, COUNT(e.id) FILTER (WHERE e.is_active) AS
			current_headcount,
         	p.department_id
  			FROM positions p LEFT JOIN employees e ON e.position_id = p.id
			WHERE p.id = $1 GROUP BY p.id`
	var cap, curr int
	var posDept uuid.UUID
	err := r.db.QueryRowContext(ctx, query, positionID).Scan(&cap, &curr, &posDept)
	if err != nil {
		return 0, err
	}

	return cap, nil
}

func (r *employeePgRepository) Update(ctx context.Context, e entities.Employee) error {
	query := `
		UPDATE employees SET
			marital_status = $2,
			phone_number   = $3,
			email          = $4,
			iban           = $5,
			department_id  = $6,
			position_id    = $7,
			address        = $8,
			district_id    = $9,
			contract_type  = $10,
			salary         = $11,
			is_active      = $12,
			updated_at     = $13
		WHERE id = $1
	`
	_, err := r.db.ExecContext(ctx, query,
		e.ID,
		e.MaritalStatus.String(),
		e.PhoneNumber.String(),
		e.Email.String(),
		e.IBAN.String(),
		e.DepartmentID,
		e.PositionID,
		e.Address.String(),
		e.DistrictID,
		e.ContractType.String(),
		e.Salary.Float64(),
		e.IsActive,
		e.UpdatedAt,
	)
	return err
}

func (r *employeePgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM employees WHERE id = $1`, id)
	return err
}

func (r *employeePgRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	var count int
	err := r.db.QueryRowContext(ctx, `SELECT COUNT(1) FROM employees WHERE id = $1`, id).Scan(&count)
	return count > 0, err
}

/*
===========================

	LIST (PagedResponse)
	===========================
*/
func (r *employeePgRepository) List(ctx context.Context, limit, offset int) (utils.PagedResponse[entities.Employee], error) {
	const countQ = `SELECT COUNT(*) FROM employees`
	const pageQ = `
		SELECT 
			id, employee_number, full_name, gender, date_of_birth, nationality, marital_status,
			phone_number, email, bi, id_date, iban, department_id, position_id, address, district_id,
			hiring_date, contract_type, salary, social_security, is_active, created_at, updated_at
		FROM employees
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2;
	`

	var total int
	if err := r.db.QueryRowContext(ctx, countQ).Scan(&total); err != nil {
		return utils.PagedResponse[entities.Employee]{Items: []entities.Employee{}, Total: 0, Limit: limit, Offset: offset}, err
	}

	rows, err := r.db.QueryContext(ctx, pageQ, limit, offset)
	if err != nil {
		return utils.PagedResponse[entities.Employee]{Items: []entities.Employee{}, Total: total, Limit: limit, Offset: offset}, err
	}
	defer rows.Close()

	items := make([]entities.Employee, 0, limit)
	for rows.Next() {
		e, err := scanEmployeeFromRow(rows)
		if err != nil {
			return utils.PagedResponse[entities.Employee]{Items: []entities.Employee{}, Total: total, Limit: limit, Offset: offset}, err
		}
		items = append(items, e)
	}
	if err := rows.Err(); err != nil {
		return utils.PagedResponse[entities.Employee]{Items: []entities.Employee{}, Total: total, Limit: limit, Offset: offset}, err
	}

	return utils.PagedResponse[entities.Employee]{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

/*
===========================

	SEARCH (PagedResponse)
	- searchText: aplica a full_name, email e employee_number (cast p/ texto)
	- filter: aceita tokens simples: "dept:<uuid>", "pos:<uuid>", "ativo:true|false"
	===========================
*/
func (r *employeePgRepository) Search(ctx context.Context, searchText, filter string, limit, offset int) (utils.PagedResponse[entities.Employee], error) {
	deptID, posID, activePtr := parseAdvancedFilter(filter)

	var conds []string
	var args []any
	argi := 1

	if strings.TrimSpace(searchText) != "" {
		pattern := "%" + searchText + "%"
		conds = append(conds, fmt.Sprintf("(full_name ILIKE $%d OR email ILIKE $%d OR CAST(employee_number AS TEXT) ILIKE $%d)", argi, argi, argi))
		args = append(args, pattern)
		argi++
	}
	if deptID != nil {
		conds = append(conds, fmt.Sprintf("department_id = $%d", argi))
		args = append(args, *deptID)
		argi++
	}
	if posID != nil {
		conds = append(conds, fmt.Sprintf("position_id = $%d", argi))
		args = append(args, *posID)
		argi++
	}
	if activePtr != nil {
		conds = append(conds, fmt.Sprintf("is_active = $%d", argi))
		args = append(args, *activePtr)
		argi++
	}

	where := "TRUE"
	if len(conds) > 0 {
		where = strings.Join(conds, " AND ")
	}

	// COUNT
	countQ := "SELECT COUNT(*) FROM employees WHERE " + where
	var total int
	if err := r.db.QueryRowContext(ctx, countQ, args...).Scan(&total); err != nil {
		return utils.PagedResponse[entities.Employee]{Items: []entities.Employee{}, Total: 0, Limit: limit, Offset: offset}, err
	}

	// PAGE
	pageQ := fmt.Sprintf(`
		SELECT 
			id, employee_number, full_name, gender, date_of_birth, nationality, marital_status,
			phone_number, email, bi, id_date, iban, department_id, position_id, address, district_id,
			hiring_date, contract_type, salary, social_security,
			is_active, created_at, updated_at
		FROM employees
		WHERE %s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, where, argi, argi+1)

	argsPage := append(args, limit, offset)

	rows, err := r.db.QueryContext(ctx, pageQ, argsPage...)
	if err != nil {
		return utils.PagedResponse[entities.Employee]{Items: []entities.Employee{}, Total: total, Limit: limit, Offset: offset}, err
	}
	defer rows.Close()

	items := make([]entities.Employee, 0, limit)
	for rows.Next() {
		e, err := scanEmployeeFromRow(rows)
		if err != nil {
			return utils.PagedResponse[entities.Employee]{Items: []entities.Employee{}, Total: total, Limit: limit, Offset: offset}, err
		}
		items = append(items, e)
	}
	if err := rows.Err(); err != nil {
		return utils.PagedResponse[entities.Employee]{Items: []entities.Employee{}, Total: total, Limit: limit, Offset: offset}, err
	}

	return utils.PagedResponse[entities.Employee]{
		Items:  items,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}, nil
}

// parseAdvancedFilter interpreta "dept:<uuid>", "pos:<uuid>", "ativo:true|false"
func parseAdvancedFilter(filter string) (*uuid.UUID, *uuid.UUID, *bool) {
	filter = strings.TrimSpace(filter)
	if filter == "" {
		return nil, nil, nil
	}

	var deptID *uuid.UUID
	var posID *uuid.UUID
	var active *bool

	parts := strings.Fields(filter)
	for _, p := range parts {
		if strings.HasPrefix(p, "dept:") {
			raw := strings.TrimPrefix(p, "dept:")
			if id, err := uuid.Parse(raw); err == nil {
				deptID = &id
			}
			continue
		}
		if strings.HasPrefix(p, "pos:") {
			raw := strings.TrimPrefix(p, "pos:")
			if id, err := uuid.Parse(raw); err == nil {
				posID = &id
			}
			continue
		}
		if strings.HasPrefix(p, "ativo:") {
			raw := strings.TrimPrefix(p, "ativo:")
			switch strings.ToLower(raw) {
			case "true", "1", "t", "yes", "sim":
				v := true
				active = &v
			case "false", "0", "f", "no", "nao", "não":
				v := false
				active = &v
			}
			continue
		}
	}
	return deptID, posID, active
}

func (r *employeePgRepository) ListOld(ctx context.Context, limit, offset int) ([]entities.Employee, error) {
	// (Apenas para referência: método antigo — não use)
	return nil, fmt.Errorf("deprecated")
}

func (r *employeePgRepository) SearchOld(ctx context.Context, searchText, departmentFilter string, limit, offset int) ([]entities.Employee, error) {
	// (Apenas para referência: método antigo — não use)
	return nil, fmt.Errorf("deprecated")
}

func scanEmployeeFromRow(scanner interface {
	Scan(dest ...any) error
}) (entities.Employee, error) {
	var e entities.Employee

	var (
		fullNameStr, genderStr, nationalityStr, maritalStr,
		phoneStr, emailStr, biStr, ibanStr, addressStr,
		contractStr, securityStr string
		salaryFloat float64
	)

	err := scanner.Scan(
		&e.ID, &e.EmployeeNumber, &fullNameStr, &genderStr, &e.DateOfBirth,
		&nationalityStr, &maritalStr, &phoneStr, &emailStr, &biStr, &e.IDValidationDate, &ibanStr,
		&e.DepartmentID, &e.PositionID, &addressStr, &e.DistrictID, &e.HiringDate,
		&contractStr, &salaryFloat, &securityStr,
		&e.IsActive, &e.CreatedAt, &e.UpdatedAt,
	)
	if err != nil {
		return e, err
	}

	var convErr error
	if e.FullName, convErr = vos.NewPersonalName(fullNameStr); convErr != nil {
		return e, convErr
	}
	if e.Gender, convErr = vos.NewGender(genderStr); convErr != nil {
		return e, convErr
	}
	if e.Nationality, convErr = vos.NewNationality(nationalityStr); convErr != nil {
		return e, convErr
	}
	if e.MaritalStatus, convErr = vos.NewMaritalStatus(maritalStr); convErr != nil {
		return e, convErr
	}
	if e.PhoneNumber, convErr = vos.NewPhoneNumber(phoneStr); convErr != nil {
		return e, convErr
	}
	if e.Email, convErr = vos.NewEmail(emailStr); convErr != nil {
		return e, convErr
	}
	if e.BI, convErr = vos.NewBI(biStr); convErr != nil {
		return e, convErr
	}
	if e.IBAN, convErr = vos.NewIBAN(ibanStr); convErr != nil {
		return e, convErr
	}
	if e.Address, convErr = vos.NewAddress(addressStr); convErr != nil {
		return e, convErr
	}
	if e.ContractType, convErr = vos.NewContractType(contractStr); convErr != nil {
		return e, convErr
	}
	if e.Salary, convErr = vos.NewSalary(salaryFloat); convErr != nil {
		return e, convErr
	}
	if e.SocialSecurity, convErr = vos.NewSocialSecurity(securityStr); convErr != nil {
		return e, convErr
	}

	return e, nil
}

var _ repos.EmployeeRepository = (*employeePgRepository)(nil)
