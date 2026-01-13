package municipios

import (
	"context"

	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type DeleteMunicipalityUseCase struct {
	Repo repos.MunicipalityRepository
}

// Execute exclui um município pelo ID.
func (uc *DeleteMunicipalityUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	return uc.Repo.Delete(ctx, id)
}
