package dtos

import (
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/vos"
	"time"

	"github.com/google/uuid"
)

type CreateMunicipioDTO struct {
	Nome        string `json:"nome" binding:"required,min=3,max=100"`
	ProvinciaID string `json:"provincia_id" binding:"required,uuid4"`
}

// CreateProvinceDoc is used ONLY for Swagger schema generation.
// It intentionally has NO validator tags, so Swagger won't enforce minLength.
type CreateMunicipioDoc struct {
	Nome        string `json:"nome"`
	ProvinciaID string `json:"provincia_id"`
}

type UpdateMunicipioDTO struct {
	Nome        string `json:"nome" binding:"omitempty,min=3,max=100"`
	ProvinciaID string `json:"provincia_id" binding:"omitempty,uuid4"`
}

type MunicipioResponseDTO struct {
	ID          string    `json:"id"`
	Nome        string    `json:"nome"`
	ProvinciaID string    `json:"provincia_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type MunicipioResultDTO struct {
	ID            string    `json:"id"`
	Nome          string    `json:"nome"`
	ProvinciaID   string    `json:"provincia_id"`
	ProvinciaNome string    `json:"provincia_nome"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func ToMunicipioResponseDTO(m entities.Municipality) MunicipioResponseDTO {
	return MunicipioResponseDTO{
		ID:          m.ID.String(),
		Nome:        m.Nome.String(),
		ProvinciaID: m.ProvinceID.String(),
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

func ToMunicipioResponseDTOList(list []entities.Municipality) []MunicipioResponseDTO {
	result := make([]MunicipioResponseDTO, len(list))
	for i, m := range list {
		result[i] = ToMunicipioResponseDTO(m)
	}
	return result
}

func ToMunicipioFromCreateDTO(input CreateMunicipioDTO) (entities.Municipality, error) {
	provinciaID, err := uuid.Parse(input.ProvinciaID)
	if err != nil {
		return entities.Municipality{}, err
	}

	nome := vos.NewMunicipality(input.Nome)

	now := time.Now()

	return entities.Municipality{
		ID:         uuid.New(),
		Nome:       nome,
		ProvinceID: provinciaID,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func ApplyUpdateToMunicipio(m *entities.Municipality, input UpdateMunicipioDTO) error {
	if input.Nome != "" {
		m.Nome = vos.NewMunicipality(input.Nome)
	}

	if input.ProvinciaID != "" {
		provinciaID, err := uuid.Parse(input.ProvinciaID)
		if err != nil {
			return err
		}
		m.ProvinceID = provinciaID
	}

	m.UpdatedAt = time.Now()
	return nil
}
