package documents_uc

import (
	"context"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type FindDocumentByIDUseCase struct {
	Repo repos.DocumentRepository
}

func (uc *FindDocumentByIDUseCase) Execute(ctx context.Context, id uuid.UUID) (entities.Document, error) {
	return uc.Repo.FindByID(ctx, id)
}
