package repos

import (
	"context"
	"rhapp/internal/domain/entities"
	"rhapp/internal/utils"

	"github.com/google/uuid"
)

type EmployeeRepository interface {
	Create(ctx context.Context, e entities.Employee) error
	Update(ctx context.Context, e entities.Employee) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (entities.Employee, error)
	List(ctx context.Context, limit, offset int) (utils.PagedResponse[entities.Employee], error)
	Search(ctx context.Context, searchText, filter string, limit, offset int) (utils.PagedResponse[entities.Employee], error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}
