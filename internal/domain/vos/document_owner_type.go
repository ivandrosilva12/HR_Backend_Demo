package vos

import (
	"errors"
)

type DocumentOwnerType string

var ErrInvalidOwnerType = errors.New("invalid document owner type")

const (
	DocumentOwnerEmployee  DocumentOwnerType = "employee"
	DocumentOwnerDependent DocumentOwnerType = "dependent"
)

func (t DocumentOwnerType) IsValid() bool {
	switch t {
	case DocumentOwnerEmployee, DocumentOwnerDependent:
		return true
	}
	return false
}

func (d DocumentOwnerType) String() string {
	return string(d)
}
