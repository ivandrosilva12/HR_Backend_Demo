// internal/usecase/workerhistory/list_by_employee.go
package workerhistory

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type ListByEmployeeIDUseCase struct{ Repo repos.WorkerHistoryRepository }

func (uc *ListByEmployeeIDUseCase) Execute(ctx context.Context, employeeID uuid.UUID, limit, offset int) ([]entities.WorkerHistory, error) {
	return uc.Repo.ListByEmployeeID(ctx, employeeID, limit, offset)
}
