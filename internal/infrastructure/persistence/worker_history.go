// internal/infra/persistence/worker_history_pg_repository.go
package persistence

import (
	"context"
	"database/sql"
	"time"

	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"

	"github.com/google/uuid"
)

type workerHistoryPgRepository struct {
	db *sql.DB
}

func NewWorkerHistoryPgRepository(db *sql.DB) *workerHistoryPgRepository {
	return &workerHistoryPgRepository{db: db}
}

func (r *workerHistoryPgRepository) Create(ctx context.Context, w entities.WorkerHistory) error {
	const q = `
		INSERT INTO worker_histories
		  (id, employee_id, position_id, start_date, end_date, status, created_at, updated_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8)`
	_, err := r.db.ExecContext(ctx, q,
		w.ID, w.EmployeeID, w.PositionID, w.StartDate, w.EndDate, string(w.Status), w.CreatedAt, w.UpdatedAt)
	return err
}

func (r *workerHistoryPgRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.WorkerHistory, error) {
	const q = `
		SELECT id, employee_id, position_id, start_date, end_date, status, created_at, updated_at
		  FROM worker_histories WHERE id = $1`
	var w entities.WorkerHistory
	var end sql.NullTime
	var status string
	err := r.db.QueryRowContext(ctx, q, id).Scan(
		&w.ID, &w.EmployeeID, &w.PositionID, &w.StartDate, &end, &status, &w.CreatedAt, &w.UpdatedAt,
	)
	if err != nil {
		return entities.WorkerHistory{}, err
	}
	if end.Valid {
		w.EndDate = &end.Time
	}
	w.Status = entities.WorkerStatus(status)
	return w, nil
}

func (r *workerHistoryPgRepository) Update(ctx context.Context, w entities.WorkerHistory) error {
	const q = `
		UPDATE worker_histories
		   SET employee_id=$2, position_id=$3, start_date=$4, end_date=$5, status=$6, updated_at=$7
		 WHERE id=$1`
	_, err := r.db.ExecContext(ctx, q,
		w.ID, w.EmployeeID, w.PositionID, w.StartDate, w.EndDate, string(w.Status), time.Now())
	return err
}

func (r *workerHistoryPgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM worker_histories WHERE id=$1`
	res, err := r.db.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *workerHistoryPgRepository) ListByEmployeeID(ctx context.Context, employeeID uuid.UUID, limit, offset int) ([]entities.WorkerHistory, error) {
	const q = `
		SELECT id, employee_id, position_id, start_date, end_date, status, created_at, updated_at
		  FROM worker_histories
		 WHERE employee_id = $1
		 ORDER BY start_date DESC
		 LIMIT $2 OFFSET $3`
	rows, err := r.db.QueryContext(ctx, q, employeeID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]entities.WorkerHistory, 0, limit)
	for rows.Next() {
		var w entities.WorkerHistory
		var end sql.NullTime
		var status string
		if err := rows.Scan(&w.ID, &w.EmployeeID, &w.PositionID, &w.StartDate, &end, &status, &w.CreatedAt, &w.UpdatedAt); err != nil {
			return nil, err
		}
		if end.Valid {
			w.EndDate = &end.Time
		}
		w.Status = entities.WorkerStatus(status)
		list = append(list, w)
	}
	return list, rows.Err()
}

var _ repos.WorkerHistoryRepository = (*workerHistoryPgRepository)(nil)
