package vos

import (
	"errors"
	"strings"
)

type Address struct {
	value string
}

func NewAddress(input string) (Address, error) {
	val := strings.TrimSpace(input)
	if len(val) < 5 {
		return Address{}, errors.New("endereço muito curto")
	}
	if len(val) > 150 {
		return Address{}, errors.New("endereço muito longo (máx. 150 caracteres)")
	}
	return Address{val}, nil
}

func (a Address) String() string {
	return a.value
}

func MustNewAddress(value string) Address {
	addr, err := NewAddress(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return addr
}
