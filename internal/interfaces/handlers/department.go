package handlers

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/usecase/departments"
	"rhapp/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DepartmentHandler struct {
	CreateUC         *departments.CreateDepartmentUseCase
	UpdateUC         *departments.UpdateDepartmentUseCase
	DeleteUC         *departments.DeleteDepartmentUseCase
	GetUC            *departments.FindDepartmentByIDUseCase
	ListUC           *departments.FindAllDepartmentsUseCase
	SearchUC         *departments.SearchDepartmentUseCase
	PositionTotalsUC *departments.DepartmentPositionTotalsUseCase
	LockConcurrency  *utils.KeyedLocker
}

func NewDepartmentHandler(
	createUC *departments.CreateDepartmentUseCase,
	updateUC *departments.UpdateDepartmentUseCase,
	deleteUC *departments.DeleteDepartmentUseCase,
	getUC *departments.FindDepartmentByIDUseCase,
	listUC *departments.FindAllDepartmentsUseCase,
	searchUC *departments.SearchDepartmentUseCase,
	positionTotalsUC *departments.DepartmentPositionTotalsUseCase,
	lockConcurrency *utils.KeyedLocker,
) *DepartmentHandler {
	return &DepartmentHandler{
		CreateUC:         createUC,
		UpdateUC:         updateUC,
		DeleteUC:         deleteUC,
		GetUC:            getUC,
		ListUC:           listUC,
		SearchUC:         searchUC,
		PositionTotalsUC: positionTotalsUC,
		LockConcurrency:  lockConcurrency,
	}
}

// Create godoc
// @Summary Criar novo departamento
// @Description Cria um novo departamento com nome único
// @Tags Departamentos
// @Accept json
// @Produce json
// @Param input body dtos.CreateDepartmentDTO true "Dados do departamento"
// @Success 201 {object} dtos.DepartmentResponseDTO
// @Failure 	 400 	{object}  utils.Payload 	"Validation error with fields detail"
// @Failure      409    {object}  utils.Payload     "Conflito (chave única)"
// @Failure      500    {object}  utils.Payload     "Erro interno"
// @Router /departments [post]
// Create godoc
func (h *DepartmentHandler) Create(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	var dto dtos.CreateDepartmentDTO
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

	// ── ParentID é opcional ───────────────────────────────────────────────
	var pID *uuid.UUID
	if dto.ParentID != nil {
		if *dto.ParentID == "" {
			pID = nil // limpar parent
		} else {
			parsed, err := uuid.Parse(*dto.ParentID)
			if err != nil {
				c.JSON(http.StatusBadRequest, utils.Payload{
					Error:   "Validação dos campos",
					Message: "Dados inválidos. Corrija os campos destacados.",
					Fields: []utils.FieldError{
						{Field: "parent_id", Label: "parent_id", Tag: "uuid4", Message: "parent_id deve ser um UUID válido.", Value: dto.ParentID},
					},
				})
				return
			}
			pID = &parsed
		}
	}
	// ──────────────────────────────────────────────────────────────────────

	// Lock por nome normalizado para evitar dupla criação simultânea
	createKey := "dept:create:" + utils.Normalize(dto.Nome)
	unlock := h.LockConcurrency.Lock(createKey)
	defer unlock()

	result, err := h.CreateUC.Execute(ctx, departments.CreateDepartmentInput{
		Nome:     dto.Nome,
		ParentID: pID,
	})
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

	c.JSON(http.StatusCreated, dtos.ToDepartmentResponseDTO(result))
}

// Update godoc
// @Summary Actualizar departamento
// @Description Actualiza os dados de um departamento existente
// @Tags Departamentos
// @Accept json
// @Produce json
// @Param id path string true "ID do departamento"
// @Param input body dtos.UpdateDepartmentDTO true "Dados atualizados"
// @Success 200 {object} dtos.DepartmentResponseDTO
// @Failure 400 {object} utils.Payload  "Erro de validação padronizado"
// @Router /departments/{id} [put]
// Update godoc
func (h *DepartmentHandler) Update(c *gin.Context) {
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

	var dto dtos.UpdateDepartmentDTO
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

	// ── ParentID é opcional ───────────────────────────────────────────────
	var pID *uuid.UUID
	if dto.ParentID != nil {
		if *dto.ParentID == "" {
			pID = nil // limpar parent
		} else {
			parsed, err := uuid.Parse(*dto.ParentID)
			if err != nil {
				c.JSON(http.StatusBadRequest, utils.Payload{
					Error:   "Validação dos campos",
					Message: "Dados inválidos. Corrija os campos destacados.",
					Fields: []utils.FieldError{
						{Field: "parent_id", Label: "parent_id", Tag: "uuid4", Message: "parent_id deve ser um UUID válido.", Value: dto.ParentID},
					},
				})
				return
			}
			pID = &parsed
		}
	}
	// ──────────────────────────────────────────────────────────────────────

	// Lock por ID para evitar writes simultâneos no mesmo recurso
	updateKey := "dept:id:" + uid.String()
	unlock := h.LockConcurrency.Lock(updateKey)
	defer unlock()

	result, err := h.UpdateUC.Execute(ctx, departments.UpdateDepartmentInput{
		ID:       uid,
		Nome:     dto.Nome,
		ParentID: pID,
	})
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

	c.JSON(http.StatusOK, dtos.ToDepartmentResponseDTO(result))
}

// Delete godoc
// @Summary Remover departamento
// @Description Exclui um departamento existente pelo ID
// @Tags Departamentos
// @Produce json
// @Param id path string true "ID do departamento"
// @Success 204 {string} string "No Content"
// @Failure 400 {object} utils.Payload "Erro de validação padronizado (UUID inválido, departamento não encontrada, etc.)"
// @Router /departments/{id} [delete]
func (h *DepartmentHandler) Delete(c *gin.Context) {
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

	// Lock por ID para evitar delete concorrente com update
	deleteKey := "dept:id:" + uid.String()
	unlock := h.LockConcurrency.Lock(deleteKey)
	defer unlock()

	if err := h.DeleteUC.Execute(ctx, departments.DeleteDepartmentInput{ID: uid}); err != nil {
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

// FindByID godoc
// @Summary Buscar departamento por ID
// @Description Retorna os dados de um departamento específico
// @Tags Departamentos
// @Produce json
// @Param id path string true "ID do departamento"
// @Success 200 {object} dtos.DepartmentResponseDTO
// @Failure 400  {object} utils.Payload  "Erro de validação padronizado (UUID inválido, departamento não encontrado, etc.)"
// @Failure 404  {object} utils.Payload
// @Router /departments/{id} [get]
func (h *DepartmentHandler) FindByID(c *gin.Context) {
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

	result, err := h.GetUC.Execute(ctx, departments.FindDepartmentByIDInput{ID: uid})
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

	c.JSON(http.StatusOK, dtos.ToDepartmentResponseDTO(result))
}

// @Summary Listar departamentos
// @Tags Departamentos
// @Produce json
// @Success 200 {object} utils.PagedResponse
// @Router /departments [get]
func (h *DepartmentHandler) FindAll(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	p := utils.PaginationInput(c)
	paged, err := h.ListUC.Execute(ctx, p.Limit, p.Offset)

	if err != nil {

		fmt.Println("Erro - ", err)

		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos dados dos departamentos.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	itemsDTO := toDepartmentDTOsConcurrent(paged.Items) // já retorna [] quando vazio
	out := utils.PagedResponse[dtos.DepartmentResponseDTO]{
		Items:  itemsDTO,
		Total:  paged.Total,
		Limit:  paged.Limit,
		Offset: paged.Offset,
	}
	c.JSON(http.StatusOK, out)
}

// @Summary Buscar departamentos
// @Tags Departamentos
// @Accept json
// @Produce json
// @Param search query string false "Texto de busca"
// @Param filter query string false "Filtro"
// @Param limit query int false "Limite"
// @Param offset query int false "Offset"
// @Success 200 {object} utils.PagedResponse
// @Router /departments/search [get]
func (h *DepartmentHandler) Search(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	input := utils.ParseSearchInput(c)
	paged, err := h.SearchUC.Execute(ctx, input)
	if err != nil {
		fmt.Println("ERRO - ", err)
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos dados dos departamentos.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	itemsDTO := toDepartmentDTOsConcurrent(paged.Items)
	out := utils.PagedResponse[dtos.DepartmentResponseDTO]{
		Items:  itemsDTO,
		Total:  paged.Total,
		Limit:  paged.Limit,
		Offset: paged.Offset,
	}
	c.JSON(http.StatusOK, out)
}

// Conversor concorrente já garante slice vazio se não houver itens
func toDepartmentDTOsConcurrent(items []entities.Department) []dtos.DepartmentResponseDTO {
	n := len(items)
	if n == 0 {
		return []dtos.DepartmentResponseDTO{}
	}
	out := make([]dtos.DepartmentResponseDTO, n)
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
				out[i] = dtos.ToDepartmentResponseDTO(items[i])
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

// GET /departments/:id/position-totals?include_children=true|false

// @Summary Totais de posições por departamento
// @Description Retorna, para o departamento informado (e opcionalmente a subárvore), o total de posições, ocupadas e disponíveis.
// @Tags Departamentos
// @Produce json
// @Param id path string true "ID do departamento (raiz)"
// @Param include_children query bool false "Incluir subdepartamentos" default(false)
// @Success 200 {array} object
// @Failure 400 {object} utils.Payload "UUID inválido"
// @Failure 500 {object} utils.Payload "Erro interno"
// @Router /departments/{id}/position-totals [get]
func (h *DepartmentHandler) PositionTotals(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawID := c.Param("id")
	deptID, err := uuid.Parse(rawID)
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

	includeChildren := utils.ParseBool(c.DefaultQuery("include_children", "false"))

	totals, err := h.PositionTotalsUC.Execute(ctx, departments.DepartmentPositionTotalsInput{
		DepartmentRoot:  deptID,
		IncludeChildren: includeChildren,
	})
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

	// resposta em snake_case
	type resp struct {
		DepartmentID       uuid.UUID `json:"department_id"`
		DepartmentNome     string    `json:"department_nome"`
		TotalPositions     int       `json:"total_positions"`
		OccupiedPositions  int       `json:"occupied_positions"`
		AvailablePositions int       `json:"available_positions"`
	}

	out := make([]resp, 0, len(totals))
	for _, t := range totals {
		out = append(out, resp{
			DepartmentID:       t.DepartmentID,
			DepartmentNome:     t.DepartmentNome,
			TotalPositions:     t.TotalPositions,
			OccupiedPositions:  t.OccupiedPositions,
			AvailablePositions: t.AvailablePositions,
		})
	}

	c.JSON(http.StatusOK, out)
}
