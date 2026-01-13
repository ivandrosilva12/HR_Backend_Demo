package repos

import (
	"context"
	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"

	"github.com/google/uuid"
)

type MunicipalityRepository interface {
	Create(ctx context.Context, m entities.Municipality) error
	Update(ctx context.Context, m entities.Municipality) error
	Delete(ctx context.Context, id uuid.UUID) error
	FindByID(ctx context.Context, id uuid.UUID) (entities.Municipality, error)
	FindAll(ctx context.Context, limit, offset int) ([]entities.Municipality, int, error)
	ExistsByID(ctx context.Context, id uuid.UUID) (bool, error)
	ExistsByNomeAndProvince(ctx context.Context, nome string, provinceID uuid.UUID) (bool, error)
	Search(ctx context.Context, searchText, filter string, limit, offset int) ([]dtos.MunicipioResultDTO, int, error)
}
