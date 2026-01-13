package education

import (
	"context"

	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type DeleteEducationHistoryUseCase struct {
	Repo repos.EducationHistoryRepository
}

func (uc *DeleteEducationHistoryUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	return uc.Repo.Delete(ctx, id)
}
