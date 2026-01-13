package vos

import (
	"errors"
	"strings"
)

type StatusReason struct {
	value string
}

// Lista de motivos permitidos (opcional, ou apenas validação por tamanho e conteúdo)
var allowedReasons = map[string]bool{
	"processo disciplinar": true,
	"licença médica":       true,
	"aposentadoria":        true,
	"licença maternidade":  true,
	"licença paternidade":  true,
	"falecimento familiar": true,
	"outros":               true,
}

func NewStatusReason(input string) (StatusReason, error) {

	normalized := strings.ToLower(strings.TrimSpace(input))
	if !allowedReasons[normalized] {
		return StatusReason{}, errors.New("motivo inválido")
	}

	if len(normalized) < 3 {
		return StatusReason{}, errors.New("o motivo deve ter ao menos 3 caracteres")
	}

	if len(normalized) > 100 {
		return StatusReason{}, errors.New("o motivo não pode exceder 100 caracteres")
	}

	return StatusReason{normalized}, nil

}

func (r StatusReason) String() string {
	return r.value
}

func MustNewStatusReason(value string) StatusReason {
	ret, err := NewStatusReason(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return ret
}
