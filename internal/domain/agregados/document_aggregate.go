package agregados

import (
	"context"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/vos"
	"time"

	"github.com/google/uuid"
)

type DocumentAggregate struct {
	OwnerType vos.DocumentOwnerType
	OwnerID   uuid.UUID
	Documents []entities.Document
}

func (agg *DocumentAggregate) AddDocument(doc entities.Document) {
	doc.UploadedAt = time.Now()
	agg.Documents = append(agg.Documents, doc)
}

func (agg *DocumentAggregate) ListActiveDocuments() []entities.Document {
	var result []entities.Document
	for _, doc := range agg.Documents {
		if doc.IsActive {
			result = append(result, doc)
		}
	}
	return result
}

type DocumentAggregateRepository interface {
	GetByOwner(ctx context.Context, ownerType vos.DocumentOwnerType, ownerID uuid.UUID) (*DocumentAggregate, error)
}
