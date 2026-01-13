package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/usecase/education"
	"rhapp/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EducationHandler struct {
	CreateUC        *education.CreateEducationHistoryUseCase
	UpdateUC        *education.UpdateEducationHistoryUseCase
	DeleteUC        *education.DeleteEducationHistoryUseCase
	FindUC          *education.FindEducationHistoryByIDUseCase
	ListUC          *education.ListEducationHistoriesUseCase
	LockConcurrency *utils.KeyedLocker
}

func NewEducationHandler(
	create *education.CreateEducationHistoryUseCase,
	update *education.UpdateEducationHistoryUseCase,
	deleteUC *education.DeleteEducationHistoryUseCase,
	find *education.FindEducationHistoryByIDUseCase,
	list *education.ListEducationHistoriesUseCase,
	lock *utils.KeyedLocker,
) *EducationHandler {
	return &EducationHandler{
		CreateUC:        create,
		UpdateUC:        update,
		DeleteUC:        deleteUC,
		FindUC:          find,
		ListUC:          list,
		LockConcurrency: lock,
	}
}

// @Summary Criar histórico educacional
// @Tags Education
// @Accept json
// @Produce json
// @Param input body dtos.CreateEducationDTO true "Dados do histórico educacional"
// @Success 201 {object} dtos.EducationResponseDTO
// @Failure 400,409 {object} utils.Payload
// @Router /education [post]
func (h *EducationHandler) Create(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	var dto dtos.CreateEducationDTO

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

	areaEstudoID, err := uuid.Parse(dto.FieldOfStudy)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id da área de estudo deve ser um UUID válido.", Value: dto.FieldOfStudy},
			},
		})
		return
	}

	const layout = "2006-01-02"
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

	// Lock para evitar criação duplicada concorrente para o mesmo funcionário/curso/período
	// Chave: empID + instituição(normalizada) + grau(normalizado) + startDate
	lockKey := "edu:create:" + empID.String() + ":" + utils.Normalize(dto.Institution) + ":" + utils.Normalize(dto.Degree) + ":" + startDate.Format("2006-01-02")
	unlock := h.LockConcurrency.Lock(lockKey)
	defer unlock()

	result, err := h.CreateUC.Execute(ctx, education.CreateEducationHistoryInput{
		EmployeeID:   empID,
		Institution:  dto.Institution,
		Degree:       dto.Degree,
		AreaEstudoID: areaEstudoID,
		StartDate:    startDate,
		EndDate:      endDate,
		Description:  dto.Description,
	})
	if err != nil {
		fmt.Println("ERRO : ", err)
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusCreated, dtos.ToEducationResponseDTO(result))
}

// @Summary Atualizar histórico educacional
// @Tags Education
// @Accept json
// @Produce json
// @Param id path string true "ID do histórico"
// @Param input body dtos.UpdateEducationDTO true "Dados atualizados"
// @Success 200 {object} dtos.EducationResponseDTO
// @Failure 400,404 {object} utils.Payload
// @Router /education/{id} [put]
func (h *EducationHandler) Update(c *gin.Context) {
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

	var dto dtos.UpdateEducationDTO
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

	const layout = "2006-01-02"
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

	areaEstudoID, err := uuid.Parse(dto.FieldOfStudy)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id da área de estudo deve ser um UUID válido.", Value: dto.FieldOfStudy},
			},
		})
		return
	}

	// Lock por ID para serializar writes no mesmo recurso
	updateKey := "edu:id:" + id.String()
	unlock := h.LockConcurrency.Lock(updateKey)
	defer unlock()

	result, err := h.UpdateUC.Execute(ctx, education.UpdateEducationHistoryInput{
		ID:           id,
		Institution:  dto.Institution,
		Degree:       dto.Degree,
		AreaEstudoID: areaEstudoID,
		StartDate:    startDate,
		EndDate:      endDate,
		Description:  dto.Description,
	})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, dtos.ToEducationResponseDTO(result))
}

// @Summary Remover histórico educacional
// @Tags Education
// @Produce json
// @Param id path string true "ID do histórico"
// @Success 204
// @Failure 400,404 {object} utils.Payload
// @Router /education/{id} [delete]
func (h *EducationHandler) Delete(c *gin.Context) {
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
	deleteKey := "edu:id:" + uid.String()
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

// @Summary Buscar histórico por ID
// @Tags Education
// @Produce json
// @Param id path string true "ID do histórico"
// @Success 200 {object} dtos.EducationResponseDTO
// @Failure 400,404 {object} utils.Payload
// @Router /education/{id} [get]
func (h *EducationHandler) FindByID(c *gin.Context) {
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

	res, err := h.FindUC.Execute(ctx, uid)
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, dtos.ToEducationResponseDTO(res))
}

// @Summary Listar históricos educacionais por funcionário
// @Tags Education
// @Param employee_id query string true "ID do funcionário"
// @Param limit query int false "Limite de resultados"
// @Param offset query int false "Offset de paginação"
// @Success 200 {array} dtos.EducationResponseDTO
// @Failure 400 {object} utils.Payload
// @Router /education [get]
func (h *EducationHandler) List(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawEmpID := c.Query("employee_id")
	employeeID, err := uuid.Parse(rawEmpID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields:  []utils.FieldError{{Field: "ID", Label: "id", Tag: "uuid", Message: "id do funcionário deve ser um UUID válido.", Value: rawEmpID}},
		})
		return
	}

	pagination := utils.PaginationInput(c)

	res, err := h.ListUC.Execute(ctx, employeeID, pagination.Limit, pagination.Offset)
	if err != nil {
		fmt.Println("ERRO - ", err)
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos dados.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	c.JSON(http.StatusOK, dtos.ToEducationResponseDTOList(res))
}
