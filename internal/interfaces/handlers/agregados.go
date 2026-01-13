package handlers

import (
	"context"
	"net/http"

	"rhapp/internal/domain/vos"
	"rhapp/internal/usecase/agregados"
	"rhapp/internal/utils"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AgregadosHandler struct {
	EmployeeAggUC *agregados.GetEmployeeAggregateByIDUseCase
	DocumentAggUC *agregados.GetDocumentsByOwnerUseCase
	LocationAggUC *agregados.GetLocationByProvinceIDUseCase
	OrgAggUC      *agregados.GetOrgStructureByDepartmentUseCase
}

func NewAgregadosHandler(
	empUC *agregados.GetEmployeeAggregateByIDUseCase,
	docUC *agregados.GetDocumentsByOwnerUseCase,
	locUC *agregados.GetLocationByProvinceIDUseCase,
	orgUC *agregados.GetOrgStructureByDepartmentUseCase,
) *AgregadosHandler {
	return &AgregadosHandler{
		EmployeeAggUC: empUC,
		DocumentAggUC: docUC,
		LocationAggUC: locUC,
		OrgAggUC:      orgUC,
	}
}

// GetEmployeeAggregate godoc
// @Summary      Buscar dados completos de um funcionário
// @Tags         Agregados
// @Produce      json
// @Param        id   path      string  true  "ID do Funcionário"
// @Success      200  {object}  agregados.EmployeeAggregate
// @Failure      400  {object}  utils.Payload
// @Failure      404  {object}  utils.Payload
// @Failure      500  {object}  utils.Payload
// @Router       /employees/{id}/aggregate [get]
func (h *AgregadosHandler) GetEmployeeAggregate(c *gin.Context) {
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

	agg, err := h.EmployeeAggUC.Execute(ctx, agregados.GetEmployeeAggregateInput{ID: uid})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, agg)
}

// GetDocumentsByOwner godoc
// @Summary      Buscar documentos por proprietário
// @Tags         Agregados
// @Produce      json
// @Param        owner_type  query     string  true  "employee ou dependent"
// @Param        owner_id    query     string  true  "UUID do proprietário"
// @Success      200  {object}  agregados.DocumentAggregate
// @Failure      400  {object}  utils.Payload
// @Failure      404  {object}  utils.Payload
// @Failure      500  {object}  utils.Payload
// @Router       /documents/aggregate [get]
func (h *AgregadosHandler) GetDocumentsByOwner(c *gin.Context) {
	ctx, cancel := context.WithTimeout(c.Request.Context(), utils.HandlerTimeout)
	defer cancel()

	ownerTypeStr := c.Query("owner_type")
	ownerIDStr := c.Query("owner_id")

	// validação do owner_id
	ownerID, err := uuid.Parse(ownerIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "OwnerID", Label: "owner_id", Tag: "uuid", Message: "owner_id deve ser um UUID válido.", Value: ownerIDStr},
			},
		})
		return
	}

	// validação do owner_type
	if !isValidOwnerType(ownerTypeStr) {
		c.JSON(http.StatusBadRequest, utils.Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields: []utils.FieldError{
				{Field: "OwnerType", Label: "owner_type", Tag: "oneof", Message: "owner_type deve ser 'employee' ou 'dependent'.", Value: ownerTypeStr},
			},
		})
		return
	}

	agg, err := h.DocumentAggUC.Execute(ctx, agregados.GetDocumentsByOwnerInput{
		OwnerType: vos.DocumentOwnerType(ownerTypeStr),
		OwnerID:   ownerID,
	})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, agg)
}

// GetLocationAggregate godoc
// @Summary      Buscar província com municípios e distritos
// @Tags         Agregados
// @Produce      json
// @Param        id   path      string  true  "ID da Província"
// @Success      200  {object}  agregados.LocationAggregate
// @Failure      400  {object}  utils.Payload
// @Failure      404  {object}  utils.Payload
// @Failure      500  {object}  utils.Payload
// @Router       /provinces/{id}/aggregate [get]
func (h *AgregadosHandler) GetLocationAggregate(c *gin.Context) {
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

	agg, err := h.LocationAggUC.Execute(ctx, agregados.GetLocationByProvinceInput{ProvinceID: uid})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, agg)
}

// GetOrgStructureAggregate godoc
// @Summary      Buscar departamento com posições e funcionários
// @Tags         Agregados
// @Produce      json
// @Param        id   path      string  true  "ID do Departamento"
// @Success      200  {object}  agregados.OrgStructureAggregate
// @Failure      400  {object}  utils.Payload
// @Failure      404  {object}  utils.Payload
// @Failure      500  {object}  utils.Payload
// @Router       /departments/{id}/aggregate [get]
func (h *AgregadosHandler) GetOrgStructureAggregate(c *gin.Context) {
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

	agg, err := h.OrgAggUC.Execute(ctx, agregados.GetOrgStructureByDepartmentInput{DepartmentID: uid})
	if err != nil {
		if ok, payload, status := utils.HumanizeDB(err); ok {
			c.JSON(status, payload)
			return
		}
		c.JSON(http.StatusInternalServerError, utils.Payload{Error: "internal", Message: "Erro interno do servidor"})
		return
	}

	c.JSON(http.StatusOK, agg)
}

// --------- helpers ---------

func isValidOwnerType(t string) bool {
	switch t {
	case "employee", "dependent":
		return true
	default:
		return false
	}
}
