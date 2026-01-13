package repos

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/utils"

	"github.com/google/uuid"
)

type AreaEstudoRepository interface {
	Create(ctx context.Context, a entities.AreaEstudo) error
	Update(ctx context.Context, a entities.AreaEstudo) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (entities.AreaEstudo, error)
	FindAll(ctx context.Context, limit, offset int) (utils.PagedResponse[entities.AreaEstudo], error)
	Search(ctx context.Context, searchText, filter string, limit, offset int) (utils.PagedResponse[entities.AreaEstudo], error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByNome(ctx context.Context, nome string) (bool, error)
}
