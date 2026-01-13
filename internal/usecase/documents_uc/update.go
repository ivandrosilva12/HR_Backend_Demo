package documents_uc

import (
	"context"
	"time"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type UpdateDocumentInput struct {
	ID          uuid.UUID
	DocumentDTO dtos.UpdateDocumentDTO
}

type UpdateDocumentUseCase struct {
	Repo repos.DocumentRepository
}

func (uc *UpdateDocumentUseCase) Execute(ctx context.Context, input UpdateDocumentInput) (entities.Document, error) {

	// Buscar o funcionário existente
	doc, err := uc.Repo.FindByID(ctx, input.ID)
	if err != nil {
		return entities.Document{}, err
	}

	// Aplicar atualizações no documento
	if err := dtos.ApplyUpdateToDocument(&doc, input.DocumentDTO); err != nil {
		return entities.Document{}, err
	}

	doc.UploadedAt = time.Now()
	if err := uc.Repo.Update(ctx, doc); err != nil {
		return entities.Document{}, err
	}

	return doc, nil
}
