package employees

import (
	"context"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
)

type CreateEmployeeInput struct {
	CreateEmployeeDTO dtos.CreateEmployeeDTO
}

type CreateEmployeeUseCase struct {
	Repo repos.EmployeeRepository
}

func (uc *CreateEmployeeUseCase) Execute(ctx context.Context, input CreateEmployeeInput) (entities.Employee, error) {

	emp, err := dtos.ToEmployeeFromCreateDTO(input.CreateEmployeeDTO)
	if err != nil {
		return entities.Employee{}, err
	}

	if err := uc.Repo.Create(ctx, emp); err != nil {
		return entities.Employee{}, err
	}

	return emp, nil
}
