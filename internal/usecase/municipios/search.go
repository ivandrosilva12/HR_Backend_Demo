package municipios

import (
	"context"
	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/repos"
	"rhapp/internal/utils"
)

type SearchMunicipalityUseCase struct {
	Repo repos.MunicipalityRepository
}

func (uc *SearchMunicipalityUseCase) Execute(ctx context.Context, input utils.SearchInput) ([]dtos.MunicipioResultDTO, int, error) {
	utils.ApplyDefaults(&input)
	return uc.Repo.Search(ctx, input.SearchText, input.Filter, input.Limit, input.Offset)
}
