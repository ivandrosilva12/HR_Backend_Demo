package distritos

import (
	"context"
	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/repos"
	"rhapp/internal/utils"
)

type SearchDistrictUseCase struct {
	Repo repos.DistrictRepository
}

func (uc *SearchDistrictUseCase) Execute(ctx context.Context, input utils.SearchInput) ([]dtos.DistritoResultDTO, int, error) {
	utils.ApplyDefaults(&input)
	return uc.Repo.Search(ctx, input.SearchText, input.Filter, input.Limit, input.Offset)
}
