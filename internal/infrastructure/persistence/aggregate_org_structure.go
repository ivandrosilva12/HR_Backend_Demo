package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"rhapp/internal/domain/agregados"
	"rhapp/internal/domain/entities"

	"github.com/google/uuid"
)

type orgStructureAggregatePgRepository struct {
	db *sql.DB
}

func NewOrgStructureAggregatePgRepository(db *sql.DB) *orgStructureAggregatePgRepository {
	return &orgStructureAggregatePgRepository{db: db}
}

func (r *orgStructureAggregatePgRepository) GetByDepartmentID(ctx context.Context, deptID uuid.UUID) (*agregados.OrgStructureAggregate, error) {
	var dept entities.Department
	query := `SELECT id, nome, created_at, updated_at FROM departments WHERE id = $1`
	err := r.db.QueryRowContext(ctx, query, deptID).Scan(&dept.ID, &dept.Nome, &dept.CreatedAt, &dept.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("departamento com ID %s não encontrado", deptID)
		}
		return nil, fmt.Errorf("erro ao buscar departamento: %w", err)
	}

	agg := &agregados.OrgStructureAggregate{
		Department: dept,
	}

	// Posições
	posQuery := `SELECT id, nome, department_id, created_at, updated_at FROM positions WHERE department_id = $1`
	posRows, err := r.db.QueryContext(ctx, posQuery, deptID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar posições: %w", err)
	}
	defer posRows.Close()

	for posRows.Next() {
		var pos entities.Position
		if scanErr := posRows.Scan(&pos.ID, &pos.Nome, &pos.DepartmentID, &pos.CreatedAt, &pos.UpdatedAt); scanErr != nil {
			return nil, fmt.Errorf("erro ao escanear posição: %w", scanErr)
		}
		agg.Positions = append(agg.Positions, pos)
	}
	if err = posRows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar posições: %w", err)
	}

	// Funcionários
	empQuery := `
		SELECT id, employee_number, full_name, gender, date_of_birth, nationality, marital_status,
		       phone_number, email, bi, iban, department_id, position_id, address, district_id,
		       hiring_date, contract_type, salary, social_security, supervisor_id,
		       is_active, created_at, updated_at
		FROM employees
		WHERE department_id = $1
	`
	empRows, err := r.db.QueryContext(ctx, empQuery, deptID)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar funcionários: %w", err)
	}
	defer empRows.Close()

	for empRows.Next() {
		var emp entities.Employee
		var supervisorID sql.NullString

		if scanErr := empRows.Scan(
			&emp.ID,
			&emp.EmployeeNumber,
			&emp.FullName,
			&emp.Gender,
			&emp.DateOfBirth,
			&emp.Nationality,
			&emp.MaritalStatus,
			&emp.PhoneNumber,
			&emp.Email,
			&emp.BI,
			&emp.IBAN,
			&emp.DepartmentID,
			&emp.PositionID,
			&emp.Address,
			&emp.DistrictID,
			&emp.HiringDate,
			&emp.ContractType,
			&emp.Salary,
			&emp.SocialSecurity,
			&supervisorID,
			&emp.IsActive,
			&emp.CreatedAt,
			&emp.UpdatedAt,
		); scanErr != nil {
			return nil, fmt.Errorf("erro ao escanear funcionário: %w", scanErr)
		}

		agg.Employees = append(agg.Employees, emp)
	}
	if err = empRows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar funcionários: %w", err)
	}

	return agg, nil
}

var _ agregados.OrgStructureAggregateRepository = (*orgStructureAggregatePgRepository)(nil)
