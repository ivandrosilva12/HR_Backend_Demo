package repos

import (
	"context"
	"rhapp/internal/domain/entities"

	"github.com/google/uuid"
)

type DocumentRepository interface {
	Create(ctx context.Context, doc entities.Document) error
	Update(ctx context.Context, doc entities.Document) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (entities.Document, error)
	FindAllByOwner(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]entities.Document, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}
