package persistence

import (
	"context"
	"database/sql"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/domain/vos"

	"github.com/google/uuid"
)

type provincePgRepository struct {
	db *sql.DB
}

func NewProvincePgRepository(db *sql.DB) *provincePgRepository {
	return &provincePgRepository{db: db}
}

func (r *provincePgRepository) Create(ctx context.Context, p entities.Province) error {
	query := `INSERT INTO provinces (id, nome, created_at, updated_at) VALUES ($1, $2, $3, $4)`
	_, err := r.db.ExecContext(ctx, query, p.ID, p.Nome.String(), p.CreatedAt, p.UpdatedAt)
	return err
}

func (r *provincePgRepository) Update(ctx context.Context, p entities.Province) error {
	query := `UPDATE provinces SET nome = $1, updated_at = $2 WHERE id = $3`
	res, err := r.db.ExecContext(ctx, query, p.Nome.String(), p.UpdatedAt, p.ID)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *provincePgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM provinces WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *provincePgRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.Province, error) {
	query := `SELECT id, nome, created_at, updated_at FROM provinces WHERE id = $1`
	var p entities.Province
	var nome string

	err := r.db.QueryRowContext(ctx, query, id).Scan(&p.ID, &nome, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return entities.Province{}, err
	}
	v := vos.NewProvince(nome)
	p.Nome = v
	return p, nil
}

func (r *provincePgRepository) FindAll(ctx context.Context, limit, offset int) ([]entities.Province, int, error) {
	// total sem filtros
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM provinces`).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `SELECT id, nome, created_at, updated_at FROM provinces ORDER BY nome ASC LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []entities.Province
	for rows.Next() {
		var p entities.Province
		var nome string
		if err := rows.Scan(&p.ID, &nome, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		v := vos.NewProvince(nome)
		p.Nome = v
		list = append(list, p)
	}
	return list, total, rows.Err()
}

func (r *provincePgRepository) Search(ctx context.Context, searchText string, limit, offset int) ([]entities.Province, int, error) {
	// total com o mesmo filtro
	var total int
	if err := r.db.QueryRowContext(ctx,
		`SELECT COUNT(*) FROM provinces
          WHERE $1 = '' OR LOWER(nome) LIKE '%' || LOWER($1) || '%'`,
		searchText,
	).Scan(&total); err != nil {
		return nil, 0, err
	}

	rows, err := r.db.QueryContext(ctx,
		`SELECT id, nome, created_at, updated_at
		   FROM provinces
		  WHERE $1 = '' OR LOWER(nome) LIKE '%' || LOWER($1) || '%'
		  ORDER BY created_at DESC
		  LIMIT $2 OFFSET $3`,
		searchText, limit, offset,
	)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []entities.Province
	for rows.Next() {
		var p entities.Province
		var nome string
		if err := rows.Scan(&p.ID, &nome, &p.CreatedAt, &p.UpdatedAt); err != nil {
			return nil, 0, err
		}
		v := vos.NewProvince(nome)
		p.Nome = v
		results = append(results, p)
	}
	return results, total, rows.Err()
}

func (r *provincePgRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM provinces WHERE id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	return exists, err
}

func (r *provincePgRepository) ExistsByNome(ctx context.Context, nome string) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM provinces WHERE LOWER(nome) = LOWER($1))`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, nome).Scan(&exists)
	return exists, err
}

var _ repos.ProvinceRepository = (*provincePgRepository)(nil)
