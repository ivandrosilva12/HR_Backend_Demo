package dtos

import (
	"rhapp/internal/domain/entities"
	"time"

	"github.com/google/uuid"
)

type CreateWorkDTO struct {
	EmployeeID       string `json:"employee_id" binding:"required,uuid4"`
	Company          string `json:"company" binding:"required,min=2,max=100"`
	Position         string `json:"position" binding:"required,min=2,max=100"`
	StartDate        string `json:"start_date" binding:"required,datetime=2006-01-02"`
	EndDate          string `json:"end_date" binding:"omitempty,datetime=2006-01-02"`
	Responsibilities string `json:"responsibilities" binding:"omitempty,max=500"`
}

type UpdateWorkDTO struct {
	Company          string `json:"company" binding:"omitempty,min=2,max=100"`
	Position         string `json:"position" binding:"omitempty,min=2,max=100"`
	StartDate        string `json:"start_date" binding:"omitempty,datetime=2006-01-02"`
	EndDate          string `json:"end_date" binding:"omitempty,datetime=2006-01-02"`
	Responsibilities string `json:"responsibilities" binding:"omitempty,max=500"`
}

type WorkResponseDTO struct {
	ID               string    `json:"id"`
	EmployeeID       string    `json:"employee_id"`
	Company          string    `json:"company"`
	Position         string    `json:"position"`
	StartDate        string    `json:"start_date"`
	EndDate          *string   `json:"end_date,omitempty"`
	Responsibilities string    `json:"responsibilities"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func ToWorkResponseDTO(w entities.WorkHistory) WorkResponseDTO {
	var endDate *string
	if !w.EndDate.IsZero() {
		formatted := w.EndDate.Format("2006-01-02")
		endDate = &formatted
	}

	return WorkResponseDTO{
		ID:               w.ID.String(),
		EmployeeID:       w.EmployeeID.String(),
		Company:          w.Company,
		Position:         w.Position,
		StartDate:        w.StartDate.Format("2006-01-02"),
		EndDate:          endDate,
		Responsibilities: w.Responsibilities,
		CreatedAt:        w.CreatedAt,
		UpdatedAt:        w.UpdatedAt,
	}
}

func ToWorkResponseDTOList(list []entities.WorkHistory) []WorkResponseDTO {
	result := make([]WorkResponseDTO, len(list))
	for i, w := range list {
		result[i] = ToWorkResponseDTO(w)
	}
	return result
}

func ToWorkFromCreateDTO(input CreateWorkDTO) (entities.WorkHistory, error) {
	employeeID, err := uuid.Parse(input.EmployeeID)
	if err != nil {
		return entities.WorkHistory{}, err
	}

	company := input.Company

	position := input.Position

	startDate, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		return entities.WorkHistory{}, err
	}

	var endDate *time.Time
	if input.EndDate != "" {
		ed, err := time.Parse("2006-01-02", input.EndDate)
		if err != nil {
			return entities.WorkHistory{}, err
		}
		endDate = &ed
	}

	now := time.Now()

	return entities.WorkHistory{
		ID:               uuid.New(),
		EmployeeID:       employeeID,
		Company:          company,
		Position:         position,
		StartDate:        startDate,
		EndDate:          *endDate,
		Responsibilities: input.Responsibilities,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func ApplyUpdateToWork(w *entities.WorkHistory, input UpdateWorkDTO) error {
	if input.Company != "" {
		w.Company = input.Company
	}

	if input.Position != "" {
		w.Position = input.Position
	}

	if input.StartDate != "" {
		start, err := time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			return err
		}
		w.StartDate = start
	}

	if input.EndDate != "" {
		end, err := time.Parse("2006-01-02", input.EndDate)
		if err != nil {
			return err
		}
		w.EndDate = end
	}

	if input.Responsibilities != "" {
		w.Responsibilities = input.Responsibilities
	}

	w.UpdatedAt = time.Now()
	return nil
}
