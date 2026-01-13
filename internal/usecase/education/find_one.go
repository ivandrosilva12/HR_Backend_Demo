package education

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type FindEducationHistoryByIDUseCase struct {
	Repo repos.EducationHistoryRepository
}

func (uc *FindEducationHistoryByIDUseCase) Execute(ctx context.Context, id uuid.UUID) (entities.EducationHistory, error) {
	return uc.Repo.FindByID(ctx, id)
}
