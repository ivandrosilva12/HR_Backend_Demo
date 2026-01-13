package vos

import (
	"errors"
	"strings"
)

type DocumentType struct {
	value string
}

var allowedTypes = map[string]bool{
	"BI":       true,
	"Contrato": true,
	"Diploma":  true,
	"Foto":     true,
	"Outro":    true,
}

func NewDocumentType(value string) (DocumentType, error) {
	v := strings.TrimSpace(value)
	if !allowedTypes[v] {
		return DocumentType{}, errors.New("tipo de documento inválido")
	}
	return DocumentType{v}, nil
}

func (d DocumentType) String() string {
	return d.value
}

func MustNewDocumentType(value string) DocumentType {
	ret, err := NewDocumentType(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return ret
}
