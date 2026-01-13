package persistence

import (
	"context"
	"database/sql"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/domain/vos"

	"github.com/google/uuid"
)

type municipalityPgRepository struct {
	db *sql.DB
}

func NewMunicipalityPgRepository(db *sql.DB) *municipalityPgRepository {
	return &municipalityPgRepository{db: db}
}

func (r *municipalityPgRepository) Create(ctx context.Context, m entities.Municipality) error {
	const query = `
		INSERT INTO municipalities (id, nome, province_id, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)`
	_, err := r.db.ExecContext(ctx, query, m.ID, m.Nome.String(), m.ProvinceID, m.CreatedAt, m.UpdatedAt)
	return err
}

func (r *municipalityPgRepository) Update(ctx context.Context, m entities.Municipality) error {
	const query = `
		UPDATE municipalities
		SET nome = $1, province_id = $2, updated_at = $3
		WHERE id = $4`
	res, err := r.db.ExecContext(ctx, query, m.Nome.String(), m.ProvinceID, m.UpdatedAt, m.ID)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *municipalityPgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM municipalities WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *municipalityPgRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.Municipality, error) {
	const query = `
		SELECT id, nome, province_id, created_at, updated_at
		FROM municipalities
		WHERE id = $1`
	var m entities.Municipality
	var nome string
	err := r.db.QueryRowContext(ctx, query, id).Scan(&m.ID, &nome, &m.ProvinceID, &m.CreatedAt, &m.UpdatedAt)
	if err != nil {
		return entities.Municipality{}, err
	}
	v := vos.NewMunicipality(nome)
	m.Nome = v
	return m, nil
}

func (r *municipalityPgRepository) FindAll(ctx context.Context, limit, offset int) ([]entities.Municipality, int, error) {
	// total sem filtros
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT COUNT(*) FROM municipalities`).Scan(&total); err != nil {
		return nil, 0, err
	}

	const query = `
		SELECT id, nome, province_id, created_at, updated_at
		FROM municipalities
		ORDER BY nome ASC
		LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var list []entities.Municipality
	for rows.Next() {
		var m entities.Municipality
		var nome string
		if err := rows.Scan(&m.ID, &nome, &m.ProvinceID, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, 0, err
		}
		v := vos.NewMunicipality(nome)
		m.Nome = v
		list = append(list, m)
	}
	return list, total, rows.Err()
}

func (r *municipalityPgRepository) Search(ctx context.Context, searchText, provinceFilter string, limit, offset int) ([]dtos.MunicipioResultDTO, int, error) {
	// total com o mesmo filtro que a função search_municipalities
	// (search_text = '' OR m.nome ILIKE %search%) AND (province_filter = '' OR p.nome = province_filter)
	var total int
	if err := r.db.QueryRowContext(ctx, `
		SELECT COUNT(*)
		FROM municipalities m
		JOIN provinces p ON m.province_id = p.id
		WHERE
		  ($1 = '' OR LOWER(m.nome) LIKE '%' || LOWER($1) || '%')
		  AND ($2 = '' OR LOWER(p.nome) = LOWER($2))
	`, searchText, provinceFilter).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Itens: podemos usar a function ou replicar a query; aqui replicamos com projeção do DTO
	rows, err := r.db.QueryContext(ctx, `
		SELECT 
			m.id::text AS id,
			m.nome::text AS nome,
			m.province_id::text AS province_id,
			p.nome::text AS province_nome,
			m.created_at,
			m.updated_at
		FROM municipalities m
		JOIN provinces p ON m.province_id = p.id
		WHERE
		  ($1 = '' OR LOWER(m.nome) LIKE '%' || LOWER($1) || '%')
		  AND ($2 = '' OR LOWER(p.nome) = LOWER($2))
		ORDER BY m.created_at DESC
		LIMIT $3 OFFSET $4
	`, searchText, provinceFilter, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []dtos.MunicipioResultDTO
	for rows.Next() {
		var dto dtos.MunicipioResultDTO
		if err := rows.Scan(&dto.ID, &dto.Nome, &dto.ProvinciaID, &dto.ProvinciaNome, &dto.CreatedAt, &dto.UpdatedAt); err != nil {
			return nil, 0, err
		}
		results = append(results, dto)
	}
	return results, total, rows.Err()
}

func (r *municipalityPgRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM municipalities WHERE id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	return exists, err
}

func (r *municipalityPgRepository) ExistsByNomeAndProvince(ctx context.Context, nome string, provinceID uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1 FROM municipalities
			WHERE LOWER(nome) = LOWER($1) AND province_id = $2)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, nome, provinceID).Scan(&exists)
	return exists, err
}

var _ repos.MunicipalityRepository = (*municipalityPgRepository)(nil)
