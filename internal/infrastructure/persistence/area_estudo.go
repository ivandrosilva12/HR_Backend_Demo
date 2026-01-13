// internal/infra/persistence/area_estudo_pg_repository.go
package persistence

import (
	"context"
	"database/sql"
	"errors"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/utils"

	"github.com/google/uuid"
)

type areaEstudoPgRepository struct {
	db *sql.DB
}

func NewAreaEstudoPgRepository(db *sql.DB) *areaEstudoPgRepository {
	return &areaEstudoPgRepository{db: db}
}

func (r *areaEstudoPgRepository) Create(ctx context.Context, a entities.AreaEstudo) error {
	query := `
		INSERT INTO areas_estudo (id, nome, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query, a.ID, a.Nome, a.Description, a.CreatedAt, a.UpdatedAt)
	return err
}

func (r *areaEstudoPgRepository) Update(ctx context.Context, a entities.AreaEstudo) error {
	query := `
		UPDATE areas_estudo
		SET nome = $1, description = $2, updated_at = $3
		WHERE id = $4
	`
	res, err := r.db.ExecContext(ctx, query, a.Nome, a.Description, a.UpdatedAt, a.ID)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *areaEstudoPgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM areas_estudo WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *areaEstudoPgRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.AreaEstudo, error) {
	query := `
		SELECT id, nome, description, created_at, updated_at
		FROM areas_estudo
		WHERE id = $1
	`
	var a entities.AreaEstudo
	err := r.db.QueryRowContext(ctx, query, id).Scan(&a.ID, &a.Nome, &a.Description, &a.CreatedAt, &a.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return a, sql.ErrNoRows
		}
		return a, err
	}
	return a, nil
}

// 🔁 Agora devolve PagedResponse e garante items != nil
func (r *areaEstudoPgRepository) FindAll(ctx context.Context, limit, offset int) (utils.PagedResponse[entities.AreaEstudo], error) {
	resp := utils.PagedResponse[entities.AreaEstudo]{Items: []entities.AreaEstudo{}, Limit: limit, Offset: offset}

	// total
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM areas_estudo`).Scan(&resp.Total); err != nil {
		return resp, err
	}

	// page
	query := `
		SELECT id, nome, description, created_at, updated_at
		FROM areas_estudo
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return resp, err
	}
	defer rows.Close()

	for rows.Next() {
		var a entities.AreaEstudo
		if err := rows.Scan(&a.ID, &a.Nome, &a.Description, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return resp, err
		}
		resp.Items = append(resp.Items, a)
	}
	return resp, rows.Err()
}

// 🔁 Agora devolve PagedResponse e garante items != nil.
// Contagem usa ILIKE em nome/description para total consistente.
func (r *areaEstudoPgRepository) Search(ctx context.Context, searchText, _ string, limit, offset int) (utils.PagedResponse[entities.AreaEstudo], error) {
	resp := utils.PagedResponse[entities.AreaEstudo]{Items: []entities.AreaEstudo{}, Limit: limit, Offset: offset}

	// total pelo critério de pesquisa
	countQ := `
		SELECT COUNT(*)
		FROM areas_estudo
		WHERE $1 = '' OR (nome ILIKE '%'||$1||'%' OR description ILIKE '%'||$1||'%')
	`
	if err := r.db.QueryRowContext(ctx, countQ, searchText).Scan(&resp.Total); err != nil {
		return resp, err
	}

	// page (mantém função existente)
	rows, err := r.db.QueryContext(ctx, `SELECT * FROM search_areas_estudo($1, $2, $3)`, searchText, limit, offset)
	if err != nil {
		return resp, err
	}
	defer rows.Close()

	for rows.Next() {
		var a entities.AreaEstudo
		if err := rows.Scan(&a.ID, &a.Nome, &a.Description, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return resp, err
		}
		resp.Items = append(resp.Items, a)
	}
	return resp, rows.Err()
}

func (r *areaEstudoPgRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM areas_estudo WHERE id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	return exists, err
}

func (r *areaEstudoPgRepository) ExistsByNome(ctx context.Context, nome string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM areas_estudo WHERE LOWER(nome) = LOWER($1))`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, nome).Scan(&exists)
	return exists, err
}

var _ repos.AreaEstudoRepository = (*areaEstudoPgRepository)(nil)
