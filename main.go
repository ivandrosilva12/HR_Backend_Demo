package main

// @title HRMS API
// @version 1.0
// @description API de Gestão de Recursos Humanos (Províncias, Municípios, Departamentos, etc.)
// @termsOfService http://localhost:8080/terms
// @contact.name Equipa de Desenvolvimento
// @contact.email suporte@empresa.com
// @license.name MIT
// @license.url https://opensource.org/licenses/MIT
// @host localhost:8080
// @BasePath /api
// @schemes http

import (
	"context"
	"log"
	"os"

	"rhapp/internal/infrastructure/persistence"
	"rhapp/internal/infrastructure/postgresdb"
	"rhapp/internal/infrastructure/redisdb"
	"rhapp/internal/infrastructure/storage/s3storage"
	"rhapp/internal/interfaces/handlers"
	"rhapp/internal/interfaces/router"
	"rhapp/internal/usecase/areas_estudo"
	"rhapp/internal/usecase/departments"
	"rhapp/internal/usecase/dependents"
	"rhapp/internal/usecase/distritos"
	"rhapp/internal/usecase/documents_uc"
	"rhapp/internal/usecase/education"
	"rhapp/internal/usecase/employee_status"
	"rhapp/internal/usecase/employees"
	"rhapp/internal/usecase/municipios"
	"rhapp/internal/usecase/positions"
	"rhapp/internal/usecase/provincias"
	"rhapp/internal/usecase/workerhistory"
	"rhapp/internal/usecase/workhistory"
	"rhapp/internal/utils"

	_ "github.com/lib/pq"
)

func main() {

	postgresdb.InitPostgres()
	defer postgresdb.DB.Close()

	// Redis
	redisdb.InitRedis()

	// Repositório
	provinceRepo := persistence.NewProvincePgRepository(postgresdb.DB)
	departmentRepo := persistence.NewDepartmentPgRepository(postgresdb.DB)
	positionRepo := persistence.NewPositionPgRepository(postgresdb.DB)
	areaEstudoRepo := persistence.NewAreaEstudoPgRepository(postgresdb.DB)
	municipalityRepo := persistence.NewMunicipalityPgRepository(postgresdb.DB)
	distRepo := persistence.NewDistrictPgRepository(postgresdb.DB)
	employeeRepo := persistence.NewEmployeePgRepository(postgresdb.DB)
	dependentRepo := persistence.NewDependentPgRepository(postgresdb.DB)
	documentRepo := persistence.NewDocumentPgRepository(postgresdb.DB)
	educationRepo := persistence.NewEducationPgRepository(postgresdb.DB)
	empStatusRepo := persistence.NewEmployeeStatusPgRepository(postgresdb.DB)
	supervisorRepo := persistence.NewWorkerHistoryPgRepository(postgresdb.DB)
	workRepo := persistence.NewWorkHistoryPgRepository(postgresdb.DB)

	// ===== S3 Object Storage =====
	s3Bucket := mustGetenv("S3_BUCKET")
	s3Region := mustGetenv("S3_REGION")
	s3Endpoint := os.Getenv("S3_ENDPOINT")      // opcional p/ MinIO/Wasabi
	s3BaseURL := os.Getenv("S3_BASE_URL")       // opcional — para URL pública
	s3ACL := getenvDefault("S3_ACL", "private") // "private" ou "public-read"
	forcePath := getenvDefault("S3_FORCE_PATH", "false") == "true"

	objStorage, err := s3storage.New(
		context.Background(),
		s3storage.Options{
			Bucket:         s3Bucket,
			Region:         s3Region,
			Endpoint:       s3Endpoint,
			BaseURL:        s3BaseURL,
			ForcePathStyle: forcePath,
			ACL:            s3ACL,
		},
	)
	if err != nil {
		log.Fatalf("falha ao inicializar S3 storage: %v", err)
	}

	createWorkerHistoryUC := &workerhistory.CreateUseCase{Repo: supervisorRepo}
	updateWorkerHistoryUC := &workerhistory.UpdateUseCase{Repo: supervisorRepo}
	deleteWorkerHistoryUC := &workerhistory.DeleteUseCase{Repo: supervisorRepo}
	getWorkerHistoryUC := &workerhistory.FindByIDUseCase{Repo: supervisorRepo}
	listWorkerHistoryUC := &workerhistory.ListByEmployeeIDUseCase{Repo: supervisorRepo}

	createWorkUC := &workhistory.CreateWorkHistoryUseCase{Repo: workRepo}
	updateWorkUC := &workhistory.UpdateWorkHistoryUseCase{Repo: workRepo}
	deleteWorkUC := &workhistory.DeleteWorkHistoryUseCase{Repo: workRepo}
	getWorkUC := &workhistory.FindWorkHistoryByIDUseCase{Repo: workRepo}
	listWorkUC := &workhistory.ListWorkHistoryByEmployeeUseCase{Repo: workRepo}

	// Use cases
	createEmployeeStatusUC := &employee_status.CreateEmployeeStatusUseCase{Repo: empStatusRepo}
	updateEmployeeStatusUC := &employee_status.UpdateEmployeeStatusUseCase{Repo: empStatusRepo}
	deleteEmployeeStatusUC := &employee_status.DeleteEmployeeStatusUseCase{Repo: empStatusRepo}
	getEmployeeStatusUC := &employee_status.FindEmployeeStatusByIDUseCase{Repo: empStatusRepo}
	listEmployeeStatusUC := &employee_status.ListEmployeeStatusByEmployeeUseCase{Repo: empStatusRepo}

	// Use cases
	createEducationUC := &education.CreateEducationHistoryUseCase{Repo: educationRepo}
	updateEducationUC := &education.UpdateEducationHistoryUseCase{Repo: educationRepo}
	deleteEducationUC := &education.DeleteEducationHistoryUseCase{Repo: educationRepo}
	getEducationUC := &education.FindEducationHistoryByIDUseCase{Repo: educationRepo}
	listEducationUC := &education.ListEducationHistoriesUseCase{Repo: educationRepo}

	// Use cases
	createDocumentUC := &documents_uc.CreateDocumentUseCase{Repo: documentRepo}
	updateDocumentUC := &documents_uc.UpdateDocumentUseCase{Repo: documentRepo}
	deleteDocumentUC := &documents_uc.DeleteDocumentUseCase{Repo: documentRepo}
	getDocumentUC := &documents_uc.FindDocumentByIDUseCase{Repo: documentRepo}
	listDocumentsUC := &documents_uc.ListDocumentsUseCase{Repo: documentRepo, DependentRepo: dependentRepo, EmployeesRepo: employeeRepo}

	// Use cases
	createProvinceUC := &provincias.CreateProvinceUseCase{Repo: provinceRepo}
	updateProvinceUC := &provincias.UpdateProvinceUseCase{Repo: provinceRepo}
	deleteProvinceUC := &provincias.DeleteProvinceUseCase{Repo: provinceRepo}
	getProvinceUC := &provincias.FindProvinceByIDUseCase{Repo: provinceRepo}
	listProvincesUC := &provincias.FindAllProvincesUseCase{Repo: provinceRepo}
	searchProvinceUC := &provincias.SearchProvinceUseCase{Repo: provinceRepo}

	// Use cases - Municipality
	createMunicipalityUC := &municipios.CreateMunicipalityUseCase{Repo: municipalityRepo}
	updateMunicipalityUC := &municipios.UpdateMunicipalityUseCase{Repo: municipalityRepo}
	deleteMunicipalityUC := &municipios.DeleteMunicipalityUseCase{Repo: municipalityRepo}
	findMunicipalityByIDUC := &municipios.FindMunicipalityByIDUseCase{Repo: municipalityRepo}
	listMunicipalityUC := &municipios.ListMunicipalitiesUseCase{Repo: municipalityRepo}
	searchMunicipalityUC := &municipios.SearchMunicipalityUseCase{Repo: municipalityRepo}

	// 🧠 Use Cases - Department
	createDeptUC := &departments.CreateDepartmentUseCase{Repo: departmentRepo}
	updateDeptUC := &departments.UpdateDepartmentUseCase{Repo: departmentRepo}
	deleteDeptUC := &departments.DeleteDepartmentUseCase{Repo: departmentRepo}
	getDeptUC := &departments.FindDepartmentByIDUseCase{Repo: departmentRepo}
	listDeptsUC := &departments.FindAllDepartmentsUseCase{Repo: departmentRepo}
	searchDeptsUC := &departments.SearchDepartmentUseCase{Repo: departmentRepo}
	deptPositionTotalsUC := &departments.DepartmentPositionTotalsUseCase{Repo: departmentRepo}

	// 🧠 Use Cases - Position
	createPosUC := &positions.CreatePositionUseCase{Repo: positionRepo, DeptRepo: departmentRepo}
	updatePosUC := &positions.UpdatePositionUseCase{Repo: positionRepo, DeptRepo: departmentRepo}
	deletePosUC := &positions.DeletePositionUseCase{Repo: positionRepo}
	getPosUC := &positions.FindPositionByIDUseCase{Repo: positionRepo}
	listPosUC := &positions.FindAllPositionsUseCase{Repo: positionRepo}
	searchPosUC := &positions.SearchPositionUseCase{Repo: positionRepo}

	// 🧠 Use Cases - AreaEstudo
	createAreaUC := &areas_estudo.CreateAreaEstudoUseCase{Repo: areaEstudoRepo}
	updateAreaUC := &areas_estudo.UpdateAreaEstudoUseCase{Repo: areaEstudoRepo}
	deleteAreaUC := &areas_estudo.DeleteAreaEstudoUseCase{Repo: areaEstudoRepo}
	getAreaUC := &areas_estudo.GetAreaEstudoByIDUseCase{Repo: areaEstudoRepo}
	listAreaUC := &areas_estudo.ListAllAreasEstudoUseCase{Repo: areaEstudoRepo}
	searchAreaUC := &areas_estudo.SearchAreaEstudoUseCase{Repo: areaEstudoRepo}

	// Use Cases de District
	createDistrictUC := &distritos.CreateDistrictUseCase{Repo: distRepo}
	updateDistrictUC := &distritos.UpdateDistrictUseCase{Repo: distRepo}
	deleteDistrictUC := &distritos.DeleteDistrictUseCase{Repo: distRepo}
	findDistrictUC := &distritos.FindDistrictByIDUseCase{Repo: distRepo}
	listDistrictUC := &distritos.ListAllDistrictsUseCase{Repo: distRepo}
	searchDistrictUC := &distritos.SearchDistrictUseCase{Repo: distRepo}

	// Use Cases de Employee
	createEmployeeUC := &employees.CreateEmployeeUseCase{Repo: employeeRepo}
	updateEmployeeUC := &employees.UpdateEmployeeUseCase{Repo: employeeRepo}
	deleteEmployeeUC := &employees.DeleteEmployeeUseCase{Repo: employeeRepo}
	findEmployeeUC := &employees.FindEmployeeByIDUseCase{Repo: employeeRepo}
	listEmployeeUC := &employees.ListEmployeesUseCase{Repo: employeeRepo}
	searchEmployeeUC := &employees.SearchEmployeesUseCase{Repo: employeeRepo}

	// Use Cases de Employee
	createDependentUC := &dependents.CreateDependentUseCase{Repo: dependentRepo}
	updateDependentUC := &dependents.UpdateDependentUseCase{Repo: dependentRepo}
	deleteDependentUC := &dependents.DeleteDependentUseCase{Repo: dependentRepo}
	findDependentUC := &dependents.FindDependentByIDUseCase{Repo: dependentRepo}
	listDependentUC := &dependents.ListDependentsUseCase{Repo: dependentRepo}

	lockConcurrency := utils.NewKeyedLocker()

	provinceHandler := handlers.NewProvinceHandler(createProvinceUC, updateProvinceUC, deleteProvinceUC, getProvinceUC, listProvincesUC, searchProvinceUC, lockConcurrency)
	departmentHandler := handlers.NewDepartmentHandler(createDeptUC, updateDeptUC, deleteDeptUC, getDeptUC, listDeptsUC, searchDeptsUC, deptPositionTotalsUC, lockConcurrency)
	positionHandler := handlers.NewPositionHandler(createPosUC, updatePosUC, deletePosUC, getPosUC, listPosUC, searchPosUC, lockConcurrency)
	areaEstudoHandler := handlers.NewAreaEstudoHandler(createAreaUC, updateAreaUC, deleteAreaUC, getAreaUC, listAreaUC, searchAreaUC, lockConcurrency)
	municipalityHandler := handlers.NewMunicipalityHandler(createMunicipalityUC, updateMunicipalityUC, deleteMunicipalityUC, findMunicipalityByIDUC, listMunicipalityUC, searchMunicipalityUC, lockConcurrency)
	distritoHandler := handlers.NewDistrictHandler(createDistrictUC, updateDistrictUC, deleteDistrictUC, findDistrictUC, listDistrictUC, searchDistrictUC, lockConcurrency)
	employeeHandler := handlers.NewEmployeeHandler(createEmployeeUC, updateEmployeeUC, deleteEmployeeUC, findEmployeeUC, listEmployeeUC, searchEmployeeUC, lockConcurrency)
	dependentHandler := handlers.NewDependentHandler(createDependentUC, updateDependentUC, deleteDependentUC, findDependentUC, listDependentUC, lockConcurrency)
	//	searchHandler := handlers.NewSearchHandler(searchMunicipalityUC, searchDepartmentUC, searchPositionUC, searchProvinceUC, searchAreaEstudoUC, searchDistrito, searchEmployee)
	supervisorHandler := handlers.NewWorkerHistoryHandler(createWorkerHistoryUC, updateWorkerHistoryUC, deleteWorkerHistoryUC, getWorkerHistoryUC, listWorkerHistoryUC, lockConcurrency)
	educationHandler := handlers.NewEducationHandler(createEducationUC, updateEducationUC, deleteEducationUC, getEducationUC, listEducationUC, lockConcurrency)

	// ⬇️ Agora o DocumentHandler recebe também o ObjectStorage
	documentHandler := handlers.NewDocumentHandler(createDocumentUC, updateDocumentUC, deleteDocumentUC, getDocumentUC, listDocumentsUC, lockConcurrency, objStorage)

	empStatusHandler := handlers.NewEmployeeStatusHandler(createEmployeeStatusUC, updateEmployeeStatusUC, deleteEmployeeStatusUC, getEmployeeStatusUC, listEmployeeStatusUC, lockConcurrency)
	workHandler := handlers.NewWorkHandler(createWorkUC, updateWorkUC, deleteWorkUC, getWorkUC, listWorkUC, lockConcurrency)

	// Iniciar servidor
	r := router.SetupRouter(
		provinceHandler,
		areaEstudoHandler,
		departmentHandler,
		positionHandler,
		municipalityHandler,
		//		searchHandler,
		distritoHandler,
		employeeHandler,
		dependentHandler,
		documentHandler,
		educationHandler,
		empStatusHandler,
		supervisorHandler,
		workHandler)

	log.Println("Servidor iniciado na porta 8080")
	r.Run(":8080")
}

func mustGetenv(k string) string {
	v := os.Getenv(k)
	if v == "" {
		log.Fatalf("variável de ambiente obrigatória ausente: %s", k)
	}
	return v
}

func getenvDefault(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
