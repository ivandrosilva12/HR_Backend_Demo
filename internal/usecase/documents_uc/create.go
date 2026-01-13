package documents_uc

import (
	"context"
	"fmt"
	"time"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/domain/vos"

	"github.com/google/uuid"
)

type CreateDocumentInput struct {
	OwnerType string
	OwnerID   uuid.UUID
	Type      string
	FileName  string
	FileURL   string
	Extension string
	IsActive  bool
	ObjectKey string // ✅
}

type CreateDocumentUseCase struct {
	Repo repos.DocumentRepository
}

func (uc *CreateDocumentUseCase) Execute(ctx context.Context, input CreateDocumentInput) (entities.Document, error) {

	ot := vos.DocumentOwnerType(input.OwnerType)
	if !ot.IsValid() {
		return entities.Document{}, fmt.Errorf("invalid owner_type")
	}
	doc := entities.Document{
		ID:         uuid.New(),
		OwnerType:  ot,
		OwnerID:    input.OwnerID,
		Type:       vos.MustNewDocumentType(input.Type),
		FileName:   vos.MustNewFilename(input.FileName),
		FileURL:    vos.MustNewDocumentURL(input.FileURL),
		Extension:  vos.MustNewFileExtension(input.Extension),
		IsActive:   input.IsActive,
		UploadedAt: time.Now(),
		ObjectKey:  input.ObjectKey, // ✅
	}

	if err := uc.Repo.Create(ctx, doc); err != nil {
		return entities.Document{}, err
	}
	return doc, nil
}
