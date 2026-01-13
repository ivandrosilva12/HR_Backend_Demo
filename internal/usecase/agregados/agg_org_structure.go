package agregados

import (
	"context"

	"rhapp/internal/domain/agregados"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type GetOrgStructureByDepartmentInput struct {
	DepartmentID uuid.UUID
}

type GetOrgStructureByDepartmentUseCase struct {
	Repo repos.OrgStructureAggregateRepository
}

func NewGetOrgStructureByDepartmentUseCase(repo repos.OrgStructureAggregateRepository) *GetOrgStructureByDepartmentUseCase {
	return &GetOrgStructureByDepartmentUseCase{Repo: repo}
}

func (uc *GetOrgStructureByDepartmentUseCase) Execute(ctx context.Context, input GetOrgStructureByDepartmentInput) (*agregados.OrgStructureAggregate, error) {
	return uc.Repo.GetByDepartmentID(ctx, input.DepartmentID)
}
