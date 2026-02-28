package ason

import (
	"reflect"
	"strconv"
	"strings"
	"unsafe"
)

// ---------------------------------------------------------------------------
// Decode — deserialize from ASON
// ---------------------------------------------------------------------------

// Decode parses ASON data and stores the result in v.
// v must be a pointer to a struct or a pointer to a slice of structs.
// Single struct input:  {field1,field2,...}:(val1,val2,...)
// Slice input:         [{field1,field2,...}]:(val1,val2,...),(val3,val4,...)
// The [{...}]: bracket wrapper is required for slice targets.
func Decode(data []byte, v any) error {
	d := decoder{
		data: data,
		pos:  0,
	}
	d.skipWhitespaceAndComments()

	// Detect target type and input format
	rv := reflect.ValueOf(v)
	isSliceTarget := rv.Kind() == reflect.Ptr && rv.Elem().Kind() == reflect.Slice
	startsWithBracket := d.pos < len(d.data) && d.data[d.pos] == '[' &&
		d.pos+1 < len(d.data) && d.data[d.pos+1] == '{'

	// Strict format enforcement: slices require [{...}]:, structs require {...}:
	if isSliceTarget && !startsWithBracket {
		return &UnmarshalError{d.pos, "Decode requires '[{...}]:' format for slice types"}
	}
	if startsWithBracket {
		return d.decodeSliceTop(v)
	}
	if err := d.unmarshalTop(v); err != nil {
		return err
	}
	d.skipWhitespaceAndComments()
	if d.pos < len(d.data) {
		return &UnmarshalError{d.pos, "trailing characters after decoded value"}
	}
	return nil
}

// decodeSliceTop parses [{schema}]:(v1,...),(v2,...) format
func (d *decoder) decodeSliceTop(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.Elem().Kind() != reflect.Slice {
		return &UnmarshalError{d.pos, "Decode requires a pointer to slice for [{...}]: format"}
	}
	sliceVal := rv.Elem()
	elemType := sliceVal.Type().Elem()
	for elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	if elemType.Kind() != reflect.Struct {
		return &UnmarshalError{d.pos, "Decode requires a slice of structs"}
	}

	// Skip '['
	d.pos++

	// Parse schema
	if d.pos >= len(d.data) || d.data[d.pos] != '{' {
		return d.errorf("expected '{'")
	}
	fields, err := d.parseSchema()
	if err != nil {
		return err
	}
	d.skipWhitespaceAndComments()
	if d.pos >= len(d.data) || d.data[d.pos] != ']' {
		return d.errorf("expected ']'")
	}
	d.pos++
	d.skipWhitespaceAndComments()
	if d.pos >= len(d.data) || d.data[d.pos] != ':' {
		return d.errorf("expected ':'")
	}
	d.pos++

	si := getStructInfo(elemType)
	fieldMap := buildFieldMap(si, fields)

	// Parse rows
	for {
		d.skipWhitespaceAndComments()
		if d.pos >= len(d.data) {
			break
		}
		if d.data[d.pos] != '(' {
			break
		}

		elem := reflect.New(elemType).Elem()
		if err := d.unmarshalTuple(elem, si, fieldMap); err != nil {
			return err
		}
		sliceVal = reflect.Append(sliceVal, elem)

		d.skipWhitespaceAndComments()
		if d.pos < len(d.data) && d.data[d.pos] == ',' {
			d.pos++
			d.skipWhitespaceAndComments()
			if d.pos >= len(d.data) || d.data[d.pos] != '(' {
				break
			}
		}
	}

	rv.Elem().Set(sliceVal)
	return nil
}

// ---------------------------------------------------------------------------
// decoder
// ---------------------------------------------------------------------------

type decoder struct {
	data []byte
	pos  int
}

func (d *decoder) errorf(format string, args ...any) *UnmarshalError {
	msg := format
	if len(args) > 0 {
		msg = sprintf(format, args...)
	}
	return &UnmarshalError{Pos: d.pos, Message: msg}
}

func sprintf(format string, args ...any) string {
	// Minimal sprintf to avoid fmt dependency in hot path
	result := format
	for _, a := range args {
		switch v := a.(type) {
		case string:
			result = strings.Replace(result, "%s", v, 1)
		case byte:
			result = strings.Replace(result, "%c", string(rune(v)), 1)
		}
	}
	return result
}

func (d *decoder) skipWhitespace() {
	for d.pos < len(d.data) {
		switch d.data[d.pos] {
		case ' ', '\t', '\n', '\r':
			d.pos++
		default:
			return
		}
	}
}

func (d *decoder) skipWhitespaceAndComments() {
	for {
		d.skipWhitespace()
		if d.pos+1 < len(d.data) && d.data[d.pos] == '/' && d.data[d.pos+1] == '*' {
			d.pos += 2
			for d.pos+1 < len(d.data) {
				if d.data[d.pos] == '*' && d.data[d.pos+1] == '/' {
					d.pos += 2
					break
				}
				d.pos++
			}
		} else {
			return
		}
	}
}

// parseSchema parses {field1,field2,...} or {field1:type1,field2:type2,...}
// Returns field names only (type annotations are skipped).
func (d *decoder) parseSchema() ([]string, error) {
	if d.data[d.pos] != '{' {
		return nil, d.errorf("expected '{'")
	}
	d.pos++

	var fields []string
	for {
		d.skipWhitespace()
		if d.pos >= len(d.data) {
			return nil, d.errorf("unexpected EOF in schema")
		}
		if d.data[d.pos] == '}' {
			d.pos++
			break
		}
		if len(fields) > 0 {
			if d.data[d.pos] != ',' {
				return nil, d.errorf("expected ','")
			}
			d.pos++
			d.skipWhitespace()
		}

		// Parse field name
		start := d.pos
		for d.pos < len(d.data) {
			b := d.data[d.pos]
			if b == ',' || b == '}' || b == ':' || b == ' ' || b == '\t' {
				break
			}
			d.pos++
		}
		name := unsafeString(d.data[start:d.pos])
		d.skipWhitespace()

		// Skip optional type annotation after ':'
		if d.pos < len(d.data) && d.data[d.pos] == ':' {
			d.pos++
			d.skipWhitespace()
			if d.pos < len(d.data) && d.data[d.pos] == '{' {
				if err := d.skipBalanced('{', '}'); err != nil {
					return nil, err
				}
			} else if d.pos < len(d.data) && d.data[d.pos] == '[' {
				if err := d.skipBalanced('[', ']'); err != nil {
					return nil, err
				}
			} else if d.pos+3 <= len(d.data) && string(d.data[d.pos:d.pos+3]) == "map" {
				d.pos += 3
				if d.pos < len(d.data) && d.data[d.pos] == '[' {
					if err := d.skipBalanced('[', ']'); err != nil {
						return nil, err
					}
				}
			} else {
				for d.pos < len(d.data) {
					b := d.data[d.pos]
					if b == ',' || b == '}' || b == ' ' || b == '\t' {
						break
					}
					d.pos++
				}
			}
		}

		fields = append(fields, name)
	}
	return fields, nil
}

func (d *decoder) skipBalanced(open, close byte) error {
	depth := 0
	for d.pos < len(d.data) {
		b := d.data[d.pos]
		d.pos++
		if b == open {
			depth++
		} else if b == close {
			depth--
			if depth == 0 {
				return nil
			}
		}
	}
	return d.errorf("unbalanced brackets")
}

// ---------------------------------------------------------------------------
// buildFieldMap maps schema field names to struct field indices
// ---------------------------------------------------------------------------

func buildFieldMap(si *structInfo, schemaFields []string) []int {
	m := make([]int, len(schemaFields))
	for i, name := range schemaFields {
		if idx, ok := si.nameIndex[name]; ok {
			m[i] = idx
		} else {
			m[i] = -1
		}
	}
	return m
}

// ---------------------------------------------------------------------------
// unmarshalTop — parse schema header then data for a single struct
// ---------------------------------------------------------------------------

func (d *decoder) unmarshalTop(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return d.errorf("Unmarshal requires a pointer")
	}
	rv = rv.Elem()
	if rv.Kind() != reflect.Struct {
		return d.errorf("Unmarshal requires a pointer to struct")
	}

	si := getStructInfo(rv.Type())

	// Parse schema
	if d.pos >= len(d.data) || d.data[d.pos] != '{' {
		return d.errorf("expected '{'")
	}
	fields, err := d.parseSchema()
	if err != nil {
		return err
	}

	d.skipWhitespaceAndComments()
	if d.pos >= len(d.data) || d.data[d.pos] != ':' {
		return d.errorf("expected ':'")
	}
	d.pos++
	d.skipWhitespaceAndComments()

	fieldMap := buildFieldMap(si, fields)
	return d.unmarshalTuple(rv, si, fieldMap)
}

// ---------------------------------------------------------------------------
// unmarshalTuple — parse (val1,val2,...) positionally
// ---------------------------------------------------------------------------

func (d *decoder) unmarshalTuple(rv reflect.Value, si *structInfo, fieldMap []int) error {
	d.skipWhitespaceAndComments()
	if d.pos >= len(d.data) || d.data[d.pos] != '(' {
		return d.errorf("expected '('")
	}
	d.pos++

	for i := 0; i < len(fieldMap); i++ {
		d.skipWhitespaceAndComments()
		if d.pos >= len(d.data) {
			return d.errorf("unexpected EOF in tuple")
		}
		if d.data[d.pos] == ')' {
			break
		}
		if i > 0 {
			if d.data[d.pos] == ',' {
				d.pos++
				d.skipWhitespaceAndComments()
				if d.pos < len(d.data) && d.data[d.pos] == ')' {
					break
				}
			} else if d.data[d.pos] == ')' {
				break
			} else {
				return d.errorf("expected ',' or ')'")
			}
		}

		fi := fieldMap[i]
		if fi < 0 {
			// Unknown field — skip value
			if err := d.skipValue(); err != nil {
				return err
			}
			continue
		}

		fv := rv.FieldByIndex(si.fields[fi].index)
		if err := d.unmarshalValue(fv); err != nil {
			return err
		}
	}

	// Skip any remaining values (source struct may have more fields than target)
	d.skipWhitespaceAndComments()
	for d.pos < len(d.data) && d.data[d.pos] != ')' {
		if d.data[d.pos] == ',' {
			d.pos++
			d.skipWhitespaceAndComments()
			if d.pos < len(d.data) && d.data[d.pos] == ')' {
				break
			}
		}
		if d.pos < len(d.data) && d.data[d.pos] != ')' {
			if err := d.skipValue(); err != nil {
				return err
			}
			d.skipWhitespaceAndComments()
		}
	}

	if d.pos < len(d.data) && d.data[d.pos] == ')' {
		d.pos++
	}
	return nil
}

// ---------------------------------------------------------------------------
// unmarshalValue — dispatch based on target field type
// ---------------------------------------------------------------------------

func (d *decoder) unmarshalValue(fv reflect.Value) error {
	d.skipWhitespaceAndComments()
	if d.pos >= len(d.data) {
		return nil
	}

	// Handle pointer types
	if fv.Kind() == reflect.Ptr {
		if d.atValueEnd() {
			// nil
			fv.Set(reflect.Zero(fv.Type()))
			return nil
		}
		if fv.IsNil() {
			fv.Set(reflect.New(fv.Type().Elem()))
		}
		return d.unmarshalValue(fv.Elem())
	}

	// Handle interface{}
	if fv.Kind() == reflect.Interface {
		val, err := d.parseAnyValue()
		if err != nil {
			return err
		}
		if val != nil {
			fv.Set(reflect.ValueOf(val))
		}
		return nil
	}

	switch fv.Kind() {
	case reflect.Bool:
		v, err := d.parseBool()
		if err != nil {
			return err
		}
		fv.SetBool(v)

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		v, err := d.parseInt64()
		if err != nil {
			return err
		}
		fv.SetInt(v)

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		v, err := d.parseUint64()
		if err != nil {
			return err
		}
		fv.SetUint(v)

	case reflect.Float32, reflect.Float64:
		v, err := d.parseFloat64()
		if err != nil {
			return err
		}
		fv.SetFloat(v)

	case reflect.String:
		s, err := d.parseString()
		if err != nil {
			return err
		}
		fv.SetString(s)

	case reflect.Slice:
		return d.unmarshalSlice(fv)

	case reflect.Map:
		return d.unmarshalMap(fv)

	case reflect.Struct:
		return d.unmarshalNestedStruct(fv)

	default:
		return d.errorf("unsupported type: %s", fv.Type().String())
	}
	return nil
}

// ---------------------------------------------------------------------------
// Parsing primitives
// ---------------------------------------------------------------------------

func (d *decoder) atValueEnd() bool {
	if d.pos >= len(d.data) {
		return true
	}
	b := d.data[d.pos]
	return b == ',' || b == ')' || b == ']'
}

func (d *decoder) parseBool() (bool, error) {
	d.skipWhitespaceAndComments()
	if d.pos+4 <= len(d.data) && string(d.data[d.pos:d.pos+4]) == "true" {
		if d.pos+4 >= len(d.data) || isDelim(d.data[d.pos+4]) {
			d.pos += 4
			return true, nil
		}
	}
	if d.pos+5 <= len(d.data) && string(d.data[d.pos:d.pos+5]) == "false" {
		if d.pos+5 >= len(d.data) || isDelim(d.data[d.pos+5]) {
			d.pos += 5
			return false, nil
		}
	}
	return false, d.errorf("invalid bool")
}

func isDelim(b byte) bool {
	return b == ',' || b == ')' || b == ']' || b == ' ' || b == '\t' || b == '\n' || b == '\r'
}

func (d *decoder) parseInt64() (int64, error) {
	d.skipWhitespaceAndComments()
	neg := false
	if d.pos < len(d.data) && d.data[d.pos] == '-' {
		neg = true
		d.pos++
	}
	var val uint64
	digits := 0
	for d.pos < len(d.data) && d.data[d.pos] >= '0' && d.data[d.pos] <= '9' {
		val = val*10 + uint64(d.data[d.pos]-'0')
		d.pos++
		digits++
	}
	if digits == 0 {
		return 0, d.errorf("invalid number")
	}
	if neg {
		return -int64(val), nil
	}
	return int64(val), nil
}

func (d *decoder) parseUint64() (uint64, error) {
	d.skipWhitespaceAndComments()
	var val uint64
	digits := 0
	for d.pos < len(d.data) && d.data[d.pos] >= '0' && d.data[d.pos] <= '9' {
		val = val*10 + uint64(d.data[d.pos]-'0')
		d.pos++
		digits++
	}
	if digits == 0 {
		return 0, d.errorf("invalid number")
	}
	return val, nil
}

func (d *decoder) parseFloat64() (float64, error) {
	d.skipWhitespaceAndComments()
	start := d.pos
	neg := false
	if d.pos < len(d.data) && d.data[d.pos] == '-' {
		neg = true
		d.pos++
	}
	intStart := d.pos
	for d.pos < len(d.data) && d.data[d.pos] >= '0' && d.data[d.pos] <= '9' {
		d.pos++
	}
	if d.pos == intStart && neg {
		return 0, d.errorf("invalid number")
	}
	hasDot := d.pos < len(d.data) && d.data[d.pos] == '.'
	if hasDot {
		d.pos++
		for d.pos < len(d.data) && d.data[d.pos] >= '0' && d.data[d.pos] <= '9' {
			d.pos++
		}
	}
	if d.pos == start || (d.pos == start+1 && neg) {
		return 0, d.errorf("invalid number")
	}
	// Fast path: simple decimal without exponent
	hasExp := d.pos < len(d.data) && (d.data[d.pos] == 'e' || d.data[d.pos] == 'E')
	if !hasExp && d.pos-start <= 18 {
		var intPart uint64
		p := start
		if neg {
			p++
		}
		for p < d.pos && d.data[p] != '.' {
			intPart = intPart*10 + uint64(d.data[p]-'0')
			p++
		}
		if !hasDot {
			v := float64(intPart)
			if neg {
				v = -v
			}
			return v, nil
		}
		p++ // skip '.'
		var frac float64
		scale := 0.1
		for p < d.pos {
			frac += float64(d.data[p]-'0') * scale
			scale *= 0.1
			p++
		}
		v := float64(intPart) + frac
		if neg {
			v = -v
		}
		return v, nil
	}
	// Fallback: handle exponent / very long numbers
	if hasExp {
		d.pos++
		if d.pos < len(d.data) && (d.data[d.pos] == '+' || d.data[d.pos] == '-') {
			d.pos++
		}
		for d.pos < len(d.data) && d.data[d.pos] >= '0' && d.data[d.pos] <= '9' {
			d.pos++
		}
	}
	s := unsafeString(d.data[start:d.pos])
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0, d.errorf("invalid float: %s", s)
	}
	return v, nil
}

func (d *decoder) parseString() (string, error) {
	d.skipWhitespaceAndComments()
	if d.pos >= len(d.data) {
		return "", nil
	}
	if d.data[d.pos] == '"' {
		return d.parseQuotedString()
	}
	return d.parsePlainValue()
}

// parsePlainValue reads unquoted string until delimiter, returns zerocopy slice.
func (d *decoder) parsePlainValue() (string, error) {
	start := d.pos
	hasEscape := false
	for d.pos < len(d.data) {
		b := d.data[d.pos]
		if b == ',' || b == ')' || b == ']' {
			break
		}
		if b == '\\' {
			hasEscape = true
			d.pos += 2
			continue
		}
		d.pos++
	}
	raw := d.data[start:d.pos]
	// Trim trailing whitespace
	for len(raw) > 0 && (raw[len(raw)-1] == ' ' || raw[len(raw)-1] == '\t') {
		raw = raw[:len(raw)-1]
	}
	// Trim leading whitespace
	for len(raw) > 0 && (raw[0] == ' ' || raw[0] == '\t') {
		raw = raw[1:]
	}
	if hasEscape {
		return unescapePlain(raw)
	}
	// Zerocopy
	return unsafeString(raw), nil
}

// parseQuotedString handles "..." with escape sequences.
// Zerocopy when no escapes found.
func (d *decoder) parseQuotedString() (string, error) {
	d.pos++ // skip opening "
	start := d.pos

	// Fast scan for closing quote without escapes
	scan := d.pos
	for scan < len(d.data) {
		if d.data[scan] == '"' {
			// No escapes — zerocopy
			s := unsafeString(d.data[start:scan])
			d.pos = scan + 1
			return s, nil
		}
		if d.data[scan] == '\\' {
			break
		}
		scan++
	}

	// Slow path: has escapes
	var buf []byte
	if scan > start {
		buf = append(buf, d.data[start:scan]...)
	}
	d.pos = scan

	for d.pos < len(d.data) {
		b := d.data[d.pos]
		if b == '"' {
			d.pos++
			return string(buf), nil
		}
		if b == '\\' {
			d.pos++
			if d.pos >= len(d.data) {
				return "", d.errorf("unclosed string")
			}
			esc := d.data[d.pos]
			d.pos++
			switch esc {
			case '"':
				buf = append(buf, '"')
			case '\\':
				buf = append(buf, '\\')
			case 'n':
				buf = append(buf, '\n')
			case 't':
				buf = append(buf, '\t')
			case ',':
				buf = append(buf, ',')
			case '(':
				buf = append(buf, '(')
			case ')':
				buf = append(buf, ')')
			case '[':
				buf = append(buf, '[')
			case ']':
				buf = append(buf, ']')
			case 'u':
				if d.pos+4 > len(d.data) {
					return "", d.errorf("invalid unicode escape")
				}
				hexStr := unsafeString(d.data[d.pos : d.pos+4])
				cp, err := strconv.ParseUint(hexStr, 16, 32)
				if err != nil {
					return "", d.errorf("invalid unicode escape")
				}
				buf = append(buf, string(rune(cp))...)
				d.pos += 4
			default:
				return "", d.errorf("invalid escape: \\%c", esc)
			}
		} else {
			buf = append(buf, b)
			d.pos++
		}
	}
	return "", d.errorf("unclosed string")
}

func unescapePlain(raw []byte) (string, error) {
	buf := make([]byte, 0, len(raw))
	i := 0
	for i < len(raw) {
		if raw[i] == '\\' {
			i++
			if i >= len(raw) {
				return "", &UnmarshalError{Message: "unexpected EOF in escape"}
			}
			switch raw[i] {
			case ',':
				buf = append(buf, ',')
			case '(':
				buf = append(buf, '(')
			case ')':
				buf = append(buf, ')')
			case '[':
				buf = append(buf, '[')
			case ']':
				buf = append(buf, ']')
			case '"':
				buf = append(buf, '"')
			case '\\':
				buf = append(buf, '\\')
			case 'n':
				buf = append(buf, '\n')
			case 't':
				buf = append(buf, '\t')
			case 'u':
				if i+4 >= len(raw) {
					return "", &UnmarshalError{Message: "invalid unicode escape"}
				}
				hexStr := unsafeString(raw[i+1 : i+5])
				cp, err := strconv.ParseUint(hexStr, 16, 32)
				if err != nil {
					return "", &UnmarshalError{Message: "invalid unicode escape"}
				}
				buf = append(buf, string(rune(cp))...)
				i += 4
			default:
				return "", &UnmarshalError{Message: "invalid escape: \\" + string(rune(raw[i]))}
			}
		} else {
			buf = append(buf, raw[i])
		}
		i++
	}
	return string(buf), nil
}

// ---------------------------------------------------------------------------
// parseAnyValue — parse value when target type is interface{}
// ---------------------------------------------------------------------------

func (d *decoder) parseAnyValue() (any, error) {
	d.skipWhitespaceAndComments()
	if d.pos >= len(d.data) || d.atValueEnd() {
		return nil, nil
	}
	b := d.data[d.pos]

	switch {
	case b == '"':
		return d.parseQuotedString()
	case b == '[':
		var arr []any
		d.pos++
		first := true
		for {
			d.skipWhitespaceAndComments()
			if d.pos >= len(d.data) || d.data[d.pos] == ']' {
				d.pos++
				break
			}
			if !first {
				if d.data[d.pos] == ',' {
					d.pos++
					d.skipWhitespaceAndComments()
					if d.pos < len(d.data) && d.data[d.pos] == ']' {
						d.pos++
						break
					}
				}
			}
			first = false
			v, err := d.parseAnyValue()
			if err != nil {
				return nil, err
			}
			arr = append(arr, v)
		}
		return arr, nil
	case b == '(':
		// tuple
		d.pos++
		var arr []any
		first := true
		for {
			d.skipWhitespaceAndComments()
			if d.pos >= len(d.data) || d.data[d.pos] == ')' {
				d.pos++
				break
			}
			if !first {
				if d.data[d.pos] == ',' {
					d.pos++
					d.skipWhitespaceAndComments()
					if d.pos < len(d.data) && d.data[d.pos] == ')' {
						d.pos++
						break
					}
				}
			}
			first = false
			v, err := d.parseAnyValue()
			if err != nil {
				return nil, err
			}
			arr = append(arr, v)
		}
		return arr, nil
	case b == 't' || b == 'f':
		v, err := d.parseBool()
		if err == nil {
			return v, nil
		}
		// fallthrough to string
		return d.parsePlainValue()
	case b == '-' && d.pos+1 < len(d.data) && d.data[d.pos+1] >= '0' && d.data[d.pos+1] <= '9':
		return d.parseNumberAny()
	case b >= '0' && b <= '9':
		return d.parseNumberAny()
	default:
		return d.parsePlainValue()
	}
}

func (d *decoder) parseNumberAny() (any, error) {
	start := d.pos
	if d.pos < len(d.data) && d.data[d.pos] == '-' {
		d.pos++
	}
	for d.pos < len(d.data) && d.data[d.pos] >= '0' && d.data[d.pos] <= '9' {
		d.pos++
	}
	isFloat := false
	if d.pos < len(d.data) && d.data[d.pos] == '.' {
		isFloat = true
		d.pos++
		for d.pos < len(d.data) && d.data[d.pos] >= '0' && d.data[d.pos] <= '9' {
			d.pos++
		}
	}
	s := unsafeString(d.data[start:d.pos])
	if isFloat {
		v, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return nil, d.errorf("invalid number")
		}
		return v, nil
	}
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return nil, d.errorf("invalid number")
	}
	return v, nil
}

// ---------------------------------------------------------------------------
// Compound type parsing
// ---------------------------------------------------------------------------

func (d *decoder) unmarshalSlice(fv reflect.Value) error {
	d.skipWhitespaceAndComments()
	if d.pos >= len(d.data) || d.data[d.pos] != '[' {
		return d.errorf("expected '['")
	}
	d.pos++

	elemType := fv.Type().Elem()
	slice := reflect.MakeSlice(fv.Type(), 0, 4)

	first := true
	for {
		d.skipWhitespaceAndComments()
		if d.pos >= len(d.data) || d.data[d.pos] == ']' {
			d.pos++
			break
		}
		if !first {
			if d.data[d.pos] == ',' {
				d.pos++
				d.skipWhitespaceAndComments()
				if d.pos < len(d.data) && d.data[d.pos] == ']' {
					d.pos++
					break
				}
			} else {
				break
			}
		}
		first = false

		elem := reflect.New(elemType).Elem()
		// If element is a struct inside an array, it needs nested parsing
		if elemType.Kind() == reflect.Struct {
			if err := d.unmarshalNestedStruct(elem); err != nil {
				return err
			}
		} else {
			if err := d.unmarshalValue(elem); err != nil {
				return err
			}
		}
		slice = reflect.Append(slice, elem)
	}

	fv.Set(slice)
	return nil
}

func (d *decoder) unmarshalMap(fv reflect.Value) error {
	d.skipWhitespaceAndComments()
	if d.pos >= len(d.data) || d.data[d.pos] != '[' {
		return d.errorf("expected '['")
	}
	d.pos++

	keyType := fv.Type().Key()
	valType := fv.Type().Elem()
	m := reflect.MakeMap(fv.Type())

	first := true
	for {
		d.skipWhitespaceAndComments()
		if d.pos >= len(d.data) || d.data[d.pos] == ']' {
			d.pos++
			break
		}
		if !first {
			if d.data[d.pos] == ',' {
				d.pos++
				d.skipWhitespaceAndComments()
				if d.pos < len(d.data) && d.data[d.pos] == ']' {
					d.pos++
					break
				}
			} else {
				break
			}
		}
		first = false

		// Expect (key,val)
		if d.pos >= len(d.data) || d.data[d.pos] != '(' {
			return d.errorf("expected '(' in map entry")
		}
		d.pos++

		key := reflect.New(keyType).Elem()
		if err := d.unmarshalValue(key); err != nil {
			return err
		}

		d.skipWhitespaceAndComments()
		if d.pos < len(d.data) && d.data[d.pos] == ',' {
			d.pos++
		}

		val := reflect.New(valType).Elem()
		if err := d.unmarshalValue(val); err != nil {
			return err
		}

		d.skipWhitespaceAndComments()
		if d.pos < len(d.data) && d.data[d.pos] == ')' {
			d.pos++
		}

		m.SetMapIndex(key, val)
	}

	fv.Set(m)
	return nil
}

func (d *decoder) unmarshalNestedStruct(fv reflect.Value) error {
	d.skipWhitespaceAndComments()
	if d.pos >= len(d.data) {
		return d.errorf("unexpected EOF")
	}

	// Check for inline schema: {field1,field2,...}:(val1,val2,...)
	if d.data[d.pos] == '{' {
		si := getStructInfo(fv.Type())
		fields, err := d.parseSchema()
		if err != nil {
			return err
		}
		d.skipWhitespaceAndComments()
		if d.pos >= len(d.data) || d.data[d.pos] != ':' {
			return d.errorf("expected ':'")
		}
		d.pos++
		d.skipWhitespaceAndComments()
		fieldMap := buildFieldMap(si, fields)
		return d.unmarshalTuple(fv, si, fieldMap)
	}

	// Positional tuple: (val1,val2,...)
	if d.data[d.pos] == '(' {
		si := getStructInfo(fv.Type())
		// Build identity fieldMap
		fieldMap := make([]int, len(si.fields))
		for i := range fieldMap {
			fieldMap[i] = i
		}
		return d.unmarshalTuple(fv, si, fieldMap)
	}

	return d.errorf("expected '{' or '(' for struct")
}

// ---------------------------------------------------------------------------
// skipValue — skip any value without allocating
// ---------------------------------------------------------------------------

func (d *decoder) skipValue() error {
	d.skipWhitespaceAndComments()
	if d.pos >= len(d.data) {
		return nil
	}

	switch d.data[d.pos] {
	case '"':
		d.pos++
		for d.pos < len(d.data) {
			if d.data[d.pos] == '"' {
				d.pos++
				return nil
			}
			if d.data[d.pos] == '\\' {
				d.pos++
			}
			d.pos++
		}
		return d.errorf("unclosed string")
	case '(':
		return d.skipBalanced('(', ')')
	case '[':
		return d.skipBalanced('[', ']')
	default:
		for d.pos < len(d.data) {
			b := d.data[d.pos]
			if b == ',' || b == ')' || b == ']' {
				return nil
			}
			d.pos++
		}
		return nil
	}
}

// ---------------------------------------------------------------------------
// unsafe helpers
// ---------------------------------------------------------------------------

func unsafeBytes(s string) []byte {
	if len(s) == 0 {
		return nil
	}
	return unsafe.Slice(unsafe.StringData(s), len(s))
}
