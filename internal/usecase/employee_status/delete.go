package employee_status

import (
	"context"

	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type DeleteEmployeeStatusUseCase struct {
	Repo repos.EmployeeStatusRepository
}

func (uc *DeleteEmployeeStatusUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	return uc.Repo.Delete(ctx, id)
}
