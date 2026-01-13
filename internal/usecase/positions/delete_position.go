package positions

import (
	"context"

	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type DeletePositionUseCase struct {
	Repo repos.PositionRepository
}

func (uc *DeletePositionUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	return uc.Repo.Delete(ctx, id)
}
