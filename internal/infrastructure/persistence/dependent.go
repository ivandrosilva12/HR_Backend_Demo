package persistence

import (
	"context"
	"database/sql"
	"errors"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/domain/vos"
	"time"

	"github.com/google/uuid"
)

type dependentPgRepository struct {
	db *sql.DB
}

func NewDependentPgRepository(db *sql.DB) *dependentPgRepository {
	return &dependentPgRepository{db: db}
}

func (r *dependentPgRepository) Create(ctx context.Context, d entities.Dependent) error {
	query := `
		INSERT INTO dependents (
			id, employee_id, full_name, relationship, gender, date_of_birth, is_active, created_at, updated_at
		) VALUES (
			$1,$2,$3,$4,$5,$6,$7,$8,$9
		)
	`
	_, err := r.db.ExecContext(ctx, query,
		d.ID, d.EmployeeID, d.FullName.String(), d.Relationship.String(), d.Gender.String(), d.DateOfBirth.String(),
		d.IsActive, d.CreatedAt, d.UpdatedAt,
	)
	return err
}

func (r *dependentPgRepository) Update(ctx context.Context, d entities.Dependent) error {
	query := `
		UPDATE dependents SET
			full_name = $1, relationship = $2, gender = $3,
			date_of_birth = $4, is_active = $5, updated_at = $6
		WHERE id = $7
	`
	res, err := r.db.ExecContext(ctx, query,
		d.FullName.String(), d.Relationship.String(), d.Gender.String(),
		d.DateOfBirth.String(), d.IsActive,
		d.UpdatedAt, d.ID,
	)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *dependentPgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	res, err := r.db.ExecContext(ctx, `DELETE FROM dependents WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *dependentPgRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.Dependent, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, employee_id, full_name, relationship, gender, date_of_birth, is_active, created_at, updated_at
		FROM dependents WHERE id = $1`, id)
	d, err := scanDependentFromRow(row)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.Dependent{}, sql.ErrNoRows
		}
		return entities.Dependent{}, err
	}
	return d, nil
}

func (r *dependentPgRepository) FindAllByEmployee(ctx context.Context, empID uuid.UUID, limit, offset int) ([]entities.Dependent, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, full_name, relationship, gender, date_of_birth, is_active, created_at, updated_at
		FROM dependents
		WHERE employee_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`, empID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []entities.Dependent
	for rows.Next() {
		d, err := scanDependentFromRow(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, rows.Err()
}

func (r *dependentPgRepository) Search(ctx context.Context, searchText, filter string, limit, offset int, empID *uuid.UUID) ([]entities.Dependent, error) {
	rows, err := r.db.QueryContext(ctx,
		`SELECT id, employee_id, full_name, relationship, gender, date_of_birth, is_active, created_at, updated_at
		 FROM search_dependents($1, $2, $3, $4)`,
		searchText, empID, limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []entities.Dependent
	for rows.Next() {
		d, err := scanDependentFromRow(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, d)
	}
	return list, rows.Err()
}

func (r *dependentPgRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM dependents WHERE id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	return exists, err
}

func scanDependentFromRow(scanner interface {
	Scan(dest ...any) error
}) (entities.Dependent, error) {
	var d entities.Dependent

	var (
		fullNameStr  string
		relStr       string
		genderStr    string
		birthDateStr time.Time
	)

	if err := scanner.Scan(
		&d.ID, &d.EmployeeID, &fullNameStr, &relStr, &genderStr, &birthDateStr,
		&d.IsActive, &d.CreatedAt, &d.UpdatedAt,
	); err != nil {
		return d, err
	}

	var err error
	if d.FullName, err = vos.NewPersonalName(fullNameStr); err != nil {
		return d, err
	}
	if d.Relationship, err = vos.NewRelationshipType(relStr); err != nil {
		return d, err
	}
	if d.Gender, err = vos.NewGender(genderStr); err != nil {
		return d, err
	}
	if d.DateOfBirth, err = vos.NewBirthDate(birthDateStr.Format("2006-01-02")); err != nil {
		return d, err
	}

	return d, nil
}

var _ repos.DependentRepository = (*dependentPgRepository)(nil)
