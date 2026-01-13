package handlers

import (
	"context"
	"net/http"
	"runtime"
	"sync"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/usecase/positions"
	"rhapp/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type PositionHandler struct {
	CreateUC        *positions.CreatePositionUseCase
	UpdateUC        *positions.UpdatePositionUseCase
	DeleteUC        *positions.DeletePositionUseCase
	FindOneUC       *positions.FindPositionByIDUseCase
	FindAllUC       *positions.FindAllPositionsUseCase
	PositionUC      *positions.SearchPositionUseCase
	LockConcurrency *utils.KeyedLocker
}

func NewPositionHandler(
	createUC *positions.CreatePositionUseCase,
	updateUC *positions.UpdatePositionUseCase,
	deleteUC *positions.DeletePositionUseCase,
	findOneUC *positions.FindPositionByIDUseCase,
	findAllUC *positions.FindAllPositionsUseCase,
	positionUC *positions.SearchPositionUseCase,
	lockConcurrency *utils.KeyedLocker,
) *PositionHandler {
	return &PositionHandler{
		CreateUC:        createUC,
		UpdateUC:        updateUC,
		DeleteUC:        deleteUC,
		FindOneUC:       findOneUC,
		FindAllUC:       findAllUC,
		PositionUC:      positionUC,
		LockConcurrency: lockConcurrency,
	}
}

// Create godoc
// @Summary Criar nova posição
// @Description Cria uma nova posição (cargo) associada a um departamento
// @Tags Posições
// @Accept json
// @Produce json
// @Param payload body dtos.CreatePositionDTO true "Dados da posição"
// @Success 201 {object} dtos.PositionResponseDTO
// @Failure 400 {object}  utils.Payload
// @Failure 409 {object}  utils.Payload
// @Router /positions [post]
func (h *PositionHandler) Create(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	var dto dtos.CreatePositionDTO
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

	deptID, err := uuid.Parse(dto.DepartmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id deve ser um UUID válido.", Value: dto.DepartmentID},
			},
		})
		return
	}

	// Lock por (departmentID + nome normalizado) para evitar criação duplicada concorrente
	key := "pos:create:" + deptID.String() + ":" + utils.Normalize(dto.Nome)
	unlock := h.LockConcurrency.Lock(key)
	defer unlock()

	position, err := h.CreateUC.Execute(ctx, positions.CreatePositionInput{
		Name:         dto.Nome,
		DepartmentID: deptID,
		MaxHeadcount: dto.MaxHeadcount,
		Tipo:         dto.Tipo, // NEW
	})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusCreated, dtos.ToPositionResponseDTO(position))
}

// Update godoc
// @Summary Atualizar posição
// @Description Atualiza os dados de uma posição existente
// @Tags Posições
// @Accept json
// @Produce json
// @Param id path string true "ID da posição"
// @Param payload body dtos.UpdatePositionDTO true "Campos atualizados"
// @Success 200 {object} dtos.PositionResponseDTO
// @Failure 400 {object} utils.Payload
// @Failure 404 {object} utils.Payload
// @Router /positions/{id} [put]
func (h *PositionHandler) Update(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	var dto dtos.UpdatePositionDTO
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

	idStr := c.Param("id")
	posID, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id deve ser um UUID válido.", Value: idStr},
			},
		})
		return
	}

	depID, err := uuid.Parse(dto.DepartmentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id deve ser um UUID válido.", Value: dto.DepartmentID},
			},
		})
		return
	}

	// Lock por ID para serializar alterações na mesma posição
	updateKey := "pos:id:" + posID.String()
	unlock := h.LockConcurrency.Lock(updateKey)
	defer unlock()

	position, err := h.UpdateUC.Execute(ctx, positions.UpdatePositionInput{
		ID:           posID,
		Nome:         dto.Nome,
		DepartmentID: depID,
		MaxHeadcount: dto.MaxHeadcount,
		Tipo:         dto.Tipo, // NEW
	})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, dtos.ToPositionResponseDTO(position))
}

// Delete godoc
// @Summary Remover posição
// @Description Remove uma posição com base no ID
// @Tags Posições
// @Produce json
// @Param id path string true "ID da posição"
// @Success 204 "No Content"
// @Failure 400 {object} utils.Payload
// @Failure 404 {object} utils.Payload
// @Router /positions/{id} [delete]
func (h *PositionHandler) Delete(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id deve ser um UUID válido.", Value: idStr},
			},
		})
		return
	}

	// Lock por ID para evitar conflito com updates/deletes simultâneos
	deleteKey := "pos:id:" + id.String()
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

// FindByID godoc
// @Summary Buscar posição por ID
// @Description Retorna os dados de uma posição específica
// @Tags Posições
// @Produce json
// @Param id path string true "ID da posição"
// @Success 200 {object} dtos.PositionResponseDTO
// @Failure 400 {object} utils.Payload
// @Failure 404 {object} utils.Payload
// @Router /positions/{id} [get]
func (h *PositionHandler) FindByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id deve ser um UUID válido.", Value: idStr},
			},
		})
		return
	}

	position, err := h.FindOneUC.Execute(ctx, positions.FindPositionByIDInput{ID: id})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, dtos.ToPositionResponseDTO(position))
}

// FindAll godoc
// @Summary Listar posições
// @Description Lista todas as posições cadastradas
// @Tags Posições
// @Produce json
// @Success 200 {object} utils.PagedResponse  "items: []PositionResponseDTO"
// @Failure 500 {object} utils.Payload
// @Router /positions [get]
func (h *PositionHandler) FindAll(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	pagination := utils.PaginationInput(c)
	res, err := h.FindAllUC.Execute(ctx, pagination.Limit, pagination.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos dados das posições.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	// Converte Items -> DTO em paralelo, preservando meta (total, limit, offset)
	itemsDTO := toPositionDTOsConcurrent(res.Items)
	out := utils.PagedResponse[dtos.PositionResponseDTO]{
		Items:  itemsDTO, // já garante [] quando vazio
		Total:  res.Total,
		Limit:  res.Limit,
		Offset: res.Offset,
	}
	c.JSON(http.StatusOK, out)
}

// SearchPositions godoc
// @Summary Buscar posições
// @Description Busca cargos/posições com filtros e paginação
// @Tags Posições
// @Accept json
// @Produce json
// @Param search query string false "Texto de busca (nome da posição)"
// @Param filter query string false "Filtro por nome do departamento"
// @Param limit query int false "Número máximo de registros a retornar"
// @Param offset query int false "Número de registros a ignorar (para paginação)"
// @Success 200 {object} utils.PagedResponse "items: []PositionResultDTO"
// @Failure 500 {object} utils.Payload
// @Router /positions/search [get]
func (h *PositionHandler) SearchPositions(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	input := utils.ParseSearchInput(c)
	res, err := h.PositionUC.Execute(ctx, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos dados das posições.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	// res já é PagedResponse[dtos.PositionResultDTO]
	// garante items não-nulos
	if res.Items == nil {
		res.Items = []dtos.PositionResultDTO{}
	}
	c.JSON(http.StatusOK, res)
}

//
// -------------------- Helpers de concorrência --------------------
//

// Converte []entities.Position -> []dtos.PositionResponseDTO usando pool de workers.
func toPositionDTOsConcurrent(items []entities.Position) []dtos.PositionResponseDTO {
	n := len(items)
	if n == 0 {
		return []dtos.PositionResponseDTO{} // nunca null
	}

	out := make([]dtos.PositionResponseDTO, n)

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
				out[i] = dtos.ToPositionResponseDTO(items[i])
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
