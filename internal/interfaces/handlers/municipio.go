package handlers

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/usecase/municipios"
	"rhapp/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type MunicipalityHandler struct {
	CreateUC        *municipios.CreateMunicipalityUseCase
	UpdateUC        *municipios.UpdateMunicipalityUseCase
	DeleteUC        *municipios.DeleteMunicipalityUseCase
	FindByIDUC      *municipios.FindMunicipalityByIDUseCase
	ListUC          *municipios.ListMunicipalitiesUseCase
	SearchUC        *municipios.SearchMunicipalityUseCase
	LockConcurrency *utils.KeyedLocker
}

func NewMunicipalityHandler(
	createUC *municipios.CreateMunicipalityUseCase,
	updateUC *municipios.UpdateMunicipalityUseCase,
	deleteUC *municipios.DeleteMunicipalityUseCase,
	findByIDUC *municipios.FindMunicipalityByIDUseCase,
	listUC *municipios.ListMunicipalitiesUseCase,
	searchUC *municipios.SearchMunicipalityUseCase,
	lockConcurrency *utils.KeyedLocker,
) *MunicipalityHandler {
	return &MunicipalityHandler{
		CreateUC:        createUC,
		UpdateUC:        updateUC,
		DeleteUC:        deleteUC,
		FindByIDUC:      findByIDUC,
		ListUC:          listUC,
		SearchUC:        searchUC,
		LockConcurrency: lockConcurrency,
	}
}

// CreateMunicipality godoc
// @Summary 	Cria um novo município
// @Description Registra um novo município associado a uma província
// @Tags 		Municípios
// @Accept 		json
// @Produce 	json
// @Param 		input body dtos.CreateMunicipioDoc true "Dados do município"
// @Success		201    {object}		dtos.MunicipioResponseDTO
// @Failure 	400	   {object}		utils.Payload 	"Validation error with fields detail"
// @Failure     409    {object}  	utils.Payload   "Conflito (chave única)"
// @Failure     500    {object}  	utils.Payload   "Erro interno"
// @Router 		/municipalities [post]
func (h *MunicipalityHandler) Create(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	var dto dtos.CreateMunicipioDTO
	if err := utils.BindAndValidateStrict(c, &dto); err != nil {

		fmt.Println("ERRO - ", err)
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

	idProv, err := uuid.Parse(dto.ProvinciaID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{
					Field:   "ID",
					Label:   "id",
					Tag:     "uuid",
					Message: "id da província deve ser um UUID válido.",
					Value:   dto.ProvinciaID},
			},
		})
		return
	}

	// Lock por (provinciaID + nome normalizado) para evitar dupla criação simultânea
	key := "mun:create:" + idProv.String() + ":" + utils.Normalize(dto.Nome)
	unlock := h.LockConcurrency.Lock(key)
	defer unlock()

	entity, err := h.CreateUC.Execute(ctx, municipios.CreateMunicipalityInput{
		Nome:       dto.Nome,
		ProvinceID: idProv,
	})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "internal",
			Message: "Erro interno do servidor"})
		return
	}
	c.JSON(http.StatusCreated, dtos.ToMunicipioResponseDTO(entity))
}

// UpdateMunicipality godoc
// @Summary 	Atualiza os dados de um município
// @Description Modifica o nome ou província associada a um município existente
// @Tags 		Municípios
// @Accept 		json
// @Produce 	json
// @Param 		id 	path string true "ID do município"
// @Param 		input 			body dtos.UpdateMunicipioDTO true "Dados atualizados"
// @Success		200 {object}	dtos.MunicipioResponseDTO
// @Failure 	400 {object}	utils.Payload
// @Failure 	404 {object}	utils.Payload
// @Router 		/municipalities/{id} [put]
func (h *MunicipalityHandler) Update(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawID := c.Param("id")
	id, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id do município deve ser um UUID válido.", Value: rawID},
			},
		})
		return
	}

	var input dtos.UpdateMunicipioDTO
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

	idProv, err := uuid.Parse(input.ProvinciaID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id da província deve ser um UUID válido.", Value: input.ProvinciaID},
			},
		})
		return
	}

	// Lock por ID do município
	updateKey := "mun:id:" + id.String()
	unlock := h.LockConcurrency.Lock(updateKey)
	defer unlock()

	entity, err := h.UpdateUC.Execute(ctx, municipios.UpdateMunicipalityInput{
		ID:         id,
		Nome:       input.Nome,
		ProvinceID: idProv,
	})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(
			http.StatusInternalServerError,
			utils.Payload{
				Error:   "internal",
				Message: "Erro interno do servidor",
			})
		return
	}

	c.JSON(http.StatusOK, dtos.ToMunicipioResponseDTO(entity))
}

// DeleteMunicipality godoc
// @Summary Remover um município
// @Description Exclui um município pelo ID
// @Tags 		Municípios
// @Produce 	json
// @Param		id 		path string true "ID do município (UUID)"
// @Success 	204 	"No Content"
// @Failure 	400 	{object} utils.Payload
// @Failure 	404 	{object} utils.Payload
// @Router 		/municipalities/{id}	[delete]
func (h *MunicipalityHandler) Delete(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawID := c.Param("id")
	uid, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id do municipio deve ser um UUID válido.", Value: rawID},
			},
		})
		return
	}

	// Lock por ID do município
	deleteKey := "mun:id:" + uid.String()
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

// GetMunicipalityByID godoc
// @Summary Busca um município pelo ID
// @Description Retorna os dados de um município específico
// @Tags Municípios
// @Produce json
// @Param id path string true "ID do município"
// @Success 200 {object} dtos.MunicipioResponseDTO
// @Failure 400 {object} utils.Payload
// @Failure 404 {object} utils.Payload
// @Router /municipalities/{id} [get]
func (h *MunicipalityHandler) FindByID(c *gin.Context) {
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

	municipio, err := h.FindByIDUC.Execute(ctx, uid)
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, dtos.ToMunicipioResponseDTO(municipio))
}

// ListMunicipalities godoc
// @Summary Lista todos os municípios
// @Description Retorna municípios paginados
// @Tags Municípios
// @Produce json
// @Success 200 {object} utils.PagedResponse[dtos.MunicipioResponseDTO]
// @Failure 500 {object} utils.Payload
// @Router /municipalities [get]
func (h *MunicipalityHandler) List(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	pagination := utils.PaginationInput(c)
	items, total, err := h.ListUC.Execute(ctx, pagination.Limit, pagination.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos dados dos municípios.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	// Garantir slice não-nil (toMunicipioDTOsConcurrent já devolve [] quando len==0)
	out := toMunicipioDTOsConcurrent(items)
	if out == nil {
		out = []dtos.MunicipioResponseDTO{}
	}

	c.JSON(http.StatusOK, utils.PagedResponse[dtos.MunicipioResponseDTO]{
		Items:  out, // nunca null
		Total:  total,
		Limit:  pagination.Limit,
		Offset: pagination.Offset,
	})
}

// SearchMunicipalities godoc
// @Summary Buscar municípios
// @Description Busca municípios com filtros e paginação
// @Tags Municípios
// @Accept json
// @Produce json
// @Param search query string false "Texto de busca (nome do município)"
// @Param filter query string false "Filtro por nome da província"
// @Param limit query int false "Número máximo de registros a retornar"
// @Param offset query int false "Número de registros a ignorar (para paginação)"
// @Success 200 {object} utils.PagedResponse[dtos.MunicipioResultDTO]
// @Failure 500 {object} utils.Payload
// @Router /municipalities/search [get]
func (h *MunicipalityHandler) SearchMunicipalities(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	input := utils.ParseSearchInput(c)
	results, total, err := h.SearchUC.Execute(ctx, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos dados dos municípios.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	// >>> Garantir [] em vez de null <<<
	if results == nil {
		results = []dtos.MunicipioResultDTO{}
	}

	c.JSON(http.StatusOK, utils.PagedResponse[dtos.MunicipioResultDTO]{
		Items:  results, // nunca null
		Total:  total,
		Limit:  input.Limit,
		Offset: input.Offset,
	})
}

//
// -------------------- Helpers de concorrência --------------------
//

// Converte []entities.Municipality -> []dtos.MunicipioResponseDTO usando um pool de workers.
func toMunicipioDTOsConcurrent(items []entities.Municipality) []dtos.MunicipioResponseDTO {
	n := len(items)
	if n == 0 {
		return []dtos.MunicipioResponseDTO{}
	}

	out := make([]dtos.MunicipioResponseDTO, n)

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
				out[i] = dtos.ToMunicipioResponseDTO(items[i])
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
