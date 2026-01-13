package provincias

import (
	"context"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
)

type FindAllProvincesUseCase struct {
	Repo repos.ProvinceRepository
}

// Execute retorna todas as províncias com suporte a paginação + total.
func (uc *FindAllProvincesUseCase) Execute(ctx context.Context, limit, offset int) ([]entities.Province, int, error) {
	provinces, total, err := uc.Repo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return provinces, total, nil
}
