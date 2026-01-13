// internal/usecase/departments/find_all.go
package departments

import (
	"context"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/utils"
)

type FindAllDepartmentsUseCase struct {
	Repo repos.DepartmentRepository
}

func (uc *FindAllDepartmentsUseCase) Execute(ctx context.Context, limit, offset int) (utils.PagedResponse[entities.Department], error) {
	return uc.Repo.FindAll(ctx, limit, offset)
}
