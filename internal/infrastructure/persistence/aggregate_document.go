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

type documentAggregatePgRepository struct {
	db *sql.DB
}

func NewDocumentAggregatePgRepository(db *sql.DB) *documentAggregatePgRepository {
	return &documentAggregatePgRepository{db: db}
}

func (r *documentAggregatePgRepository) GetByOwner(ctx context.Context, ownerType vos.DocumentOwnerType, ownerID uuid.UUID) (*agregados.DocumentAggregate, error) {
	query := `
		SELECT id, type, file_name, file_url, extension, is_active, uploaded_at
		FROM documents
		WHERE owner_type = $1 AND owner_id = $2
	`

	rows, err := r.db.QueryContext(ctx, query, ownerType.String(), ownerID)
	if err != nil {
		return nil, fmt.Errorf("erro ao consultar documentos: %w", err)
	}
	defer rows.Close()

	var docs []entities.Document
	for rows.Next() {
		var doc entities.Document
		if err := rows.Scan(
			&doc.ID,
			&doc.Type,
			&doc.FileName,
			&doc.FileURL,
			&doc.Extension,
			&doc.IsActive,
			&doc.UploadedAt,
		); err != nil {
			return nil, fmt.Errorf("erro ao escanear documento: %w", err)
		}
		doc.OwnerType = ownerType
		doc.OwnerID = ownerID
		docs = append(docs, doc)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("erro ao iterar resultados: %w", err)
	}

	return &agregados.DocumentAggregate{
		OwnerType: ownerType,
		OwnerID:   ownerID,
		Documents: docs,
	}, nil
}

var _ agregados.DocumentAggregateRepository = (*documentAggregatePgRepository)(nil)
