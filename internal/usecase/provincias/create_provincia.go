package provincias

import (
	"context"
	"time"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/domain/vos"

	"github.com/google/uuid"
)

type CreateProvinceUseCase struct {
	Repo repos.ProvinceRepository
}

type CreateProvinceInput struct {
	Name string
}

func (uc *CreateProvinceUseCase) Execute(ctx context.Context, input CreateProvinceInput) (entities.Province, error) {

	province := entities.Province{
		ID:        uuid.New(),
		Nome:      vos.NewProvince(input.Name),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := uc.Repo.Create(ctx, province); err != nil {
		return entities.Province{}, err
	}

	return province, nil
}
