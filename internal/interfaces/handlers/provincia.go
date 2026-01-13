package handlers

import (
	"context"
	"net/http"
	"runtime"
	"sync"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/usecase/provincias"
	"rhapp/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ProvinceHandler struct {
	CreateUC        *provincias.CreateProvinceUseCase
	UpdateUC        *provincias.UpdateProvinceUseCase
	DeleteUC        *provincias.DeleteProvinceUseCase
	GetByIDUC       *provincias.FindProvinceByIDUseCase
	ListAllUC       *provincias.FindAllProvincesUseCase
	SearchUC        *provincias.SearchProvinceUseCase
	LockConcurrency *utils.KeyedLocker
}

func NewProvinceHandler(
	createUC *provincias.CreateProvinceUseCase,
	updateUC *provincias.UpdateProvinceUseCase,
	deleteUC *provincias.DeleteProvinceUseCase,
	getByIDUC *provincias.FindProvinceByIDUseCase,
	listAllUC *provincias.FindAllProvincesUseCase,
	searchUC *provincias.SearchProvinceUseCase,
	lockConcurrency *utils.KeyedLocker,
) *ProvinceHandler {
	return &ProvinceHandler{
		CreateUC:        createUC,
		UpdateUC:        updateUC,
		DeleteUC:        deleteUC,
		GetByIDUC:       getByIDUC,
		ListAllUC:       listAllUC,
		SearchUC:        searchUC,
		LockConcurrency: lockConcurrency,
	}
}

// Create godoc
// @Summary      Criar Província
// @Description  Cria uma nova província
// @Tags         Províncias
// @Accept       json
// @Produce      json
// @Param        input  body    dtos.CreateProvinceDoc   true  "Dados da província"
// @Success      201    {object}  dtos.ProvinceResponseDTO
// @Failure 	 400 	{object}  utils.Payload 	"Validation error with fields detail"
// @Failure      409    {object}  utils.Payload     "Conflito (chave única)"
// @Failure      500    {object}  utils.Payload     "Erro interno"
// @Router       /provinces [post]
func (h *ProvinceHandler) Create(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	var dto dtos.CreateProvinceDTO
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

	// Lock por nome normalizado para evitar dupla criação simultânea
	createKey := "prov:create:" + utils.Normalize(dto.Nome)
	unlock := h.LockConcurrency.Lock(createKey)
	defer unlock()

	province, err := h.CreateUC.Execute(ctx, provincias.CreateProvinceInput{
		Name: dto.Nome,
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

	c.JSON(http.StatusCreated, dtos.ToProvinceResponseDTO(province))
}

// Update godoc
// @Summary        Atualizar Província
// @Description    Atualiza o nome de uma província existente.
// @Tags           Províncias
// @Accept         json
// @Produce        json
// @Param          id    path     string                  true  "ID da província (UUID)"
// @Param          input  body     dtos.UpdateProvinceDTO  true  "Payload com o campo 'nome'"
// @Success        200   {object} dtos.ProvinceResponseDTO
// @Failure        400   {object} utils.Payload  "Erro de validação padronizado"
// @Router         /provinces/{id} [put]
func (h *ProvinceHandler) Update(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	var dto dtos.UpdateProvinceDTO
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

	rawID := c.Param("id")
	uid, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{
					Field:   "ID",
					Label:   "id",
					Tag:     "uuid",
					Message: "id deve ser um UUID válido.",
					Value:   rawID,
				},
			},
		})
		return
	}

	// Lock por ID para evitar writes simultâneos no mesmo recurso
	updateKey := "prov:id:" + uid.String()
	unlock := h.LockConcurrency.Lock(updateKey)
	defer unlock()

	province, err := h.UpdateUC.Execute(ctx, provincias.UpdateProvinceInput{
		ID:   uid,
		Nome: dto.Nome,
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

	c.JSON(http.StatusOK, dtos.ToProvinceResponseDTO(province))
}

// Delete godoc
// @Summary        Remover Província
// @Description    Remove uma província existente pelo ID.
// @Tags           Províncias
// @Produce        json
// @Param          id   path     string         true  "ID da província (UUID)"
// @Success        204  {string} string         "No Content"
// @Failure        400  {object} utils.Payload  "Erro de validação padronizado (UUID inválido, província não encontrada, etc.)"
// @Router         /provinces/{id} [delete]
func (h *ProvinceHandler) Delete(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawID := c.Param("id")
	uid, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{
					Field:   "ID",
					Label:   "id",
					Tag:     "uuid",
					Message: "id da provincia deve ser um UUID válido.",
					Value:   rawID,
				},
			},
		})
		return
	}

	// Lock por ID para evitar delete concorrente com update/create correlatos
	deleteKey := "prov:id:" + uid.String()
	unlock := h.LockConcurrency.Lock(deleteKey)
	defer unlock()

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

// GetByID godoc
// @Summary Buscar Província por ID
// @Description Retorna os dados de uma província pelo seu ID
// @Tags Províncias
// @Produce json
// @Param id path string true "ID da província"
// @Success 200 {object} dtos.ProvinceResponseDTO
// @Failure 400  {object} utils.Payload  "Erro de validação padronizado (UUID inválido, província não encontrada, etc.)"
// @Failure 404  {object} utils.Payload
// @Router  /provinces/{id} [get]
func (h *ProvinceHandler) GetByID(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawID := c.Param("id")
	uid, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{
					Field:   "ID",
					Label:   "id",
					Tag:     "uuid",
					Message: "id deve ser um UUID válido.",
					Value:   rawID,
				},
			},
		})
		return
	}

	province, err := h.GetByIDUC.Execute(ctx, provincias.FindProvinceByIDInput{ID: uid})
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

	c.JSON(http.StatusOK, dtos.ToProvinceResponseDTO(province))
}

// List godoc
// @Summary Listar Províncias
// @Description Retorna a lista de províncias com paginação
// @Tags Províncias
// @Produce json
// @Success 200 {object} PagedResponse[dtos.ProvinceResponseDTO]
// @Failure 500  {object} utils.Payload
// @Router /provinces [get]
func (h *ProvinceHandler) List(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	pagination := utils.PaginationInput(c)
	list, total, err := h.ListAllUC.Execute(ctx, pagination.Limit, pagination.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos dados das províncias.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	out := toProvinceDTOsConcurrent(list)
	c.JSON(http.StatusOK, utils.PagedResponse[dtos.ProvinceResponseDTO]{
		Items:  out,
		Total:  total,
		Limit:  pagination.Limit,
		Offset: pagination.Offset,
	})
}

// SearchProvinces godoc
// @Summary Buscar províncias
// @Description Busca províncias com filtros e paginação
// @Tags Províncias
// @Accept json
// @Produce json
// @Param search query string false "Texto de busca (nome da província)"
// @Param filter query string false "Filtro genérico (ignorado aqui)"
// @Param limit query int false "Número máximo de registros a retornar"
// @Param offset query int false "Número de registros a ignorar (para paginação)"
// @Success 200 {object} PagedResponse[dtos.ProvinceResponseDTO]
// @Failure 500  {object} utils.Payload
// @Router /provinces/search [get]
func (h *ProvinceHandler) SearchProvinces(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	input := utils.ParseSearchInput(c)
	results, total, err := h.SearchUC.Execute(ctx, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos dados das províncias.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	out := toProvinceDTOsConcurrent(results)
	c.JSON(http.StatusOK, utils.PagedResponse[dtos.ProvinceResponseDTO]{
		Items:  out,
		Total:  total,
		Limit:  input.Limit,
		Offset: input.Offset,
	})
}

//
// -------------------- Helpers de concorrência --------------------
//

// Converte []entities.Province -> []dtos.ProvinceResponseDTO usando um pool de workers.
func toProvinceDTOsConcurrent(items []entities.Province) []dtos.ProvinceResponseDTO {
	n := len(items)
	if n == 0 {
		return []dtos.ProvinceResponseDTO{}
	}

	out := make([]dtos.ProvinceResponseDTO, n)

	// Define número de workers de forma adaptativa
	workers := runtime.GOMAXPROCS(0)
	if workers < 2 {
		workers = 2
	}
	if workers > n {
		workers = n
	}

	jobs := make(chan int, n)
	var wg sync.WaitGroup

	// Workers
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := range jobs {
				out[i] = dtos.ToProvinceResponseDTO(items[i])
			}
		}()
	}

	// Enfileira índices
	for i := 0; i < n; i++ {
		jobs <- i
	}
	close(jobs)

	wg.Wait()
	return out
}
