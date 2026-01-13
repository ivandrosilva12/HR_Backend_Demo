package employees

import (
	"context"

	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type DeleteEmployeeUseCase struct {
	Repo repos.EmployeeRepository
}

// Execute exclui um funcionário com base no ID fornecido.
func (uc *DeleteEmployeeUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	return uc.Repo.Delete(ctx, id)
}
