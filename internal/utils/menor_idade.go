package utils

import "time"

func MenorDeIdade(birthDate time.Time) bool {
	now := time.Now()

	// Calculate years
	years := now.Year() - birthDate.Year()

	// Adjust for month and day
	if now.Month() < birthDate.Month() {
		years--
	} else if now.Month() == birthDate.Month() {
		if now.Day() < birthDate.Day() {
			years--
		} else if now.Day() == birthDate.Day() {
			// Check time on the same day
			if now.Before(birthDate.AddDate(years, 0, 0)) {
				years--
			}
		}
	}

	return years < 18
}

func ContratacaoMenorDeIdade(dataContratacao, birthDate time.Time) bool {

	// Calculate years
	years := dataContratacao.Year() - birthDate.Year()

	// Adjust for month and day
	if dataContratacao.Month() < birthDate.Month() {
		years--
	} else if dataContratacao.Month() == birthDate.Month() {
		if dataContratacao.Day() < birthDate.Day() {
			years--
		} else if dataContratacao.Day() == birthDate.Day() {
			// Check time on the same day
			if dataContratacao.Before(birthDate.AddDate(years, 0, 0)) {
				years--
			}
		}
	}

	return years < 18
}
