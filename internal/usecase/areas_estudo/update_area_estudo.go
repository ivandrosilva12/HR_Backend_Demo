package areas_estudo

import (
	"context"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type UpdateAreaEstudoUseCase struct {
	Repo repos.AreaEstudoRepository
}

type UpdateAreaEstudoInput struct {
	ID          uuid.UUID
	Nome        string
	Description string
}

func (uc *UpdateAreaEstudoUseCase) Execute(ctx context.Context, input UpdateAreaEstudoInput) (entities.AreaEstudo, error) {
	area, err := uc.Repo.FindByID(ctx, input.ID)
	if err != nil {
		return entities.AreaEstudo{}, err
	}

	dtos.ApplyUpdateToAreaEstudo(&area, dtos.UpdateAreaEstudoDTO{
		Nome:        input.Nome,
		Description: input.Description,
	})

	if err := uc.Repo.Update(ctx, area); err != nil {
		return entities.AreaEstudo{}, err
	}

	return area, nil
}
