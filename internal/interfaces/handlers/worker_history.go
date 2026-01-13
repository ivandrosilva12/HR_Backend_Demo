// internal/interfaces/http/handlers/worker_history_handler.go
package handlers

import (
	"context"
	"net/http"
	"runtime"
	"sync"
	"time"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	workerhistory "rhapp/internal/usecase/workerhistory"
	"rhapp/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type WorkerHistoryHandler struct {
	CreateUC        *workerhistory.CreateUseCase
	UpdateUC        *workerhistory.UpdateUseCase
	DeleteUC        *workerhistory.DeleteUseCase
	FindUC          *workerhistory.FindByIDUseCase
	ListUC          *workerhistory.ListByEmployeeIDUseCase
	LockConcurrency *utils.KeyedLocker
}

func NewWorkerHistoryHandler(
	create *workerhistory.CreateUseCase,
	update *workerhistory.UpdateUseCase,
	delete *workerhistory.DeleteUseCase,
	find *workerhistory.FindByIDUseCase,
	list *workerhistory.ListByEmployeeIDUseCase,
	lock *utils.KeyedLocker,
) *WorkerHistoryHandler {
	return &WorkerHistoryHandler{
		CreateUC:        create,
		UpdateUC:        update,
		DeleteUC:        delete,
		FindUC:          find,
		ListUC:          list,
		LockConcurrency: lock,
	}
}

// POST /worker-histories
func (h *WorkerHistoryHandler) Create(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	var dtoIn dtos.CreateWorkerHistoryDTO
	if err := utils.BindAndValidateStrict(c, &dtoIn); err != nil {
		if fields := utils.HumanizeValidation(dtoIn, err); len(fields) > 0 {
			c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: "Dados inválidos.", Fields: fields})
			return
		}
		c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: "Dados inválidos.", Fields: []utils.FieldError{}})
		return
	}

	empID, err := uuid.Parse(dtoIn.EmployeeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: "employee_id inválido.", Fields: []utils.FieldError{
			{Field: "employee_id", Label: "employee_id", Tag: "uuid", Message: "UUID inválido", Value: dtoIn.EmployeeID},
		}})
		return
	}
	posID, err := uuid.Parse(dtoIn.PositionID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: "position_id inválido.", Fields: []utils.FieldError{
			{Field: "position_id", Label: "position_id", Tag: "uuid", Message: "UUID inválido", Value: dtoIn.PositionID},
		}})
		return
	}

	const layout = "2006-01-02"
	start, err := time.Parse(layout, dtoIn.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: "Data inicial inválida.", Fields: []utils.FieldError{
			{Field: "start_date", Label: "start_date", Tag: "date", Message: "Use YYYY-MM-DD", Value: dtoIn.StartDate},
		}})
		return
	}
	var end *time.Time
	if dtoIn.EndDate != "" {
		e, err := time.Parse(layout, dtoIn.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: "Data final inválida.", Fields: []utils.FieldError{
				{Field: "end_date", Label: "end_date", Tag: "date", Message: "Use YYYY-MM-DD", Value: dtoIn.EndDate},
			}})
			return
		}
		end = &e
	}

	status := entities.WorkerStatus(dtoIn.Status)
	// default status
	if status == "" {
		status = entities.WorkerAtivo
	}

	lockKey := "workerhistory:create:emp:" + empID.String() + ":start:" + start.Format("2006-01-02")
	unlock := h.LockConcurrency.Lock(lockKey)
	defer unlock()

	res, err := h.CreateUC.Execute(ctx, workerhistory.CreateInput{
		EmployeeID: empID,
		PositionID: posID,
		StartDate:  start,
		EndDate:    end,
		Status:     status,
	})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusCreated, dtos.ToWorkerHistoryResponseDTO(res))
}

// PUT /worker-histories/:id
func (h *WorkerHistoryHandler) Update(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawID := c.Param("id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: "id inválido.", Fields: []utils.FieldError{
			{Field: "id", Label: "id", Tag: "uuid", Message: "UUID inválido", Value: rawID},
		}})
		return
	}

	var dtoIn dtos.UpdateWorkerHistoryDTO
	if err := utils.BindAndValidateStrict(c, &dtoIn); err != nil {
		if fields := utils.HumanizeValidation(dtoIn, err); len(fields) > 0 {
			c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: "Dados inválidos.", Fields: fields})
			return
		}
		c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: "Dados inválidos.", Fields: []utils.FieldError{}})
		return
	}

	var posIDPtr *uuid.UUID
	if dtoIn.PositionID != nil && *dtoIn.PositionID != "" {
		pid, err := uuid.Parse(*dtoIn.PositionID)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: "position_id inválido.", Fields: []utils.FieldError{
				{Field: "position_id", Label: "position_id", Tag: "uuid", Message: "UUID inválido", Value: *dtoIn.PositionID},
			}})
			return
		}
		posIDPtr = &pid
	}

	const layout = "2006-01-02"
	var startPtr *time.Time
	if dtoIn.StartDate != "" {
		s, err := time.Parse(layout, dtoIn.StartDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: "Data inicial inválida.", Fields: []utils.FieldError{
				{Field: "start_date", Label: "start_date", Tag: "date", Message: "Use YYYY-MM-DD", Value: dtoIn.StartDate},
			}})
			return
		}
		startPtr = &s
	}

	var endPtr *time.Time
	if dtoIn.EndDate != "" {
		e, err := time.Parse(layout, dtoIn.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: "Data final inválida.", Fields: []utils.FieldError{
				{Field: "end_date", Label: "end_date", Tag: "date", Message: "Use YYYY-MM-DD", Value: dtoIn.EndDate},
			}})
			return
		}
		endPtr = &e
	}

	var statusPtr *entities.WorkerStatus
	if dtoIn.Status != "" {
		st := entities.WorkerStatus(dtoIn.Status)
		statusPtr = &st
	}

	updateKey := "workerhistory:id:" + id.String()
	unlock := h.LockConcurrency.Lock(updateKey)
	defer unlock()

	// EmployeeID não vem no DTO de update; é preservado do registro atual no UC
	res, err := h.UpdateUC.Execute(ctx, workerhistory.UpdateInput{
		ID:         id,
		PositionID: posIDPtr,
		StartDate:  startPtr,
		EndDate:    endPtr,
		Status:     statusPtr,
	})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, dtos.ToWorkerHistoryResponseDTO(res))
}

// DELETE /worker-histories/:id
func (h *WorkerHistoryHandler) Delete(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawID := c.Param("id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: "id inválido.", Fields: []utils.FieldError{
			{Field: "id", Label: "id", Tag: "uuid", Message: "UUID inválido", Value: rawID},
		}})
		return
	}

	deleteKey := "workerhistory:id:" + id.String()
	unlock := h.LockConcurrency.Lock(deleteKey)
	defer unlock()

	if err := h.DeleteUC.Execute(ctx, id); err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}
	c.Status(http.StatusNoContent)
}

// GET /worker-histories/:id
func (h *WorkerHistoryHandler) FindByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawID := c.Param("id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: "id inválido.", Fields: []utils.FieldError{
			{Field: "id", Label: "id", Tag: "uuid", Message: "UUID inválido", Value: rawID},
		}})
		return
	}

	res, err := h.FindUC.Execute(ctx, id)
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}
	c.JSON(http.StatusOK, dtos.ToWorkerHistoryResponseDTO(res))
}

// GET /worker-histories?employee_id=...&limit=...&offset=...
func (h *WorkerHistoryHandler) ListByEmployee(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawEmp := c.Query("employee_id")
	empID, err := uuid.Parse(rawEmp)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{Error: "Validação dos campos", Message: "employee_id inválido.", Fields: []utils.FieldError{
			{Field: "employee_id", Label: "employee_id", Tag: "uuid", Message: "UUID inválido", Value: rawEmp},
		}})
		return
	}

	p := utils.PaginationInput(c)
	list, err := h.ListUC.Execute(ctx, empID, p.Limit, p.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "Erros na DB", Message: "Falha no carregamento dos dados.", Fields: []utils.FieldError{}})
		return
	}

	c.JSON(http.StatusOK, toWorkerDTOsConcurrent(list))
}

// helper paralelo
func toWorkerDTOsConcurrent(items []entities.WorkerHistory) []dtos.WorkerHistoryResponseDTO {
	n := len(items)
	if n == 0 {
		return []dtos.WorkerHistoryResponseDTO{}
	}
	out := make([]dtos.WorkerHistoryResponseDTO, n)
	workers := runtime.GOMAXPROCS(0)
	if workers < 2 {
		workers = 2
	}
	if workers > n {
		workers = n
	}
	jobs := make(chan int, n)
	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				out[i] = dtos.ToWorkerHistoryResponseDTO(items[i])
			}
		}()
	}
	for i := 0; i < n; i++ {
		jobs <- i
	}
	close(jobs)
	wg.Wait()
	return out
}
