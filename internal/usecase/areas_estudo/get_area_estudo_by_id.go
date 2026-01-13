package areas_estudo

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type GetAreaEstudoByIDUseCase struct {
	Repo repos.AreaEstudoRepository
}

type GetOneAreaEstudoInput struct {
	ID uuid.UUID
}

func (uc *GetAreaEstudoByIDUseCase) Execute(ctx context.Context, input GetOneAreaEstudoInput) (entities.AreaEstudo, error) {
	return uc.Repo.FindByID(ctx, input.ID)
}
