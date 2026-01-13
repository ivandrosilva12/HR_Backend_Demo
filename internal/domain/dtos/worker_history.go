package dtos

import (
	"rhapp/internal/domain/entities"
	"time"
)

type CreateWorkerHistoryDTO struct {
	EmployeeID string `json:"employee_id" binding:"required,uuid4"`
	PositionID string `json:"position_id" binding:"required,uuid4"`
	StartDate  string `json:"start_date" binding:"required,datetime=2006-01-02"`
	EndDate    string `json:"end_date" binding:"omitempty,datetime=2006-01-02"`
	Status     string `json:"status" binding:"omitempty,oneof=activo inactivo"` // default = activo
}

type UpdateWorkerHistoryDTO struct {
	PositionID *string `json:"position_id,omitempty" binding:"omitempty,uuid4"`
	StartDate  string  `json:"start_date" binding:"omitempty,datetime=2006-01-02"`
	EndDate    string  `json:"end_date" binding:"omitempty,datetime=2006-01-02"`
	Status     string  `json:"status" binding:"omitempty,oneof=activo inactivo"`
}

type WorkerHistoryResponseDTO struct {
	ID         string    `json:"id"`
	EmployeeID string    `json:"employee_id"`
	PositionID string    `json:"position_id"`
	StartDate  string    `json:"start_date"`
	EndDate    *string   `json:"end_date,omitempty"`
	Status     string    `json:"status"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

func ToWorkerHistoryResponseDTO(w entities.WorkerHistory) WorkerHistoryResponseDTO {
	var end *string
	if w.EndDate != nil {
		s := w.EndDate.Format("2006-01-02")
		end = &s
	}
	return WorkerHistoryResponseDTO{
		ID:         w.ID.String(),
		EmployeeID: w.EmployeeID.String(),
		PositionID: w.PositionID.String(),
		StartDate:  w.StartDate.Format("2006-01-02"),
		EndDate:    end,
		Status:     string(w.Status),
		CreatedAt:  w.CreatedAt,
		UpdatedAt:  w.UpdatedAt,
	}
}

func ToWorkerHistoryResponseDTOList(list []entities.WorkerHistory) []WorkerHistoryResponseDTO {
	out := make([]WorkerHistoryResponseDTO, len(list))
	for i, it := range list {
		out[i] = ToWorkerHistoryResponseDTO(it)
	}
	return out
}
