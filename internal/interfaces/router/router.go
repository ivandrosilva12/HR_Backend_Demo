package router

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	ginSwagger "github.com/swaggo/gin-swagger"

	_ "rhapp/docs" // <- certifique-se de que o path corresponde ao seu módulo
	"rhapp/internal/infrastructure/redisdb"
	"rhapp/internal/interfaces/handlers"
	"rhapp/internal/interfaces/middleware"

	swaggerFiles "github.com/swaggo/files"
)

func SetupRouter(
	provinceHandler *handlers.ProvinceHandler,
	areaEstudoHandler *handlers.AreaEstudoHandler,
	departmentHandler *handlers.DepartmentHandler,
	positionHandler *handlers.PositionHandler,
	municipalityHandler *handlers.MunicipalityHandler,
	//	searchHandler *handlers.SearchHandler,
	districtHandler *handlers.DistrictHandler,
	employeeHandler *handlers.EmployeeHandler,
	dependentHandler *handlers.DependentHandler,
	documentHandler *handlers.DocumentHandler,
	educationHandler *handlers.EducationHandler,
	empStatusHandler *handlers.EmployeeStatusHandler,
	workerHistoryHandler *handlers.WorkerHistoryHandler,
	workHandler *handlers.WorkHandler,

) *gin.Engine {
	router := gin.Default()

	// Middlewares globais
	router.Use(middleware.LoggerMiddleware())
	router.Use(middleware.CORSMiddleware())
	router.Use(middleware.Recovery())
	router.Use(middleware.StripTrailingSlash())
	// Aplica rate limit de 10 requisições por minuto por IP
	router.Use(middleware.RateLimiterRedisMiddleware(200, time.Minute, redisdb.Client)) // Redis-based limiter
	router.Use(middleware.BodySizeLimit(1 * 1024 * 1024))                               // 1MB
	router.MaxMultipartMemory = 8 << 20                                                 // 8 MiB - Upload de arquivos

	router.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
	router.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })

	api := router.Group("/api")

	provinces := api.Group("/provinces")
	{
		provinces.POST("", provinceHandler.Create)
		provinces.PUT("/:id", provinceHandler.Update)
		provinces.DELETE("/:id", provinceHandler.Delete)
		provinces.GET("/:id", provinceHandler.GetByID)
		provinces.GET("", provinceHandler.List)
		provinces.GET("/search", provinceHandler.SearchProvinces)
	}

	municipalities := api.Group("/municipalities")
	{
		municipalities.POST("", municipalityHandler.Create)
		municipalities.PUT("/:id", municipalityHandler.Update)
		municipalities.DELETE("/:id", municipalityHandler.Delete)
		municipalities.GET("/:id", municipalityHandler.FindByID)
		municipalities.GET("", municipalityHandler.List)
		municipalities.GET("/search", municipalityHandler.SearchMunicipalities)
	}

	areasestudo := api.Group("/areas-estudo")
	{
		areasestudo.POST("", areaEstudoHandler.Create)
		areasestudo.PUT("/:id", areaEstudoHandler.Update)
		areasestudo.DELETE("/:id", areaEstudoHandler.Delete)
		areasestudo.GET("/:id", areaEstudoHandler.GetByID)
		areasestudo.GET("", areaEstudoHandler.ListAll)
		areasestudo.GET("/search", areaEstudoHandler.Search)
	}

	departments := api.Group("/departments")
	{
		departments.POST("", departmentHandler.Create)
		departments.PUT("/:id", departmentHandler.Update)
		departments.DELETE("/:id", departmentHandler.Delete)
		departments.GET("/:id", departmentHandler.FindByID)
		departments.GET("", departmentHandler.FindAll)
		departments.GET("/search", departmentHandler.Search)
		departments.GET("/:id/position-totals", departmentHandler.PositionTotals)

	}

	positions := api.Group("/positions")
	{
		positions.POST("", positionHandler.Create)
		positions.PUT("/:id", positionHandler.Update)
		positions.DELETE("/:id", positionHandler.Delete)
		positions.GET("/:id", positionHandler.FindByID)
		positions.GET("", positionHandler.FindAll)
		positions.GET("/search", positionHandler.SearchPositions)
	}

	districts := api.Group("/districts")
	{
		districts.POST("", districtHandler.Create)
		districts.PUT("/:id", districtHandler.Update)
		districts.DELETE("/:id", districtHandler.Delete)
		districts.GET("/:id", districtHandler.FindByID)
		districts.GET("", districtHandler.List)
		districts.GET("/search", districtHandler.Search)
	}

	employees := api.Group("/employees")
	{
		employees.POST("", employeeHandler.Create)
		employees.PUT("/:id", employeeHandler.Update)
		employees.DELETE("/:id", employeeHandler.Delete)
		employees.GET("/:id", employeeHandler.FindByID)
		employees.GET("", employeeHandler.List)
		employees.GET("/search", employeeHandler.Search)
		employees.GET("/:id/dependents", dependentHandler.ListByEmployeePath)
	}

	dependent := api.Group("/dependents")
	{
		dependent.POST("", dependentHandler.Create)
		dependent.PUT("/:id", dependentHandler.Update)
		dependent.DELETE("/:id", dependentHandler.Delete)
		dependent.GET("/:id", dependentHandler.FindByID)
		dependent.GET("", dependentHandler.ListAllByEmployee)
	}

	document := api.Group("/documents")
	{
		document.GET("", documentHandler.ListByOwnerID)
		document.GET("/:id", documentHandler.FindByID)
		document.DELETE("/:id", documentHandler.Delete)

		// novos (conforme handlers):
		document.POST("/upload", documentHandler.Upload)        // Upload (multipart)
		document.GET("/:id/download", documentHandler.Download) // Download (stream)
		document.PUT("/:id/file", documentHandler.ReplaceFile)  // Substitui o ficheiro
	}

	education := api.Group("/educations")
	{
		education.POST("", educationHandler.Create)
		education.PUT("/:id", educationHandler.Update)
		education.DELETE("/:id", educationHandler.Delete)
		education.GET("/:id", educationHandler.FindByID)
		education.GET("", educationHandler.List)
	}

	emp_status := api.Group("/employee-status")
	{
		emp_status.POST("", empStatusHandler.Create)
		emp_status.PUT("/:id", empStatusHandler.Update)
		emp_status.DELETE("/:id", empStatusHandler.Delete)
		emp_status.GET("/:id", empStatusHandler.FindByID)
		emp_status.GET("", empStatusHandler.ListByEmployee)
	}

	worker_history := api.Group("/worker-history")
	{
		worker_history.POST("", workerHistoryHandler.Create)
		worker_history.PUT("/:id", workerHistoryHandler.Update)
		worker_history.DELETE("/:id", workerHistoryHandler.Delete)
		worker_history.GET("/:id", workerHistoryHandler.FindByID)
		worker_history.GET("", workerHistoryHandler.ListByEmployee)
	}

	work := api.Group("/works")
	{
		work.POST("", workHandler.Create)
		work.PUT("/:id", workHandler.Update)
		work.DELETE("/:id", workHandler.Delete)
		work.GET("/:id", workHandler.FindByID)
		work.GET("", workHandler.ListByEmployee)
	}
	return router
}
