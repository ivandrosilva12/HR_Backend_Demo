package vos

import (
	"errors"
	"net/url"
	"strings"
)

type DocumentURL struct {
	value string
}

func NewDocumentURL(value string) (DocumentURL, error) {
	trimmed := strings.TrimSpace(value)
	u, err := url.ParseRequestURI(trimmed)
	if err != nil {
		return DocumentURL{}, errors.New("URL de documento inválida")
	}
	return DocumentURL{u.String()}, nil
}

func (d DocumentURL) String() string {
	return string(d.value)
}

func MustNewDocumentURL(value string) DocumentURL {
	ret, err := NewDocumentURL(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return ret
}
