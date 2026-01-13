package repos

import (
	"context"

	"rhapp/internal/domain/entities"

	"github.com/google/uuid"
)

type ProvinceRepository interface {
	FindAll(ctx context.Context, limit, offset int) ([]entities.Province, int, error)
	Search(ctx context.Context, searchText string, limit, offset int) ([]entities.Province, int, error)
	FindByID(ctx context.Context, id uuid.UUID) (entities.Province, error)
	Create(ctx context.Context, p entities.Province) error
	Update(ctx context.Context, p entities.Province) error
	Delete(ctx context.Context, id uuid.UUID) error
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByNome(ctx context.Context, nome string) (bool, error)
}
