package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/usecase/employee_status"
	"rhapp/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EmployeeStatusHandler struct {
	CreateUC         *employee_status.CreateEmployeeStatusUseCase
	UpdateUC         *employee_status.UpdateEmployeeStatusUseCase
	DeleteUC         *employee_status.DeleteEmployeeStatusUseCase
	FindByIDUC       *employee_status.FindEmployeeStatusByIDUseCase
	ListByEmployeeUC *employee_status.ListEmployeeStatusByEmployeeUseCase
	LockConcurrency  *utils.KeyedLocker
}

func NewEmployeeStatusHandler(
	create *employee_status.CreateEmployeeStatusUseCase,
	update *employee_status.UpdateEmployeeStatusUseCase,
	deleteUC *employee_status.DeleteEmployeeStatusUseCase,
	find *employee_status.FindEmployeeStatusByIDUseCase,
	list *employee_status.ListEmployeeStatusByEmployeeUseCase,
	lock *utils.KeyedLocker,
) *EmployeeStatusHandler {
	return &EmployeeStatusHandler{
		CreateUC:         create,
		UpdateUC:         update,
		DeleteUC:         deleteUC,
		FindByIDUC:       find,
		ListByEmployeeUC: list,
		LockConcurrency:  lock,
	}
}

// @Summary Criar status de funcionário
// @Tags EmployeeStatus
// @Accept json
// @Produce json
// @Param input body dtos.CreateEmployeeStatusDTO true "Dados do status"
// @Success 201 {object} dtos.EmployeeStatusResponseDTO
// @Failure 400,409 {object}  utils.Payload
// @Router /employee-status [post]
func (h *EmployeeStatusHandler) Create(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	var dto dtos.CreateEmployeeStatusDTO
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

	empID, err := uuid.Parse(dto.EmployeeID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id do funcionário deve ser um UUID válido.", Value: dto.EmployeeID},
			},
		})
		return
	}

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, dto.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields:  []utils.FieldError{{Field: "start_date", Label: "start_date", Tag: "date", Message: "Data inicial inválida. Use o formato YYYY-MM-DD.", Value: dto.StartDate}},
		})
		return
	}

	var endDate *time.Time
	if dto.EndDate != "" {
		e, err := time.Parse(layout, dto.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.Payload{
				Error:   "Validação dos campos",
				Message: "Dados inválidos. Corrija os campos destacados.",
				Fields:  []utils.FieldError{{Field: "end_date", Label: "end_date", Tag: "date", Message: "Data final inválida. Use o formato YYYY-MM-DD.", Value: dto.EndDate}},
			})
			return
		}
		endDate = &e
	}

	// Lock por funcionário + data de início para evitar criações concorrentes duplicadas
	lockKey := "empstatus:create:" + empID.String() + ":" + startDate.Format("2006-01-02")
	unlock := h.LockConcurrency.Lock(lockKey)
	defer unlock()

	es, err := h.CreateUC.Execute(ctx, employee_status.CreateEmployeeStatusInput{
		EmployeeID: empID,
		Status:     dto.Status,
		Reason:     dto.Reason,
		StartDate:  startDate,
		EndDate:    endDate,
		IsCurrent:  true,
	})
	if err != nil {
		fmt.Println("ERRO - ", err)
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusCreated, dtos.ToEmployeeStatusResponseDTO(es))
}

// @Summary Atualizar status
// @Tags EmployeeStatus
// @Accept json
// @Produce json
// @Param id path string true "ID do status"
// @Param input body dtos.UpdateEmployeeStatusDTO true "Atualização"
// @Success 200 {object} dtos.EmployeeStatusResponseDTO
// @Failure 400,404 {object}  utils.Payload
// @Router /employee-status/{id} [put]
func (h *EmployeeStatusHandler) Update(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawID := c.Param("id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields:  []utils.FieldError{{Field: "ID", Label: "id", Tag: "uuid", Message: "id deve ser um UUID válido.", Value: rawID}},
		})
		return
	}

	var dto dtos.UpdateEmployeeStatusDTO
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

	layout := "2006-01-02"
	startDate, err := time.Parse(layout, dto.StartDate)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields:  []utils.FieldError{{Field: "start_date", Label: "start_date", Tag: "date", Message: "Data inicial inválida. Use o formato YYYY-MM-DD.", Value: dto.StartDate}},
		})
		return
	}

	var endDate *time.Time
	if dto.EndDate != "" {
		e, err := time.Parse(layout, dto.EndDate)
		if err != nil {
			c.JSON(http.StatusBadRequest, utils.Payload{
				Error:   "Validação dos campos",
				Message: "Dados inválidos. Corrija os campos destacados.",
				Fields:  []utils.FieldError{{Field: "end_date", Label: "end_date", Tag: "date", Message: "Data final inválida. Use o formato YYYY-MM-DD.", Value: dto.EndDate}},
			})
			return
		}
		endDate = &e
	}

	// Lock por ID para serializar alterações
	updateKey := "empstatus:id:" + id.String()
	unlock := h.LockConcurrency.Lock(updateKey)
	defer unlock()

	result, err := h.UpdateUC.Execute(ctx, employee_status.UpdateEmployeeStatusInput{
		ID:        id,
		Status:    dto.Status,
		Reason:    dto.Reason,
		StartDate: startDate,
		EndDate:   endDate,
	})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, dtos.ToEmployeeStatusResponseDTO(result))
}

// @Summary Remover status
// @Tags EmployeeStatus
// @Produce json
// @Param id path string true "ID do status"
// @Success 204
// @Failure 400,404 {object} utils.Payload
// @Router /employee-status/{id} [delete]
func (h *EmployeeStatusHandler) Delete(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawID := c.Param("id")
	uid, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields:  []utils.FieldError{{Field: "ID", Label: "id", Tag: "uuid", Message: "id deve ser um UUID válido.", Value: rawID}},
		})
		return
	}

	// Lock por ID para evitar conflitos com update/delete simultâneos
	deleteKey := "empstatus:id:" + uid.String()
	unlock := h.LockConcurrency.Lock(deleteKey)
	defer unlock()

	if err := h.DeleteUC.Execute(ctx, uid); err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}
	c.Status(http.StatusNoContent)
}

// @Summary Buscar status por ID
// @Tags EmployeeStatus
// @Produce json
// @Param id path string true "ID do status"
// @Success 200 {object} dtos.EmployeeStatusResponseDTO
// @Failure 400,404 {object} utils.Payload
// @Router /employee-status/{id} [get]
func (h *EmployeeStatusHandler) FindByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawID := c.Param("id")
	uid, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields:  []utils.FieldError{{Field: "ID", Label: "id", Tag: "uuid", Message: "id deve ser um UUID válido.", Value: rawID}},
		})
		return
	}

	es, err := h.FindByIDUC.Execute(ctx, uid)
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, dtos.ToEmployeeStatusResponseDTO(es))
}

// @Summary Listar status de funcionário
// @Tags EmployeeStatus
// @Param employee_id query string true "ID do funcionário"
// @Param limit query int false "Limite de resultados"
// @Param offset query int false "Offset de paginação"
// @Success 200 {array} dtos.EmployeeStatusResponseDTO
// @Failure 400 {object} utils.Payload
// @Router /employee-status [get]
func (h *EmployeeStatusHandler) ListByEmployee(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawEmpID := c.Query("employee_id")
	empID, err := uuid.Parse(rawEmpID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields:  []utils.FieldError{{Field: "ID", Label: "id", Tag: "uuid", Message: "id deve ser um UUID válido.", Value: rawEmpID}},
		})
		return
	}

	pagination := utils.PaginationInput(c)

	list, err := h.ListByEmployeeUC.Execute(ctx, empID, pagination.Limit, pagination.Offset)
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos dados dos status.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	c.JSON(http.StatusOK, dtos.ToEmployeeStatusResponseDTOList(list))
}
