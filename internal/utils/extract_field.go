package utils

import (
	"reflect"
	"strings"
	"time"
)

// ExtractFieldString tenta extrair o valor string de um campo de um struct
// Retorna string vazia caso o campo não exista ou não seja string.
func ExtractFieldString(v any, field string) string {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return ""
	}

	f := rv.FieldByName(field)
	if !f.IsValid() || f.Kind() != reflect.String {
		return ""
	}

	return f.String()
}

// ExtractFieldTime tenta extrair o valor time.Time de um campo de um struct
// Retorna false caso o campo não exista ou não seja time.Time.
func ExtractFieldTime(v any, field string) (time.Time, bool) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Pointer {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Struct {
		return time.Time{}, false
	}

	f := rv.FieldByName(field)
	if !f.IsValid() || f.Type() != reflect.TypeOf(time.Time{}) {
		return time.Time{}, false
	}

	return f.Interface().(time.Time), true
}

// NotEmptyString verifica se a string não está vazia e não contém apenas espaços
func NotEmptyString(s string) bool {
	return strings.TrimSpace(s) != ""
}
