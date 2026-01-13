package dtos

import (
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/vos"
	"time"

	"github.com/google/uuid"
)

type CreateDependentDTO struct {
	EmployeeID   string `json:"employee_id" binding:"required,uuid4"`
	FullName     string `json:"full_name" binding:"required,min=3,max=100"`
	Relationship string `json:"relationship" binding:"required,oneof=filho filha cônjuge pai mãe sobrinho sobrinha irmão irmã tio tia avó avô"`
	Gender       string `json:"gender" binding:"required,oneof=masculino feminino other"`
	DateOfBirth  string `json:"date_of_birth" binding:"required,datetime=2006-01-02"`
}

type UpdateDependentDTO struct {
	FullName     string `json:"full_name" binding:"omitempty,min=3,max=100"`
	Relationship string `json:"relationship" binding:"required,oneof=filho filha cônjuge pai mãe sobrinho sobrinha irmão irmã tio tia avó avô"`
	Gender       string `json:"gender" binding:"omitempty,oneof=masculino feminino other"`
	DateOfBirth  string `json:"date_of_birth" binding:"omitempty,datetime=2006-01-02"`
	IsActive     bool   `json:"is_active,omitempty"`
}

type DependentResponseDTO struct {
	ID           string    `json:"id"`
	EmployeeID   string    `json:"employee_id"`
	FullName     string    `json:"full_name"`
	Relationship string    `json:"relationship"`
	Gender       string    `json:"gender"`
	DateOfBirth  string    `json:"date_of_birth"`
	IsActive     bool      `json:"is_active"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

func ToDependentResponseDTO(d entities.Dependent) DependentResponseDTO {
	return DependentResponseDTO{
		ID:           d.ID.String(),
		EmployeeID:   d.EmployeeID.String(),
		FullName:     d.FullName.String(),
		Relationship: d.Relationship.String(),
		Gender:       d.Gender.String(),
		DateOfBirth:  d.DateOfBirth.String(),
		IsActive:     d.IsActive,
		CreatedAt:    d.CreatedAt,
		UpdatedAt:    d.UpdatedAt,
	}
}

func ToDependentFromCreateDTO(input CreateDependentDTO) (entities.Dependent, error) {
	employeeID, err := uuid.Parse(input.EmployeeID)
	if err != nil {
		return entities.Dependent{}, err
	}

	now := time.Now()

	return entities.Dependent{
		ID:           uuid.New(),
		EmployeeID:   employeeID,
		FullName:     vos.MustNewPersonalName(input.FullName),
		Relationship: vos.MustNewRelationshipType(input.Relationship),
		Gender:       vos.MustNewGender(input.Gender),
		DateOfBirth:  vos.MustNewBirthDate(input.DateOfBirth),
		IsActive:     true,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func ApplyUpdateToDependent(d *entities.Dependent, input UpdateDependentDTO) error {
	if input.FullName != "" {
		d.FullName = vos.MustNewPersonalName(input.FullName)
	}

	if input.Relationship != "" {
		d.Relationship = vos.MustNewRelationshipType(input.Relationship)
	}

	if input.Gender != "" {
		d.Gender = vos.MustNewGender(input.Gender)
	}

	if input.DateOfBirth != "" {
		d.DateOfBirth = vos.MustNewBirthDate(input.DateOfBirth)
	}

	d.IsActive = input.IsActive

	d.UpdatedAt = time.Now()
	return nil
}

func ToDependentResponseDTOList(list []entities.Dependent) []DependentResponseDTO {
	result := make([]DependentResponseDTO, len(list))
	for i, d := range list {
		result[i] = ToDependentResponseDTO(d)
	}
	return result
}
