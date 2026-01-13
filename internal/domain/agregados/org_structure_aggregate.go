package agregados

import (
	"context"
	"rhapp/internal/domain/entities"
	"time"

	"github.com/google/uuid"
)

type OrgStructureAggregate struct {
	Department entities.Department
	Positions  []entities.Position
	Employees  []entities.Employee
}

func (agg *OrgStructureAggregate) AddPosition(position entities.Position) {
	agg.Positions = append(agg.Positions, position)
}

func (agg *OrgStructureAggregate) AssignEmployeeToPosition(emp entities.Employee, positionID uuid.UUID) {
	for i := range agg.Employees {
		if agg.Employees[i].ID == emp.ID {
			agg.Employees[i].PositionID = positionID
			agg.Employees[i].UpdatedAt = time.Now()
			return
		}
	}
	emp.PositionID = positionID
	emp.UpdatedAt = time.Now()
	agg.Employees = append(agg.Employees, emp)
}

func (agg *OrgStructureAggregate) ChangeDepartmentForPosition(positionID uuid.UUID, departmentID uuid.UUID) {
	for i := range agg.Positions {
		if agg.Positions[i].ID == positionID {
			agg.Positions[i].DepartmentID = departmentID
			agg.Positions[i].UpdatedAt = time.Now()
			return
		}
	}
}

type OrgStructureAggregateRepository interface {
	GetByDepartmentID(ctx context.Context, id uuid.UUID) (*OrgStructureAggregate, error)
}
