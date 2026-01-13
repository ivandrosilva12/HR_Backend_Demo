package vos

import (
	"errors"
	"strings"
)

// Gender - VO para gênero

type MaritalStatus struct {
	value string
}

var allowedMaritalStatus = map[string]bool{
	"solteiro":       true,
	"casado":         true,
	"união de facto": true,
	"divorciado":     true,
	"viúvo":          true,
}

func NewMaritalStatus(maritalStr string) (MaritalStatus, error) {
	v := strings.ToLower(strings.TrimSpace(maritalStr))

	if !allowedMaritalStatus[v] {
		return MaritalStatus{}, errors.New("estado civil inválido")
	}

	return MaritalStatus{v}, nil

}

func (f MaritalStatus) String() string {
	return string(f.value)
}

func MustNewMaritalStatus(value string) MaritalStatus {
	ret, err := NewMaritalStatus(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return ret
}
