package distritos

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type FindDistrictByIDUseCase struct {
	Repo repos.DistrictRepository
}

type FindDistrictByIDInput struct {
	ID uuid.UUID
}

func (uc *FindDistrictByIDUseCase) Execute(ctx context.Context, input FindDistrictByIDInput) (entities.District, error) {

	return uc.Repo.FindByID(ctx, input.ID)
}
