package education

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type ListEducationHistoriesUseCase struct {
	Repo repos.EducationHistoryRepository
}

func (uc *ListEducationHistoriesUseCase) Execute(ctx context.Context, employeeID uuid.UUID, limit, offset int) ([]entities.EducationHistory, error) {
	return uc.Repo.FindAllByEmployee(ctx, employeeID, limit, offset)
}
