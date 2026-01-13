package persistence

import (
	"context"
	"database/sql"
	"errors"
	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/domain/vos"

	"github.com/google/uuid"
)

type districtPgRepository struct {
	db *sql.DB
}

func NewDistrictPgRepository(db *sql.DB) *districtPgRepository {
	return &districtPgRepository{db: db}
}

func (r *districtPgRepository) Create(ctx context.Context, d entities.District) error {
	query := `
		INSERT INTO districts (id, nome, municipio_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := r.db.ExecContext(ctx, query,
		d.ID, d.Nome.String(), d.MunicipalityID, d.CreatedAt, d.UpdatedAt,
	)
	return err
}

func (r *districtPgRepository) Update(ctx context.Context, d entities.District) error {
	query := `
		UPDATE districts
		SET nome = $1, municipio_id = $2, updated_at = $3
		WHERE id = $4
	`
	res, err := r.db.ExecContext(ctx, query,
		d.Nome.String(), d.MunicipalityID, d.UpdatedAt, d.ID,
	)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *districtPgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM districts WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *districtPgRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.District, error) {
	query := `SELECT id, nome, municipio_id, created_at, updated_at FROM districts WHERE id = $1`
	var d entities.District
	var nome string

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&d.ID, &nome, &d.MunicipalityID, &d.CreatedAt, &d.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return d, sql.ErrNoRows
		}
		return d, err
	}

	if d.Nome, err = vos.NewDistrict(nome); err != nil {
		return d, err
	}

	return d, nil
}

func (r *districtPgRepository) FindAll(ctx context.Context, limit, offset int) ([]entities.District, int, error) {
	// total sem filtros
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM districts`).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := `
		SELECT id, nome, municipio_id, created_at, updated_at
		FROM districts
		ORDER BY created_at DESC LIMIT $1 OFFSET $2;
	`

	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []entities.District
	for rows.Next() {
		var d entities.District
		var nome string

		if err := rows.Scan(&d.ID, &nome, &d.MunicipalityID, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, 0, err
		}
		if d.Nome, err = vos.NewDistrict(nome); err != nil {
			return nil, 0, err
		}
		list = append(list, d)
	}
	return list, total, rows.Err()
}

func (r *districtPgRepository) Search(ctx context.Context, searchText, municipioFilter string, limit, offset int) ([]dtos.DistritoResultDTO, int, error) {
	// total com os mesmos filtros da function search_districts
	// (search_text = '' OR d.nome ILIKE %search%) AND (municipio_filter = '' OR m.nome = municipio_filter)
	var total int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM districts d
		JOIN municipalities m ON d.municipio_id = m.id
		WHERE
		  ($1 = '' OR LOWER(d.nome) LIKE '%' || LOWER($1) || '%')
		  AND ($2 = '' OR LOWER(m.nome) = LOWER($2))
	`, searchText, municipioFilter).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Itens: replicando a projeção do DTO
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			d.id, d.nome, d.municipio_id, m.nome AS municipio_nome,
			d.created_at, d.updated_at
		FROM districts d
		JOIN municipalities m ON d.municipio_id = m.id
		WHERE
		  ($1 = '' OR LOWER(d.nome) LIKE '%' || LOWER($1) || '%')
		  AND ($2 = '' OR LOWER(m.nome) = LOWER($2))
		ORDER BY d.created_at DESC
		LIMIT $3 OFFSET $4
	`, searchText, municipioFilter, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []dtos.DistritoResultDTO
	for rows.Next() {
		var d dtos.DistritoResultDTO
		if err := rows.Scan(&d.ID, &d.Nome, &d.MunicipioID, &d.MunicipioNome, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, 0, err
		}
		results = append(results, d)
	}
	return results, total, rows.Err()
}

func (r *districtPgRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS (SELECT 1 FROM districts WHERE id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	return exists, err
}

func (r *districtPgRepository) ExistsByNomeAndMunicipio(ctx context.Context, nome string, municipioID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS (
			SELECT 1 FROM districts
			WHERE LOWER(nome) = LOWER($1) AND municipio_id = $2
		)
	`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, nome, municipioID).Scan(&exists)
	return exists, err
}

var _ repos.DistrictRepository = (*districtPgRepository)(nil)
