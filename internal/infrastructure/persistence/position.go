package persistence

import (
	"context"
	"database/sql"
	"errors"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/repos"
	"rhapp/internal/utils"

	"github.com/google/uuid"
)

type positionPgRepository struct {
	db *sql.DB
}

func NewPositionPgRepository(db *sql.DB) *positionPgRepository {
	return &positionPgRepository{db: db}
}

func (r *positionPgRepository) Create(ctx context.Context, pos entities.Position) error {
	const query = `
		INSERT INTO positions (id, nome, department_id, max_headcount, tipo, created_at, updated_at) 
		VALUES ($1, $2, $3, $4, $5, $6, $7)` // UPDATED (+tipo)
	_, err := r.db.ExecContext(ctx, query,
		pos.ID, pos.Nome, pos.DepartmentID, pos.MaxHeadcount, pos.Tipo, pos.CreatedAt, pos.UpdatedAt) // UPDATED
	return err
}

func (r *positionPgRepository) Update(ctx context.Context, pos entities.Position) error {
	const query = `
		UPDATE positions 
		   SET nome = $1, department_id = $2, max_headcount = $3, tipo = $4, updated_at = $5
		 WHERE id = $6` // UPDATED (+tipo)
	res, err := r.db.ExecContext(ctx, query,
		pos.Nome, pos.DepartmentID, pos.MaxHeadcount, pos.Tipo, pos.UpdatedAt, pos.ID) // UPDATED
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *positionPgRepository) Delete(ctx context.Context, id uuid.UUID) error {
	const query = `DELETE FROM positions WHERE id = $1`
	res, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *positionPgRepository) FindByID(ctx context.Context, id uuid.UUID) (entities.Position, error) {
	const query = `
  		SELECT p.id, p.nome, p.department_id, p.max_headcount, p.tipo,       -- UPDATED (+p.tipo)
		       COALESCE(cnt.curr, 0) AS current_headcount,
		       GREATEST(p.max_headcount - COALESCE(cnt.curr, 0), 0) AS remaining,
		       p.created_at, p.updated_at
  		FROM positions p
  		LEFT JOIN (
    		 SELECT e.position_id, COUNT(*) FILTER (WHERE e.is_active) AS curr
    		   FROM employees e
    		  GROUP BY e.position_id
  		) AS cnt ON cnt.position_id = p.id
  		WHERE p.id = $1`
	var pos entities.Position
	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&pos.ID, &pos.Nome, &pos.DepartmentID, &pos.MaxHeadcount, &pos.Tipo, // UPDATED (+Tipo)
		&pos.CurrentHeadcount, &pos.Remaining, &pos.CreatedAt, &pos.UpdatedAt)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return entities.Position{}, err
		}
		return entities.Position{}, err
	}
	return pos, nil
}

func (r *positionPgRepository) FindAll(ctx context.Context, limit, offset int) (utils.PagedResponse[entities.Position], error) {
	const countQ = `SELECT COUNT(*) FROM positions`

	const pageQ = `
		SELECT p.id, p.nome, p.department_id, p.max_headcount, p.tipo,       -- UPDATED (+p.tipo)
		       COALESCE(cnt.curr, 0) AS current_headcount,
		       GREATEST(p.max_headcount - COALESCE(cnt.curr, 0), 0) AS remaining,
		       p.created_at, p.updated_at
  		FROM positions p
  		LEFT JOIN (
    		 SELECT e.position_id, COUNT(*) FILTER (WHERE e.is_active) AS curr
    		   FROM employees e
    		  GROUP BY e.position_id
  		) AS cnt ON cnt.position_id = p.id
  		ORDER BY p.created_at DESC
  		LIMIT $1 OFFSET $2`

	var total int
	if err := r.db.QueryRowContext(ctx, countQ).Scan(&total); err != nil {
		return utils.PagedResponse[entities.Position]{Items: []entities.Position{}, Total: 0, Limit: limit, Offset: offset}, err
	}

	rows, err := r.db.QueryContext(ctx, pageQ, limit, offset)
	if err != nil {
		return utils.PagedResponse[entities.Position]{Items: []entities.Position{}, Total: total, Limit: limit, Offset: offset}, err
	}
	defer rows.Close()

	items := make([]entities.Position, 0, limit)
	for rows.Next() {
		var pos entities.Position
		if err := rows.Scan(
			&pos.ID, &pos.Nome, &pos.DepartmentID, &pos.MaxHeadcount, &pos.Tipo, // UPDATED
			&pos.CurrentHeadcount, &pos.Remaining, &pos.CreatedAt, &pos.UpdatedAt,
		); err != nil {
			return utils.PagedResponse[entities.Position]{Items: []entities.Position{}, Total: total, Limit: limit, Offset: offset}, err
		}
		items = append(items, pos)
	}
	if err := rows.Err(); err != nil {
		return utils.PagedResponse[entities.Position]{Items: []entities.Position{}, Total: total, Limit: limit, Offset: offset}, err
	}

	return utils.PagedResponse[entities.Position]{Items: items, Total: total, Limit: limit, Offset: offset}, nil
}

func (r *positionPgRepository) Search(ctx context.Context, searchText, departmentFilter string, limit, offset int) (utils.PagedResponse[dtos.PositionResultDTO], error) {
	namePattern := "%" + searchText + "%"
	deptPattern := "%" + departmentFilter + "%"

	const countQ = `
  		SELECT COUNT(*)
    	  FROM positions p
    	  JOIN departments d ON d.id = p.department_id
   	 WHERE ($1 = '%%' OR p.nome ILIKE $1)
       AND ($2 = '%%' OR d.nome ILIKE $2)
	`
	var total int
	if err := r.db.QueryRowContext(ctx, countQ, namePattern, deptPattern).Scan(&total); err != nil {
		return utils.PagedResponse[dtos.PositionResultDTO]{Items: []dtos.PositionResultDTO{}, Total: 0, Limit: limit, Offset: offset}, err
	}

	const pageQ = `
  	WITH filtered AS (
    	SELECT
      		p.id,
      		p.nome,
      		p.department_id,
      		p.max_headcount,
      		p.tipo,                    -- NEW
      		d.nome AS department_nome,
      		COUNT(e.id) FILTER (WHERE e.is_active) AS current_headcount,
      		GREATEST(p.max_headcount - COUNT(e.id) FILTER (WHERE e.is_active), 0) AS remaining,
      		p.created_at,
      		p.updated_at
    	FROM positions p
    	JOIN departments d ON d.id = p.department_id
    	LEFT JOIN employees e ON e.position_id = p.id
    	WHERE ($1 = '%%' OR p.nome ILIKE $1)
      	  AND ($2 = '%%' OR d.nome ILIKE $2)
    	GROUP BY p.id, d.nome
  	)
  	SELECT id, nome, department_id, max_headcount, tipo, department_nome,  -- UPDATED (+tipo)
           current_headcount, remaining, created_at, updated_at
  	  FROM filtered
  	ORDER BY created_at DESC
  	LIMIT $3 OFFSET $4
	`
	rows, err := r.db.QueryContext(ctx, pageQ, namePattern, deptPattern, limit, offset)
	if err != nil {
		return utils.PagedResponse[dtos.PositionResultDTO]{Items: []dtos.PositionResultDTO{}, Total: total, Limit: limit, Offset: offset}, err
	}
	defer rows.Close()

	items := make([]dtos.PositionResultDTO, 0, limit)
	for rows.Next() {
		var dto dtos.PositionResultDTO
		if err := rows.Scan(
			&dto.ID, &dto.Nome, &dto.DepartmentID, &dto.MaxHeadcount, &dto.Tipo, &dto.DepartmentNome, // UPDATED (+Tipo)
			&dto.CurrentHeadcount, &dto.Remaining, &dto.CreatedAt, &dto.UpdatedAt,
		); err != nil {
			return utils.PagedResponse[dtos.PositionResultDTO]{Items: []dtos.PositionResultDTO{}, Total: total, Limit: limit, Offset: offset}, err
		}
		items = append(items, dto)
	}
	if err := rows.Err(); err != nil {
		return utils.PagedResponse[dtos.PositionResultDTO]{Items: []dtos.PositionResultDTO{}, Total: total, Limit: limit, Offset: offset}, err
	}

	return utils.PagedResponse[dtos.PositionResultDTO]{Items: items, Total: total, Limit: limit, Offset: offset}, nil
}

func (r *positionPgRepository) ExistsByID(ctx context.Context, id uuid.UUID) (bool, error) {
	const query = `SELECT EXISTS (SELECT 1 FROM positions WHERE id = $1)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, id).Scan(&exists)
	return exists, err
}

func (r *positionPgRepository) ExistsByNomeAndDepartment(ctx context.Context, nome string, departmentID uuid.UUID) (bool, error) {
	const query = `
		SELECT EXISTS (
			SELECT 1 FROM positions 
			 WHERE LOWER(nome) = LOWER($1) AND department_id = $2
		)`
	var exists bool
	err := r.db.QueryRowContext(ctx, query, nome, departmentID).Scan(&exists)
	return exists, err
}

var _ repos.PositionRepository = (*positionPgRepository)(nil)
