package utils

import (
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgconn"
	"github.com/jackc/pgx"
	"github.com/lib/pq"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type FieldError struct {
	Field   string      `json:"field"`
	Label   string      `json:"label"`
	Tag     string      `json:"tag"`
	Message string      `json:"message"`
	Value   interface{} `json:"value,omitempty"`
}

type Payload struct {
	Error   string       `json:"error"`             // ex: validation_failed, conflict, internal
	Message string       `json:"message,omitempty"` // resumo humanizado
	Fields  []FieldError `json:"fields,omitempty"`  // detalhes por campo
}

var (
	ErrMenorDeIdade = errors.New("funcionário deve ter pelo menos 18 anos")
)

var v = validator.New()

func init() {
	// Mostra o nome do json como "campo" quando label não existir
	v.RegisterTagNameFunc(func(fld reflect.StructField) string {
		tag := fld.Tag.Get("json")
		if tag == "" || tag == "-" {
			return fld.Name
		}
		return strings.Split(tag, ",")[0]
	})
}

// fieldLabelFromDTO prefers `label:"..."`, then `json:"..."`, otherwise struct field name.
func fieldLabelFromDTO(dto any, fe validator.FieldError) string {
	if dto == nil {
		return fe.Field()
	}
	t := reflect.TypeOf(dto)
	// Handle pointer/slice/map wrappers
	for t.Kind() == reflect.Ptr || t.Kind() == reflect.Slice || t.Kind() == reflect.Array || t.Kind() == reflect.Map {
		t = t.Elem()
		if t == nil {
			return fe.Field()
		}
	}
	if t.Kind() != reflect.Struct {
		return fe.Field()
	}

	// Use the StructField by walking the namespace (handles nested structs)
	sf := structFieldByNamespace(t, fe.StructNamespace())
	if sf == nil {
		// Fallback to direct Field lookup
		if f, ok := t.FieldByName(fe.StructField()); ok {
			sf = &f
		}
	}
	if sf == nil {
		return fe.Field()
	}

	// label has priority
	if label := sf.Tag.Get("label"); label != "" {
		return label
	}

	// json tag (strip ",omitempty")
	if jsonTag := sf.Tag.Get("json"); jsonTag != "" && jsonTag != "-" {
		if idx := strings.IndexByte(jsonTag, ','); idx >= 0 {
			return jsonTag[:idx]
		}
		return jsonTag
	}

	return sf.Name
}

// structFieldByNamespace finds the StructField for a nested namespace like "CreateProvinceDTO.Nome".
func structFieldByNamespace(root reflect.Type, ns string) *reflect.StructField {
	if ns == "" {
		return nil
	}
	parts := strings.Split(ns, ".")
	// Drop the root type name if present
	if len(parts) > 0 && parts[0] == root.Name() {
		parts = parts[1:]
	}
	current := root
	var sf reflect.StructField
	for _, p := range parts {
		if current.Kind() != reflect.Struct {
			return nil
		}
		f, ok := current.FieldByName(p)
		if !ok {
			return nil
		}
		sf = f
		current = f.Type
		// Deref pointers
		for current.Kind() == reflect.Ptr {
			current = current.Elem()
		}
	}
	return &sf
}

// buildTagMessage creates friendly messages for common tags.
func buildTagMessage(label string, fe validator.FieldError) string {
	param := fe.Param()
	switch fe.Tag() {
	case "required":
		return fmt.Sprintf("%s é obrigatório.", label)
	case "min":
		return fmt.Sprintf("%s deve ter no mínimo %s caracteres.", label, param)
	case "max":
		return fmt.Sprintf("%s deve ter no máximo %s caracteres.", label, param)
	case "len":
		return fmt.Sprintf("%s deve ter exatamente %s caracteres.", label, param)
	case "email":
		return fmt.Sprintf("%s deve ser um e-mail válido.", label)
	case "oneof":
		return fmt.Sprintf("%s deve ser um dos valores: %s.", label, param)
	case "uuid", "uuid4":
		return fmt.Sprintf("%s deve ser um UUID válido.", label)
	case "gte":
		return fmt.Sprintf("%s deve ser maior ou igual a %s.", label, param)
	case "lte":
		return fmt.Sprintf("%s deve ser menor ou igual a %s.", label, param)
	case "datetime":
		// e.g., datetime=2006-01-02
		if param != "" {
			return fmt.Sprintf("%s deve estar no formato de data/hora: %s.", label, param)
		}
		return fmt.Sprintf("%s deve estar em um formato de data/hora válido.", label)
	default:
		// Fallback: show tag for debugging without leaking internals
		if param != "" {
			return fmt.Sprintf("%s é inválido (regra: %s=%s).", label, fe.Tag(), param)
		}
		return fmt.Sprintf("%s é inválido (regra: %s).", label, fe.Tag())
	}
}

// HumanizeValidation converts validator errors into friendly, labeled messages.
func HumanizeValidation(dto any, err error) []FieldError {
	if err == nil {
		return nil
	}
	var verrs validator.ValidationErrors
	if !errors.As(err, &verrs) {
		// Not a validation error → return nil to let caller handle generic errors
		return nil
	}

	fields := make([]FieldError, 0, len(verrs))
	for _, fe := range verrs {
		lbl := fieldLabelFromDTO(dto, fe)
		fields = append(fields, FieldError{
			Field:   fe.Field(),
			Label:   lbl,
			Tag:     fe.Tag(),
			Message: buildTagMessage(lbl, fe),
			Value:   safeValue(fe.Value()),
		})
	}
	return fields
}

func HumanizeDB(err error) (ok bool, payload Payload, status int) {
	caser := cases.Title(language.Portuguese)

	makeField := func(col, tag, msg string, val interface{}) FieldError {
		if col == "" {
			col = "geral"
		}
		return FieldError{
			Field:   caser.String(col),
			Label:   strings.ToLower(col),
			Tag:     tag,
			Message: msg,
			Value:   safeValue(val),
		}
	}

	// Normaliza resposta "400 Validação dos campos"
	asValidation := func(field FieldError) (bool, Payload, int) {
		return true, Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields:  []FieldError{field},
		}, http.StatusBadRequest
	}

	// Conflitos de transação (retry)
	asTxnConflict := func(msg string) (bool, Payload, int) {
		return true, Payload{
			Error:   "Conflito de transação",
			Message: msg,
			Fields:  []FieldError{},
		}, http.StatusConflict
	}

	// Erro interno/infra
	asDBFailure := func(msg string, code int) (bool, Payload, int) {
		return true, Payload{
			Error:   "Erros na DB",
			Message: msg,
			Fields:  []FieldError{},
		}, code
	}

	// ---------------------------
	// NOT FOUND (no rows)
	// ---------------------------
	// Trata os casos em que a consulta não retornou linhas.
	// Mantém o padrão de erro (400) que você usa nos handlers.
	if errors.Is(err, sql.ErrNoRows) || errors.Is(err, pgx.ErrNoRows) {
		return true, Payload{
			Error:   "Não foram encontrados resultados",
			Message: "Dados inválidos. Corrija os campos de entrada.",
			Fields:  []FieldError{}, // sem campo específico aqui
		}, http.StatusBadRequest
	}

	// ---- Coleta metadados comuns
	var (
		sqlState   string
		detail     string
		constraint string
		column     string
		msg        string
	)

	// --- pgx / pgconn ---
	var pgxErr *pgconn.PgError
	if errors.As(err, &pgxErr) {
		sqlState = pgxErr.Code
		detail = pgxErr.Detail
		constraint = pgxErr.ConstraintName
		column = pgxErr.ColumnName
		msg = pgxErr.Message
	} else {
		// --- lib/pq ---
		var pqErr *pq.Error
		if errors.As(err, &pqErr) {
			sqlState = string(pqErr.Code)
			detail = pqErr.Detail
			constraint = pqErr.Constraint
			column = pqErr.Column
			msg = pqErr.Message
		} else {
			return false, Payload{}, 0
		}
	}

	switch sqlState {
	case "23505": // unique_violation
		col, val := parseDupFromDetail(detail, constraint)
		if col == "" {
			col = column
		}
		return asValidation(makeField(col, "unique", "valor já existe.", val))

	case "23503": // foreign_key_violation
		col, val := parseDupFromDetail(detail, constraint)
		if col == "" {
			col = column
		}
		d := strings.ToLower(detail)
		if strings.Contains(d, "is not present") {
			return asValidation(makeField(col, "reference", "referência inexistente.", val))
		}
		if strings.Contains(d, "is still referenced") {
			return asValidation(makeField(col, "fk_constraint", "registro está referenciado por outros dados.", val))
		}
		return asValidation(makeField(col, "fk", "violação de chave estrangeira.", val))

	case "23502": // not_null_violation
		col := column
		if col == "" {
			col, _ = parseDupFromDetail(detail, constraint)
		}
		return asValidation(makeField(col, "required", "é obrigatório.", nil))

	case "23514": // check_violation
		col, _ := parseDupFromDetail(detail, constraint)
		return asValidation(makeField(col, "check", "violação de regra de negócio.", nil))

	case "22001": // string_data_right_truncation
		col, val := parseDupFromDetail(detail, constraint)
		if col == "" {
			col = column
		}
		return asValidation(makeField(col, "max", "excede o tamanho máximo permitido.", val))

	case "22003": // numeric_value_out_of_range
		col, val := parseDupFromDetail(detail, constraint)
		if col == "" {
			col = column
		}
		return asValidation(makeField(col, "range", "valor numérico fora do intervalo.", val))

	case "22P02": // invalid_text_representation (ex.: UUID inválido)
		col, val := parseDupFromDetail(detail, constraint)
		if col == "" {
			col = column
		}
		m := "formato inválido."
		if strings.Contains(strings.ToLower(msg+detail), "uuid") {
			m = "UUID inválido."
		}
		return asValidation(makeField(col, "format", m, val))

	case "22007": // invalid_datetime_format
		col, val := parseDupFromDetail(detail, constraint)
		if col == "" {
			col = column
		}
		return asValidation(makeField(col, "datetime", "formato de data/hora inválido.", val))

	case "22008": // datetime_field_overflow
		col, val := parseDupFromDetail(detail, constraint)
		if col == "" {
			col = column
		}
		return asValidation(makeField(col, "datetime", "data/hora fora do intervalo válido.", val))

	case "42804": // datatype_mismatch
		col, val := parseDupFromDetail(detail, constraint)
		if col == "" {
			col = column
		}
		return asValidation(makeField(col, "type", "tipo de dado incompatível.", val))

	// ---------------------------
	// CONCORRÊNCIA / TRANSACÕES
	// ---------------------------
	case "40001": // serialization_failure
		return asTxnConflict("Conflito de concorrência detectado. Tente novamente.")
	case "40P01": // deadlock_detected
		return asTxnConflict("Deadlock detectado. Tente novamente.")
	case "55P03": // lock_not_available
		return asTxnConflict("Recurso bloqueado no momento. Tente novamente.")
	case "25P02": // in_failed_sql_transaction
		return asTxnConflict("Transação anterior falhou. Refaça a operação.")

	// ---------------------------
	// PERMISSÕES
	// ---------------------------
	case "42501": // insufficient_privilege
		return true, Payload{
			Error:   "Acesso negado",
			Message: "Permissões insuficientes para executar esta operação.",
			Fields:  []FieldError{},
		}, http.StatusForbidden

	// ---------------------------
	// SINTAXE/OBJETOS (server-side)
	// ---------------------------
	case "42601", // syntax_error
		"42703", // undefined_column
		"42P01": // undefined_table
		return asDBFailure("Falha ao processar a consulta no servidor.", http.StatusInternalServerError)

	// ---------------------------
	// INFRA/RECURSOS
	// ---------------------------
	case "53300": // too_many_connections
		return asDBFailure("Muitas conexões ativas. Tente novamente mais tarde.", http.StatusServiceUnavailable)
	case "53200": // out_of_memory
		return asDBFailure("Recursos insuficientes no servidor de BD (memória).", http.StatusServiceUnavailable)
	case "53100": // disk_full
		return asDBFailure("Armazenamento insuficiente no servidor de BD.", http.StatusInsufficientStorage)
	case "53400": // configuration_limit_exceeded
		return asDBFailure("Limite de configuração do servidor de BD excedido.", http.StatusServiceUnavailable)

	// CONNECTION EXCEPTIONS (família 08xxx)
	case "08000", "08001", "08003", "08006":
		return asDBFailure("Falha de conexão com a base de dados.", http.StatusServiceUnavailable)

	// Opcional: SQLSTATE padrão para "no data" (raramente exposto)
	case "02000": // NO_DATA
		return true, Payload{
			Error:   "Validação dos campos",
			Message: "Dados inválidos. Corrija os campos destacados.",
			Fields:  []FieldError{},
		}, http.StatusBadRequest
	}

	// Fallback genérico
	return asDBFailure("Falha ao processar sua solicitação.", http.StatusInternalServerError)
}

// DETAIL examples:
// "Key (name)=(Luanda) already exists."
// "Key (tenant_id, email)=(t1, foo@bar.com) already exists."
// "Key (department_id)=(d1) is not present in table \"departments\"."
// "Key (id)=(x) is still referenced from table \"employees\"."
func parseDupFromDetail(detail, constraint string) (field, value string) {
	if detail != "" {
		start := strings.Index(detail, "Key (")
		if start != -1 {
			start += len("Key (")
			mid := strings.Index(detail[start:], ")=")
			if mid != -1 {
				cols := strings.TrimSpace(detail[start : start+mid])
				vStart := strings.Index(detail, "=(")
				if vStart != -1 {
					vStart += len("=(")
					if vEnd := strings.Index(detail[vStart:], ")"); vEnd != -1 {
						value = strings.TrimSpace(detail[vStart : vStart+vEnd])
					}
				}
				if strings.Contains(cols, ",") {
					field = strings.TrimSpace(strings.Split(cols, ",")[0])
				} else {
					field = cols
				}
				return
			}
		}
	}
	// Fallback: infer from constraint e.g. "provinces_name_key"
	if constraint != "" {
		parts := strings.Split(constraint, "_")
		if len(parts) >= 2 {
			field = parts[len(parts)-2]
		}
	}
	return
}

// safeValue trims long values to avoid noisy payloads/logs.
func safeValue(v interface{}) interface{} {
	const max = 80
	switch x := v.(type) {
	case string:
		if len(x) > max {
			return x[:max] + "…"
		}
		return x
	}
	return v
}
