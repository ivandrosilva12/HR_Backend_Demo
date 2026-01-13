package dependents

import (
	"context"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type UpdateDependentInput struct {
	ID           uuid.UUID
	FullName     string
	Relationship string
	Gender       string
	DateOfBirth  string
	IsActive     bool
}

type UpdateDependentUseCase struct {
	Repo repos.DependentRepository
}

func (uc *UpdateDependentUseCase) Execute(ctx context.Context, input UpdateDependentInput) (entities.Dependent, error) {
	d, err := uc.Repo.FindByID(ctx, input.ID)
	if err != nil {
		return entities.Dependent{}, err
	}

	err = dtos.ApplyUpdateToDependent(&d, dtos.UpdateDependentDTO{
		FullName:     input.FullName,
		Relationship: input.Relationship,
		Gender:       input.Gender,
		DateOfBirth:  input.DateOfBirth,
		IsActive:     input.IsActive,
	})

	if err != nil {
		return entities.Dependent{}, err
	}

	if err := uc.Repo.Update(ctx, d); err != nil {
		return entities.Dependent{}, err
	}
	return d, nil
}
