package municipios

import (
	"context"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
)

type ListMunicipalitiesUseCase struct {
	Repo repos.MunicipalityRepository
}

func (uc *ListMunicipalitiesUseCase) Execute(ctx context.Context, limit, offset int) ([]entities.Municipality, int, error) {
	items, total, err := uc.Repo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
