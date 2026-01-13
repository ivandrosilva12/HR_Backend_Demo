// internal/usecase/workerhistory/delete.go
package workerhistory

import (
	"context"

	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type DeleteUseCase struct{ Repo repos.WorkerHistoryRepository }

func (uc *DeleteUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	return uc.Repo.Delete(ctx, id)
}
