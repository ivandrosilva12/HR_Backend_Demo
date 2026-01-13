package entities

import (
	"time"

	"github.com/google/uuid"
)

type Position struct {
	ID               uuid.UUID
	Nome             string
	DepartmentID     uuid.UUID
	MaxHeadcount     int
	CurrentHeadcount int
	Remaining        int
	Tipo             string // NEW (matches DB enum: employee | boss)
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
