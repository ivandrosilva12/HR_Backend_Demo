package dtos

import (
	"rhapp/internal/domain/entities"
	"time"

	"github.com/google/uuid"
)

type CreatePositionDTO struct {
	Nome         string `json:"nome" binding:"required,min=2,max=100"`
	DepartmentID string `json:"department_id" binding:"required,uuid4"`
	MaxHeadcount int    `json:"max_headcount" binding:"required,min=1"`
	Tipo         string `json:"tipo" binding:"required,oneof=employee boss"` // NEW
}

type UpdatePositionDTO struct {
	Nome         string  `json:"nome" binding:"omitempty,min=2,max=100"`
	DepartmentID string  `json:"department_id" binding:"required,uuid4"`
	MaxHeadcount *int    `json:"max_headcount" binding:"omitempty,min=1"`
	Tipo         *string `json:"tipo" binding:"omitempty,oneof=employee boss"` // NEW
}

type PositionResponseDTO struct {
	ID               string    `json:"id"`
	Nome             string    `json:"nome"`
	DepartmentID     string    `json:"department_id"`
	MaxHeadcount     int       `json:"max_headcount"`
	CurrentHeadcount int       `json:"current_headcount"`
	Remaining        int       `json:"remaining"`
	Tipo             string    `json:"tipo"` // NEW
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type PositionResultDTO struct {
	ID               uuid.UUID `json:"id"`
	Nome             string    `json:"nome"`
	DepartmentID     uuid.UUID `json:"department_id"`
	MaxHeadcount     int       `json:"max_headcount"`
	CurrentHeadcount int       `json:"current_headcount"`
	Remaining        int       `json:"remaining"`
	DepartmentNome   string    `json:"department_nome"`
	Tipo             string    `json:"tipo"` // NEW
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func ToPositionResponseDTO(p entities.Position) PositionResponseDTO {
	return PositionResponseDTO{
		ID:               p.ID.String(),
		Nome:             p.Nome,
		DepartmentID:     p.DepartmentID.String(),
		MaxHeadcount:     p.MaxHeadcount,
		CurrentHeadcount: p.CurrentHeadcount,
		Remaining:        p.Remaining,
		Tipo:             p.Tipo, // NEW
		CreatedAt:        p.CreatedAt,
		UpdatedAt:        p.UpdatedAt,
	}
}

func ToPositionResponseDTOList(list []entities.Position) []PositionResponseDTO {
	result := make([]PositionResponseDTO, len(list))
	for i, p := range list {
		result[i] = ToPositionResponseDTO(p)
	}
	return result
}

func ToPositionFromCreateDTO(input CreatePositionDTO) (entities.Position, error) {
	deptID, err := uuid.Parse(input.DepartmentID)
	if err != nil {
		return entities.Position{}, err
	}

	now := time.Now()

	return entities.Position{
		ID:           uuid.New(),
		Nome:         input.Nome,
		DepartmentID: deptID,
		MaxHeadcount: input.MaxHeadcount,
		Tipo:         input.Tipo, // NEW
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func ApplyUpdateToPosition(p *entities.Position, input UpdatePositionDTO) error {
	if input.Nome != "" {
		p.Nome = input.Nome
	}

	if input.DepartmentID != "" {
		deptID, err := uuid.Parse(input.DepartmentID)
		if err != nil {
			return err
		}
		p.DepartmentID = deptID
	}

	if input.MaxHeadcount != nil {
		p.MaxHeadcount = *input.MaxHeadcount
	}

	if input.Tipo != nil { // NEW
		p.Tipo = *input.Tipo
	}

	p.UpdatedAt = time.Now()
	return nil
}
