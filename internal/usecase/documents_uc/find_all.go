package documents_uc

import (
	"context"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type ListDocumentsUseCase struct {
	Repo          repos.DocumentRepository
	DependentRepo repos.DependentRepository
	EmployeesRepo repos.EmployeeRepository
}

func (uc *ListDocumentsUseCase) Execute(ctx context.Context, ownerID uuid.UUID, limit, offset int) ([]entities.Document, error) {
	okDependent, err := uc.DependentRepo.ExistsByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	okEmployee, err := uc.EmployeesRepo.ExistsByID(ctx, ownerID)
	if err != nil {
		return nil, err
	}

	if !okDependent && !okEmployee {
		// dono não existe em nenhuma das tabelas relevantes
		return []entities.Document{}, nil
	}

	// Se o seu repo usa FindAllByOwner, troque a chamada abaixo.
	return uc.Repo.FindAllByOwner(ctx, ownerID, limit, offset)
}
