package repos

import (
	"context"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/utils"

	"github.com/google/uuid"
)

type DepartmentRepository interface {
	Create(ctx context.Context, d entities.Department) error
	Update(ctx context.Context, d entities.Department) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (entities.Department, error)
	FindAll(ctx context.Context, limit, offset int) (utils.PagedResponse[entities.Department], error)
	Search(ctx context.Context, searchText, filter string, limit, offset int) (utils.PagedResponse[entities.Department], error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByNome(ctx context.Context, nome string) (bool, error)
	DepartmentPositionTotals(ctx context.Context, departmentRoot uuid.UUID, includeChildren bool) ([]dtos.DepartmentPositionTotals, error)
}
