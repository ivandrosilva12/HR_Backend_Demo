package distritos

import (
	"context"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
)

type ListAllDistrictsUseCase struct {
	Repo repos.DistrictRepository
}

func (uc *ListAllDistrictsUseCase) Execute(ctx context.Context, limit, offset int) ([]entities.District, int, error) {
	items, total, err := uc.Repo.FindAll(ctx, limit, offset)
	if err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
