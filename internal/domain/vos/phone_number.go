package vos

import (
	"errors"
	"regexp"
)

type PhoneNumber struct {
	value string
}

func NewPhoneNumber(number string) (PhoneNumber, error) {
	if !regexp.MustCompile(`^\+?244\d{9}$`).MatchString(number) {
		return PhoneNumber{}, errors.New("telefone inválido: deve conter indicativo 244 e 9 dígitos")
	}
	return PhoneNumber{number}, nil
}

func (p PhoneNumber) String() string {
	return p.value
}

func MustNewPhoneNumber(value string) PhoneNumber {
	ret, err := NewPhoneNumber(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return ret
}
