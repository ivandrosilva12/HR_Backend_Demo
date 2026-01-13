package provincias

import (
	"context"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type UpdateProvinceUseCase struct {
	Repo repos.ProvinceRepository
}

type UpdateProvinceInput struct {
	ID   uuid.UUID
	Nome string
}

func (uc *UpdateProvinceUseCase) Execute(ctx context.Context, input UpdateProvinceInput) (entities.Province, error) {

	// Buscar província atual
	province, err := uc.Repo.FindByID(ctx, input.ID)
	if err != nil {
		return entities.Province{}, err
	}

	dtos.ApplyUpdateToProvince(&province, dtos.UpdateProvinceDTO{
		Nome: input.Nome,
	})

	if err := uc.Repo.Update(ctx, province); err != nil {
		return entities.Province{}, err
	}

	return province, nil
}
