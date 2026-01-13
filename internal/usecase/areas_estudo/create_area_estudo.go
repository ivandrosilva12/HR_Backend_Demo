package areas_estudo

import (
	"context"
	"time"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type CreateAreaEstudoUseCase struct {
	Repo repos.AreaEstudoRepository
}

type CreateAreaEstudoInput struct {
	Name        string
	Description string
}

func (uc *CreateAreaEstudoUseCase) Execute(ctx context.Context, input CreateAreaEstudoInput) (entities.AreaEstudo, error) {

	now := time.Now()
	area := entities.AreaEstudo{
		ID:          uuid.New(),
		Nome:        input.Name,
		Description: input.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := uc.Repo.Create(ctx, area); err != nil {
		return entities.AreaEstudo{}, err
	}

	return area, nil
}
