package dependents

import (
	"context"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type FindDependentByIDUseCase struct {
	Repo repos.DependentRepository
}

func (uc *FindDependentByIDUseCase) Execute(ctx context.Context, id uuid.UUID) (entities.Dependent, error) {
	return uc.Repo.FindByID(ctx, id)
}
