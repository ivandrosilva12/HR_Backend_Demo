package employee_status

import (
	"context"
	"time"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/domain/vos"

	"github.com/google/uuid"
)

type CreateEmployeeStatusInput struct {
	EmployeeID  uuid.UUID
	Status      string
	Reason      string
	Observacoes string
	StartDate   time.Time
	EndDate     *time.Time
	IsCurrent   bool
}

type CreateEmployeeStatusUseCase struct {
	Repo repos.EmployeeStatusRepository
}

func (uc *CreateEmployeeStatusUseCase) Execute(ctx context.Context, input CreateEmployeeStatusInput) (entities.EmployeeStatus, error) {
	now := time.Now()
	s := entities.EmployeeStatus{
		ID:          uuid.New(),
		EmployeeID:  input.EmployeeID,
		Status:      vos.MustNewEmployeeStatusValue(input.Status),
		Reason:      vos.MustNewStatusReason(input.Reason),
		Observacoes: input.Observacoes,
		StartDate:   input.StartDate,
		EndDate:     input.EndDate,
		IsCurrent:   input.IsCurrent,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := uc.Repo.Create(ctx, s); err != nil {
		return entities.EmployeeStatus{}, err
	}
	return s, nil
}
