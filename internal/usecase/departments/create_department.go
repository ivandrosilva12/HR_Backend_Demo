package departments

import (
	"context"
	"time"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type CreateDepartmentUseCase struct {
	Repo repos.DepartmentRepository
}

type CreateDepartmentInput struct {
	Nome     string
	ParentID *uuid.UUID
}

func (uc *CreateDepartmentUseCase) Execute(ctx context.Context, input CreateDepartmentInput) (entities.Department, error) {

	now := time.Now()
	dept := entities.Department{
		ID:        uuid.New(),
		Nome:      input.Nome,
		ParentID:  input.ParentID,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if err := uc.Repo.Create(ctx, dept); err != nil {
		return entities.Department{}, err
	}

	return dept, nil
}
