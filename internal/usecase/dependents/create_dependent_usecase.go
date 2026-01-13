package dependents

import (
	"context"
	"time"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/domain/vos"

	"github.com/google/uuid"
)

type CreateDependentInput struct {
	EmployeeID   uuid.UUID
	FullName     string
	Relationship string
	Gender       string
	DateOfBirth  string
	DocumentID   *uuid.UUID
}

type CreateDependentUseCase struct {
	Repo repos.DependentRepository
}

func (uc *CreateDependentUseCase) Execute(ctx context.Context, input CreateDependentInput) (entities.Dependent, error) {
	now := time.Now()

	dep := entities.Dependent{
		ID:           uuid.New(),
		EmployeeID:   input.EmployeeID,
		FullName:     vos.MustNewPersonalName(input.FullName),
		Relationship: vos.MustNewRelationshipType(input.Relationship),
		Gender:       vos.MustNewGender(input.Gender),
		DateOfBirth:  vos.MustNewBirthDate(input.DateOfBirth),
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.Repo.Create(ctx, dep); err != nil {
		return entities.Dependent{}, err
	}
	return dep, nil
}
