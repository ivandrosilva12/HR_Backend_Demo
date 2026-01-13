package employees

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/utils"
)

// SearchEmployeesUseCase retorna PagedResponse de funcionários
type SearchEmployeesUseCase struct {
	Repo repos.EmployeeRepository
}

func (uc *SearchEmployeesUseCase) Execute(ctx context.Context, input utils.SearchInput) (utils.PagedResponse[entities.Employee], error) {
	utils.ApplyDefaults(&input)
	return uc.Repo.Search(ctx, input.SearchText, input.Filter, input.Limit, input.Offset)
}
