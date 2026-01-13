package vos

import (
	"errors"
	"strings"
)

// RelationshipType - VO para tipo de relação com o dependente

type RelationshipType struct {
	value string
}

var allowedRelationships = map[string]bool{
	"filho":    true,
	"filha":    true,
	"cônjuge":  true,
	"pai":      true,
	"mãe":      true,
	"sobrinho": true,
	"sobrinha": true,
	"irmão":    true,
	"irmã":     true,
	"tio":      true,
	"tia":      true,
	"avó":      true,
	"avô":      true,
}

func NewRelationshipType(ext string) (RelationshipType, error) {
	e := strings.ToLower(strings.TrimSpace(ext))
	if !allowedRelationships[e] {
		return RelationshipType{}, errors.New("relacionamento inválido: deve ser Filho, Cônjuge, Pai, etc")
	}
	return RelationshipType{e}, nil
}

func (f RelationshipType) String() string {
	return f.value
}

func MustNewRelationshipType(value string) RelationshipType {
	ret, err := NewRelationshipType(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return ret
}
