package repos

import (
	"context"
	"rhapp/internal/domain/entities"

	"github.com/google/uuid"
)

type WorkerHistoryRepository interface {
	Create(ctx context.Context, s entities.WorkerHistory) error
	FindByID(ctx context.Context, id uuid.UUID) (entities.WorkerHistory, error)
	Update(ctx context.Context, s entities.WorkerHistory) error
	Delete(ctx context.Context, id uuid.UUID) error
	ListByEmployeeID(ctx context.Context, employeeID uuid.UUID, limit, offset int) ([]entities.WorkerHistory, error)
}
