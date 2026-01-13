package vos

import (
	"errors"
	"strings"
)

// Gender - VO para gênero

type Gender struct {
	value string
}

var allowedGenders = map[string]bool{
	"masculino": true,
	"feminino":  true,
}

func NewGender(genderStr string) (Gender, error) {
	genderTrimmed := strings.ToLower(strings.TrimSpace(genderStr))

	if !allowedGenders[genderTrimmed] {
		return Gender{}, errors.New("gênero deve ser masculino ou feminino")
	}

	return Gender{genderTrimmed}, nil
}

func (f Gender) String() string {
	return string(f.value)
}

func MustNewGender(value string) Gender {
	ret, err := NewGender(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return ret
}
