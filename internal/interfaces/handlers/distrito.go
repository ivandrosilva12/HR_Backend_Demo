package handlers

import (
	"context"
	"net/http"
	"runtime"
	"sync"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/usecase/distritos"
	"rhapp/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type DistrictHandler struct {
	CreateUC        *distritos.CreateDistrictUseCase
	UpdateUC        *distritos.UpdateDistrictUseCase
	DeleteUC        *distritos.DeleteDistrictUseCase
	FindByIDUC      *distritos.FindDistrictByIDUseCase
	ListUC          *distritos.ListAllDistrictsUseCase
	SearchUC        *distritos.SearchDistrictUseCase
	LockConcurrency *utils.KeyedLocker
}

func NewDistrictHandler(
	createUC *distritos.CreateDistrictUseCase,
	updateUC *distritos.UpdateDistrictUseCase,
	deleteUC *distritos.DeleteDistrictUseCase,
	findByIDUC *distritos.FindDistrictByIDUseCase,
	listUC *distritos.ListAllDistrictsUseCase,
	searchUC *distritos.SearchDistrictUseCase,
	lockConcurrency *utils.KeyedLocker,
) *DistrictHandler {
	return &DistrictHandler{
		CreateUC:        createUC,
		UpdateUC:        updateUC,
		DeleteUC:        deleteUC,
		FindByIDUC:      findByIDUC,
		ListUC:          listUC,
		SearchUC:        searchUC,
		LockConcurrency: lockConcurrency,
	}
}

// CreateDistrict godoc
// @Summary Cria um novo distrito
// @Description Registra um novo distrito associado a um município
// @Tags Distritos
// @Accept json
// @Produce json
// @Param input body dtos.CreateDistritoDTO true "Dados do distrito"
// @Success 201 {object} dtos.DistritoResponseDTO
// @Failure 400 {object} utils.Payload "Validation error with fields detail"
// @Failure 409 {object} utils.Payload "Conflito (chave única)"
// @Failure 500 {object} utils.Payload "Erro interno"
// @Router /districts [post]
func (h *DistrictHandler) Create(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	var dto dtos.CreateDistritoDTO
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

	idMun, err := uuid.Parse(dto.MunicipioID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id do municipio deve ser um UUID válido.", Value: dto.MunicipioID},
			},
		})
		return
	}

	// Lock por (municipioID + nome normalizado) para evitar dupla criação simultânea
	key := "dist:create:" + idMun.String() + ":" + utils.Normalize(dto.Nome)
	unlock := h.LockConcurrency.Lock(key)
	defer unlock()

	entity, err := h.CreateUC.Execute(ctx, distritos.CreateDistrictInput{
		Nome:        dto.Nome,
		MunicipioID: idMun,
	})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusCreated, dtos.ToDistritoResponseDTO(entity))
}

// UpdateDistrict godoc
// @Summary Atualiza os dados de um distrito
// @Description Modifica o nome e/ou município de um distrito existente
// @Tags Distritos
// @Accept json
// @Produce json
// @Param id path string true "ID do distrito"
// @Param input body dtos.UpdateDistritoDTO true "Dados atualizados"
// @Success 200 {object} dtos.DistritoResponseDTO
// @Failure 400 {object} utils.Payload
// @Failure 404 {object} utils.Payload
// @Router /districts/{id} [put]
func (h *DistrictHandler) Update(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id do distrito deve ser um UUID válido.", Value: idStr},
			},
		})
		return
	}

	var input dtos.UpdateDistritoDTO
	if err := utils.BindAndValidateStrict(c, &input); err != nil {
		if fields := utils.HumanizeValidation(input, err); len(fields) > 0 {
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

	idMun, err := uuid.Parse(input.MunicipioID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id do município deve ser um UUID válido.", Value: input.MunicipioID},
			},
		})
		return
	}

	// Lock por ID do distrito para serializar writes no mesmo recurso
	updateKey := "dist:id:" + id.String()
	unlock := h.LockConcurrency.Lock(updateKey)
	defer unlock()

	entity, err := h.UpdateUC.Execute(ctx, distritos.UpdateDistrictInput{
		ID:          id,
		Nome:        input.Nome,
		MunicipioID: idMun,
	})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, dtos.ToDistritoResponseDTO(entity))
}

// DeleteDistrict godoc
// @Summary Remove um distrito
// @Description Exclui um distrito pelo ID
// @Tags Distritos
// @Produce json
// @Param id path string true "ID do distrito"
// @Success 204 "No Content"
// @Failure 400 {object} utils.Payload
// @Failure 404 {object} utils.Payload
// @Router /districts/{id} [delete]
func (h *DistrictHandler) Delete(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawID := c.Param("id")
	uid, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id do distrito deve ser um UUID válido.", Value: rawID},
			},
		})
		return
	}

	// Lock por ID para evitar delete concorrente com update
	deleteKey := "dist:id:" + uid.String()
	unlock := h.LockConcurrency.Lock(deleteKey)
	defer unlock()

	if err := h.DeleteUC.Execute(ctx, distritos.DeleteDistrictInput{ID: uid}); err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}
	c.Status(http.StatusNoContent)
}

// GetDistrictByID godoc
// @Summary Busca um distrito pelo ID
// @Description Retorna os dados de um distrito específico
// @Tags Distritos
// @Produce json
// @Param id path string true "ID do distrito"
// @Success 200 {object} dtos.DistritoResponseDTO
// @Failure 400 {object} utils.Payload
// @Failure 404 {object} utils.Payload
// @Router /districts/{id} [get]
func (h *DistrictHandler) FindByID(c *gin.Context) {
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

	entity, err := h.FindByIDUC.Execute(ctx, distritos.FindDistrictByIDInput{ID: uid})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}
	c.JSON(http.StatusOK, dtos.ToDistritoResponseDTO(entity))
}

// ListDistricts godoc
// @Summary Lista todos os distritos
// @Description Retorna distritos paginados
// @Tags Distritos
// @Produce json
// @Success 200 {object} utils.PagedResponse[dtos.DistritoResponseDTO]
// @Failure 500 {object} utils.Payload
// @Router /districts [get]
func (h *DistrictHandler) List(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	pagination := utils.PaginationInput(c)
	items, total, err := h.ListUC.Execute(ctx, pagination.Limit, pagination.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos dados dos distritos.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	out := toDistritoDTOsConcurrent(items) // já retorna [] quando len==0
	if out == nil {
		out = []dtos.DistritoResponseDTO{} // garante nunca null
	}

	c.JSON(http.StatusOK, utils.PagedResponse[dtos.DistritoResponseDTO]{
		Items:  out, // nunca null
		Total:  total,
		Limit:  pagination.Limit,
		Offset: pagination.Offset,
	})
}

// SearchDistrito godoc
// @Summary Buscar distritos
// @Description Busca distritos com texto, filtros e paginação
// @Tags Distritos
// @Accept json
// @Produce json
// @Param search query string false "Texto de busca (nome do distrito)"
// @Param filter query string false "Filtro por nome do município"
// @Param limit query int false "Número máximo de registros a retornar"
// @Param offset query int false "Número de registros a ignorar (para paginação)"
// @Success 200 {object} utils.PagedResponse[dtos.DistritoResultDTO]
// @Failure 500 {object} utils.Payload
// @Router /districts/search [get]
func (h *DistrictHandler) Search(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	input := utils.ParseSearchInput(c)
	list, total, err := h.SearchUC.Execute(ctx, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos dados dos distritos.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	// Garante slice vazio em vez de null
	if list == nil {
		list = []dtos.DistritoResultDTO{}
	}

	c.JSON(http.StatusOK, utils.PagedResponse[dtos.DistritoResultDTO]{
		Items:  list, // nunca null
		Total:  total,
		Limit:  input.Limit,
		Offset: input.Offset,
	})
}

//
// -------------------- Helpers de concorrência --------------------
//

// Converte []entities.District -> []dtos.DistritoResponseDTO usando um pool de workers.
func toDistritoDTOsConcurrent(items []entities.District) []dtos.DistritoResponseDTO {
	n := len(items)
	if n == 0 {
		return []dtos.DistritoResponseDTO{}
	}

	out := make([]dtos.DistritoResponseDTO, n)

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
				out[i] = dtos.ToDistritoResponseDTO(items[i])
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
