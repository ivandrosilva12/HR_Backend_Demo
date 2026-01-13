package entities

import (
	"time"

	"github.com/google/uuid"
)

type Department struct {
	ID        uuid.UUID
	Nome      string
	ParentID  *uuid.UUID // novo
	CreatedAt time.Time
	UpdatedAt time.Time
}
