package repos

import (
	"context"
	"rhapp/internal/domain/entities"

	"github.com/google/uuid"
)

type EducationHistoryRepository interface {
	Create(ctx context.Context, history entities.EducationHistory) error
	Update(ctx context.Context, history entities.EducationHistory) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (entities.EducationHistory, error)
	FindAllByEmployee(ctx context.Context, employeeID uuid.UUID, limit, offset int) ([]entities.EducationHistory, error)
	Search(ctx context.Context, employeeID *uuid.UUID, searchText string, startDate, endDate *string, limit, offset int) ([]entities.EducationHistory, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
}
