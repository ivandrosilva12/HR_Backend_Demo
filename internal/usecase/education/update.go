package education

import (
	"context"
	"time"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type UpdateEducationHistoryInput struct {
	ID           uuid.UUID
	Institution  string
	Degree       string
	AreaEstudoID uuid.UUID
	StartDate    time.Time
	EndDate      *time.Time
	Description  string
}

type UpdateEducationHistoryUseCase struct {
	Repo repos.EducationHistoryRepository
}

func (uc *UpdateEducationHistoryUseCase) Execute(ctx context.Context, input UpdateEducationHistoryInput) (entities.EducationHistory, error) {
	h, err := uc.Repo.FindByID(ctx, input.ID)
	if err != nil {
		return entities.EducationHistory{}, err
	}

	err = dtos.ApplyUpdateToEducation(&h, dtos.UpdateEducationDTO{
		Institution:  input.Institution,
		Degree:       input.Degree,
		FieldOfStudy: input.AreaEstudoID.String(),
		StartDate:    input.StartDate.Format("2006-01-02"),
		EndDate:      input.EndDate.Format("2006-01-02"),
		Description:  input.Description,
	})

	if err != nil {
		return entities.EducationHistory{}, err
	}

	if err := uc.Repo.Update(ctx, h); err != nil {
		return entities.EducationHistory{}, err
	}
	return h, nil
}
