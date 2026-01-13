package persistence

import (
	"context"
	"database/sql"
	"fmt"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type workHistoryPgRepository struct {
	db *sql.DB
}

func NewWorkHistoryPgRepository(db *sql.DB) repos.WorkHistoryRepository {
	return &workHistoryPgRepository{db: db}
}

func (r *workHistoryPgRepository) Create(ctx context.Context, wh entities.WorkHistory) error {
	query := `
		INSERT INTO work_histories (
			id, employee_id, company, position, start_date, end_date, responsibilities, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	_, err := r.db.ExecContext(ctx, query,
		wh.ID, wh.EmployeeID, wh.Company, wh.Position,
		wh.StartDate, wh.EndDate, wh.Responsibilities,
	)
	return err
}

func (r *workHistoryPgRepository) Update(ctx context.Context, wh entities.WorkHistory) error {
	query := `
		UPDATE work_histories SET
			company = $1, position = $2, start_date = $3, end_date = $4,
			responsibilities = $5, updated_at = NOW()
		WHERE id = $6
	`
	_, err := r.db.ExecContext(ctx, query,
		wh.Company, wh.Position, wh.StartDate, wh.EndDate,
		wh.Responsibilities, wh.ID,
	)
	return err
}

func (r *workHistoryPgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.ExecContext(ctx, `DELETE FROM work_histories WHERE id = $1`, id)
	return err
}

func (r *workHistoryPgRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.WorkHistory, error) {
	query := `
		SELECT id, employee_id, company, position, start_date, end_date,
		       responsibilities, created_at, updated_at
		FROM work_histories WHERE id = $1
	`
	var wh entities.WorkHistory
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&wh.ID, &wh.EmployeeID, &wh.Company, &wh.Position,
		&wh.StartDate, &wh.EndDate, &wh.Responsibilities,
		&wh.CreatedAt, &wh.UpdatedAt,
	)
	return wh, err
}

func (r *workHistoryPgRepository) ListByEmployee(ctx context.Context, employeeID uuid.UUID, limit, offset int) ([]entities.WorkHistory, error) {
	query := `
		SELECT id, employee_id, company, position, start_date, end_date,
		       responsibilities, created_at, updated_at
		FROM work_histories
		WHERE employee_id = $1
		ORDER BY start_date DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.QueryContext(ctx, query, employeeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []entities.WorkHistory
	for rows.Next() {
		var wh entities.WorkHistory
		if err := rows.Scan(
			&wh.ID, &wh.EmployeeID, &wh.Company, &wh.Position,
			&wh.StartDate, &wh.EndDate, &wh.Responsibilities,
			&wh.CreatedAt, &wh.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, wh)
	}
	return list, nil
}

func (r *workHistoryPgRepository) Search(ctx context.Context, employeeID uuid.UUID, searchText string, startDate, endDate *string, limit, offset int) ([]entities.WorkHistory, error) {
	query := `
		SELECT id, employee_id, company, position, start_date, end_date,
		       responsibilities, created_at, updated_at
		FROM work_histories
		WHERE employee_id = $1
		AND (LOWER(company) LIKE LOWER($2))
	`
	params := []interface{}{employeeID, "%" + searchText + "%"}
	paramIndex := 3

	if startDate != nil {
		query += fmt.Sprintf(" AND start_date >= $%d", paramIndex)
		params = append(params, *startDate)
		paramIndex++
	}
	if endDate != nil {
		query += fmt.Sprintf(" AND end_date <= $%d", paramIndex)
		params = append(params, *endDate)
		paramIndex++
	}
	query += fmt.Sprintf(" ORDER BY start_date DESC LIMIT $%d OFFSET $%d", paramIndex, paramIndex+1)
	params = append(params, limit, offset)

	rows, err := r.db.QueryContext(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []entities.WorkHistory
	for rows.Next() {
		var wh entities.WorkHistory
		if err := rows.Scan(
			&wh.ID, &wh.EmployeeID, &wh.Company, &wh.Position,
			&wh.StartDate, &wh.EndDate, &wh.Responsibilities,
			&wh.CreatedAt, &wh.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, wh)
	}
	return list, nil
}

var _ repos.WorkHistoryRepository = (*workHistoryPgRepository)(nil)
