package employees

import (
	"context"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type UpdateEmployeeInput struct {
	ID          uuid.UUID
	EmployeeDTo dtos.UpdateEmployeeDTO
}

type UpdateEmployeeUseCase struct {
	Repo repos.EmployeeRepository
}

func (uc *UpdateEmployeeUseCase) Execute(ctx context.Context, input UpdateEmployeeInput) (entities.Employee, error) {
	// Buscar o funcionário existente
	emp, err := uc.Repo.FindByID(ctx, input.ID)
	if err != nil {
		return entities.Employee{}, err
	}

	// Aplicar atualizações no funcionário
	if err := dtos.ApplyUpdateToEmployee(&emp, input.EmployeeDTo); err != nil {
		return entities.Employee{}, err
	}

	// Persistir alterações
	if err := uc.Repo.Update(ctx, emp); err != nil {
		return entities.Employee{}, err
	}

	return emp, nil
}
