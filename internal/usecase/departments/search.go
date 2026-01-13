// internal/usecase/departments/search.go
package departments

import (
	"context"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/utils"
)

// SearchDepartmentUseCase retorna entidades completas (paginado)
type SearchDepartmentUseCase struct {
	Repo repos.DepartmentRepository
}

func (uc *SearchDepartmentUseCase) Execute(ctx context.Context, input utils.SearchInput) (utils.PagedResponse[entities.Department], error) {
	utils.ApplyDefaults(&input)
	return uc.Repo.Search(ctx, input.SearchText, input.Filter, input.Limit, input.Offset)
}
