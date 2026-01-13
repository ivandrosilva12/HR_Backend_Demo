// internal/usecase/departments/department_position_totals.go
package departments

import (
	"context"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type DepartmentPositionTotalsInput struct {
	DepartmentRoot  uuid.UUID
	IncludeChildren bool
}

type DepartmentPositionTotalsUseCase struct {
	Repo repos.DepartmentRepository
}

func (uc *DepartmentPositionTotalsUseCase) Execute(
	ctx context.Context,
	in DepartmentPositionTotalsInput,
) ([]dtos.DepartmentPositionTotals, error) {
	return uc.Repo.DepartmentPositionTotals(ctx, in.DepartmentRoot, in.IncludeChildren)
}
