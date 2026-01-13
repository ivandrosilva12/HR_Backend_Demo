package workhistory

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type ListWorkHistoryByEmployeeUseCase struct {
	Repo repos.WorkHistoryRepository
}

func (uc *ListWorkHistoryByEmployeeUseCase) Execute(ctx context.Context, employeeID uuid.UUID, limit, offset int) ([]entities.WorkHistory, error) {
	return uc.Repo.ListByEmployee(ctx, employeeID, limit, offset)
}
