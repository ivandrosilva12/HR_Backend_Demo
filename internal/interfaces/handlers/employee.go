package handlers

import (
	"context"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"rhapp/internal/domain/dtos"
	"rhapp/internal/domain/entities"
	"rhapp/internal/usecase/employees"
	"rhapp/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type EmployeeHandler struct {
	CreateUC        *employees.CreateEmployeeUseCase
	UpdateUC        *employees.UpdateEmployeeUseCase
	DeleteUC        *employees.DeleteEmployeeUseCase
	FindByIDUC      *employees.FindEmployeeByIDUseCase
	ListUC          *employees.ListEmployeesUseCase
	SearchUC        *employees.SearchEmployeesUseCase
	LockConcurrency *utils.KeyedLocker
}

func NewEmployeeHandler(
	create *employees.CreateEmployeeUseCase,
	update *employees.UpdateEmployeeUseCase,
	deleteUC *employees.DeleteEmployeeUseCase,
	find *employees.FindEmployeeByIDUseCase,
	list *employees.ListEmployeesUseCase,
	search *employees.SearchEmployeesUseCase,
	lockConcurrency *utils.KeyedLocker,
) *EmployeeHandler {
	return &EmployeeHandler{
		CreateUC:        create,
		UpdateUC:        update,
		DeleteUC:        deleteUC,
		FindByIDUC:      find,
		ListUC:          list,
		SearchUC:        search,
		LockConcurrency: lockConcurrency,
	}
}

// CreateEmployee godoc
// @Summary Cria um novo funcionário
// @Description Registra um novo funcionário no sistema com todos os dados obrigatórios
// @Tags Funcionários
// @Accept json
// @Produce json
// @Param input body dtos.CreateEmployeeDTO true "Dados do funcionário"
// @Success 201 {object} dtos.EmployeeResponseDTO
// @Failure 400 {object} utils.Payload "Validação dos campos (Fields detalhados)"
// @Failure 409 {object} utils.Payload "Conflito (chave única)"
// @Failure 500 {object} utils.Payload "Erro interno"
// @Router /employees [post]
func (h *EmployeeHandler) Create(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	var dto dtos.CreateEmployeeDTO
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

	// Chave de lock para evitar criação concorrente do mesmo funcionário.
	// Prioriza BI; em seguida Email; como fallback usa Nome+DataNascimento.
	lockKey := "emp:create:"
	switch {
	case utils.NotEmptyString(getBI(dto)):
		lockKey += "bi:" + utils.Normalize(getBI(dto))
	case utils.NotEmptyString(getEmail(dto)):
		lockKey += "email:" + utils.Normalize(getEmail(dto))
	default:
		lockKey += "name:" + utils.Normalize(getFullName(dto)) + ":" + getDOBISO(dto)
	}
	unlock := h.LockConcurrency.Lock(lockKey)
	defer unlock()

	entity, err := h.CreateUC.Execute(ctx, employees.CreateEmployeeInput{
		CreateEmployeeDTO: dto,
	})
	if err != nil {

		fmt.Println("ERRO : ", err)
		// Exemplo de regra de negócio específica (menor de idade)
		if err == utils.ErrMenorDeIdade {
			c.JSON(http.StatusBadRequest, utils.Payload{
				Error:   "Validação de campos",
				Message: "Funcionário menor de idade não permitido.",
				Fields:  []utils.FieldError{},
			})
			return
		}
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		fmt.Println("ERRO - ", err)
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusCreated, dtos.ToEmployeeResponseDTO(entity))
}

// UpdateEmployee godoc
// @Summary Atualiza os dados de um funcionário
// @Description Modifica os dados de um funcionário existente
// @Tags Funcionários
// @Accept json
// @Produce json
// @Param id path string true "ID do funcionário"
// @Param input body dtos.UpdateEmployeeDTO true "Dados atualizados"
// @Success 200 {object} dtos.EmployeeResponseDTO
// @Failure 400 {object} utils.Payload
// @Failure 404 {object} utils.Payload
// @Failure 409 {object} utils.Payload
// @Failure 500 {object} utils.Payload
// @Router /employees/{id} [put]
func (h *EmployeeHandler) Update(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id do funcionário deve ser um UUID válido.", Value: idStr},
			},
		})
		return
	}

	var dto dtos.UpdateEmployeeDTO
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

	// Serializa writes por ID
	updateKey := "emp:id:" + id.String()
	unlock := h.LockConcurrency.Lock(updateKey)
	defer unlock()

	entity, err := h.UpdateUC.Execute(ctx, employees.UpdateEmployeeInput{
		ID:          id,
		EmployeeDTo: dto,
	})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}
	c.JSON(http.StatusOK, dtos.ToEmployeeResponseDTO(entity))
}

// DeleteEmployee godoc
// @Summary Remove um funcionário
// @Description Exclui um funcionário pelo ID
// @Tags Funcionários
// @Produce json
// @Param id path string true "ID do funcionário"
// @Success 204 "No Content"
// @Failure 400 {object} utils.Payload
// @Failure 404 {object} utils.Payload
// @Failure 500 {object} utils.Payload
// @Router /employees/{id} [delete]
func (h *EmployeeHandler) Delete(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	rawID := c.Param("id")
	uid, err := uuid.Parse(rawID)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "ID", Label: "id", Tag: "uuid", Message: "id do funcionário deve ser um UUID válido.", Value: rawID},
			},
		})
		return
	}

	// Serializa delete por ID (evita corrida com update)
	deleteKey := "emp:id:" + uid.String()
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

// GetEmployeeByID godoc
// @Summary Busca um funcionário pelo ID
// @Description Retorna os dados de um funcionário específico
// @Tags Funcionários
// @Produce json
// @Param id path string true "ID do funcionário"
// @Success 200 {object} dtos.EmployeeResponseDTO
// @Failure 400 {object} utils.Payload
// @Failure 404 {object} utils.Payload
// @Failure 500 {object} utils.Payload
// @Router /employees/{id} [get]
func (h *EmployeeHandler) FindByID(c *gin.Context) {
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

	entity, err := h.FindByIDUC.Execute(ctx, uid)
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}
	c.JSON(http.StatusOK, dtos.ToEmployeeResponseDTO(entity))
}

// ListEmployees godoc
// @Summary Lista funcionários
// @Description Retorna uma lista paginada de funcionários
// @Tags Funcionários
// @Produce json
// @Param limit query int false "Número máximo de registros a retornar"
// @Param offset query int false "Número de registros a ignorar (para paginação)"
// @Success 200 {object} utils.PagedResponse "items: []EmployeeResponseDTO"
// @Failure 500 {object} utils.Payload
// @Router /employees [get]
func (h *EmployeeHandler) List(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	pagination := utils.PaginationInput(c)
	result, err := h.ListUC.Execute(ctx, pagination.Limit, pagination.Offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos dados dos funcionários.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	itemsDTO := toEmployeeDTOsConcurrent(result.Items)
	out := utils.PagedResponse[dtos.EmployeeResponseDTO]{
		Items:  itemsDTO,
		Total:  result.Total,
		Limit:  result.Limit,
		Offset: result.Offset,
	}
	c.JSON(http.StatusOK, out)
}

// SearchEmployees godoc
// @Summary Buscar funcionários
// @Description Busca funcionários com texto, filtros e paginação
// @Tags Funcionários
// @Accept json
// @Produce json
// @Param search query string false "Texto de busca (nome, email, nº funcionário, etc.)"
// @Param filter query string false "Filtro (ex.: dept:UUID pos:UUID ativo:true)"
// @Param limit query int false "Número máximo de registros a retornar"
// @Param offset query int false "Número de registros a ignorar (para paginação)"
// @Success 200 {object} utils.PagedResponse "items: []EmployeeResponseDTO"
// @Failure 500 {object} utils.Payload
// @Router /employees/search [get]
func (h *EmployeeHandler) Search(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	input := utils.ParseSearchInput(c)
	results, err := h.SearchUC.Execute(ctx, input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, utils.Payload{
			Error:   "Erros na DB",
			Message: "Falha no carregamento dos dados dos funcionários.",
			Fields:  []utils.FieldError{},
		})
		return
	}

	itemsDTO := toEmployeeDTOsConcurrent(results.Items)
	out := utils.PagedResponse[dtos.EmployeeResponseDTO]{
		Items:  itemsDTO,
		Total:  results.Total,
		Limit:  results.Limit,
		Offset: results.Offset,
	}
	c.JSON(http.StatusOK, out)
}

// -------------------- Helpers --------------------

func getBI(dto dtos.CreateEmployeeDTO) string {
	type biCarrier interface{ GetBI() string }
	if v, ok := any(dto).(biCarrier); ok {
		return v.GetBI()
	}
	return utils.ExtractFieldString(dto, "BI")
}

func getEmail(dto dtos.CreateEmployeeDTO) string {
	type emailCarrier interface{ GetEmail() string }
	if v, ok := any(dto).(emailCarrier); ok {
		return v.GetEmail()
	}
	return utils.ExtractFieldString(dto, "Email")
}

func getFullName(dto dtos.CreateEmployeeDTO) string {
	type nameCarrier interface{ GetFullName() string }
	if v, ok := any(dto).(nameCarrier); ok {
		return v.GetFullName()
	}
	return utils.ExtractFieldString(dto, "FullName")
}

func getDOBISO(dto dtos.CreateEmployeeDTO) string {
	if t, ok := utils.ExtractFieldTime(dto, "DateOfBirth"); ok {
		return t.Format(time.DateOnly)
	}
	return "unknown"
}

func toEmployeeDTOsConcurrent(items []entities.Employee) []dtos.EmployeeResponseDTO {
	n := len(items)
	if n == 0 {
		return []dtos.EmployeeResponseDTO{}
	}

	out := make([]dtos.EmployeeResponseDTO, n)

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
				out[i] = dtos.ToEmployeeResponseDTO(items[i])
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
