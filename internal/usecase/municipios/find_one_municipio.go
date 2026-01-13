package municipios

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type FindMunicipalityByIDUseCase struct {
	Repo repos.MunicipalityRepository
}

func (uc *FindMunicipalityByIDUseCase) Execute(ctx context.Context, id uuid.UUID) (entities.Municipality, error) {
	return uc.Repo.FindByID(ctx, id)
}
