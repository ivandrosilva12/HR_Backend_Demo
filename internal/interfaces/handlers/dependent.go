package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/usecase/dependents"
	"rhapp/internal/utils"
)

type DependentHandler struct {
	CreateUC        *dependents.CreateDependentUseCase
	UpdateUC        *dependents.UpdateDependentUseCase
	DeleteUC        *dependents.DeleteDependentUseCase
	FindUC          *dependents.FindDependentByIDUseCase
	ListUC          *dependents.ListDependentsUseCase
	LockConcurrency *utils.KeyedLocker
}

func NewDependentHandler(
	createUC *dependents.CreateDependentUseCase,
	updateUC *dependents.UpdateDependentUseCase,
	deleteUC *dependents.DeleteDependentUseCase,
	getUC *dependents.FindDependentByIDUseCase,
	listUC *dependents.ListDependentsUseCase,
	lock *utils.KeyedLocker,
) *DependentHandler {
	return &DependentHandler{
		CreateUC:        createUC,
		UpdateUC:        updateUC,
		DeleteUC:        deleteUC,
		FindUC:          getUC,
		ListUC:          listUC,
		LockConcurrency: lock,
	}
}

// @Summary Criar dependente
// @Tags Dependentes
// @Accept json
// @Produce json
// @Param input body dtos.CreateDependentDTO true "Dados do dependente"
// @Success 201 {object} dtos.DependentResponseDTO
// @Failure 400,409 {object} utils.Payload
// @Router /dependents [post]
func (h *DependentHandler) Create(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	var dto dtos.CreateDependentDTO
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

	// Lock para evitar criação duplicada concorrente para o mesmo funcionário + nome + data
	lockKey := "dep:create:" + empID.String() + ":" + utils.Normalize(dto.FullName) + ":" + dto.DateOfBirth
	unlock := h.LockConcurrency.Lock(lockKey)
	defer unlock()

	result, err := h.CreateUC.Execute(ctx, dependents.CreateDependentInput{
		EmployeeID:   empID,
		FullName:     dto.FullName,
		Relationship: dto.Relationship,
		Gender:       dto.Gender,
		DateOfBirth:  dto.DateOfBirth,
	})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusCreated, dtos.ToDependentResponseDTO(result))
}

// @Summary Atualizar dependente
// @Tags Dependentes
// @Accept json
// @Produce json
// @Param id path string true "ID do dependente"
// @Param input body dtos.UpdateDependentDTO true "Dados do dependente"
// @Success 200 {object} dtos.DependentResponseDTO
// @Failure 400,404 {object} utils.Payload
// @Router /dependents/{id} [put]
func (h *DependentHandler) Update(c *gin.Context) {
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

	var dto dtos.UpdateDependentDTO
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

	// Lock por ID para serializar updates no mesmo dependente
	updateKey := "dep:id:" + id.String()
	unlock := h.LockConcurrency.Lock(updateKey)
	defer unlock()

	updated, err := h.UpdateUC.Execute(ctx, dependents.UpdateDependentInput{
		ID:           id,
		FullName:     dto.FullName,
		Relationship: dto.Relationship,
		Gender:       dto.Gender,
		DateOfBirth:  dto.DateOfBirth,
		IsActive:     dto.IsActive,
	})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, dtos.ToDependentResponseDTO(updated))
}

// @Summary Deletar dependente
// @Tags Dependentes
// @Produce json
// @Param id path string true "ID do dependente"
// @Success 204
// @Failure 400,404 {object} utils.Payload
// @Router /dependents/{id} [delete]
func (h *DependentHandler) Delete(c *gin.Context) {
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

	// Lock por ID para evitar conflito com update/delete simultâneos
	deleteKey := "dep:id:" + uid.String()
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

// @Summary Buscar dependente por ID
// @Tags Dependentes
// @Produce json
// @Param id path string true "ID do dependente"
// @Success 200 {object} dtos.DependentResponseDTO
// @Failure 400,404 {object} utils.Payload
// @Router /dependents/{id} [get]
func (h *DependentHandler) FindByID(c *gin.Context) {
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

	d, err := h.FindUC.Execute(ctx, uid)
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}
	c.JSON(http.StatusOK, dtos.ToDependentResponseDTO(d))
}

// @Summary Listar dependentes por funcionário
// @Tags Dependentes
// @Produce json
// @Param employee_id query string true "ID do funcionário"
// @Param limit query int false "Limite"
// @Param offset query int false "Offset"
// @Success 200 {array} dtos.DependentResponseDTO
// @Failure 400 {object} utils.Payload
// @Router /dependents [get]
// GET /dependents?employee_id=...&limit=&offset=
func (h *DependentHandler) ListAllByEmployee(c *gin.Context) {
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
		fmt.Println("ERRO - ", err)
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos dados.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	c.JSON(http.StatusOK, dtos.ToDependentResponseDTOList(list))
}

// NOVO: GET /employees/:id/dependents
func (h *DependentHandler) ListByEmployeePath(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawEmpID := c.Param("id") // lido do path
	employeeID, err := uuid.Parse(rawEmpID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "id", Label: "id", Tag: "uuid", Message: "id do funcionário deve ser um UUID válido.", Value: rawEmpID},
			},
		})
		return
	}

	pagination := utils.PaginationInput(c) // continua a aceitar ?limit=&offset=
	list, err := h.ListUC.Execute(ctx, employeeID, pagination.Limit, pagination.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos dados.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	c.JSON(http.StatusOK, dtos.ToDependentResponseDTOList(list))
}
