package entities

import (
	"rhapp/internal/domain/vos"
	"time"

	"github.com/google/uuid"
)

type Municipality struct {
	ID         uuid.UUID
	Nome       vos.Municipality
	ProvinceID uuid.UUID
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
