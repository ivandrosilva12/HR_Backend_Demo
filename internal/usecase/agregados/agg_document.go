package agregados

import (
	"context"
	"rhapp/internal/domain/agregados"
	"rhapp/internal/domain/vos"

	"github.com/google/uuid"
)

type GetDocumentsByOwnerInput struct {
	OwnerType vos.DocumentOwnerType
	OwnerID   uuid.UUID
}

type GetDocumentsByOwnerUseCase struct {
	Repo agregados.DocumentAggregateRepository
}

func NewGetDocumentsByOwnerUseCase(repo agregados.DocumentAggregateRepository) *GetDocumentsByOwnerUseCase {
	return &GetDocumentsByOwnerUseCase{Repo: repo}
}

func (uc *GetDocumentsByOwnerUseCase) Execute(ctx context.Context, input GetDocumentsByOwnerInput) (*agregados.DocumentAggregate, error) {
	return uc.Repo.GetByOwner(ctx, input.OwnerType, input.OwnerID)
}
