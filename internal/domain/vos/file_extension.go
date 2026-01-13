package vos

import (
	"errors"
	"strings"
)

type FileExtension struct {
	value string
}

var allowedExtensions = map[string]bool{
	"pdf":  true,
	"jpg":  true,
	"jpeg": true,
	"png":  true,
	"docx": true,
	"doc":  true,
}

func NewFileExtension(ext string) (FileExtension, error) {
	e := strings.ToLower(strings.TrimSpace(ext))
	if !allowedExtensions[e] {
		return FileExtension{}, errors.New("extensão de ficheiro não suportada")
	}
	return FileExtension{e}, nil
}

func (e FileExtension) String() string {
	return string(e.value)
}

func MustNewFileExtension(value string) FileExtension {
	ret, err := NewFileExtension(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return ret
}
