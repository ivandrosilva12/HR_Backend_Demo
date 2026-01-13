package positions

import (
	"context"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/repos"
	"rhapp/internal/utils"
)

// Agora retorna PagedResponse[dtos.PositionResultDTO]
type SearchPositionUseCase struct {
	Repo repos.PositionRepository
}

func (uc *SearchPositionUseCase) Execute(ctx context.Context, input utils.SearchInput) (utils.PagedResponse[dtos.PositionResultDTO], error) {
	utils.ApplyDefaults(&input)
	return uc.Repo.Search(ctx, input.SearchText, input.Filter, input.Limit, input.Offset)
}
