package vos

import (
	"errors"
	"strings"
)

type ContractType string

var validContracts = map[string]bool{
	"definitivo": true,
	"temporário": true,
	"interno":    true,
	"consultor":  true,
}

func NewContractType(ct string) (ContractType, error) {
	ct = strings.TrimSpace(ct)
	if !validContracts[ct] {
		return "", errors.New("tipo de contrato inválido")
	}
	return ContractType(ct), nil
}

func (c ContractType) String() string {
	return string(c)
}

func MustNewContractType(value string) ContractType {
	ret, err := NewContractType(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return ret
}
