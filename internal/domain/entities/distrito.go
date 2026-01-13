package entities

import (
	"rhapp/internal/domain/vos"
	"time"

	"github.com/google/uuid"
)

type District struct {
	ID             uuid.UUID
	Nome           vos.District
	MunicipalityID uuid.UUID
	CreatedAt      time.Time
	UpdatedAt      time.Time
}
