package positions

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type FindPositionByIDUseCase struct {
	Repo repos.PositionRepository
}

type FindPositionByIDInput struct {
	ID uuid.UUID
}

func (uc *FindPositionByIDUseCase) Execute(ctx context.Context, input FindPositionByIDInput) (entities.Position, error) {
	return uc.Repo.FindByID(ctx, input.ID)
}
