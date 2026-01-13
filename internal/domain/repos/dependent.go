// internal/domain/repos/dependent_repository.go
package repos

import (
	"context"

	"rhapp/internal/domain/entities"

	"github.com/google/uuid"
)

type DependentRepository interface {
	Create(ctx context.Context, d entities.Dependent) error
	Update(ctx context.Context, d entities.Dependent) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (entities.Dependent, error)
	FindAllByEmployee(ctx context.Context, empID uuid.UUID, limit, offset int) ([]entities.Dependent, error)
	Search(ctx context.Context, searchText, filter string, limit, offset int, employeeID *uuid.UUID) ([]entities.Dependent, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}
