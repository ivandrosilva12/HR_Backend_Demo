package vos

import (
	"errors"
	"regexp"
	"strings"
)

type PersonalName struct {
	value string
}

func NewPersonalName(input string) (PersonalName, error) {
	nameTrimmed := strings.TrimSpace(input)
	if nameTrimmed == "" {
		return PersonalName{}, errors.New("nome completo obrigatório")
	}
	if !regexp.MustCompile(`^[A-Za-z\sÀ-ü]+$`).MatchString(nameTrimmed) {
		return PersonalName{}, errors.New("nome completo deve conter apenas letras")
	}
	return PersonalName{nameTrimmed}, nil
}

func (f PersonalName) String() string {
	return string(f.value)
}

func MustNewPersonalName(value string) PersonalName {
	ret, err := NewPersonalName(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return ret
}
