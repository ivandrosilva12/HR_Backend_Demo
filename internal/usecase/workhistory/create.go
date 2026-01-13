package workhistory

import (
	"context"
	"time"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type CreateWorkHistoryInput struct {
	EmployeeID       uuid.UUID
	Company          string
	Position         string
	StartDate        time.Time
	EndDate          *time.Time
	Responsibilities string
}

type CreateWorkHistoryUseCase struct {
	Repo repos.WorkHistoryRepository
}

func (uc *CreateWorkHistoryUseCase) Execute(ctx context.Context, input entities.WorkHistory) (entities.WorkHistory, error) {
	now := time.Now()
	work := entities.WorkHistory{
		ID:               uuid.New(),
		EmployeeID:       input.EmployeeID,
		Company:          input.Company,
		Position:         input.Position,
		StartDate:        input.StartDate,
		EndDate:          input.EndDate,
		Responsibilities: input.Responsibilities,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	if err := uc.Repo.Create(ctx, work); err != nil {
		return entities.WorkHistory{}, err
	}
	return work, nil
}
