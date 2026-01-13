package documents_uc

import (
	"context"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type DeleteDocumentUseCase struct {
	Repo repos.DocumentRepository
}

func (uc *DeleteDocumentUseCase) Execute(ctx context.Context, id uuid.UUID) error {
	return uc.Repo.Delete(ctx, id)
}
