package persistence

import (
	"context"
	"database/sql"
	"fmt"

	"rhapp/internal/domain/agregados"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/vos"

	"github.com/google/uuid"
)

type locationAggregatePgRepository struct {
	db *sql.DB
}

func NewLocationAggregatePgRepository(db *sql.DB) *locationAggregatePgRepository {
	return &locationAggregatePgRepository{db: db}
}

func (r *locationAggregatePgRepository) GetByProvinceID(ctx context.Context, id uuid.UUID) (*agregados.LocationAggregate, error) {
	var prov entities.Province
	var nome string

	query := `SELECT id, nome, created_at, updated_at FROM provinces WHERE id = $1`
	if err := r.db.QueryRowContext(ctx, query, id).Scan(&prov.ID, &nome, &prov.CreatedAt, &prov.UpdatedAt); err != nil {
		return nil, fmt.Errorf("erro ao buscar província: %w", err)
	}
	nomeVO := vos.NewProvince(nome)

	prov.Nome = nomeVO

	agg := &agregados.LocationAggregate{Province: prov}

	// Municípios
	munRows, err := r.db.QueryContext(ctx, `
		SELECT id, nome, province_id, created_at, updated_at 
		FROM municipalities 
		WHERE province_id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar municípios: %w", err)
	}
	defer munRows.Close()

	for munRows.Next() {
		var m entities.Municipality
		var mNome string
		if err := munRows.Scan(&m.ID, &mNome, &m.ProvinceID, &m.CreatedAt, &m.UpdatedAt); err != nil {
			return nil, fmt.Errorf("erro ao escanear município: %w", err)
		}
		m.Nome = vos.NewMunicipality(mNome)
		agg.Municipalities = append(agg.Municipalities, m)
	}
	if err := munRows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar municípios: %w", err)
	}

	// Distritos
	distRows, err := r.db.QueryContext(ctx, `
		SELECT d.id, d.nome, d.municipio_id, d.created_at, d.updated_at
		FROM districts d
		INNER JOIN municipalities m ON m.id = d.municipio_id
		WHERE m.province_id = $1`, id)
	if err != nil {
		return nil, fmt.Errorf("erro ao buscar distritos: %w", err)
	}
	defer distRows.Close()

	for distRows.Next() {
		var d entities.District
		var dName string
		if err := distRows.Scan(&d.ID, &dName, &d.MunicipalityID, &d.CreatedAt, &d.UpdatedAt); err != nil {
			return nil, fmt.Errorf("erro ao escanear distrito: %w", err)
		}
		if d.Nome, err = vos.NewDistrict(dName); err != nil {
			return nil, fmt.Errorf("erro ao validar distrito: %w", err)
		}
		agg.Districts = append(agg.Districts, d)
	}
	if err := distRows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar distritos: %w", err)
	}

	return agg, nil
}

var _ agregados.LocationAggregateRepository = (*locationAggregatePgRepository)(nil)
