package dtos

import (
	"rhapp/internal/domain/entities"
	"time"

	"github.com/google/uuid"
)

type CreateAreaEstudoDTO struct {
	Nome        string `json:"nome" binding:"required,min=3,max=100"`
	Description string `json:"descricao" binding:"required,min=5,max=255"`
}

type UpdateAreaEstudoDTO struct {
	Nome        string `json:"nome" binding:"omitempty,min=3,max=100"`
	Description string `json:"descricao" binding:"omitempty,min=5,max=255"`
}

type AreaEstudoResponseDTO struct {
	ID          string    `json:"id"`
	Nome        string    `json:"nome"`
	Description string    `json:"descricao"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToAreaEstudoResponseDTO(a entities.AreaEstudo) AreaEstudoResponseDTO {
	return AreaEstudoResponseDTO{
		ID:          a.ID.String(),
		Nome:        a.Nome,
		Description: a.Description,
		CreatedAt:   a.CreatedAt,
		UpdatedAt:   a.UpdatedAt,
	}
}

func ToAreaEstudoFromCreateDTO(input CreateAreaEstudoDTO) entities.AreaEstudo {
	now := time.Now()
	return entities.AreaEstudo{
		ID:          uuid.New(),
		Nome:        input.Nome,
		Description: input.Description,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
}

func ToAreaEstudoResponseDTOList(areas []entities.AreaEstudo) []AreaEstudoResponseDTO {
	result := make([]AreaEstudoResponseDTO, len(areas))
	for i, a := range areas {
		result[i] = ToAreaEstudoResponseDTO(a)
	}
	return result
}

func ApplyUpdateToAreaEstudo(area *entities.AreaEstudo, input UpdateAreaEstudoDTO) {

	if input.Nome != "" {
		area.Nome = input.Nome
	}
	if input.Description != "" {
		area.Description = input.Description
	}
	area.UpdatedAt = time.Now()
}
