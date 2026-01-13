package workhistory

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type FindWorkHistoryByIDUseCase struct {
	Repo repos.WorkHistoryRepository
}

func (uc *FindWorkHistoryByIDUseCase) Execute(ctx context.Context, id uuid.UUID) (entities.WorkHistory, error) {
	return uc.Repo.FindByID(ctx, id)
}
