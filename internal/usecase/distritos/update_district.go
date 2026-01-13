package distritos

import (
	"context"
	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type UpdateDistrictUseCase struct {
	Repo repos.DistrictRepository
}

type UpdateDistrictInput struct {
	ID          uuid.UUID
	Nome        string
	MunicipioID uuid.UUID
}

func (uc *UpdateDistrictUseCase) Execute(ctx context.Context, input UpdateDistrictInput) (entities.District, error) {
	d, err := uc.Repo.FindByID(ctx, input.ID)
	if err != nil {
		return entities.District{}, err
	}

	err = dtos.ApplyUpdateToDistrito(&d, dtos.UpdateDistritoDTO{
		Nome:        input.Nome,
		MunicipioID: input.MunicipioID.String(),
	})

	if err != nil {
		return entities.District{}, err
	}

	if err := uc.Repo.Update(ctx, d); err != nil {
		return entities.District{}, err
	}

	return d, nil
}
