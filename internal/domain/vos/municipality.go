package vos

import (
	"strings"
)

type Municipality struct {
	value string
}

func NewMunicipality(input string) Municipality {
	val := strings.TrimSpace(strings.Title(strings.ToLower(input)))
	return Municipality{val}
}

func (m Municipality) String() string {
	return m.value
}

func MustNewMunicipality(value string) Municipality {
	return NewMunicipality(value)
}
