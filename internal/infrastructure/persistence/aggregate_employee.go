package persistence

import (
	"context"
	"database/sql"

	"rhapp/internal/domain/agregados"
	"rhapp/internal/domain/entities"

	"github.com/google/uuid"
)

type employeeAggregatePgRepository struct {
	db *sql.DB
}

func NewEmployeeAggregatePgRepository(db *sql.DB) *employeeAggregatePgRepository {
	return &employeeAggregatePgRepository{db: db}
}

func (r *employeeAggregatePgRepository) GetFullByID(ctx context.Context, id uuid.UUID) (*agregados.EmployeeAggregate, error) {
	var agg agregados.EmployeeAggregate

	employeeRepo := NewEmployeePgRepository(r.db)
	emp, err := employeeRepo.FindByID(ctx, id)
	if err != nil {
		return nil, err
	}
	agg.Employee = emp

	// Dependents
	dependentsRows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, full_name, relationship, gender, date_of_birth, document_id, is_active, created_at, updated_at 
		FROM dependents 
		WHERE employee_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer dependentsRows.Close()

	for dependentsRows.Next() {
		var d entities.Dependent
		if scanErr := dependentsRows.Scan(&d.ID, &d.EmployeeID, &d.FullName, &d.Relationship, &d.Gender, &d.DateOfBirth, &d.IsActive, &d.CreatedAt, &d.UpdatedAt); scanErr != nil {
			return nil, scanErr
		}
		agg.Dependents = append(agg.Dependents, d)
	}
	if err := dependentsRows.Err(); err != nil {
		return nil, err
	}

	// Education Histories
	eduRows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, institution, degree, field_of_study, start_date, end_date, description 
		FROM education_histories 
		WHERE employee_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer eduRows.Close()

	for eduRows.Next() {
		var e entities.EducationHistory
		if scanErr := eduRows.Scan(&e.ID, &e.EmployeeID, &e.Institution, &e.Degree, &e.AreaEstudoID, &e.StartDate, &e.EndDate, &e.Description); scanErr != nil {
			return nil, scanErr
		}
		agg.EducationHistories = append(agg.EducationHistories, e)
	}
	if err := eduRows.Err(); err != nil {
		return nil, err
	}

	// Work Histories
	workRows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, company, position, start_date, end_date, responsibilities 
		FROM work_histories 
		WHERE employee_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer workRows.Close()

	for workRows.Next() {
		var w entities.WorkHistory
		if scanErr := workRows.Scan(&w.ID, &w.EmployeeID, &w.Company, &w.Position, &w.StartDate, &w.EndDate, &w.Responsibilities); scanErr != nil {
			return nil, scanErr
		}
		agg.WorkHistories = append(agg.WorkHistories, w)
	}
	if err := workRows.Err(); err != nil {
		return nil, err
	}

	// Employee Statuses
	statusRows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, status, start_date, end_date, is_current, created_at, updated_at 
		FROM employee_statuses 
		WHERE employee_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer statusRows.Close()

	for statusRows.Next() {
		var s entities.EmployeeStatus
		if scanErr := statusRows.Scan(&s.ID, &s.EmployeeID, &s.Status, &s.StartDate, &s.EndDate, &s.IsCurrent, &s.CreatedAt, &s.UpdatedAt); scanErr != nil {
			return nil, scanErr
		}
		agg.Statuses = append(agg.Statuses, s)
	}
	if err := statusRows.Err(); err != nil {
		return nil, err
	}

	// Supervisor History
	supRows, err := r.db.QueryContext(ctx, `
		SELECT id, employee_id, supervisor_id, start_date, end_date, created_at, updated_at 
		FROM supervisor_histories 
		WHERE employee_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer supRows.Close()

	for supRows.Next() {
		var sh entities.WorkerHistory
		if scanErr := supRows.Scan(&sh.ID, &sh.EmployeeID, &sh.StartDate, &sh.EndDate, &sh.CreatedAt, &sh.UpdatedAt); scanErr != nil {
			return nil, scanErr
		}
		agg.WorkerHistory = append(agg.WorkerHistory, sh)
	}
	if err := supRows.Err(); err != nil {
		return nil, err
	}

	// Documents
	docRows, err := r.db.QueryContext(ctx, `
		SELECT id, owner_type, owner_id, type, file_name, file_url, extension, is_active, uploaded_at 
		FROM documents 
		WHERE owner_type = 'employee' AND owner_id = $1`, id)
	if err != nil {
		return nil, err
	}
	defer docRows.Close()

	for docRows.Next() {
		var d entities.Document
		if scanErr := docRows.Scan(&d.ID, &d.OwnerType, &d.OwnerID, &d.Type, &d.FileName, &d.FileURL, &d.Extension, &d.IsActive, &d.UploadedAt); scanErr != nil {
			return nil, scanErr
		}
		agg.Documents = append(agg.Documents, d)
	}
	if err := docRows.Err(); err != nil {
		return nil, err
	}

	return &agg, nil
}

var _ agregados.EmployeeAggregateRepository = (*employeeAggregatePgRepository)(nil)
