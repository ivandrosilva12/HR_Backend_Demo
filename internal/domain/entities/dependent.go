package entities

import (
	"rhapp/internal/domain/vos"
	"time"

	"github.com/google/uuid"
)

type Dependent struct {
	ID           uuid.UUID
	EmployeeID   uuid.UUID
	FullName     vos.PersonalName
	Relationship vos.RelationshipType
	Gender       vos.Gender
	DateOfBirth  vos.BirthDate
	IsActive     bool
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
