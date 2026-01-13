package handlers

import (
	"context"
	"net/http"
	"runtime"
	"sync"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/usecase/areas_estudo"
	"rhapp/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AreaEstudoHandler struct {
	CreateUC        *areas_estudo.CreateAreaEstudoUseCase
	UpdateUC        *areas_estudo.UpdateAreaEstudoUseCase
	DeleteUC        *areas_estudo.DeleteAreaEstudoUseCase
	GetUC           *areas_estudo.GetAreaEstudoByIDUseCase
	ListUC          *areas_estudo.ListAllAreasEstudoUseCase
	SearchUC        *areas_estudo.SearchAreaEstudoUseCase
	LockConcurrency *utils.KeyedLocker
}

func NewAreaEstudoHandler(
	createUC *areas_estudo.CreateAreaEstudoUseCase,
	updateUC *areas_estudo.UpdateAreaEstudoUseCase,
	deleteUC *areas_estudo.DeleteAreaEstudoUseCase,
	getUC *areas_estudo.GetAreaEstudoByIDUseCase,
	listUC *areas_estudo.ListAllAreasEstudoUseCase,
	searchUC *areas_estudo.SearchAreaEstudoUseCase,
	lockConcurrency *utils.KeyedLocker,
) *AreaEstudoHandler {
	return &AreaEstudoHandler{
		CreateUC:        createUC,
		UpdateUC:        updateUC,
		DeleteUC:        deleteUC,
		GetUC:           getUC,
		ListUC:          listUC,
		SearchUC:        searchUC,
		LockConcurrency: lockConcurrency,
	}
}

// Create godoc
// @Summary      Criar nova Área de Estudo
// @Description  Cria uma nova área de estudo com nome e descrição obrigatórios.
// @Tags         Áreas de Estudo
// @Accept       json
// @Produce      json
// @Param        input  body      dtos.CreateAreaEstudoDTO  true  "Dados da Área de Estudo"
// @Success      201    {object}  dtos.AreaEstudoResponseDTO
// @Failure 	 400 	{object}  utils.Payload 	"Validation error with fields detail"
// @Failure      409    {object}  utils.Payload     "Conflito (chave única)"
// @Failure      500    {object}  utils.Payload     "Erro interno"
// @Router       /areas-estudo [post]
func (h *AreaEstudoHandler) Create(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	var dto dtos.CreateAreaEstudoDTO
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

	// Lock por nome normalizado para evitar criação simultânea duplicada
	key := "area:create:" + utils.Normalize(dto.Nome)
	unlock := h.LockConcurrency.Lock(key)
	defer unlock()

	result, err := h.CreateUC.Execute(ctx, areas_estudo.CreateAreaEstudoInput{
		Name:        dto.Nome,
		Description: dto.Description,
	})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusCreated, dtos.ToAreaEstudoResponseDTO(result))
}

// Update godoc
// @Summary Atualizar área de estudo
// @Tags Áreas de Estudo
// @Accept json
// @Produce json
// @Param id path string true "ID da área de estudo"
// @Param input body dtos.UpdateAreaEstudoDTO true "Dados atualizados"
// @Success 200 {object} dtos.AreaEstudoResponseDTO
// @Failure 400 {object} utils.Payload
// @Failure 404 {object} utils.Payload
// @Router /areas-estudo/{id} [put]
func (h *AreaEstudoHandler) Update(c *gin.Context) {
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

	var dto dtos.UpdateAreaEstudoDTO
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

	// Lock por ID para serializar alterações no mesmo recurso
	updateKey := "area:id:" + uid.String()
	unlock := h.LockConcurrency.Lock(updateKey)
	defer unlock()

	result, err := h.UpdateUC.Execute(ctx, areas_estudo.UpdateAreaEstudoInput{
		ID:          uid,
		Nome:        dto.Nome,
		Description: dto.Description,
	})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, dtos.ToAreaEstudoResponseDTO(result))
}

// Delete godoc
// @Summary Excluir área de estudo
// @Description Remove uma área de estudo pelo ID.
// @Tags Áreas de Estudo
// @Produce json
// @Param id path string true "ID da área de estudo"
// @Success 204 "Removido com sucesso"
// @Failure 400 {object} utils.Payload
// @Failure 404 {object} utils.Payload
// @Router /areas-estudo/{id} [delete]
func (h *AreaEstudoHandler) Delete(c *gin.Context) {
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

	// Lock por ID para evitar deletes concorrentes com updates
	deleteKey := "area:id:" + uid.String()
	unlock := h.LockConcurrency.Lock(deleteKey)
	defer unlock()

	if err := h.DeleteUC.Execute(ctx, areas_estudo.DeleteAreaEstudoInput{ID: uid}); err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.Status(http.StatusNoContent)
}

// GetByID godoc
// @Summary Obter área de estudo por ID
// @Tags Áreas de Estudo
// @Produce json
// @Param id path string true "ID da área de estudo"
// @Success 200 {object} dtos.AreaEstudoResponseDTO
// @Failure 400 {object} utils.Payload
// @Failure 404 {object} utils.Payload
// @Router /areas-estudo/{id} [get]
func (h *AreaEstudoHandler) GetByID(c *gin.Context) {
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

	result, err := h.GetUC.Execute(ctx, areas_estudo.GetOneAreaEstudoInput{ID: uid})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, dtos.ToAreaEstudoResponseDTO(result))
}

// @Summary Listar áreas de estudo
// @Tags Áreas de Estudo
// @Produce json
// @Success 200 {object} utils.PagedResponse
// @Router /areas-estudo [get]
func (h *AreaEstudoHandler) ListAll(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	p := utils.PaginationInput(c)
	paged, err := h.ListUC.Execute(ctx, p.Limit, p.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos dados das áreas de estudo.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	itemsDTO := toAreaEstudoDTOsConcurrent(paged.Items)
	out := utils.PagedResponse[dtos.AreaEstudoResponseDTO]{
		Items:  itemsDTO,
		Total:  paged.Total,
		Limit:  paged.Limit,
		Offset: paged.Offset,
	}
	c.JSON(http.StatusOK, out)
}

// @Summary Buscar áreas de estudo
// @Tags Áreas de Estudo
// @Accept json
// @Produce json
// @Param search query string false "Texto de busca"
// @Param filter query string false "Filtro"
// @Param limit query int false "Limite"
// @Param offset query int false "Offset"
// @Success 200 {object} utils.PagedResponse
// @Router /areas-estudo/search [get]
func (h *AreaEstudoHandler) Search(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	input := utils.ParseSearchInput(c)
	paged, err := h.SearchUC.Execute(ctx, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos dados das áreas de estudo.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	itemsDTO := toAreaEstudoDTOsConcurrent(paged.Items)
	out := utils.PagedResponse[dtos.AreaEstudoResponseDTO]{
		Items:  itemsDTO,
		Total:  paged.Total,
		Limit:  paged.Limit,
		Offset: paged.Offset,
	}
	c.JSON(http.StatusOK, out)
}

func toAreaEstudoDTOsConcurrent(items []entities.AreaEstudo) []dtos.AreaEstudoResponseDTO {
	n := len(items)
	if n == 0 {
		return []dtos.AreaEstudoResponseDTO{}
	}
	out := make([]dtos.AreaEstudoResponseDTO, n)
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
				out[i] = dtos.ToAreaEstudoResponseDTO(items[i])
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
