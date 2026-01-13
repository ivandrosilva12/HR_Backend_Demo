package workhistory

import (
	"context"
	"time"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type UpdateWorkHistoryInput struct {
	ID      uuid.UUID
	WorkDTO dtos.UpdateWorkDTO
}

type UpdateWorkHistoryUseCase struct {
	Repo repos.WorkHistoryRepository
}

func (uc *UpdateWorkHistoryUseCase) Execute(ctx context.Context, input UpdateWorkHistoryInput) (entities.WorkHistory, error) {
	work, err := uc.Repo.FindByID(ctx, input.ID)
	if err != nil {
		return entities.WorkHistory{}, err
	}

	if err := dtos.ApplyUpdateToWork(&work, input.WorkDTO); err != nil {
		return entities.WorkHistory{}, err
	}

	work.UpdatedAt = time.Now()
	if err := uc.Repo.Update(ctx, work); err != nil {
		return entities.WorkHistory{}, err
	}
	return work, nil
}
