package workerhistory

import (
	"context"
	"time"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type CreateInput struct {
	EmployeeID uuid.UUID
	PositionID uuid.UUID
	StartDate  time.Time
	EndDate    *time.Time
	Status     entities.WorkerStatus // default: activo
}

type CreateUseCase struct{ Repo repos.WorkerHistoryRepository }

func (uc *CreateUseCase) Execute(ctx context.Context, in CreateInput) (entities.WorkerHistory, error) {
	status := in.Status
	if status == "" {
		status = entities.WorkerAtivo
	}
	w := entities.WorkerHistory{
		ID:         uuid.New(),
		EmployeeID: in.EmployeeID,
		PositionID: in.PositionID,
		StartDate:  in.StartDate,
		EndDate:    in.EndDate,
		Status:     status,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	if err := w.Validate(); err != nil {
		return entities.WorkerHistory{}, err
	}
	if err := uc.Repo.Create(ctx, w); err != nil {
		return entities.WorkerHistory{}, err
	}
	return w, nil
}
