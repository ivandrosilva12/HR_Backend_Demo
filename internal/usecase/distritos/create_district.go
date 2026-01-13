package distritos

import (
	"context"
	"time"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/domain/vos"

	"github.com/google/uuid"
)

type CreateDistrictUseCase struct {
	Repo repos.DistrictRepository
}

type CreateDistrictInput struct {
	Nome        string
	MunicipioID uuid.UUID
}

func (uc *CreateDistrictUseCase) Execute(ctx context.Context, input CreateDistrictInput) (entities.District, error) {

	district := entities.District{
		ID:             uuid.New(),
		Nome:           vos.MustNewDistrict(input.Nome),
		MunicipalityID: input.MunicipioID,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	if err := uc.Repo.Create(ctx, district); err != nil {
		return entities.District{}, err
	}

	return district, nil
}
