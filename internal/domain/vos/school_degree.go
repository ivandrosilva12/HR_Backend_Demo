package vos

import (
	"errors"
	"strings"
)

type SchoolDegree struct {
	value string
}

var allowedSchoolDegrees = map[string]bool{
	"ensino de base": true,
	"2º ciclo":       true,
	"3º ciclo":       true,
	"ensino médio":   true,
	"bacharelato":    true,
	"licenciatura":   true,
	"mestrado":       true,
	"doutoramento":   true,
}

func NewSchoolDegree(ext string) (SchoolDegree, error) {
	e := strings.ToLower(strings.TrimSpace(ext))
	if !allowedSchoolDegrees[e] {
		return SchoolDegree{}, errors.New("grau acadêmico não suportado")
	}
	return SchoolDegree{e}, nil
}

func (e SchoolDegree) String() string {
	return e.value
}

func MustNewSchoolDegree(value string) SchoolDegree {
	sd, err := NewSchoolDegree(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return sd
}
