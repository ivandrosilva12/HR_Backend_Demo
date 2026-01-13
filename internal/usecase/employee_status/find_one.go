package employee_status

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type FindEmployeeStatusByIDUseCase struct {
	Repo repos.EmployeeStatusRepository
}

func (uc *FindEmployeeStatusByIDUseCase) Execute(ctx context.Context, id uuid.UUID) (entities.EmployeeStatus, error) {
	return uc.Repo.FindByID(ctx, id)
}
