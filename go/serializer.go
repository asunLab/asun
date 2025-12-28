package ason

import (
	"strconv"
	"strings"
	"unicode"
)

// Serialize converts a Value to its ASON string representation.
func Serialize(v *Value) string {
	if v == nil {
		return "null"
	}

	var sb strings.Builder
	serializeValue(&sb, v)
	return sb.String()
}

func serializeValue(sb *strings.Builder, v *Value) {
	if v == nil {
		sb.WriteString("null")
		return
	}

	switch v.typ {
	case TypeNull:
		sb.WriteString("null")
	case TypeBool:
		if v.boolVal {
			sb.WriteString("true")
		} else {
			sb.WriteString("false")
		}
	case TypeInteger:
		sb.WriteString(strconv.FormatInt(v.intVal, 10))
	case TypeFloat:
		sb.WriteString(strconv.FormatFloat(v.floatVal, 'g', -1, 64))
	case TypeString:
		serializeString(sb, v.strVal)
	case TypeArray:
		sb.WriteRune('[')
		for i, item := range v.arrVal {
			if i > 0 {
				sb.WriteRune(',')
			}
			serializeValue(sb, item)
		}
		sb.WriteRune(']')
	case TypeObject:
		sb.WriteRune('(')
		for i, key := range v.objKeys {
			if i > 0 {
				sb.WriteRune(',')
			}
			serializeValue(sb, v.objVals[key])
		}
		sb.WriteRune(')')
	}
}

func serializeString(sb *strings.Builder, s string) {
	// Check if quoting is needed
	needsQuote := false
	if len(s) == 0 {
		needsQuote = true
	} else {
		for _, r := range s {
			if r == '"' || r == '\\' || r == '(' || r == ')' ||
				r == '[' || r == ']' || r == '{' || r == '}' ||
				r == ',' || r == ':' || unicode.IsSpace(r) {
				needsQuote = true
				break
			}
		}
		// Check if it looks like a keyword or number
		if !needsQuote {
			if s == "null" || s == "true" || s == "false" {
				needsQuote = true
			} else if len(s) > 0 && (s[0] == '-' || s[0] == '+' || (s[0] >= '0' && s[0] <= '9')) {
				needsQuote = true
			}
		}
	}

	if !needsQuote {
		sb.WriteString(s)
		return
	}

	sb.WriteRune('"')
	for _, r := range s {
		switch r {
		case '"':
			sb.WriteString("\\\"")
		case '\\':
			sb.WriteString("\\\\")
		case '\n':
			sb.WriteString("\\n")
		case '\r':
			sb.WriteString("\\r")
		case '\t':
			sb.WriteString("\\t")
		default:
			if r < 0x20 {
				sb.WriteString("\\u")
				sb.WriteString(strconv.FormatInt(int64(r), 16))
			} else {
				sb.WriteRune(r)
			}
		}
	}
	sb.WriteRune('"')
}

// SerializeWithSchema serializes a value with its schema.
func SerializeWithSchema(v *Value) string {
	if v == nil {
		return "null"
	}
	if !v.IsObject() && !v.IsArray() {
		return Serialize(v)
	}

	var sb strings.Builder

	// Handle array of objects
	if v.IsArray() && v.Len() > 0 && v.Get(0).IsObject() {
		// Build schema from first object
		first := v.Get(0)
		sb.WriteRune('{')
		for i, key := range first.Keys() {
			if i > 0 {
				sb.WriteRune(',')
			}
			sb.WriteString(key)
			writeFieldSchema(&sb, first.Field(key))
		}
		sb.WriteRune('}')
		sb.WriteRune(':')

		// Serialize all objects
		for i := 0; i < v.Len(); i++ {
			if i > 0 {
				sb.WriteRune(',')
			}
			serializeObjectData(&sb, v.Get(i))
		}
		return sb.String()
	}

	// Single object
	if v.IsObject() {
		sb.WriteRune('{')
		for i, key := range v.Keys() {
			if i > 0 {
				sb.WriteRune(',')
			}
			sb.WriteString(key)
			writeFieldSchema(&sb, v.Field(key))
		}
		sb.WriteRune('}')
		sb.WriteRune(':')
		serializeObjectData(&sb, v)
		return sb.String()
	}

	return Serialize(v)
}

func writeFieldSchema(sb *strings.Builder, v *Value) {
	if v == nil {
		return
	}
	if v.IsArray() {
		sb.WriteString("[]")
		if v.Len() > 0 && v.Get(0).IsObject() {
			sb.WriteRune('{')
			first := v.Get(0)
			for i, key := range first.Keys() {
				if i > 0 {
					sb.WriteRune(',')
				}
				sb.WriteString(key)
				writeFieldSchema(sb, first.Field(key))
			}
			sb.WriteRune('}')
		}
	} else if v.IsObject() {
		sb.WriteRune('{')
		for i, key := range v.Keys() {
			if i > 0 {
				sb.WriteRune(',')
			}
			sb.WriteString(key)
			writeFieldSchema(sb, v.Field(key))
		}
		sb.WriteRune('}')
	}
}

func serializeObjectData(sb *strings.Builder, v *Value) {
	if v == nil || !v.IsObject() {
		serializeValue(sb, v)
		return
	}

	sb.WriteRune('(')
	for i, key := range v.Keys() {
		if i > 0 {
			sb.WriteRune(',')
		}
		field := v.Field(key)
		if field.IsObject() {
			serializeObjectData(sb, field)
		} else if field.IsArray() {
			serializeArrayData(sb, field)
		} else {
			serializeValue(sb, field)
		}
	}
	sb.WriteRune(')')
}

func serializeArrayData(sb *strings.Builder, v *Value) {
	if v == nil || !v.IsArray() {
		serializeValue(sb, v)
		return
	}

	sb.WriteRune('[')
	for i := 0; i < v.Len(); i++ {
		if i > 0 {
			sb.WriteRune(',')
		}
		item := v.Get(i)
		if item.IsObject() {
			serializeObjectData(sb, item)
		} else {
			serializeValue(sb, item)
		}
	}
	sb.WriteRune(']')
}
