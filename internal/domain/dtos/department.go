package dtos

import (
	"rhapp/internal/domain/entities"
	"time"

	"github.com/google/uuid"
)

type CreateDepartmentDTO struct {
	Nome     string  `json:"nome" binding:"required,min=2,max=100"`
	ParentID *string `json:"parent_id" binding:"omitempty,uuid4"`
}

type UpdateDepartmentDTO struct {
	Nome     string  `json:"nome" binding:"omitempty,min=2,max=100"`
	ParentID *string `json:"parent_id" binding:"omitempty,uuid4"`
}

type DepartmentResponseDTO struct {
	ID        string    `json:"id"`
	Nome      string    `json:"nome"`
	ParentID  *string   `json:"parent_id,omitempty"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type DepartmentPositionTotals struct {
	DepartmentID       uuid.UUID `json:"departmentId"`
	DepartmentNome     string    `json:"departmentNome"`
	TotalPositions     int       `json:"totalPositions"`
	OccupiedPositions  int       `json:"occupiedPositions"`
	AvailablePositions int       `json:"availablePositions"`
}

func ToDepartmentFromCreateDTO(input CreateDepartmentDTO) entities.Department {
	now := time.Now()
	var pid *uuid.UUID
	if input.ParentID != nil && *input.ParentID != "" {
		if v, err := uuid.Parse(*input.ParentID); err == nil {
			pid = &v
		}
	}
	return entities.Department{
		ID:        uuid.New(),
		Nome:      input.Nome,
		ParentID:  pid,
		CreatedAt: now,
		UpdatedAt: now,
	}
}

func ToDepartmentResponseDTO(d entities.Department) DepartmentResponseDTO {
	var pid *string
	if d.ParentID != nil {
		v := d.ParentID.String()
		pid = &v
	}
	return DepartmentResponseDTO{
		ID:        d.ID.String(),
		Nome:      d.Nome,
		ParentID:  pid,
		CreatedAt: d.CreatedAt,
		UpdatedAt: d.UpdatedAt,
	}
}

func ApplyUpdateToDepartment(d *entities.Department, input UpdateDepartmentDTO) {
	if input.Nome != "" {
		d.Nome = input.Nome
	}
	if input.ParentID != nil {
		if *input.ParentID == "" {
			d.ParentID = nil
		} else if v, err := uuid.Parse(*input.ParentID); err == nil {
			d.ParentID = &v
		}
	}
	d.UpdatedAt = time.Now()
}

func ToDepartmentResponseDTOList(list []entities.Department) []DepartmentResponseDTO {
	result := make([]DepartmentResponseDTO, len(list))
	for i, d := range list {
		result[i] = ToDepartmentResponseDTO(d)
	}
	return result
}
