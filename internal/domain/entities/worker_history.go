// internal/domain/entities/worker_history.go
package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type WorkerStatus string

const (
	WorkerAtivo   WorkerStatus = "activo"
	WorkerInativo WorkerStatus = "inactivo"
)

type WorkerHistory struct {
	ID         uuid.UUID
	EmployeeID uuid.UUID
	PositionID uuid.UUID
	StartDate  time.Time
	EndDate    *time.Time
	Status     WorkerStatus
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

func (w *WorkerHistory) Validate() error {
	if w.EndDate != nil && w.EndDate.Before(w.StartDate) {
		return errors.New("data de término não pode ser anterior à data de início")
	}
	if w.Status != WorkerAtivo && w.Status != WorkerInativo {
		return errors.New("status inválido (use 'activo' ou 'inactivo')")
	}
	return nil
}
