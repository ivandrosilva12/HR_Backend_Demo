// internal/usecase/workerhistory/find_by_id.go
package workerhistory

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type FindByIDUseCase struct{ Repo repos.WorkerHistoryRepository }

func (uc *FindByIDUseCase) Execute(ctx context.Context, id uuid.UUID) (entities.WorkerHistory, error) {
	return uc.Repo.FindByID(ctx, id)
}
