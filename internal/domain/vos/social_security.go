package vos

import (
	"errors"
	"regexp"
	"strings"
)

type SocialSecurity struct {
	value string
}

var ssRegex = regexp.MustCompile(`^\d{6,12}$`)

func NewSocialSecurity(input string) (SocialSecurity, error) {
	val := strings.TrimSpace(input)
	if !ssRegex.MatchString(val) {
		return SocialSecurity{}, errors.New("número da segurança social inválido")
	}
	return SocialSecurity{val}, nil
}

func (s SocialSecurity) String() string {
	return s.value
}

func MustNewSocialSecurity(value string) SocialSecurity {
	ss, err := NewSocialSecurity(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return ss
}
