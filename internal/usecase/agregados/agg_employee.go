package agregados

import (
	"context"

	"rhapp/internal/domain/agregados"

	"github.com/google/uuid"
)

/*
***********

/employees/:id/aggregate

***************
*/
type GetEmployeeAggregateByIDUseCase struct {
	Repo agregados.EmployeeAggregateRepository
}

type GetEmployeeAggregateInput struct {
	ID uuid.UUID
}

func (uc *GetEmployeeAggregateByIDUseCase) Execute(ctx context.Context, input GetEmployeeAggregateInput) (*agregados.EmployeeAggregate, error) {
	return uc.Repo.GetFullByID(ctx, input.ID)
}
