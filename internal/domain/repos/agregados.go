package repos

import (
	"context"
	"rhapp/internal/domain/agregados"
	"rhapp/internal/domain/vos"

	"github.com/google/uuid"
)

type DocumentAggregateRepository interface {
	GetByOwner(ctx context.Context, ownerType vos.DocumentOwnerType, ownerID uuid.UUID) (*agregados.DocumentAggregate, error)
}

type EmployeeAggregateRepository interface {
	GetFullByID(ctx context.Context, id uuid.UUID) (*agregados.EmployeeAggregate, error)
}

type LocationAggregateRepository interface {
	GetByProvinceID(ctx context.Context, id uuid.UUID) (*agregados.LocationAggregate, error)
}

type OrgStructureAggregateRepository interface {
	GetByDepartmentID(ctx context.Context, departmentID uuid.UUID) (*agregados.OrgStructureAggregate, error)
}
