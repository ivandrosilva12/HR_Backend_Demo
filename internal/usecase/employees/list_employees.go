package employees

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/utils"
)

type ListEmployeesUseCase struct {
	Repo repos.EmployeeRepository
}

func (uc *ListEmployeesUseCase) Execute(ctx context.Context, limit, offset int) (utils.PagedResponse[entities.Employee], error) {
	return uc.Repo.List(ctx, limit, offset)
}
