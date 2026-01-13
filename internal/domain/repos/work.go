package repos

import (
	"context"
	"rhapp/internal/domain/entities"

	"github.com/google/uuid"
)

type WorkHistoryRepository interface {
	Create(ctx context.Context, wh entities.WorkHistory) error
	Update(ctx context.Context, wh entities.WorkHistory) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (entities.WorkHistory, error)
	ListByEmployee(ctx context.Context, employeeID uuid.UUID, limit, offset int) ([]entities.WorkHistory, error)
	Search(ctx context.Context, employeeID uuid.UUID, searchText string, startDate, endDate *string, limit, offset int) ([]entities.WorkHistory, error)
}
