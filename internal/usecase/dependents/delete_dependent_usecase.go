package dependents

import (
	"context"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type DeleteDependentUseCase struct {
	Repo repos.DependentRepository
}

func (uc *DeleteDependentUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	return uc.Repo.Delete(ctx, id)
}
