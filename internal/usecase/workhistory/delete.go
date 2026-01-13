package workhistory

import (
	"context"

	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type DeleteWorkHistoryUseCase struct {
	Repo repos.WorkHistoryRepository
}

func (uc *DeleteWorkHistoryUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	return uc.Repo.Delete(ctx, id)
}
