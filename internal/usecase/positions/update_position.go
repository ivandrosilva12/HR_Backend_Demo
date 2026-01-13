package positions

import (
	"context"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type UpdatePositionUseCase struct {
	Repo     repos.PositionRepository
	DeptRepo repos.DepartmentRepository
}

type UpdatePositionInput struct {
	ID           uuid.UUID
	Nome         string
	DepartmentID uuid.UUID
	MaxHeadcount *int
	Tipo         *string // NEW
}

func (uc *UpdatePositionUseCase) Execute(ctx context.Context, input UpdatePositionInput) (entities.Position, error) {
	position, err := uc.Repo.FindByID(ctx, input.ID)
	if err != nil {
		return entities.Position{}, err
	}

	err = dtos.ApplyUpdateToPosition(&position, dtos.UpdatePositionDTO{
		Nome:         input.Nome,
		DepartmentID: input.DepartmentID.String(),
		MaxHeadcount: input.MaxHeadcount,
		Tipo:         input.Tipo, // NEW
	})
	if err != nil {
		return entities.Position{}, err
	}

	if err := uc.Repo.Update(ctx, position); err != nil {
		return entities.Position{}, err
	}

	return position, nil
}
