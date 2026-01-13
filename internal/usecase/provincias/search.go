package provincias

import (
	"context"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/utils"
)

// SearchProvinceUseCase retorna províncias (itens + total)
type SearchProvinceUseCase struct {
	Repo repos.ProvinceRepository
}

func (uc *SearchProvinceUseCase) Execute(ctx context.Context, input utils.SearchInput) ([]entities.Province, int, error) {
	utils.ApplyDefaults(&input)
	return uc.Repo.Search(ctx, input.SearchText, input.Limit, input.Offset)
}
