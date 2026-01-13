package handlers

/*
import (
	"net/http"
	"rhapp/internal/domain/dtos"
	usecases "rhapp/internal/usecase"
	"rhapp/internal/utils"

	"context"

	"github.com/gin-gonic/gin"
)

type SearchHandler struct {
	MunicipalityUC *usecases.SearchMunicipalityUseCase
	DepartmentUC   *usecases.SearchDepartmentUseCase
	PositionUC     *usecases.SearchPositionUseCase

	AreaEstudoUC *usecases.SearchAreaEstudoUseCase
	DistritoUC   *usecases.SearchDistrictUseCase
	EmployeeUC   *usecases.SearchEmployeesUseCase
}

func NewSearchHandler(
	municipalityUC *usecases.SearchMunicipalityUseCase,
	departmentUC *usecases.SearchDepartmentUseCase,

	provinceUC *usecases.SearchProvinceUseCase,
	areaEstudoUC *usecases.SearchAreaEstudoUseCase,
	distritoUC *usecases.SearchDistrictUseCase,
	employeeUC *usecases.SearchEmployeesUseCase,

) *SearchHandler {
	return &SearchHandler{
		MunicipalityUC: municipalityUC,
		DepartmentUC:   departmentUC,
		PositionUC:     positionUC,
		ProvinceUC:     provinceUC,
		AreaEstudoUC:   areaEstudoUC,
		DistritoUC:     distritoUC,
		EmployeeUC:     employeeUC,
	}
}








func (h *SearchHandler) SearchEmployee(c *gin.Context) {
	input := utils.ParseSearchInput(c)

	list, err := h.EmployeeUC.Execute(context.Background(), input)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, list)
}
*/
