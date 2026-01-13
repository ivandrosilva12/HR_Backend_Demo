package areas_estudo

import (
	"context"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type DeleteAreaEstudoUseCase struct {
	Repo repos.AreaEstudoRepository
}

type DeleteAreaEstudoInput struct {
	ID uuid.UUID
}

func (uc *DeleteAreaEstudoUseCase) Execute(ctx context.Context, input DeleteAreaEstudoInput) error {
	return uc.Repo.Delete(ctx, input.ID)
}
