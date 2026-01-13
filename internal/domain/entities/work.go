package entities

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

type WorkHistory struct {
	ID               uuid.UUID
	EmployeeID       uuid.UUID
	Company          string
	Position         string
	StartDate        time.Time
	EndDate          time.Time
	Responsibilities string
	IsCurrent        bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Validação de consistência de datas
func (s *WorkHistory) Validate() error {
	if s.EndDate.Before(s.StartDate) {
		return errors.New("data de término não pode ser anterior à data de início")
	}
	return nil
}
