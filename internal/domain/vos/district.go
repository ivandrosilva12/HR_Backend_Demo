package vos

import (
	"errors"
	"strings"
)

type District struct {
	value string
}

func NewDistrict(input string) (District, error) {
	val := strings.TrimSpace(strings.Title((strings.ToLower(input))))
	if len(val) < 3 {
		return District{}, errors.New("nome de distrito (ou comuna) muito curto")
	}
	if len(val) > 60 {
		return District{}, errors.New("nome de distrito (ou comuna) muito longo")
	}
	return District{val}, nil
}

func (d District) String() string {
	return d.value
}

func MustNewDistrict(value string) District {
	ret, err := NewDistrict(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return ret
}
