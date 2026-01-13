package dtos

import (
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/vos"
	"time"

	"github.com/google/uuid"
)

type CreateEmployeeStatusDTO struct {
	EmployeeID  string `json:"employee_id" binding:"required,uuid4"`
	Status      string `json:"status" binding:"required,oneof=activo suspenso reformado demitido convalescente"`
	Reason      string `json:"reason" binding:"required"`
	Observacoes string `json:"observacoes" binding:"omitempty,max=255"`
	StartDate   string `json:"start_date" binding:"required,datetime=2006-01-02"`
	EndDate     string `json:"end_date" binding:"omitempty,datetime=2006-01-02"`
}

type UpdateEmployeeStatusDTO struct {
	ID          string `json:"id"`
	Status      string `json:"status" binding:"omitempty,oneof=activo suspenso reformado demitido convalescente"`
	Reason      string `json:"reason" binding:"required"`
	Observacoes string `json:"observacoes" binding:"omitempty,max=255"`
	StartDate   string `json:"start_date" binding:"omitempty,datetime=2006-01-02"`
	EndDate     string `json:"end_date" binding:"omitempty,datetime=2006-01-02"`
}

type EmployeeStatusResponseDTO struct {
	ID          string    `json:"id"`
	EmployeeID  string    `json:"employee_id"`
	Status      string    `json:"status"`
	Reason      string    `json:"reason"`
	Observacoes string    `json:"observacoes"`
	StartDate   string    `json:"start_date"`
	EndDate     *string   `json:"end_date,omitempty"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

func ToEmployeeStatusResponseDTO(e entities.EmployeeStatus) EmployeeStatusResponseDTO {
	var endDate *string
	if e.EndDate != nil {
		formatted := e.EndDate.Format("2006-01-02")
		endDate = &formatted
	}

	return EmployeeStatusResponseDTO{
		ID:         e.ID.String(),
		EmployeeID: e.EmployeeID.String(),
		Status:     e.Status.String(),
		Reason:     e.Reason.String(),
		StartDate:  e.StartDate.Format("2006-01-02"),
		EndDate:    endDate,
		CreatedAt:  e.CreatedAt,
		UpdatedAt:  e.UpdatedAt,
	}
}

func ToEmployeeStatusResponseDTOList(list []entities.EmployeeStatus) []EmployeeStatusResponseDTO {
	result := make([]EmployeeStatusResponseDTO, len(list))
	for i, e := range list {
		result[i] = ToEmployeeStatusResponseDTO(e)
	}
	return result
}

func ToEmployeeStatusFromCreateDTO(input CreateEmployeeStatusDTO) (entities.EmployeeStatus, error) {
	employeeID, err := uuid.Parse(input.EmployeeID)
	if err != nil {
		return entities.EmployeeStatus{}, err
	}

	start, err := time.Parse("2006-01-02", input.StartDate)
	if err != nil {
		return entities.EmployeeStatus{}, err
	}

	status, err := vos.NewEmployeeStatusValue(input.Status)
	if err != nil {
		return entities.EmployeeStatus{}, err
	}

	reason, err := vos.NewStatusReason(input.Reason)
	if err != nil {
		return entities.EmployeeStatus{}, err
	}

	var end *time.Time
	if input.EndDate != "" {
		e, err := time.Parse("2006-01-02", input.EndDate)
		if err != nil {
			return entities.EmployeeStatus{}, err
		}
		end = &e
	}

	now := time.Now()

	return entities.EmployeeStatus{
		ID:         uuid.New(),
		EmployeeID: employeeID,
		Status:     status,
		Reason:     reason,
		StartDate:  start,
		EndDate:    end,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

func ApplyUpdateToEmployeeStatus(es *entities.EmployeeStatus, input UpdateEmployeeStatusDTO) error {
	if input.Status != "" {
		status, err := vos.NewEmployeeStatusValue(input.Status)
		if err != nil {
			return err
		}
		es.Status = status
	}

	if input.Reason != "" {
		reason, err := vos.NewStatusReason(input.Reason)
		if err != nil {
			return err
		}
		es.Reason = reason
	}

	if input.StartDate != "" {
		start, err := time.Parse("2006-01-02", input.StartDate)
		if err != nil {
			return err
		}
		es.StartDate = start
	}

	if input.EndDate != "" {
		end, err := time.Parse("2006-01-02", input.EndDate)
		if err != nil {
			return err
		}
		es.EndDate = &end
	}

	es.UpdatedAt = time.Now()
	return nil
}
