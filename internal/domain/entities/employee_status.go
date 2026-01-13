package entities

import (
	"rhapp/internal/domain/vos"
	"time"

	"github.com/google/uuid"
)

type EmployeeStatus struct {
	ID          uuid.UUID
	EmployeeID  uuid.UUID
	Status      vos.EmployeeStatus
	Reason      vos.StatusReason
	Observacoes string
	StartDate   time.Time
	EndDate     *time.Time
	IsCurrent   bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}
