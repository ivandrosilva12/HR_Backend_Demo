package persistence

import (
	"context"
	"database/sql"
	"errors"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/domain/vos"

	"github.com/google/uuid"
)

type employeeStatusPgRepository struct {
	db *sql.DB
}

func NewEmployeeStatusPgRepository(db *sql.DB) repos.EmployeeStatusRepository {
	return &employeeStatusPgRepository{db: db}
}

func (r *employeeStatusPgRepository) Create(ctx context.Context, s entities.EmployeeStatus) error {
	query := `
		INSERT INTO employee_statuses (
			id, employee_id, status, reason,
			start_date, end_date, is_current,
			created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`
	_, err := r.db.ExecContext(ctx, query,
		s.ID, s.EmployeeID, s.Status.String(), s.Reason.String(),
		s.StartDate, s.EndDate, s.IsCurrent,
		s.CreatedAt, s.UpdatedAt,
	)
	return err
}

func (r *employeeStatusPgRepository) Update(ctx context.Context, s entities.EmployeeStatus) error {
	query := `
		UPDATE employee_statuses SET
			status = $1, reason = $2, start_date = $3,
			end_date = $4, is_current = $5, updated_at = $6
		WHERE id = $7`
	res, err := r.db.ExecContext(ctx, query,
		s.Status.String(), s.Reason.String(), s.StartDate,
		s.EndDate, s.IsCurrent, s.UpdatedAt,
		s.ID,
	)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *employeeStatusPgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM employee_statuses WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *employeeStatusPgRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.EmployeeStatus, error) {
	query := `
        SELECT id, employee_id, status, reason,
               start_date, end_date, is_current,
               created_at, updated_at
        FROM employee_statuses
        WHERE id = $1`
	var s entities.EmployeeStatus
	var statusStr, reasonStr string
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&s.ID, &s.EmployeeID, &statusStr, &reasonStr,
		&s.StartDate, &s.EndDate, &s.IsCurrent,
		&s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.EmployeeStatus{}, sql.ErrNoRows
		}
		return entities.EmployeeStatus{}, err
	}
	if v, err := vos.NewEmployeeStatusValue(statusStr); err == nil {
		s.Status = v
	} else {
		return entities.EmployeeStatus{}, err
	}

	if v, err := vos.NewStatusReason(reasonStr); err == nil {
		s.Reason = v
	} else {
		return entities.EmployeeStatus{}, err
	}

	return s, nil
}

func (r *employeeStatusPgRepository) FindAllByEmployee(ctx context.Context, employeeID uuid.UUID, limit, offset int) ([]entities.EmployeeStatus, error) {
	query := `
		SELECT id, employee_id, status, reason,
		       start_date, end_date, is_current,
		       created_at, updated_at
		FROM employee_statuses
		WHERE employee_id = $1
		ORDER BY start_date DESC
		LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, employeeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var statusStr, reasonStr string
	var list []entities.EmployeeStatus
	for rows.Next() {
		var s entities.EmployeeStatus
		if err := rows.Scan(
			&s.ID, &s.EmployeeID, &statusStr, &reasonStr,
			&s.StartDate, &s.EndDate, &s.IsCurrent,
			&s.CreatedAt, &s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		if v, err := vos.NewEmployeeStatusValue(statusStr); err == nil {
			s.Status = v
		} else {
			return []entities.EmployeeStatus{}, err
		}

		if v, err := vos.NewStatusReason(reasonStr); err == nil {
			s.Reason = v
		} else {
			return []entities.EmployeeStatus{}, err
		}
		list = append(list, s)
	}
	return list, nil
}

func (r *employeeStatusPgRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM employee_statuses WHERE id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	return exists, err
}

var _ repos.EmployeeStatusRepository = (*employeeStatusPgRepository)(nil)
