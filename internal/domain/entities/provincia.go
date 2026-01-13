package entities

import (
	"rhapp/internal/domain/vos"
	"time"

	"github.com/google/uuid"
)

type Province struct {
	ID        uuid.UUID
	Nome      vos.Province
	CreatedAt time.Time
	UpdatedAt time.Time
}
