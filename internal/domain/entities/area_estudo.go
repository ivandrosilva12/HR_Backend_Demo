package entities

import (
	"time"

	"github.com/google/uuid"
)

type AreaEstudo struct {
	ID          uuid.UUID
	Nome        string
	Description string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
