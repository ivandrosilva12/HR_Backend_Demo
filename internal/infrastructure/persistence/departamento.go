// internal/infra/persistence/department_pg_repository.go
package persistence

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/utils"

	"github.com/google/uuid"
)

type departmentPgRepository struct {
	db *sql.DB
}

func NewDepartmentPgRepository(db *sql.DB) *departmentPgRepository {
	return &departmentPgRepository{db: db}
}

func (r *departmentPgRepository) Create(ctx context.Context, d entities.Department) error {
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO departments (id, nome, parent_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`, d.ID, d.Nome, d.ParentID, d.CreatedAt, d.UpdatedAt)
	return err
}

func (r *departmentPgRepository) Update(ctx context.Context, d entities.Department) error {

	res, err := r.db.ExecContext(ctx, `
		UPDATE departments
		SET nome = $1, parent_id = $2, updated_at = $3
		WHERE id = $4
	`, d.Nome, d.ParentID, d.UpdatedAt, d.ID)

	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *departmentPgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM departments WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *departmentPgRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.Department, error) {

	var d entities.Department
	err := r.db.QueryRowContext(ctx, `
		SELECT id, nome, parent_id, created_at, updated_at
		FROM departments WHERE id = $1
	`, id).Scan(&d.ID, &d.Nome, &d.ParentID, &d.CreatedAt, &d.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return d, sql.ErrNoRows
		}
		return d, err
	}
	return d, nil
}

// 🔁 Agora devolve PagedResponse e garante items != nil
func (r *departmentPgRepository) FindAll(ctx context.Context, limit, offset int) (utils.PagedResponse[entities.Department], error) {
	resp := utils.PagedResponse[entities.Department]{Items: []entities.Department{}, Limit: limit, Offset: offset}

	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM departments`).Scan(&resp.Total); err != nil {
		return resp, err
	}

	rows, err := r.db.QueryContext(ctx, `
		SELECT id, nome, parent_id, created_at, updated_at
		FROM departments
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`, limit, offset)

	if err != nil {
		return resp, err
	}
	defer rows.Close()

	for rows.Next() {
		var d entities.Department
		if err := rows.Scan(&d.ID, &d.Nome, &d.ParentID, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return resp, err
		}
		resp.Items = append(resp.Items, d)
	}
	return resp, rows.Err()
}

// 🔁 Agora devolve PagedResponse e garante items != nil.
// Contagem usa ILIKE em nome para total consistente.
func (r *departmentPgRepository) Search(ctx context.Context, searchText, _ string, limit, offset int) (utils.PagedResponse[entities.Department], error) {
	resp := utils.PagedResponse[entities.Department]{Items: []entities.Department{}, Limit: limit, Offset: offset}

	countQ := `
		SELECT COUNT(*)
		FROM departments
		WHERE $1 = '' OR (nome ILIKE '%'||$1||'%')
	`
	if err := r.db.QueryRowContext(ctx, countQ, searchText).Scan(&resp.Total); err != nil {
		return resp, err
	}

	rows, err := r.db.QueryContext(ctx, `SELECT * FROM search_departments($1, $2, $3)`, searchText, limit, offset)
	if err != nil {
		return resp, err
	}
	defer rows.Close()

	for rows.Next() {
		var d entities.Department
		if err := rows.Scan(&d.ID, &d.Nome, &d.ParentID, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return resp, fmt.Errorf("erro ao escanear linha: %w", err)
		}

		resp.Items = append(resp.Items, d)
	}
	return resp, rows.Err()
}

func (r *departmentPgRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM departments WHERE id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	return exists, err
}

func (r *departmentPgRepository) ExistsByNome(ctx context.Context, nome string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM departments WHERE LOWER(nome) = LOWER($1))`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, nome).Scan(&exists)
	return exists, err
}

func (r *departmentPgRepository) DepartmentPositionTotals(
	ctx context.Context,
	departmentRoot uuid.UUID,
	includeChildren bool,
) ([]dtos.DepartmentPositionTotals, error) {

	rows, err := r.db.QueryContext(ctx, `
		SELECT department_id,
		       department_nome,
		       total_positions,
		       occupied_positions,
		       available_positions
		FROM department_position_totals($1, $2)
	`, departmentRoot, includeChildren)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []dtos.DepartmentPositionTotals
	for rows.Next() {
		var t dtos.DepartmentPositionTotals
		if err := rows.Scan(
			&t.DepartmentID,
			&t.DepartmentNome,
			&t.TotalPositions,
			&t.OccupiedPositions,
			&t.AvailablePositions,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

var _ repos.DepartmentRepository = (*departmentPgRepository)(nil)
