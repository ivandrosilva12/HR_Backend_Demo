package handlers

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/usecase/workhistory"
	"rhapp/internal/utils"
)

type WorkHandler struct {
	CreateUC        *workhistory.CreateWorkHistoryUseCase
	UpdateUC        *workhistory.UpdateWorkHistoryUseCase
	DeleteUC        *workhistory.DeleteWorkHistoryUseCase
	FindUC          *workhistory.FindWorkHistoryByIDUseCase
	ListUC          *workhistory.ListWorkHistoryByEmployeeUseCase
	LockConcurrency *utils.KeyedLocker
}

func NewWorkHandler(
	create *workhistory.CreateWorkHistoryUseCase,
	update *workhistory.UpdateWorkHistoryUseCase,
	deleteUC *workhistory.DeleteWorkHistoryUseCase,
	find *workhistory.FindWorkHistoryByIDUseCase,
	list *workhistory.ListWorkHistoryByEmployeeUseCase,
	lock *utils.KeyedLocker,
) *WorkHandler {
	return &WorkHandler{
		CreateUC:        create,
		UpdateUC:        update,
		DeleteUC:        deleteUC,
		FindUC:          find,
		ListUC:          list,
		LockConcurrency: lock,
	}
}

// CreateWorkHistory godoc
// @Summary Criar novo histórico profissional
// @Tags Histórico Profissional
// @Accept json
// @Produce json
// @Param data body dtos.CreateWorkDTO true "Dados do histórico profissional"
// @Success 201 {object} dtos.WorkResponseDTO
// @Failure 400,409 {object} utils.Payload
// @Failure 500 {object} utils.Payload
// @Router /work_history [post]
func (h *WorkHandler) Create(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	var dto dtos.CreateWorkDTO
	if err := utils.BindAndValidateStrict(c, &dto); err != nil {
		if fields := utils.HumanizeValidation(dto, err); len(fields) > 0 {
			c.JSON(http.StatusBadRequest, utils.Payload{
				Error:   "Validação dos campos",
				Message: "Dados inválidos. Corrija os campos destacados.",
				Fields:  fields,
			})
			return
		}
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	// Para o lock de criação, usamos um escopo conservador por funcionário
	// (evita duplicações concorrentes durante múltiplos POSTs simultâneos).
	// Se houver uma constraint mais específica (ex.: empresa + início), ajuste a chave.
	if empIDStr := dto.EmployeeID; empIDStr != "" {
		if _, err := uuid.Parse(empIDStr); err == nil && h.LockConcurrency != nil {
			lockKey := "work:create:emp:" + empIDStr
			unlock := h.LockConcurrency.Lock(lockKey)
			defer unlock()
		}
	}

	work, err := dtos.ToWorkFromCreateDTO(dto)
	if err != nil {
		// Erros de parsing/transformação são tratados como 400 se forem previsíveis,
		// mas como não temos granularidade aqui, retornamos 500.
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "internal",
			Message: "Erro interno do servidor",
		})
		return
	}

	created, err := h.CreateUC.Execute(ctx, work)
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "internal",
			Message: "Erro interno do servidor",
		})
		return
	}

	c.JSON(http.StatusCreated, dtos.ToWorkResponseDTO(created))
}

// UpdateWorkHistory godoc
// @Summary Atualizar histórico profissional
// @Tags Histórico Profissional
// @Accept json
// @Produce json
// @Param id path string true "ID do histórico"
// @Param data body dtos.UpdateWorkDTO true "Campos a atualizar"
// @Success 200 {object} dtos.WorkResponseDTO
// @Failure 400,404 {object} utils.Payload
// @Failure 500 {object} utils.Payload
// @Router /work_history/{id} [put]
func (h *WorkHandler) Update(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id do histórico deve ser um UUID válido.", Value: idStr},
			},
		})
		return
	}

	var dto dtos.UpdateWorkDTO
	if err := utils.BindAndValidateStrict(c, &dto); err != nil {
		if fields := utils.HumanizeValidation(dto, err); len(fields) > 0 {
			c.JSON(http.StatusBadRequest, utils.Payload{
				Error:   "Validação dos campos",
				Message: "Dados inválidos. Corrija os campos destacados.",
				Fields:  fields,
			})
			return
		}
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	// Lock por ID para serializar updates no mesmo recurso
	if h.LockConcurrency != nil {
		updateKey := "work:id:" + id.String()
		unlock := h.LockConcurrency.Lock(updateKey)
		defer unlock()
	}

	updated, err := h.UpdateUC.Execute(ctx, workhistory.UpdateWorkHistoryInput{ID: id, WorkDTO: dto})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "internal",
			Message: "Erro interno do servidor",
		})
		return
	}

	c.JSON(http.StatusOK, dtos.ToWorkResponseDTO(updated))
}

// DeleteWorkHistory godoc
// @Summary Remover histórico profissional
// @Tags Histórico Profissional
// @Produce json
// @Param id path string true "ID do histórico"
// @Success 204
// @Failure 400,404 {object} utils.Payload
// @Failure 500 {object} utils.Payload
// @Router /work_history/{id} [delete]
func (h *WorkHandler) Delete(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawID := c.Param("id")
	uid, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id do histórico deve ser um UUID válido.", Value: rawID},
			},
		})
		return
	}

	// Lock por ID para evitar conflito com update/delete simultâneos
	if h.LockConcurrency != nil {
		deleteKey := "work:id:" + uid.String()
		unlock := h.LockConcurrency.Lock(deleteKey)
		defer unlock()
	}

	if err := h.DeleteUC.Execute(ctx, uid); err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "internal",
			Message: "Erro interno do servidor",
		})
		return
	}
	c.Status(http.StatusNoContent)
}

// FindWorkHistoryByID godoc
// @Summary Buscar histórico profissional por ID
// @Tags Histórico Profissional
// @Produce json
// @Param id path string true "ID do histórico"
// @Success 200 {object} dtos.WorkResponseDTO
// @Failure 400,404 {object} utils.Payload
// @Failure 500 {object} utils.Payload
// @Router /work_history/{id} [get]
func (h *WorkHandler) FindByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawID := c.Param("id")
	uid, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id deve ser um UUID válido.", Value: rawID},
			},
		})
		return
	}

	work, err := h.FindUC.Execute(ctx, uid)
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "internal",
			Message: "Erro interno do servidor",
		})
		return
	}
	c.JSON(http.StatusOK, dtos.ToWorkResponseDTO(work))
}

// ListWorkHistoryByEmployee godoc
// @Summary Listar históricos profissionais de um funcionário
// @Tags Histórico Profissional
// @Produce json
// @Param employee_id path string true "ID do Funcionário"
// @Param limit query int false "Limite"
// @Param offset query int false "Offset"
// @Success 200 {array} dtos.WorkResponseDTO
// @Failure 400 {object} utils.Payload
// @Failure 500 {object} utils.Payload
// @Router /employees/{employee_id}/work_history [get]
func (h *WorkHandler) ListByEmployee(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawEmpID := c.Query("employee_id") // <-- CORRIGIDO: vem da query string
	if rawEmpID == "" {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "employee_id", Label: "employee_id", Tag: "required", Message: "employee_id é obrigatório.", Value: rawEmpID},
			},
		})
		return
	}

	employeeID, err := uuid.Parse(rawEmpID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "employee_id", Label: "employee_id", Tag: "uuid", Message: "employee_id deve ser um UUID válido.", Value: rawEmpID},
			},
		})
		return
	}

	pagination := utils.PaginationInput(c)

	list, err := h.ListUC.Execute(ctx, employeeID, pagination.Limit, pagination.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos dados dos históricos profissionais.",
			Fields:  []utils.FieldError{},
		})
		return
	}
	c.JSON(http.StatusOK, dtos.ToWorkResponseDTOList(list))
}
