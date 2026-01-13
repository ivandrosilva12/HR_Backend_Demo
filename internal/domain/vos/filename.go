package vos

import (
	"errors"
	"regexp"
	"strings"
)

type Filename struct {
	value string
}

func NewFilename(value string) (Filename, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return Filename{}, errors.New("nome do ficheiro é obrigatório")
	}
	if len(value) > 255 {
		return Filename{}, errors.New("nome do ficheiro é muito longo")
	}
	if !regexp.MustCompile(`^[\w\-. ]+$`).MatchString(value) {
		return Filename{}, errors.New("nome do ficheiro inválido (caracteres não permitidos)")
	}
	return Filename{value}, nil
}

func (f Filename) String() string {
	return string(f.value)
}

func MustNewFilename(value string) Filename {
	ret, err := NewFilename(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return ret
}
