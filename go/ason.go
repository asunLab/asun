// Package ason provides ASON (Array-Schema Object Notation) serialization.
package ason

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// Marshal serializes a Go value to ASON format with schema.
func Marshal(v interface{}) (string, error) {
	return marshalInternal(reflect.ValueOf(v), true, "", 0)
}

// MarshalData serializes a Go value to ASON format without schema (data only).
func MarshalData(v interface{}) (string, error) {
	return marshalInternal(reflect.ValueOf(v), false, "", 0)
}

// MarshalIndent serializes a Go value to ASON format with indentation.
func MarshalIndent(v interface{}, prefix, indent string) (string, error) {
	return marshalInternal(reflect.ValueOf(v), true, indent, 0)
}

func marshal(rv reflect.Value, withSchema bool) (string, error) {
	return marshalInternal(rv, withSchema, "", 0)
}

type marshalOpts struct {
	withSchema bool
	indent     string
	depth      int
}

func marshalInternal(rv reflect.Value, withSchema bool, indent string, depth int) (string, error) {
	opts := &marshalOpts{withSchema: withSchema, indent: indent, depth: depth}
	return marshalValue(rv, opts)
}

func marshalValue(rv reflect.Value, opts *marshalOpts) (string, error) {
	// Handle pointer/interface
	for rv.Kind() == reflect.Ptr || rv.Kind() == reflect.Interface {
		if rv.IsNil() {
			return "null", nil
		}
		rv = rv.Elem()
	}

	switch rv.Kind() {
	case reflect.Bool:
		if rv.Bool() {
			return "true", nil
		}
		return "false", nil

	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10), nil

	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10), nil

	case reflect.Float32:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 32), nil
	case reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'g', -1, 64), nil

	case reflect.String:
		return formatString(rv.String()), nil

	case reflect.Slice, reflect.Array:
		return marshalSliceOpts(rv, opts)

	case reflect.Struct:
		return marshalStructOpts(rv, opts)

	case reflect.Map:
		return marshalMapOpts(rv, opts)

	default:
		return "", fmt.Errorf("unsupported type: %v", rv.Kind())
	}
}

func formatString(s string) string {
	if s == "" || s == "null" || s == "true" || s == "false" {
		return `"` + s + `"`
	}
	first := s[0]
	if first == '-' || first == '+' || (first >= '0' && first <= '9') {
		return `"` + escapeString(s) + `"`
	}
	for _, c := range s {
		if c == '"' || c == '\\' || c == '(' || c == ')' ||
			c == '[' || c == ']' || c == '{' || c == '}' ||
			c == ',' || c == ':' || c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			return `"` + escapeString(s) + `"`
		}
	}
	return s
}

func escapeString(s string) string {
	var sb strings.Builder
	for _, c := range s {
		switch c {
		case '"':
			sb.WriteString(`\"`)
		case '\\':
			sb.WriteString(`\\`)
		case '\n':
			sb.WriteString(`\n`)
		case '\r':
			sb.WriteString(`\r`)
		case '\t':
			sb.WriteString(`\t`)
		default:
			sb.WriteRune(c)
		}
	}
	return sb.String()
}

func marshalSlice(rv reflect.Value, withSchema bool) (string, error) {
	return marshalSliceOpts(rv, &marshalOpts{withSchema: withSchema})
}

func marshalSliceOpts(rv reflect.Value, opts *marshalOpts) (string, error) {
	if rv.Len() == 0 {
		return "[]", nil
	}

	var elements []string
	var schema string
	// Serialize children WITHOUT indent to extract schema cleanly
	childOpts := &marshalOpts{withSchema: true, indent: "", depth: 0}

	for i := 0; i < rv.Len(); i++ {
		elem, err := marshalValue(rv.Index(i), childOpts)
		if err != nil {
			return "", err
		}

		// Check if element has schema: {fields}:(data)
		if strings.HasPrefix(elem, "{") && strings.Contains(elem, "}:(") {
			colonPos := strings.Index(elem, "}:(")
			if schema == "" {
				schema = elem[1:colonPos]
			}
			elements = append(elements, elem[colonPos+2:]) // includes (...)
		} else {
			elements = append(elements, elem)
		}
	}

	if schema != "" && opts.withSchema {
		if opts.indent != "" {
			childInd := strings.Repeat(opts.indent, opts.depth+1)
			return "{" + schema + "}:\n" + childInd + strings.Join(elements, ",\n"+childInd), nil
		}
		return "{" + schema + "}:" + strings.Join(elements, ","), nil
	}
	if opts.indent != "" {
		childInd := strings.Repeat(opts.indent, opts.depth+1)
		ind := strings.Repeat(opts.indent, opts.depth)
		return "[\n" + childInd + strings.Join(elements, ",\n"+childInd) + "\n" + ind + "]", nil
	}
	return "[" + strings.Join(elements, ",") + "]", nil
}

func marshalStruct(rv reflect.Value, withSchema bool) (string, error) {
	return marshalStructOpts(rv, &marshalOpts{withSchema: withSchema})
}

func marshalStructOpts(rv reflect.Value, opts *marshalOpts) (string, error) {
	t := rv.Type()
	var schemaFields []string
	var dataFields []string
	childOpts := &marshalOpts{withSchema: true, indent: opts.indent, depth: opts.depth + 1}

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue // unexported
		}

		name := getFieldName(field)
		if name == "-" {
			continue
		}

		fv := rv.Field(i)
		value, err := marshalValue(fv, childOpts)
		if err != nil {
			return "", err
		}

		// Check for nested schema
		if strings.HasPrefix(value, "{") && strings.Contains(value, "}:(") {
			colonPos := strings.Index(value, "}:(")
			nestedSchema := value[1:colonPos]
			nestedData := value[colonPos+3 : len(value)-1]
			schemaFields = append(schemaFields, name+"{"+nestedSchema+"}")
			dataFields = append(dataFields, "("+nestedData+")")
		} else if strings.HasPrefix(value, "{") && strings.Contains(value, "}:") {
			// Array of structs
			colonPos := strings.Index(value, "}:")
			nestedSchema := value[1:colonPos]
			nestedData := value[colonPos+2:]
			schemaFields = append(schemaFields, name+"[]{"+nestedSchema+"}")
			dataFields = append(dataFields, "["+nestedData+"]")
		} else if strings.HasPrefix(value, "[") {
			schemaFields = append(schemaFields, name+"[]")
			dataFields = append(dataFields, value)
		} else {
			schemaFields = append(schemaFields, name)
			dataFields = append(dataFields, value)
		}
	}

	if opts.indent != "" {
		indent := strings.Repeat(opts.indent, opts.depth)
		childIndent := strings.Repeat(opts.indent, opts.depth+1)
		data := "(\n" + childIndent + strings.Join(dataFields, ",\n"+childIndent) + "\n" + indent + ")"
		if opts.withSchema {
			return "{" + strings.Join(schemaFields, ",") + "}:\n" + indent + data, nil
		}
		return data, nil
	}

	data := "(" + strings.Join(dataFields, ",") + ")"
	if opts.withSchema {
		return "{" + strings.Join(schemaFields, ",") + "}:" + data, nil
	}
	return data, nil
}

func marshalMap(rv reflect.Value, withSchema bool) (string, error) {
	return marshalMapOpts(rv, &marshalOpts{withSchema: withSchema})
}

func marshalMapOpts(rv reflect.Value, opts *marshalOpts) (string, error) {
	keys := rv.MapKeys()
	if len(keys) == 0 {
		return "{}", nil
	}

	var schemaFields []string
	var dataFields []string
	childOpts := &marshalOpts{withSchema: true, indent: opts.indent, depth: opts.depth + 1}

	for _, key := range keys {
		name := fmt.Sprintf("%v", key.Interface())
		value, err := marshalValue(rv.MapIndex(key), childOpts)
		if err != nil {
			return "", err
		}
		schemaFields = append(schemaFields, name)
		dataFields = append(dataFields, value)
	}

	data := "(" + strings.Join(dataFields, ",") + ")"
	if opts.withSchema {
		return "{" + strings.Join(schemaFields, ",") + "}:" + data, nil
	}
	return data, nil
}

func getFieldName(field reflect.StructField) string {
	if tag := field.Tag.Get("ason"); tag != "" {
		if idx := strings.Index(tag, ","); idx != -1 {
			return tag[:idx]
		}
		return tag
	}
	if tag := field.Tag.Get("json"); tag != "" {
		if idx := strings.Index(tag, ","); idx != -1 {
			return tag[:idx]
		}
		return tag
	}
	return field.Name
}

// Unmarshal parses ASON data and stores the result in v.
func Unmarshal(data string, v interface{}) error {
	p := &parser{input: data}
	return p.unmarshal(reflect.ValueOf(v))
}

type parser struct {
	input  string
	pos    int
	schema []schemaField
}

type schemaField struct {
	name     string
	isArray  bool
	children []schemaField
}

func (p *parser) peek() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	return p.input[p.pos]
}

func (p *parser) advance() byte {
	if p.pos >= len(p.input) {
		return 0
	}
	c := p.input[p.pos]
	p.pos++
	return c
}

func (p *parser) skipWs() {
	for p.pos < len(p.input) && (p.input[p.pos] == ' ' || p.input[p.pos] == '\t' ||
		p.input[p.pos] == '\n' || p.input[p.pos] == '\r') {
		p.pos++
	}
}

func (p *parser) expect(c byte) error {
	p.skipWs()
	if p.peek() != c {
		return fmt.Errorf("expected '%c', got '%c' at pos %d", c, p.peek(), p.pos)
	}
	p.advance()
	return nil
}

func (p *parser) unmarshal(rv reflect.Value) error {
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return fmt.Errorf("unmarshal requires non-nil pointer")
	}
	rv = rv.Elem()

	p.skipWs()
	// Check for schema
	if p.peek() == '{' {
		schema, err := p.parseSchema()
		if err != nil {
			return err
		}
		p.schema = schema
		if err := p.expect(':'); err != nil {
			return err
		}
	}

	return p.parseValue(rv)
}

func (p *parser) parseSchema() ([]schemaField, error) {
	if err := p.expect('{'); err != nil {
		return nil, err
	}

	var fields []schemaField
	for {
		p.skipWs()
		if p.peek() == '}' {
			p.advance()
			break
		}
		if len(fields) > 0 {
			if err := p.expect(','); err != nil {
				return nil, err
			}
		}
		field, err := p.parseSchemaField()
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)
	}
	return fields, nil
}

func (p *parser) parseSchemaField() (schemaField, error) {
	p.skipWs()
	name := p.parseIdent()
	if name == "" {
		return schemaField{}, fmt.Errorf("expected field name at pos %d", p.pos)
	}

	field := schemaField{name: name}
	p.skipWs()

	if p.peek() == '[' {
		p.advance()
		if err := p.expect(']'); err != nil {
			return schemaField{}, err
		}
		field.isArray = true
	}

	p.skipWs()
	if p.peek() == '{' {
		children, err := p.parseSchema()
		if err != nil {
			return schemaField{}, err
		}
		field.children = children
	}

	return field, nil
}

func (p *parser) parseIdent() string {
	p.skipWs()
	start := p.pos
	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '-' {
			p.pos++
		} else {
			break
		}
	}
	return p.input[start:p.pos]
}

func (p *parser) parseValue(rv reflect.Value) error {
	p.skipWs()

	// Handle pointer
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return p.parseValue(rv.Elem())
	}

	switch p.peek() {
	case '(':
		return p.parseTuple(rv)
	case '[':
		return p.parseArray(rv)
	case '"':
		return p.parseQuotedString(rv)
	case 'n':
		return p.parseNull(rv)
	case 't', 'f':
		return p.parseBool(rv)
	default:
		return p.parseUnquoted(rv)
	}
}

func (p *parser) parseTuple(rv reflect.Value) error {
	if err := p.expect('('); err != nil {
		return err
	}

	switch rv.Kind() {
	case reflect.Struct:
		return p.parseStruct(rv)
	case reflect.Slice:
		// Single tuple as slice element
		elem := reflect.New(rv.Type().Elem()).Elem()
		if err := p.parseStruct(elem); err != nil {
			return err
		}
		if err := p.expect(')'); err != nil {
			return err
		}
		rv.Set(reflect.Append(rv, elem))
		// Check for more tuples
		for {
			p.skipWs()
			if p.peek() != ',' {
				break
			}
			p.advance()
			p.skipWs()
			if p.peek() != '(' {
				break
			}
			p.advance()
			elem := reflect.New(rv.Type().Elem()).Elem()
			if err := p.parseStruct(elem); err != nil {
				return err
			}
			if err := p.expect(')'); err != nil {
				return err
			}
			rv.Set(reflect.Append(rv, elem))
		}
		return nil
	default:
		return fmt.Errorf("cannot parse tuple into %v", rv.Kind())
	}
}

func (p *parser) parseStruct(rv reflect.Value) error {
	t := rv.Type()
	fieldMap := make(map[string]int)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.PkgPath != "" {
			continue
		}
		name := getFieldName(field)
		fieldMap[name] = i
		fieldMap[strings.ToLower(name)] = i
	}

	// Use schema if available, otherwise use struct field order
	schema := p.schema
	if len(schema) == 0 {
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" {
				continue
			}
			schema = append(schema, schemaField{name: getFieldName(field)})
		}
	}

	for i, sf := range schema {
		if i > 0 {
			if err := p.expect(','); err != nil {
				return err
			}
		}

		idx, ok := fieldMap[sf.name]
		if !ok {
			idx, ok = fieldMap[strings.ToLower(sf.name)]
		}
		if !ok {
			// Skip unknown field
			p.skipValue()
			continue
		}

		// Set child schema for nested struct
		oldSchema := p.schema
		if len(sf.children) > 0 {
			p.schema = sf.children
		}

		if err := p.parseValue(rv.Field(idx)); err != nil {
			return err
		}

		p.schema = oldSchema
	}

	return nil
}

func (p *parser) parseArray(rv reflect.Value) error {
	if err := p.expect('['); err != nil {
		return err
	}

	if rv.Kind() != reflect.Slice {
		return fmt.Errorf("cannot parse array into %v", rv.Kind())
	}

	elemType := rv.Type().Elem()
	for {
		p.skipWs()
		if p.peek() == ']' {
			p.advance()
			break
		}
		if rv.Len() > 0 {
			if err := p.expect(','); err != nil {
				return err
			}
		}
		elem := reflect.New(elemType).Elem()
		if err := p.parseValue(elem); err != nil {
			return err
		}
		rv.Set(reflect.Append(rv, elem))
	}
	return nil
}

func (p *parser) parseQuotedString(rv reflect.Value) error {
	p.advance() // opening "
	start := p.pos
	for p.pos < len(p.input) && p.input[p.pos] != '"' {
		if p.input[p.pos] == '\\' {
			p.pos++
		}
		p.pos++
	}
	s := p.input[start:p.pos]
	p.advance() // closing "

	return p.setString(rv, s)
}

func (p *parser) parseUnquoted(rv reflect.Value) error {
	start := p.pos
	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if c == '(' || c == ')' || c == '[' || c == ']' ||
			c == '{' || c == '}' || c == ',' || c == ':' ||
			c == ' ' || c == '\t' || c == '\n' || c == '\r' {
			break
		}
		p.pos++
	}
	s := p.input[start:p.pos]

	switch rv.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		rv.SetInt(n)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		n, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		rv.SetUint(n)
	case reflect.Float32, reflect.Float64:
		n, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		rv.SetFloat(n)
	case reflect.String:
		rv.SetString(s)
	default:
		return fmt.Errorf("cannot parse %q into %v", s, rv.Kind())
	}
	return nil
}

func (p *parser) parseNull(rv reflect.Value) error {
	if p.pos+4 <= len(p.input) && p.input[p.pos:p.pos+4] == "null" {
		p.pos += 4
		rv.Set(reflect.Zero(rv.Type()))
		return nil
	}
	return p.parseUnquoted(rv)
}

func (p *parser) parseBool(rv reflect.Value) error {
	if p.pos+4 <= len(p.input) && p.input[p.pos:p.pos+4] == "true" {
		p.pos += 4
		if rv.Kind() == reflect.Bool {
			rv.SetBool(true)
		}
		return nil
	}
	if p.pos+5 <= len(p.input) && p.input[p.pos:p.pos+5] == "false" {
		p.pos += 5
		if rv.Kind() == reflect.Bool {
			rv.SetBool(false)
		}
		return nil
	}
	return p.parseUnquoted(rv)
}

func (p *parser) setString(rv reflect.Value, s string) error {
	if rv.Kind() == reflect.String {
		rv.SetString(s)
		return nil
	}
	return fmt.Errorf("cannot set string to %v", rv.Kind())
}

func (p *parser) skipValue() {
	depth := 0
	for p.pos < len(p.input) {
		c := p.input[p.pos]
		if c == '(' || c == '[' || c == '{' {
			depth++
		} else if c == ')' || c == ']' || c == '}' {
			if depth == 0 {
				break
			}
			depth--
		} else if c == ',' && depth == 0 {
			break
		}
		p.pos++
	}
}
