package dependents

import (
	"context"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/utils"

	"github.com/google/uuid"
)

// SearchDependentsUseCase retorna entidades de dependentes
type SearchDependentsUseCase struct {
	Repo repos.DependentRepository
}

func (uc *SearchDependentsUseCase) Execute(ctx context.Context, input utils.SearchInput, empID *uuid.UUID) ([]entities.Dependent, error) {
	utils.ApplyDefaults(&input)
	return uc.Repo.Search(ctx, input.SearchText, input.Filter, input.Limit, input.Offset, empID)
}
