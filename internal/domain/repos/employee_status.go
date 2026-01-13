package repos

import (
	"context"
	"rhapp/internal/domain/entities"

	"github.com/google/uuid"
)

type EmployeeStatusRepository interface {
	Create(ctx context.Context, status entities.EmployeeStatus) error
	Update(ctx context.Context, status entities.EmployeeStatus) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (entities.EmployeeStatus, error)
	FindAllByEmployee(ctx context.Context, employeeID uuid.UUID, limit, offset int) ([]entities.EmployeeStatus, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}
