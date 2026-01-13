package dtos

import (
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/vos"
	"time"

	"github.com/google/uuid"
)

type CreateProvinceDTO struct {
	Nome string `json:"nome" binding:"required,min=3,max=100"`
}

// CreateProvinceDoc is used ONLY for Swagger schema generation.
// It intentionally has NO validator tags, so Swagger won't enforce minLength.
type CreateProvinceDoc struct {
	// Nome da Província
	// example: Luanda
	Nome string `json:"nome" example:"Luanda"`
}

type UpdateProvinceDTO struct {
	Nome string `json:"nome" binding:"required,min=3,max=100"`
}

type ProvinceResponseDTO struct {
	ID        string    `json:"id"`
	Nome      string    `json:"nome"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func ToProvinceFromCreateDTO(input CreateProvinceDTO) (entities.Province, error) {
	nome := vos.NewProvince(input.Nome)
	now := time.Now()

	return entities.Province{
		ID:        uuid.New(),
		Nome:      nome,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

func ApplyUpdateToProvince(p *entities.Province, input UpdateProvinceDTO) {
	if input.Nome != "" {
		nome := vos.NewProvince(input.Nome)
		p.Nome = nome
	}

	p.UpdatedAt = time.Now()
}

func ToProvinceResponseDTO(p entities.Province) ProvinceResponseDTO {
	return ProvinceResponseDTO{
		ID:        p.ID.String(),
		Nome:      p.Nome.String(),
		CreatedAt: p.CreatedAt,
		UpdatedAt: p.UpdatedAt,
	}
}

func ToProvinceResponseDTOList(list []entities.Province) []ProvinceResponseDTO {
	result := make([]ProvinceResponseDTO, len(list))
	for i, p := range list {
		result[i] = ToProvinceResponseDTO(p)
	}
	return result
}
