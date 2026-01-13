package municipios

import (
	"context"
	"time"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/domain/vos"

	"github.com/google/uuid"
)

type CreateMunicipalityInput struct {
	Nome       string
	ProvinceID uuid.UUID
}

type CreateMunicipalityUseCase struct {
	Repo repos.MunicipalityRepository
}

func (uc *CreateMunicipalityUseCase) Execute(ctx context.Context, input CreateMunicipalityInput) (entities.Municipality, error) {

	m := entities.Municipality{
		ID:         uuid.New(),
		Nome:       vos.NewMunicipality(input.Nome),
		ProvinceID: input.ProvinceID,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	if err := uc.Repo.Create(ctx, m); err != nil {
		return entities.Municipality{}, err
	}

	return m, nil
}
