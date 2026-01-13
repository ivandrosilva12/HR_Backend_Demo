package vos

import (
	"errors"
	"regexp"
)

type Email struct {
	value string
}

func NewEmail(email string) (Email, error) {
	if !regexp.MustCompile(`^[\w\.-]+@[\w\.-]+\.\w+$`).MatchString(email) {
		return Email{}, errors.New("email inválido")
	}
	return Email{email}, nil
}

func (e Email) String() string {
	return e.value
}

func MustNewEmail(value string) Email {
	ret, err := NewEmail(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return ret
}
