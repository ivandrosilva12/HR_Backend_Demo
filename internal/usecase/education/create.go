package education

import (
	"context"
	"time"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/domain/vos"

	"github.com/google/uuid"
)

type CreateEducationHistoryInput struct {
	EmployeeID   uuid.UUID
	Institution  string
	Degree       string
	AreaEstudoID uuid.UUID
	StartDate    time.Time
	EndDate      *time.Time
	Description  string
}

type CreateEducationHistoryUseCase struct {
	Repo repos.EducationHistoryRepository
}

func (uc *CreateEducationHistoryUseCase) Execute(ctx context.Context, input CreateEducationHistoryInput) (entities.EducationHistory, error) {
	history := entities.EducationHistory{
		ID:           uuid.New(),
		EmployeeID:   input.EmployeeID,
		Institution:  input.Institution,
		Degree:       vos.MustNewSchoolDegree(input.Degree),
		AreaEstudoID: input.AreaEstudoID,
		StartDate:    input.StartDate,
		EndDate:      *input.EndDate,
		Description:  input.Description,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := uc.Repo.Create(ctx, history); err != nil {
		return entities.EducationHistory{}, err
	}
	return history, nil
}
