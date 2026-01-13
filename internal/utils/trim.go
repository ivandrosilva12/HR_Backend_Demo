package utils

import (
	"reflect"
	"strings"
	"time"
)

// TrimStrings percorre recursivamente o DTO e aplica strings.TrimSpace
// em todos os campos de tipo string (incluindo aliases).
// Aceita ponteiro para struct (ou ponteiro para slice/map/array de structs).
func TrimStrings(dto any) {
	if dto == nil {
		return
	}
	v := reflect.ValueOf(dto)
	if v.Kind() != reflect.Ptr || v.IsNil() {
		// Requer ponteiro settable para modificar in-place.
		return
	}
	trimValue(v)
}

var timeType = reflect.TypeOf(time.Time{})

func trimValue(v reflect.Value) {
	// Desreferencia ponteiros
	for v.Kind() == reflect.Ptr {
		if v.IsNil() {
			return
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString(strings.TrimSpace(v.String()))

	case reflect.Struct:
		// Ignora time.Time
		if v.Type() == timeType {
			return
		}
		for i := 0; i < v.NumField(); i++ {
			f := v.Field(i)
			// Apenas campos exportados são settable fora do pacote
			if !f.CanSet() {
				// Se for ponteiro/exportado, tentamos descer mesmo assim
				if f.Kind() == reflect.Ptr && !f.IsNil() {
					trimValue(f)
				}
				continue
			}
			trimValue(f.Addr())
		}

	case reflect.Slice, reflect.Array:
		for i := 0; i < v.Len(); i++ {
			elem := v.Index(i)
			// Para elementos não-endereçáveis (ex.: array), trabalhamos numa cópia settable
			if elem.CanAddr() {
				trimValue(elem.Addr())
			} else {
				tmp := reflect.New(elem.Type()).Elem()
				tmp.Set(elem)
				trimValue(tmp.Addr())
				if v.Kind() == reflect.Slice {
					v.Index(i).Set(tmp)
				}
			}
		}

	case reflect.Map:
		// Reconstrói o mapa quando a chave for string (ou alias)
		keyKind := v.Type().Key().Kind()
		newMap := reflect.MakeMapWithSize(v.Type(), v.Len())
		iter := v.MapRange()
		for iter.Next() {
			k := iter.Key()
			val := iter.Value()

			newK := reflect.New(k.Type()).Elem()
			newK.Set(k)

			// Trima a chave se for string-like
			if keyKind == reflect.String {
				newK.SetString(strings.TrimSpace(newK.Convert(reflect.TypeOf("")).String()))
			}

			// Trima o valor recursivamente
			newV := reflect.New(val.Type()).Elem()
			newV.Set(val)
			trimValue(newV.Addr())

			newMap.SetMapIndex(newK, newV)
		}
		v.Set(newMap)

	case reflect.Interface:
		if v.IsNil() {
			return
		}
		inner := v.Elem()
		tmp := reflect.New(inner.Type()).Elem()
		tmp.Set(inner)
		trimValue(tmp.Addr())
		v.Set(tmp)

	default:
		// outros tipos (int, bool, etc.): nada a fazer
	}
}
