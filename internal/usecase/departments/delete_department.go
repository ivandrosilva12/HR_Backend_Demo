package departments

import (
	"context"

	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type DeleteDepartmentUseCase struct {
	Repo repos.DepartmentRepository
}

type DeleteDepartmentInput struct {
	ID uuid.UUID
}

func (uc *DeleteDepartmentUseCase) Execute(ctx context.Context, input DeleteDepartmentInput) error {
	return uc.Repo.Delete(ctx, input.ID)
}
