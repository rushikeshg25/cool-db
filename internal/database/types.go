package database

import (
	"strconv"
	"strings"
)

// DataType is the on-disk and query-facing type of a column value.
type DataType string

const (
	TypeInteger DataType = "INTEGER"
	TypeText    DataType = "TEXT"
	TypeBoolean DataType = "BOOLEAN"
	TypeFloat   DataType = "FLOAT"
)

func (d DataType) valid() bool {
	switch d {
	case TypeInteger, TypeText, TypeBoolean, TypeFloat:
		return true
	default:
		return false
	}
}

// Column describes a column stored in the database catalog.
type Column struct {
	Name       string   `json:"name"`
	Type       DataType `json:"type"`
	MaxLength  int      `json:"max_length,omitempty"`
	NotNull    bool     `json:"not_null,omitempty"`
	Unique     bool     `json:"unique,omitempty"`
	PrimaryKey bool     `json:"primary_key,omitempty"`
}

// Value stores one typed SQL value without losing integer precision during
// JSON persistence.
type Value struct {
	Type    DataType `json:"type,omitempty"`
	Null    bool     `json:"null,omitempty"`
	Integer int64    `json:"integer,omitempty"`
	Text    string   `json:"text,omitempty"`
	Boolean bool     `json:"boolean,omitempty"`
	Float   float64  `json:"float,omitempty"`
}

func (v Value) String() string {
	if v.Null {
		return "NULL"
	}

	switch v.Type {
	case TypeInteger:
		return strconv.FormatInt(v.Integer, 10)
	case TypeText:
		return v.Text
	case TypeBoolean:
		return strconv.FormatBool(v.Boolean)
	case TypeFloat:
		return strconv.FormatFloat(v.Float, 'g', -1, 64)
	default:
		return ""
	}
}

// valuesEqual implements SQL three-valued logic: a comparison involving NULL
// is UNKNOWN, which never satisfies a predicate. Use IS NULL to match NULLs.
func valuesEqual(left, right Value) bool {
	if left.Null || right.Null {
		return false
	}
	if left.Type != right.Type {
		return false
	}

	switch left.Type {
	case TypeInteger:
		return left.Integer == right.Integer
	case TypeText:
		return left.Text == right.Text
	case TypeBoolean:
		return left.Boolean == right.Boolean
	case TypeFloat:
		return left.Float == right.Float
	default:
		return false
	}
}

func coerceLiteral(literal sqlLiteral, column Column) (Value, error) {
	if literal.kind == literalNull {
		if column.NotNull || column.PrimaryKey {
			return Value{}, newError(CodeConstraint, "column %q cannot be NULL", column.Name)
		}
		return Value{Type: column.Type, Null: true}, nil
	}

	value := Value{Type: column.Type}
	switch column.Type {
	case TypeInteger:
		if literal.kind != literalNumber || strings.ContainsAny(literal.value, ".eE") {
			return Value{}, newError(CodeType, "column %q expects INTEGER", column.Name)
		}
		integer, err := strconv.ParseInt(literal.value, 10, 64)
		if err != nil {
			return Value{}, newError(CodeType, "invalid INTEGER value %q", literal.value)
		}
		value.Integer = integer
	case TypeText:
		if literal.kind != literalString {
			return Value{}, newError(CodeType, "column %q expects a quoted string", column.Name)
		}
		if column.MaxLength > 0 && len([]rune(literal.value)) > column.MaxLength {
			return Value{}, newError(CodeConstraint, "value for column %q exceeds VARCHAR(%d)", column.Name, column.MaxLength)
		}
		value.Text = literal.value
	case TypeBoolean:
		if literal.kind != literalBoolean {
			return Value{}, newError(CodeType, "column %q expects BOOLEAN", column.Name)
		}
		value.Boolean = strings.EqualFold(literal.value, "true")
	case TypeFloat:
		if literal.kind != literalNumber {
			return Value{}, newError(CodeType, "column %q expects FLOAT", column.Name)
		}
		floatValue, err := strconv.ParseFloat(literal.value, 64)
		if err != nil {
			return Value{}, newError(CodeType, "invalid FLOAT value %q", literal.value)
		}
		value.Float = floatValue
	default:
		return Value{}, newError(CodeStorage, "column %q has unsupported stored type %q", column.Name, column.Type)
	}

	return value, nil
}
