package vos

import (
	"errors"
	"regexp"
	"strings"
)

type Nationality struct {
	value string
}

var nationalityRegex = regexp.MustCompile(`^[A-Za-zÀ-ÿ\s\-]{3,}$`)

func NewNationality(input string) (Nationality, error) {
	input = strings.TrimSpace(input)
	if !nationalityRegex.MatchString(input) {
		return Nationality{}, errors.New("nacionalidade inválida: use apenas letras e mínimo de 3 caracteres")
	}
	return Nationality{input}, nil
}

func (n Nationality) String() string {
	return n.value
}

func MustNewNationality(value string) Nationality {
	nat, err := NewNationality(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return nat
}
