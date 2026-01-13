package employee_status

import (
	"context"
	"time"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type UpdateEmployeeStatusInput struct {
	ID          uuid.UUID
	Status      string
	Reason      string
	Observacoes string
	StartDate   time.Time
	EndDate     *time.Time
	IsCurrent   bool
}

type UpdateEmployeeStatusUseCase struct {
	Repo repos.EmployeeStatusRepository
}

func (uc *UpdateEmployeeStatusUseCase) Execute(ctx context.Context, input UpdateEmployeeStatusInput) (entities.EmployeeStatus, error) {
	s, err := uc.Repo.FindByID(ctx, input.ID)
	if err != nil {
		return entities.EmployeeStatus{}, err
	}

	err = dtos.ApplyUpdateToEmployeeStatus(&s, dtos.UpdateEmployeeStatusDTO{
		Status:      input.Status,
		Reason:      input.Reason,
		Observacoes: input.Observacoes,
		StartDate:   input.StartDate.Format("2006-01-02"),
		EndDate:     input.EndDate.Format("2006-01-02"),
	})

	if err != nil {
		return entities.EmployeeStatus{}, err
	}

	if err := uc.Repo.Update(ctx, s); err != nil {
		return entities.EmployeeStatus{}, err
	}

	return s, nil
}
