package vos

import (
	"errors"
	"regexp"
	"strings"
)

type IBAN struct {
	value string
}

var ibanRegex = regexp.MustCompile(`^AO06\d{21}$`)

func NewIBAN(iban string) (IBAN, error) {
	iban = strings.ToUpper(strings.TrimSpace(iban))
	if !ibanRegex.MatchString(iban) {
		return IBAN{}, errors.New("IBAN inválido: deve começar com AO06 e ter 25 caracteres no total")
	}
	return IBAN{iban}, nil
}

func (i IBAN) String() string {
	return i.value
}

func MustNewIBAN(value string) IBAN {
	ret, err := NewIBAN(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return ret
}
