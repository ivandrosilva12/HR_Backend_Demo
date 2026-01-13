package provincias

import (
	"context"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type DeleteProvinceUseCase struct {
	Repo repos.ProvinceRepository
}

// Execute remove uma província pelo ID
func (uc *DeleteProvinceUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	return uc.Repo.Delete(ctx, id)
}
