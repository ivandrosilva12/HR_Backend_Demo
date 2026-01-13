package positions

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/utils"
)

type FindAllPositionsUseCase struct {
	Repo repos.PositionRepository
}

// Agora retorna PagedResponse[entities.Position]
func (uc *FindAllPositionsUseCase) Execute(ctx context.Context, limit, offset int) (utils.PagedResponse[entities.Position], error) {
	return uc.Repo.FindAll(ctx, limit, offset)
}
