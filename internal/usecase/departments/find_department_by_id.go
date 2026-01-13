package departments

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type FindDepartmentByIDUseCase struct {
	Repo repos.DepartmentRepository
}

type FindDepartmentByIDInput struct {
	ID uuid.UUID
}

func (uc *FindDepartmentByIDUseCase) Execute(ctx context.Context, input FindDepartmentByIDInput) (entities.Department, error) {
	return uc.Repo.FindByID(ctx, input.ID)
}
