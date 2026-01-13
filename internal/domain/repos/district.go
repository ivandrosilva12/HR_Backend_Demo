package repos

import (
	"context"
	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"

	"github.com/google/uuid"
)

type DistrictRepository interface {
	Create(ctx context.Context, d entities.District) error
	Update(ctx context.Context, d entities.District) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (entities.District, error)
	FindAll(ctx context.Context, limit, offset int) ([]entities.District, int, error)
	Search(ctx context.Context, searchText, municipioFilter string, limit, offset int) ([]dtos.DistritoResultDTO, int, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByNomeAndMunicipio(ctx context.Context, name string, municipioID uuid.UUID) (bool, error)
}
