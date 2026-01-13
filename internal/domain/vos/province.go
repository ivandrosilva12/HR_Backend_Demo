package vos

import (
	"strings"
)

type Province struct {
	value string
}

func NewProvince(input string) Province {
	nameTrimmed := strings.TrimSpace(strings.Title(strings.ToLower(input)))
	return Province{nameTrimmed}
}

func (p Province) String() string {
	return p.value
}
