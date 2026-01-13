package dtos

import (
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/vos"
	"time"

	"github.com/google/uuid"
)

type CreateDistritoDTO struct {
	Nome        string `json:"nome" binding:"required,min=2,max=100"`
	MunicipioID string `json:"municipio_id" binding:"required,uuid4"`
}

type UpdateDistritoDTO struct {
	Nome        string `json:"nome" binding:"omitempty,min=2,max=100"`
	MunicipioID string `json:"municipio_id" binding:"omitempty,uuid4"`
}

type DistritoResponseDTO struct {
	ID          string    `json:"id"`
	Nome        string    `json:"nome"`
	MunicipioID string    `json:"municipio_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type DistritoResultDTO struct {
	ID            uuid.UUID `json:"id"`
	Nome          string    `json:"nome"`
	MunicipioID   uuid.UUID `json:"municipio_id"`
	MunicipioNome string    `json:"municipio_nome"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

func ToDistritoResponseDTO(d entities.District) DistritoResponseDTO {
	return DistritoResponseDTO{
		ID:          d.ID.String(),
		Nome:        d.Nome.String(),
		MunicipioID: d.MunicipalityID.String(),
		CreatedAt:   d.CreatedAt,
		UpdatedAt:   d.UpdatedAt,
	}
}

func ToDistritoFromCreateDTO(input CreateDistritoDTO) (entities.District, error) {
	municipioID, err := uuid.Parse(input.MunicipioID)
	if err != nil {
		return entities.District{}, err
	}

	nome, err := vos.NewDistrict(input.Nome)
	if err != nil {
		return entities.District{}, err
	}

	now := time.Now()

	return entities.District{
		ID:             uuid.New(),
		Nome:           nome,
		MunicipalityID: municipioID,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func ApplyUpdateToDistrito(d *entities.District, input UpdateDistritoDTO) error {
	if input.Nome != "" {
		nome, err := vos.NewDistrict(input.Nome)
		if err != nil {
			return err
		}
		d.Nome = nome
	}

	if input.MunicipioID != "" {
		municipioID, err := uuid.Parse(input.MunicipioID)
		if err != nil {
			return err
		}
		d.MunicipalityID = municipioID
	}

	d.UpdatedAt = time.Now()
	return nil
}

func ToDistritoResponseDTOList(list []entities.District) []DistritoResponseDTO {
	result := make([]DistritoResponseDTO, len(list))
	for i, d := range list {
		result[i] = ToDistritoResponseDTO(d)
	}
	return result
}
