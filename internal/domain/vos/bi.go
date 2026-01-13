package vos

import (
	"errors"
	"regexp"
)

type BI struct {
	value string
}

var biRegex = regexp.MustCompile(`^\d{9}[A-Z]{2}\d{3}$`)

func NewBI(bi string) (BI, error) {
	if !biRegex.MatchString(bi) {
		return BI{}, errors.New("BI inválido: deve conter 9 dígitos + 2 letras + 3 dígitos (ex: 123456789LA045)")
	}
	return BI{bi}, nil
}

func (b BI) String() string {
	return b.value
}

func MustNewBI(value string) BI {
	bi, err := NewBI(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return bi
}
