package entities

import (
	"errors"
	"rhapp/internal/domain/vos"
	"time"

	"github.com/google/uuid"
)

type Employee struct {
	ID               uuid.UUID
	EmployeeNumber   int
	FullName         vos.PersonalName
	Gender           vos.Gender
	DateOfBirth      time.Time
	Nationality      vos.Nationality
	MaritalStatus    vos.MaritalStatus
	PhoneNumber      vos.PhoneNumber
	Email            vos.Email
	BI               vos.BI
	IDValidationDate time.Time
	IBAN             vos.IBAN
	DepartmentID     uuid.UUID
	PositionID       uuid.UUID
	Address          vos.Address
	DistrictID       uuid.UUID
	HiringDate       time.Time
	ContractType     vos.ContractType
	Salary           vos.Salary
	SocialSecurity   vos.SocialSecurity
	IsActive         bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// Validação de consistência de datas
func (s *EmployeeStatus) Validate() error {
	if s.EndDate != nil && s.EndDate.Before(s.StartDate) {
		return errors.New("data de término não pode ser anterior à data de início")
	}
	return nil
}
