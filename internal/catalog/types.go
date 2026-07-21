package catalog

import (
	"fmt"
	"strconv"
	"strings"
)

type TypeTag int

const (
	TypeText    TypeTag = iota // texto libre
	TypeInteger               // enteros
	TypeDecimal               // punto flotante
	TypeBool                  // true/false
)

func (t TypeTag) String() string {
	switch t {
	case TypeText:
		return "TEXT"
	case TypeInteger:
		return "INTEGER"
	case TypeDecimal:
		return "DECIMAL"
	case TypeBool:
		return "BOOL"
	default:
		return "UNKNOWN"
	}
}

type Column struct {
	Name string
	Type TypeTag
}

type Schema struct {
	Columns []Column
	index   map[string]int // nombre -> posición
}

func NewSchema(columns []Column) Schema {
	idx := make(map[string]int, len(columns))
	for i, c := range columns {
		idx[strings.ToLower(c.Name)] = i
	}
	return Schema{Columns: columns, index: idx}
}

func (s Schema) HasColumn(name string) bool {
	_, ok := s.index[strings.ToLower(name)]
	return ok
}

func (s Schema) ColumnIndex(name string) (int, bool) {
	idx, ok := s.index[strings.ToLower(name)]
	return idx, ok
}

func (s Schema) ColumnByName(name string) (Column, bool) {
	idx, ok := s.index[strings.ToLower(name)]
	if !ok {
		return Column{}, false
	}
	return s.Columns[idx], true
}

type Row map[string]interface{}

func ParseValue(raw string, tag TypeTag) (interface{}, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" || strings.EqualFold(raw, "null") || strings.EqualFold(raw, "nil") {
		return nil, nil
	}

	switch tag {
	case TypeInteger:
		v, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot parse %q as INTEGER: %w", raw, err)
		}
		return v, nil

	case TypeDecimal:
		v, err := strconv.ParseFloat(raw, 64)
		if err != nil {
			return nil, fmt.Errorf("cannot parse %q as DECIMAL: %w", raw, err)
		}
		return v, nil

	case TypeBool:
		lower := strings.ToLower(raw)
		switch lower {
		case "true", "1", "yes":
			return true, nil
		case "false", "0", "no":
			return false, nil
		default:
			return nil, fmt.Errorf("cannot parse %q as BOOL", raw)
		}

	case TypeText:
		return raw, nil

	default:
		return raw, nil
	}
}

func InferType(values []string) TypeTag {
	boolCount, intCount, decimalCount := 0, 0, 0
	nonEmpty := 0

	for _, v := range values {
		v = strings.TrimSpace(v)
		if v == "" || strings.EqualFold(v, "null") {
			continue
		}
		nonEmpty++

		if _, err := strconv.ParseInt(v, 10, 64); err == nil {
			intCount++
			continue
		}
		if _, err := strconv.ParseFloat(v, 64); err == nil {
			decimalCount++
			continue
		}
		lower := strings.ToLower(v)
		if lower == "true" || lower == "false" || lower == "1" || lower == "0" || lower == "yes" || lower == "no" {
			boolCount++
			continue
		}
	}

	if nonEmpty == 0 {
		return TypeText
	}

	if intCount == nonEmpty {
		return TypeInteger
	}
	if decimalCount+intCount == nonEmpty {
		return TypeDecimal
	}
	if boolCount == nonEmpty {
		return TypeBool
	}

	return TypeText
}
