// internal/usecase/workerhistory/update.go
package workerhistory

import (
	"context"
	"time"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type UpdateInput struct {
	ID         uuid.UUID
	EmployeeID uuid.UUID // garantimos que permanece do registro existente
	PositionID *uuid.UUID
	StartDate  *time.Time
	EndDate    *time.Time
	Status     *entities.WorkerStatus
}

type UpdateUseCase struct{ Repo repos.WorkerHistoryRepository }

func (uc *UpdateUseCase) Execute(ctx context.Context, in UpdateInput) (entities.WorkerHistory, error) {
	// carrega atual
	curr, err := uc.Repo.FindByID(ctx, in.ID)
	if err != nil {
		return entities.WorkerHistory{}, err
	}

	// aplica mudanças
	if in.PositionID != nil {
		curr.PositionID = *in.PositionID
	}
	if in.StartDate != nil {
		curr.StartDate = *in.StartDate
	}
	if in.EndDate != nil {
		curr.EndDate = in.EndDate
	}
	if in.Status != nil {
		curr.Status = *in.Status
	}
	curr.UpdatedAt = time.Now()

	if err := curr.Validate(); err != nil {
		return entities.WorkerHistory{}, err
	}

	// persiste
	if err := uc.Repo.Update(ctx, curr); err != nil {
		return entities.WorkerHistory{}, err
	}
	return curr, nil
}
