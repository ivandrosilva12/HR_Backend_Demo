package distritos

import (
	"context"

	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type DeleteDistrictUseCase struct {
	Repo repos.DistrictRepository
}

type DeleteDistrictInput struct {
	ID uuid.UUID
}

func (uc *DeleteDistrictUseCase) Execute(ctx context.Context, input DeleteDistrictInput) error {
	return uc.Repo.Delete(ctx, input.ID)
}
