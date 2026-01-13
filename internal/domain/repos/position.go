// internal/domain/repos/position_repository.go
package repos

import (
	"context"
	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/utils"

	"github.com/google/uuid"
)

type PositionRepository interface {
	Create(ctx context.Context, pos entities.Position) error
	Update(ctx context.Context, pos entities.Position) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (entities.Position, error)
	FindAll(ctx context.Context, limit, offset int) (utils.PagedResponse[entities.Position], error)
	Search(ctx context.Context, searchText, departmentFilter string, limit, offset int) (utils.PagedResponse[dtos.PositionResultDTO], error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByNomeAndDepartment(ctx context.Context, nome string, departmentID uuid.UUID) (bool, error)
}
