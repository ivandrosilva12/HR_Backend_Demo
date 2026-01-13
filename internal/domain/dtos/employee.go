package dtos

import (
	"errors"
	"rhapp/internal/domain/entities"
	"rhapp/internal/domain/vos"
	"rhapp/internal/utils"
	"time"

	"github.com/google/uuid"
)

type CreateEmployeeDTO struct {
	FullName         string  `json:"full_name" binding:"required,min=3,max=100"`
	Gender           string  `json:"gender" binding:"required,oneof=masculino feminino"`
	DateOfBirth      string  `json:"date_of_birth" binding:"required,datetime=2006-01-02"`
	Nationality      string  `json:"nationality" binding:"required,min=2,max=50"`
	MaritalStatus    string  `json:"marital_status" binding:"required,oneof=solteiro casado divorciado viúvo"`
	PhoneNumber      string  `json:"phone_number" binding:"required,e164"`
	Email            string  `json:"email" binding:"required,email"`
	BI               string  `json:"bi" binding:"required,len=14"`
	IDValidationDate string  `json:"id_date" binding:"required,datetime=2006-01-02"`
	IBAN             string  `json:"iban" binding:"required,min=10,max=34"`
	DepartmentID     string  `json:"department_id" binding:"required,uuid4"`
	PositionID       string  `json:"position_id" binding:"required,uuid4"`
	Address          string  `json:"address" binding:"required,min=5,max=200"`
	DistrictID       string  `json:"district_id" binding:"required,uuid4"`
	HiringDate       string  `json:"hiring_date" binding:"required,datetime=2006-01-02"`
	ContractType     string  `json:"contract_type" binding:"required,oneof=definitivo temporário interno consultor"`
	Salary           float64 `json:"salary" binding:"required,gt=0"`
	SocialSecurity   string  `json:"social_security" binding:"required,min=6,max=12"`
}

type UpdateEmployeeDTO struct {
	MaritalStatus string  `json:"marital_status" binding:"omitempty,oneof=solteiro casado divorciado viúvo"`
	PhoneNumber   string  `json:"phone_number" binding:"omitempty,e164"`
	Email         string  `json:"email" binding:"omitempty,email"`
	IBAN          string  `json:"iban" binding:"omitempty,min=10,max=34"`
	DepartmentID  string  `json:"department_id" binding:"omitempty,uuid4"`
	PositionID    string  `json:"position_id" binding:"omitempty,uuid4"`
	Address       string  `json:"address" binding:"omitempty,min=5,max=200"`
	DistrictID    string  `json:"district_id" binding:"required,uuid4"`
	ContractType  string  `json:"contract_type" binding:"omitempty,oneof=definitivo temporário interno consultor"`
	Salary        float64 `json:"salary" binding:"omitempty,gt=0"`
	IsActive      *bool   `json:"is_active,omitempty"`
}

type EmployeeResponseDTO struct {
	ID               string    `json:"id"`
	EmployeeNumber   int       `json:"employee_number"`
	FullName         string    `json:"full_name"`
	Gender           string    `json:"gender"`
	DateOfBirth      string    `json:"date_of_birth"`
	Nationality      string    `json:"nationality"`
	MaritalStatus    string    `json:"marital_status"`
	PhoneNumber      string    `json:"phone_number"`
	Email            string    `json:"email"`
	BI               string    `json:"bi"`
	IDValidationDate string    `json:"id_date"`
	IBAN             string    `json:"iban"`
	DepartmentID     string    `json:"department_id"`
	PositionID       string    `json:"position_id"`
	Address          string    `json:"address"`
	DistrictID       string    `json:"district_id"`
	HiringDate       string    `json:"hiring_date"`
	ContractType     string    `json:"contract_type"`
	Salary           float64   `json:"salary"`
	SocialSecurity   string    `json:"social_security"`
	IsActive         bool      `json:"is_active"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func ToEmployeeResponseDTO(e entities.Employee) EmployeeResponseDTO {
	return EmployeeResponseDTO{
		ID:               e.ID.String(),
		EmployeeNumber:   e.EmployeeNumber,
		FullName:         e.FullName.String(),
		Gender:           e.Gender.String(),
		DateOfBirth:      e.DateOfBirth.Format("2006-01-02"),
		Nationality:      e.Nationality.String(),
		MaritalStatus:    e.MaritalStatus.String(),
		PhoneNumber:      e.PhoneNumber.String(),
		Email:            e.Email.String(),
		BI:               e.BI.String(),
		IDValidationDate: e.IDValidationDate.Format("2006-01-02"),
		IBAN:             e.IBAN.String(),
		DepartmentID:     e.DepartmentID.String(),
		PositionID:       e.PositionID.String(),
		Address:          e.Address.String(),
		DistrictID:       e.DistrictID.String(),
		HiringDate:       e.HiringDate.Format("2006-01-02"),
		ContractType:     string(e.ContractType),
		Salary:           e.Salary.Float64(),
		SocialSecurity:   e.SocialSecurity.String(),
		IsActive:         e.IsActive,
		CreatedAt:        e.CreatedAt,
		UpdatedAt:        e.UpdatedAt,
	}
}

func ToEmployeeResponseDTOList(list []entities.Employee) []EmployeeResponseDTO {
	result := make([]EmployeeResponseDTO, len(list))
	for i, e := range list {
		result[i] = ToEmployeeResponseDTO(e)
	}
	return result
}

func ToEmployeeFromCreateDTO(input CreateEmployeeDTO) (entities.Employee, error) {
	dateOfBirth, err := vos.NewBirthDate(input.DateOfBirth)
	if err != nil {
		return entities.Employee{}, err
	}

	if utils.MenorDeIdade(dateOfBirth.Time()) {
		return entities.Employee{}, errors.New("menor de idade")
	}

	if time.Now().Year()-dateOfBirth.Time().Year() > 80 {
		return entities.Employee{}, errors.New("nao podem ter acima de 80 anos")
	}

	/* Hiring Date Validation */
	hiringDate, err := time.Parse("2006-01-02", input.HiringDate)
	if err != nil {
		return entities.Employee{}, err
	}

	idDate, err := time.Parse("2006-01-02", input.IDValidationDate)
	if err != nil || idDate.Before(time.Now()) {
		return entities.Employee{}, errors.New("reveja a data de validade do seu BI")
	}

	if time.Now().Before(hiringDate) {
		return entities.Employee{}, errors.New("data de contratação não pode estar no futuro")
	}

	if hiringDate.Before(dateOfBirth.Time()) {
		return entities.Employee{}, errors.New("data de contratação não pode ser menor que a data de nascimento")
	}

	if utils.ContratacaoMenorDeIdade(hiringDate, dateOfBirth.Time()) {
		return entities.Employee{}, errors.New("nao altura da contratacao era menor de idade")
	}

	/* End Hiring Date Validation */

	departmentID, err := uuid.Parse(input.DepartmentID)
	if err != nil {
		return entities.Employee{}, err
	}

	positionID, err := uuid.Parse(input.PositionID)
	if err != nil {
		return entities.Employee{}, err
	}

	fullName, err := vos.NewPersonalName(input.FullName)
	if err != nil {
		return entities.Employee{}, err
	}

	gender, err := vos.NewGender(input.Gender)
	if err != nil {
		return entities.Employee{}, err
	}

	nationality, err := vos.NewNationality(input.Nationality)
	if err != nil {
		return entities.Employee{}, err
	}

	maritalStatus, err := vos.NewMaritalStatus(input.MaritalStatus)
	if err != nil {
		return entities.Employee{}, err
	}

	phone, err := vos.NewPhoneNumber(input.PhoneNumber)
	if err != nil {
		return entities.Employee{}, err
	}

	email, err := vos.NewEmail(input.Email)
	if err != nil {
		return entities.Employee{}, err
	}

	bi, err := vos.NewBI(input.BI)
	if err != nil {
		return entities.Employee{}, err
	}

	iban, err := vos.NewIBAN(input.IBAN)
	if err != nil {
		return entities.Employee{}, err
	}

	address, err := vos.NewAddress(input.Address)
	if err != nil {
		return entities.Employee{}, err
	}

	district, err := uuid.Parse(input.DistrictID)
	if err != nil {
		return entities.Employee{}, err
	}

	contractType, err := vos.NewContractType(input.ContractType)
	if err != nil {
		return entities.Employee{}, err
	}

	salary, err := vos.NewSalary(input.Salary)
	if err != nil {
		return entities.Employee{}, err
	}

	ssn, err := vos.NewSocialSecurity(input.SocialSecurity)
	if err != nil {
		return entities.Employee{}, err
	}

	now := time.Now()

	return entities.Employee{
		ID:               uuid.New(),
		FullName:         fullName,
		Gender:           gender,
		DateOfBirth:      dateOfBirth.Time(),
		Nationality:      nationality,
		MaritalStatus:    maritalStatus,
		PhoneNumber:      phone,
		Email:            email,
		BI:               bi,
		IDValidationDate: idDate,
		IBAN:             iban,
		DepartmentID:     departmentID,
		PositionID:       positionID,
		Address:          address,
		DistrictID:       district,
		HiringDate:       hiringDate,
		ContractType:     contractType,
		Salary:           salary,
		SocialSecurity:   ssn,
		IsActive:         true,
		CreatedAt:        now,
		UpdatedAt:        now,
	}, nil
}

func ApplyUpdateToEmployee(e *entities.Employee, input UpdateEmployeeDTO) error {
	if input.MaritalStatus != "" {
		status, err := vos.NewMaritalStatus(input.MaritalStatus)
		if err != nil {
			return err
		}
		e.MaritalStatus = status
	}

	if input.PhoneNumber != "" {
		phone, err := vos.NewPhoneNumber(input.PhoneNumber)
		if err != nil {
			return err
		}
		e.PhoneNumber = phone
	}

	if input.Email != "" {
		email, err := vos.NewEmail(input.Email)
		if err != nil {
			return err
		}
		e.Email = email
	}

	if input.IBAN != "" {
		iban, err := vos.NewIBAN(input.IBAN)
		if err != nil {
			return err
		}
		e.IBAN = iban
	}

	if input.DepartmentID != "" {
		deptID, err := uuid.Parse(input.DepartmentID)
		if err != nil {
			return err
		}
		e.DepartmentID = deptID
	}

	if input.PositionID != "" {
		posID, err := uuid.Parse(input.PositionID)
		if err != nil {
			return err
		}
		e.PositionID = posID
	}

	if input.Address != "" {
		address, err := vos.NewAddress(input.Address)
		if err != nil {
			return err
		}
		e.Address = address
	}

	if input.DistrictID != "" {
		district, err := uuid.Parse(input.DistrictID)
		if err != nil {
			return err
		}
		e.DistrictID = district
	}

	if input.ContractType != "" {
		contract, err := vos.NewContractType(input.ContractType)
		if err != nil {
			return err
		}
		e.ContractType = contract
	}

	if input.Salary > 0 {
		salary, err := vos.NewSalary(input.Salary)
		if err != nil {
			return err
		}
		e.Salary = salary
	}

	if input.IsActive != nil {
		e.IsActive = *input.IsActive
	}

	e.UpdatedAt = time.Now()
	return nil
}
