package municipios

import (
	"context"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type UpdateMunicipalityInput struct {
	ID         uuid.UUID
	Nome       string
	ProvinceID uuid.UUID
}

type UpdateMunicipalityUseCase struct {
	Repo repos.MunicipalityRepository
}

func (uc *UpdateMunicipalityUseCase) Execute(ctx context.Context, input UpdateMunicipalityInput) (entities.Municipality, error) {
	municipality, err := uc.Repo.FindByID(ctx, input.ID)
	if err != nil {
		return entities.Municipality{}, err
	}

	err = dtos.ApplyUpdateToMunicipio(&municipality, dtos.UpdateMunicipioDTO{
		Nome:        input.Nome,
		ProvinciaID: input.ProvinceID.String(),
	})
	if err != nil {
		return entities.Municipality{}, err
	}

	if err := uc.Repo.Update(ctx, municipality); err != nil {
		return entities.Municipality{}, err
	}

	return municipality, nil
}
