package provincias

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type FindProvinceByIDUseCase struct {
	Repo repos.ProvinceRepository
}

type FindProvinceByIDInput struct {
	ID uuid.UUID
}

func (uc *FindProvinceByIDUseCase) Execute(ctx context.Context, input FindProvinceByIDInput) (entities.Province, error) {

	prov, err := uc.Repo.FindByID(ctx, input.ID)
	if err != nil {
		return entities.Province{}, err
	}

	return prov, nil
}
