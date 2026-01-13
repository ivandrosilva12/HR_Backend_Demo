package dtos

import (
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/vos"
	"time"

	"github.com/google/uuid"
)

type CreateEducationDTO struct {
	EmployeeID   string `json:"employee_id" binding:"required,uuid4"`
	Institution  string `json:"institution" binding:"required,min=3,max=100"`
	Degree       string `json:"degree" binding:"required,min=2,max=50"`
	FieldOfStudy string `json:"field_of_study" binding:"required,uuid4"`
	StartDate    string `json:"start_date" binding:"required,datetime=2006-01-02"`
	EndDate      string `json:"end_date" binding:"required,datetime=2006-01-02"`
	Description  string `json:"description" binding:"omitempty,max=255"`
}

type UpdateEducationDTO struct {
	Institution  string `json:"institution" binding:"omitempty,min=3,max=100"`
	Degree       string `json:"degree" binding:"omitempty,min=2,max=50"`
	FieldOfStudy string `json:"field_of_study" binding:"omitempty,uuid4"`
	StartDate    string `json:"start_date" binding:"omitempty,datetime=2006-01-02"`
	EndDate      string `json:"end_date" binding:"omitempty,datetime=2006-01-02"`
	Description  string `json:"description" binding:"omitempty,max=255"`
}

type EducationResponseDTO struct {
	ID           string    `json:"id"`
	EmployeeID   string    `json:"employee_id"`
	Institution  string    `json:"institution"`
	Degree       string    `json:"degree"`
	FieldOfStudy string    `json:"field_of_study"`
	StartDate    string    `json:"start_date"`
	EndDate      string    `json:"end_date"`
	Description  string    `json:"description"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func ToEducationResponseDTO(e entities.EducationHistory) EducationResponseDTO {
	return EducationResponseDTO{
		ID:           e.ID.String(),
		EmployeeID:   e.EmployeeID.String(),
		Institution:  e.Institution,
		Degree:       e.Degree.String(),
		FieldOfStudy: e.AreaEstudoID.String(),
		StartDate:    e.StartDate.Format("2006-01-02"),
		EndDate:      e.EndDate.Format("2006-01-02"),
		Description:  e.Description,
		CreatedAt:    e.CreatedAt,
		UpdatedAt:    e.UpdatedAt,
	}
}

func ToEducationResponseDTOList(list []entities.EducationHistory) []EducationResponseDTO {
	result := make([]EducationResponseDTO, len(list))
	for i, e := range list {
		result[i] = ToEducationResponseDTO(e)
	}
	return result
}

func ToEducationFromCreateDTO(input CreateEducationDTO) (entities.EducationHistory, error) {
	employeeID, err := uuid.Parse(input.EmployeeID)
	if err != nil {
		return entities.EducationHistory{}, err
	}

	fieldID, err := uuid.Parse(input.FieldOfStudy)
	if err != nil {
		return entities.EducationHistory{}, err
	}

	degree, err := vos.NewSchoolDegree(input.Degree)
	if err != nil {
		return entities.EducationHistory{}, err
	}

	start, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		return entities.EducationHistory{}, err
	}

	end, err := time.Parse("2006-01-02", input.EndDate)
	if err != nil {
		return entities.EducationHistory{}, err
	}

	return entities.EducationHistory{
		ID:           uuid.New(),
		EmployeeID:   employeeID,
		Institution:  input.Institution,
		Degree:       degree,
		AreaEstudoID: fieldID,
		StartDate:    start,
		EndDate:      end,
		Description:  input.Description,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func ApplyUpdateToEducation(e *entities.EducationHistory, input UpdateEducationDTO) error {
	if input.Institution != "" {
		e.Institution = input.Institution
	}
	if input.Degree != "" {
		deg, err := vos.NewSchoolDegree(input.Degree)
		if err != nil {
			return err
		}
		e.Degree = deg
	}
	if input.FieldOfStudy != "" {
		fieldId, err := uuid.Parse(input.FieldOfStudy)
		if err != nil {
			return err
		}
		e.AreaEstudoID = fieldId
	}
	if input.StartDate != "" {
		start, err := time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			return err
		}
		e.StartDate = start
	}
	if input.EndDate != "" {
		end, err := time.Parse("2006-01-02", input.EndDate)
		if err != nil {
			return err
		}
		e.EndDate = end
	}
	if input.Description != "" {
		e.Description = input.Description
	}

	e.UpdatedAt = time.Now()
	return nil
}
