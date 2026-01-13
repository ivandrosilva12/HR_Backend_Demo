package agregados

import (
	"context"

	"rhapp/internal/domain/agregados"

	"github.com/google/uuid"
)

type GetLocationByProvinceIDUseCase struct {
	Repo agregados.LocationAggregateRepository
}

type GetLocationByProvinceInput struct {
	ProvinceID uuid.UUID
}

func (uc *GetLocationByProvinceIDUseCase) Execute(ctx context.Context, input GetLocationByProvinceInput) (*agregados.LocationAggregate, error) {
	return uc.Repo.GetByProvinceID(ctx, input.ProvinceID)
}
