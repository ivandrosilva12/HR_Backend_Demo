package vos

import (
	"errors"
)

type Salary struct {
	value float64
}

const MinimumLegalSalaryAOA = 32000.0

func NewSalary(input float64) (Salary, error) {
	if input < 0 {
		return Salary{}, errors.New("salário não pode ser negativo")
	}
	if input < MinimumLegalSalaryAOA {
		return Salary{}, errors.New("salário abaixo do mínimo legal em Angola")
	}
	return Salary{input}, nil
}

func (s Salary) Float64() float64 {
	return s.value
}

func MustNewSalary(value float64) Salary {
	sal, err := NewSalary(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return sal
}
