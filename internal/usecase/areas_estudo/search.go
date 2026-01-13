// internal/usecase/areas_estudo/search.go
package areas_estudo

import (
	"context"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/utils"
)

// SearchAreaEstudoUseCase retorna áreas de estudo (paginado)
type SearchAreaEstudoUseCase struct {
	Repo repos.AreaEstudoRepository
}

func (uc *SearchAreaEstudoUseCase) Execute(ctx context.Context, input utils.SearchInput) (utils.PagedResponse[entities.AreaEstudo], error) {
	utils.ApplyDefaults(&input)
	return uc.Repo.Search(ctx, input.SearchText, input.Filter, input.Limit, input.Offset)
}
