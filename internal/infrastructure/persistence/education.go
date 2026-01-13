package persistence

import (
	"context"
	"database/sql"
	"errors"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/domain/vos"

	"github.com/google/uuid"
)

type educationPgRepository struct {
	db *sql.DB
}

func NewEducationPgRepository(db *sql.DB) repos.EducationHistoryRepository {
	return &educationPgRepository{db: db}
}

func (r *educationPgRepository) Create(ctx context.Context, h entities.EducationHistory) error {
	query := `
		INSERT INTO education_histories (
			id, employee_id, institution, degree, field_of_study,
			start_date, end_date, description, created_at, updated_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`
	_, err := r.db.ExecContext(ctx, query,
		h.ID, h.EmployeeID, h.Institution, h.Degree.String(), // <- String()
		h.AreaEstudoID, h.StartDate, h.EndDate, h.Description, h.CreatedAt, h.UpdatedAt,
	)
	return err
}

func (r *educationPgRepository) Update(ctx context.Context, h entities.EducationHistory) error {
	query := `
		UPDATE education_histories SET
			institution = $1, degree = $2, field_of_study = $3,
			start_date = $4, end_date = $5, description = $6,
			updated_at = $7
		WHERE id = $8`
	res, err := r.db.ExecContext(ctx, query,
		h.Institution, h.Degree.String(), // <- String()
		h.AreaEstudoID, h.StartDate, h.EndDate, h.Description, h.UpdatedAt, h.ID,
	)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *educationPgRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.EducationHistory, error) {
	query := `
		SELECT id, employee_id, institution, degree, field_of_study,
		       start_date, end_date, description, created_at, updated_at
		FROM education_histories WHERE id = $1`
	var h entities.EducationHistory
	var degreeStr string

	err := r.db.QueryRowContext(ctx, query, id).Scan( // <- passa id
		&h.ID, &h.EmployeeID, &h.Institution, &degreeStr, &h.AreaEstudoID, // <- string
		&h.StartDate, &h.EndDate, &h.Description, &h.CreatedAt, &h.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.EducationHistory{}, sql.ErrNoRows
		}
		return entities.EducationHistory{}, err
	}

	deg, err := vos.NewSchoolDegree(degreeStr) // <- converte para VO
	if err != nil {
		return entities.EducationHistory{}, err
	}
	h.Degree = deg
	return h, nil
}

func (r *educationPgRepository) FindAllByEmployee(ctx context.Context, employeeID uuid.UUID, limit, offset int) ([]entities.EducationHistory, error) {
	query := `
		SELECT id, employee_id, institution, degree, field_of_study,
		       start_date, end_date, description, created_at, updated_at
		FROM education_histories
		WHERE employee_id = $1
		ORDER BY start_date DESC LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, query, employeeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []entities.EducationHistory
	for rows.Next() {
		var h entities.EducationHistory
		var degreeStr string

		if err := rows.Scan(
			&h.ID, &h.EmployeeID, &h.Institution, &degreeStr, &h.AreaEstudoID,
			&h.StartDate, &h.EndDate, &h.Description, &h.CreatedAt, &h.UpdatedAt,
		); err != nil {
			return nil, err
		}
		deg, err := vos.NewSchoolDegree(degreeStr)
		if err != nil {
			return nil, err
		}
		h.Degree = deg
		list = append(list, h)
	}
	return list, rows.Err()
}

func (r *educationPgRepository) Search(ctx context.Context, employeeID *uuid.UUID, searchText string, startDate, endDate *string, limit, offset int) ([]entities.EducationHistory, error) {
	query := `SELECT * FROM search_education_histories($1, $2, $3, $4, $5, $6)`
	var empID any
	if employeeID != nil {
		empID = *employeeID
	}
	rows, err := r.db.QueryContext(ctx, query, empID, searchText, startDate, endDate, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []entities.EducationHistory
	for rows.Next() {
		var h entities.EducationHistory
		var degreeStr string

		if err := rows.Scan(
			&h.ID, &h.EmployeeID, &h.Institution, &degreeStr, &h.AreaEstudoID,
			&h.StartDate, &h.EndDate, &h.Description, &h.CreatedAt, &h.UpdatedAt,
		); err != nil {
			return nil, err
		}
		deg, err := vos.NewSchoolDegree(degreeStr)
		if err != nil {
			return nil, err
		}
		h.Degree = deg
		results = append(results, h)
	}
	return results, nil
}

func (r *educationPgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM education_histories WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rows, _ := res.RowsAffected()
	if rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *educationPgRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM education_histories WHERE id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	return exists, err
}

var _ repos.EducationHistoryRepository = (*educationPgRepository)(nil)
