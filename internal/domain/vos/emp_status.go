package vos

import (
	"errors"
	"strings"
)

type EmployeeStatus struct {
	value string
}

var allowedEmployeeStatus = map[string]bool{
	"activo":        true,
	"suspenso":      true,
	"reformado":     true,
	"demitido":      true,
	"convalescente": true,
}

func NewEmployeeStatusValue(input string) (EmployeeStatus, error) {
	normalized := strings.ToLower(strings.TrimSpace(input))
	if !allowedEmployeeStatus[normalized] {
		return EmployeeStatus{}, errors.New("status inválido (valores permitidos: Activo, Suspenso, Reformado, Demitido, Convalescente)")
	}
	return EmployeeStatus{normalized}, nil
}

func (s EmployeeStatus) String() string {
	return s.value
}

func MustNewEmployeeStatusValue(value string) EmployeeStatus {
	ret, err := NewEmployeeStatusValue(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return ret
}
