// internal/usecase/areas_estudo/list_all.go
package areas_estudo

import (
	"context"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/utils"
)

type ListAllAreasEstudoUseCase struct {
	Repo repos.AreaEstudoRepository
}

func (uc *ListAllAreasEstudoUseCase) Execute(ctx context.Context, limit, offset int) (utils.PagedResponse[entities.AreaEstudo], error) {
	return uc.Repo.FindAll(ctx, limit, offset)
}
