package entities

import (
	"errors"
	"rhapp/internal/domain/vos"
	"time"

	"github.com/google/uuid"
)

type EducationHistory struct {
	ID           uuid.UUID
	EmployeeID   uuid.UUID
	Institution  string
	Degree       vos.SchoolDegree
	AreaEstudoID uuid.UUID
	StartDate    time.Time
	EndDate      time.Time
	Description  string
	IsCurrent    bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Validação de consistência de datas
func (s *EducationHistory) Validate() error {
	if s.EndDate.Before(s.StartDate) {
		return errors.New("data de término não pode ser anterior à data de início")
	}
	return nil
}
