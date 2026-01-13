package departments

import (
	"context"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/utils"

	"github.com/google/uuid"
)

type UpdateDepartmentUseCase struct {
	Repo repos.DepartmentRepository
}

type UpdateDepartmentInput struct {
	ID       uuid.UUID
	Nome     string
	ParentID *uuid.UUID
}

func (uc *UpdateDepartmentUseCase) Execute(ctx context.Context, input UpdateDepartmentInput) (entities.Department, error) {
	// Buscar o departamento actual
	dept, err := uc.Repo.FindByID(ctx, input.ID)
	if err != nil {
		return entities.Department{}, err
	}

	dtos.ApplyUpdateToDepartment(&dept, dtos.UpdateDepartmentDTO{
		Nome:     input.Nome,
		ParentID: utils.ToOptionalString(input.ParentID),
	})

	if err := uc.Repo.Update(ctx, dept); err != nil {
		return entities.Department{}, err
	}

	return dept, nil
}
