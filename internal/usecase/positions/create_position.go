package positions

import (
	"context"
	"fmt"
	"time"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type CreatePositionUseCase struct {
	Repo     repos.PositionRepository
	DeptRepo repos.DepartmentRepository
}

type CreatePositionInput struct {
	Name         string
	DepartmentID uuid.UUID
	MaxHeadcount int
	Tipo         string // NEW
}

func (uc *CreatePositionUseCase) Execute(ctx context.Context, input CreatePositionInput) (entities.Position, error) {
	if input.MaxHeadcount < 1 {
		return entities.Position{}, fmt.Errorf("max_headcount deve ser >=1")
	}
	if input.Tipo != "employee" && input.Tipo != "boss" { // NEW
		return entities.Position{}, fmt.Errorf("tipo inválido: use 'employee' ou 'boss'")
	}

	now := time.Now()
	position := entities.Position{
		ID:           uuid.New(),
		Nome:         input.Name,
		DepartmentID: input.DepartmentID,
		MaxHeadcount: input.MaxHeadcount,
		Tipo:         input.Tipo, // NEW
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := uc.Repo.Create(ctx, position); err != nil {
		return entities.Position{}, err
	}

	return position, nil
}
