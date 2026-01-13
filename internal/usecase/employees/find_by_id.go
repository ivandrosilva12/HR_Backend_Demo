package employees

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type FindEmployeeByIDUseCase struct {
	Repo repos.EmployeeRepository
}

// Execute busca um funcionário pelo ID fornecido.
func (uc *FindEmployeeByIDUseCase) Execute(ctx context.Context, id uuid.UUID) (entities.Employee, error) {
	return uc.Repo.FindByID(ctx, id)
}
