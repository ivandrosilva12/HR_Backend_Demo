package agregados

import (
	"context"
	"rhapp/internal/domain/entities"

	"github.com/google/uuid"
)

type LocationAggregate struct {
	Province       entities.Province
	Municipalities []entities.Municipality
	Districts      []entities.District
}

func (agg *LocationAggregate) AddMunicipality(m entities.Municipality) {
	agg.Municipalities = append(agg.Municipalities, m)
}

func (agg *LocationAggregate) AddDistrict(d entities.District) {
	agg.Districts = append(agg.Districts, d)
}

func (agg *LocationAggregate) ListDistrictsByMunicipalityID(municipalityID uuid.UUID) []entities.District {
	var result []entities.District
	for _, d := range agg.Districts {
		if d.MunicipalityID == municipalityID {
			result = append(result, d)
		}
	}
	return result
}

type LocationAggregateRepository interface {
	GetByProvinceID(ctx context.Context, id uuid.UUID) (*LocationAggregate, error)
}
