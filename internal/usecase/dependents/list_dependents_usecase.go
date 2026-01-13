package dependents

import (
	"context"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type ListDependentsUseCase struct {
	Repo repos.DependentRepository
}

func (uc *ListDependentsUseCase) Execute(ctx context.Context, empID uuid.UUID, limit, offset int) ([]entities.Dependent, error) {
	return uc.Repo.FindAllByEmployee(ctx, empID, limit, offset)
}
