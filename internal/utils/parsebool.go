package utils

import (
	"strconv"
	"strings"
)

// ParseBool interpreta strings comuns como booleanos.
// True:  "1", "t", "true", "y", "yes", "on"   (case-insensitive)
// False: "0", "f", "false", "n", "no", "off", "" (vazio)
// Para qualquer outro valor inválido, devolve false.
func ParseBool(s string) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return false
	}
	// cobre 1/0, t/f, true/false, etc.
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	switch strings.ToLower(s) {
	case "y", "yes", "on":
		return true
	case "n", "no", "off":
		return false
	default:
		return false
	}
}

// (Opcional) Variante com valor por defeito customizável.
// Útil se quiseres usar "não reconhecido => default".
func ParseBoolDefault(s string, def bool) bool {
	s = strings.TrimSpace(s)
	if s == "" {
		return def
	}
	if b, err := strconv.ParseBool(s); err == nil {
		return b
	}
	switch strings.ToLower(s) {
	case "y", "yes", "on":
		return true
	case "n", "no", "off":
		return false
	default:
		return def
	}
}
