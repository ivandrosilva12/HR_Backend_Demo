package vos

import (
	"errors"
	"time"
)

// BirthDate - VO para datas de nascimento

type BirthDate struct {
	value time.Time
}

func NewBirthDate(dateStr string) (BirthDate, error) {
	t, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return BirthDate{}, errors.New("data de nascimento inválida. Use o formato YYYY-MM-DD")
	}
	if t.After(time.Now()) {
		return BirthDate{}, errors.New("data de nascimento não pode ser futura")
	}
	return BirthDate{value: t}, nil
}

func (b BirthDate) String() string {
	return b.value.Format("2006-01-02")
}

func (b BirthDate) Time() time.Time {
	return b.value
}

func MustNewBirthDate(value string) BirthDate {
	ret, err := NewBirthDate(value)
	if err != nil {
		panic(err) // ou retorne erro padronizado
	}
	return ret
}
