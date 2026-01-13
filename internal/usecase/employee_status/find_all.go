package employee_status

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type ListEmployeeStatusByEmployeeUseCase struct {
	Repo repos.EmployeeStatusRepository
}

func (uc *ListEmployeeStatusByEmployeeUseCase) Execute(ctx context.Context, employeeID uuid.UUID, limit, offset int) ([]entities.EmployeeStatus, error) {
	return uc.Repo.FindAllByEmployee(ctx, employeeID, limit, offset)
}
