package agregados

import (
	"context"
	"rhapp/internal/domain/entities"
	"time"

	"github.com/google/uuid"
)

type EmployeeAggregate struct {
	Employee           entities.Employee
	Statuses           []entities.EmployeeStatus
	Dependents         []entities.Dependent
	EducationHistories []entities.EducationHistory
	WorkHistories      []entities.WorkHistory
	Documents          []entities.Document
	WorkerHistory      []entities.WorkerHistory
}

func (agg *EmployeeAggregate) AddDependent(dep entities.Dependent) {
	agg.Dependents = append(agg.Dependents, dep)
}

func (agg *EmployeeAggregate) AddWorkHistory(work entities.WorkHistory) {
	agg.WorkHistories = append(agg.WorkHistories, work)
}

func (agg *EmployeeAggregate) AddEducationHistory(edu entities.EducationHistory) {
	agg.EducationHistories = append(agg.EducationHistories, edu)
}

func (agg *EmployeeAggregate) AddStatus(status entities.EmployeeStatus) {
	for i := range agg.Statuses {
		agg.Statuses[i].IsCurrent = false
	}
	status.IsCurrent = true
	agg.Statuses = append(agg.Statuses, status)
}

func (agg *EmployeeAggregate) AddDocument(doc entities.Document) {
	agg.Documents = append(agg.Documents, doc)
}

func (agg *EmployeeAggregate) AddSupervisor(super entities.WorkerHistory) {
	for i := range agg.WorkerHistory {
		if agg.WorkerHistory[i].EndDate == nil {
			now := time.Now()
			agg.WorkerHistory[i].EndDate = &now
		}
	}
	agg.WorkerHistory = append(agg.WorkerHistory, super)
}

type EmployeeAggregateRepository interface {
	GetFullByID(ctx context.Context, id uuid.UUID) (*EmployeeAggregate, error)
}
