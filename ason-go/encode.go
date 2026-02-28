package ason

import (
	"math"
	"reflect"
	"strconv"
	"sync"
	"unsafe"
)

// ---------------------------------------------------------------------------
// Lookup tables
// ---------------------------------------------------------------------------

// needsQuote[b] is true if byte b forces a string to be wrapped in "..."
var needsQuote = func() [256]bool {
	var t [256]bool
	for i := 0; i < 32; i++ {
		t[i] = true
	}
	t[','] = true
	t['('] = true
	t[')'] = true
	t['['] = true
	t[']'] = true
	t['"'] = true
	t['\\'] = true
	return t
}()

// escapeChar[b] is the char after '\' for byte b inside a quoted string.
// 0 means the byte needs no escaping.
var escapeChar = func() [256]byte {
	var t [256]byte
	t['"'] = '"'
	t['\\'] = '\\'
	t['\n'] = 'n'
	t['\t'] = 't'
	return t
}()

// decDigits is a two-digit lookup for fast integer formatting.
const decDigits = "00010203040506070809" +
	"10111213141516171819" +
	"20212223242526272829" +
	"30313233343536373839" +
	"40414243444546474849" +
	"50515253545556575859" +
	"60616263646566676869" +
	"70717273747576777879" +
	"80818283848586878889" +
	"90919293949596979899"

// ---------------------------------------------------------------------------
// Buffer pool
// ---------------------------------------------------------------------------

var bufPool = sync.Pool{
	New: func() any {
		b := make([]byte, 0, 256)
		return &b
	},
}

func getBuf() *[]byte {
	bp := bufPool.Get().(*[]byte)
	*bp = (*bp)[:0]
	return bp
}

func putBuf(bp *[]byte) {
	if cap(*bp) <= 1<<16 {
		bufPool.Put(bp)
	}
}

// ---------------------------------------------------------------------------
// Fast number formatting
// ---------------------------------------------------------------------------

func appendU64(buf []byte, v uint64) []byte {
	if v < 10 {
		return append(buf, byte('0'+v))
	}
	if v < 100 {
		idx := v * 2
		return append(buf, decDigits[idx], decDigits[idx+1])
	}
	var tmp [20]byte
	i := 20
	for v >= 100 {
		rem := v % 100
		v /= 100
		i -= 2
		tmp[i] = decDigits[rem*2]
		tmp[i+1] = decDigits[rem*2+1]
	}
	if v >= 10 {
		idx := v * 2
		i -= 2
		tmp[i] = decDigits[idx]
		tmp[i+1] = decDigits[idx+1]
	} else {
		i--
		tmp[i] = byte('0' + v)
	}
	return append(buf, tmp[i:]...)
}

func appendI64(buf []byte, v int64) []byte {
	if v < 0 {
		buf = append(buf, '-')
		if v == math.MinInt64 {
			return append(buf, "9223372036854775808"...)
		}
		return appendU64(buf, uint64(-v))
	}
	return appendU64(buf, uint64(v))
}

func appendFloat64(buf []byte, v float64) []byte {
	if math.IsInf(v, 0) || math.IsNaN(v) {
		return append(buf, '0')
	}
	// Integer-valued float: write as int + ".0"
	_, frac := math.Modf(v)
	if frac == 0.0 {
		iv := int64(v)
		if float64(iv) == v {
			buf = appendI64(buf, iv)
			return append(buf, '.', '0')
		}
	}
	// Fast path: one decimal place
	v10 := v * 10.0
	_, frac10 := math.Modf(v10)
	if frac10 == 0.0 && math.Abs(v10) < 1e18 {
		vi := int64(v10)
		if vi < 0 {
			buf = append(buf, '-')
			vi = -vi
		}
		intPart := uint64(vi) / 10
		fracPart := byte(uint64(vi) % 10)
		buf = appendU64(buf, intPart)
		buf = append(buf, '.', '0'+fracPart)
		return buf
	}
	// Fast path: two decimal places
	v100 := v * 100.0
	_, frac100 := math.Modf(v100)
	if frac100 == 0.0 && math.Abs(v100) < 1e18 {
		vi := int64(v100)
		if vi < 0 {
			buf = append(buf, '-')
			vi = -vi
		}
		intPart := uint64(vi) / 100
		fracIdx := uint64(vi) % 100
		buf = appendU64(buf, intPart)
		buf = append(buf, '.')
		buf = append(buf, decDigits[fracIdx*2])
		d2 := decDigits[fracIdx*2+1]
		if d2 != '0' {
			buf = append(buf, d2)
		}
		return buf
	}
	// Fallback: use strconv-like formatting via fmt
	// We use a simple approach: format with enough precision
	return appendFloatGeneral(buf, v)
}

func appendFloatGeneral(buf []byte, v float64) []byte {
	// Use strconv for full correctness on edge cases
	buf = strconv.AppendFloat(buf, v, 'f', -1, 64)
	// Ensure decimal point for float identity
	hasDecimal := false
	for i := len(buf) - 1; i >= 0; i-- {
		if buf[i] == '.' {
			hasDecimal = true
			break
		}
		if buf[i] == '-' || buf[i] < '0' || buf[i] > '9' {
			break
		}
	}
	if !hasDecimal {
		buf = append(buf, '.', '0')
	}
	return buf
}

// ---------------------------------------------------------------------------
// String quoting / escaping
// ---------------------------------------------------------------------------

func stringNeedsQuoting(s string) bool {
	if len(s) == 0 {
		return true
	}
	b := *(*[]byte)(unsafe.Pointer(&s))
	if b[0] == ' ' || b[len(b)-1] == ' ' {
		return true
	}
	if s == "true" || s == "false" {
		return true
	}
	couldBeNumber := true
	numStart := 0
	if b[0] == '-' {
		numStart = 1
	}
	if numStart >= len(b) {
		couldBeNumber = false
	}
	for i := 0; i < len(b); i++ {
		if needsQuote[b[i]] {
			return true
		}
		if couldBeNumber && i >= numStart && !(b[i] >= '0' && b[i] <= '9') && b[i] != '.' {
			couldBeNumber = false
		}
	}
	return couldBeNumber && len(b) > numStart
}

func appendEscaped(buf []byte, s string) []byte {
	buf = append(buf, '"')
	b := *(*[]byte)(unsafe.Pointer(&s))
	start := 0
	for i := 0; i < len(b); i++ {
		esc := escapeChar[b[i]]
		if esc != 0 {
			buf = append(buf, b[start:i]...)
			buf = append(buf, '\\', esc)
			start = i + 1
		}
	}
	buf = append(buf, b[start:]...)
	buf = append(buf, '"')
	return buf
}

func appendStr(buf []byte, s string) []byte {
	if stringNeedsQuoting(s) {
		return appendEscaped(buf, s)
	}
	return append(buf, s...)
}

// ---------------------------------------------------------------------------
// Struct info cache
// ---------------------------------------------------------------------------

type fieldInfo struct {
	name      string
	index     []int
	tagged    bool
	fieldType reflect.Type // resolved field type (for schema generation)
}

type structInfo struct {
	fields     []fieldInfo
	structType reflect.Type   // the struct type these fields belong to
	nameIndex  map[string]int // field name → index in fields slice
}

var structCache sync.Map // map[reflect.Type]*structInfo

func getStructInfo(t reflect.Type) *structInfo {
	if v, ok := structCache.Load(t); ok {
		return v.(*structInfo)
	}
	si := buildStructInfo(t)
	actual, _ := structCache.LoadOrStore(t, si)
	return actual.(*structInfo)
}

func buildStructInfo(t reflect.Type) *structInfo {
	n := t.NumField()
	fields := make([]fieldInfo, 0, n)
	for i := 0; i < n; i++ {
		f := t.Field(i)
		if !f.IsExported() {
			continue
		}
		fi := fieldInfo{index: f.Index, fieldType: f.Type}

		// Check ason tag first, then json tag
		if tag, ok := f.Tag.Lookup("ason"); ok {
			if tag == "-" {
				continue
			}
			name := tag
			if idx := indexOf(name, ','); idx >= 0 {
				name = name[:idx]
			}
			if name != "" {
				fi.name = name
				fi.tagged = true
			}
		}
		if fi.name == "" {
			if tag, ok := f.Tag.Lookup("json"); ok {
				if tag == "-" {
					continue
				}
				name := tag
				if idx := indexOf(name, ','); idx >= 0 {
					name = name[:idx]
				}
				if name != "" {
					fi.name = name
					fi.tagged = true
				}
			}
		}
		if fi.name == "" {
			fi.name = f.Name
		}

		// Handle embedded structs
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			embedded := getStructInfo(f.Type)
			for _, ef := range embedded.fields {
				ef2 := ef
				ef2.index = append(append([]int{}, f.Index...), ef.index...)
				fields = append(fields, ef2)
			}
			continue
		}

		fields = append(fields, fi)
	}
	si := &structInfo{fields: fields, structType: t}
	si.nameIndex = make(map[string]int, len(fields))
	for i, fi := range fields {
		si.nameIndex[fi.name] = i
	}
	return si
}

func indexOf(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// ---------------------------------------------------------------------------
// Recursive schema generation
// ---------------------------------------------------------------------------

// appendStructSchema writes {field1,field2,...} for a struct type.
// For nested struct/slice-of-struct fields, it recurses into the nested schema.
func appendStructSchema(buf []byte, si *structInfo, typed bool) []byte {
	buf = append(buf, '{')
	for i, fi := range si.fields {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = appendFieldSchema(buf, fi, typed)
	}
	buf = append(buf, '}')
	return buf
}

// appendFieldSchema writes a field's name and optional schema/type annotation.
// Struct and slice-of-struct fields always get nested schema (structural info).
// Primitive fields only get type annotations in typed mode.
func appendFieldSchema(buf []byte, fi fieldInfo, typed bool) []byte {
	buf = append(buf, fi.name...)

	ft := fi.fieldType
	for ft.Kind() == reflect.Ptr {
		ft = ft.Elem()
	}

	switch ft.Kind() {
	case reflect.Struct:
		// Always output nested struct schema: field:{f1,f2,...}
		nested := getStructInfo(ft)
		buf = append(buf, ':')
		buf = appendStructSchema(buf, nested, typed)
	case reflect.Slice:
		elemType := ft.Elem()
		for elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}
		if elemType.Kind() == reflect.Struct {
			// Always output nested struct-array schema: field:[{f1,f2,...}]
			nested := getStructInfo(elemType)
			buf = append(buf, ':')
			buf = append(buf, '[')
			buf = appendStructSchema(buf, nested, typed)
			buf = append(buf, ']')
		} else if typed {
			hint := typeHintForKind(elemType.Kind())
			if hint != "" {
				buf = append(buf, ':', '[')
				buf = append(buf, hint...)
				buf = append(buf, ']')
			}
		}
	case reflect.Map:
		if typed {
			keyHint := typeHintForKind(ft.Key().Kind())
			valType := ft.Elem()
			for valType.Kind() == reflect.Ptr {
				valType = valType.Elem()
			}
			valHint := typeHintForKind(valType.Kind())
			if keyHint != "" && valHint != "" {
				buf = append(buf, ':')
				buf = append(buf, "map["...)
				buf = append(buf, keyHint...)
				buf = append(buf, ',')
				buf = append(buf, valHint...)
				buf = append(buf, ']')
			}
		}
	default:
		if typed {
			hint := typeHintForKind(ft.Kind())
			if hint != "" {
				buf = append(buf, ':')
				buf = append(buf, hint...)
			}
		}
	}
	return buf
}

func typeHintForKind(k reflect.Kind) string {
	switch k {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return "int"
	case reflect.Float32, reflect.Float64:
		return "float"
	case reflect.Bool:
		return "bool"
	case reflect.String:
		return "str"
	default:
		return ""
	}
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

// Marshal serializes a struct to ASON format.
// Output: {field1,field2,...}:(val1,val2,...)
// Encode serializes a struct or slice of structs to ASON format.
// Single struct output: {field1,field2,...}:(val1,val2,...)
// Slice output: [{field1,field2,...}]:(val1,val2,...),(val3,val4,...)
func Encode(v any) ([]byte, error) {
	return encodeInner(v, false)
}

// EncodeTyped serializes a struct or slice of structs to ASON format with type annotations.
// Single: {field1:type1,field2:type2,...}:(val1,val2,...)
// Slice: [{field1:type1,...}]:(val1,val2,...),(val3,val4,...)
func EncodeTyped(v any) ([]byte, error) {
	return encodeInner(v, true)
}

func encodeInner(v any, typed bool) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	// Auto-detect slice of structs
	if rv.Kind() == reflect.Slice {
		elemType := rv.Type().Elem()
		for elemType.Kind() == reflect.Ptr {
			elemType = elemType.Elem()
		}
		if elemType.Kind() == reflect.Struct {
			return encodeSliceInner(v, typed)
		}
	}
	if rv.Kind() != reflect.Struct {
		return nil, &MarshalError{"Encode requires a struct or slice of structs"}
	}

	bp := getBuf()
	buf := *bp

	si := getStructInfo(rv.Type())

	// Serialize data first into a temp region
	dataBuf := make([]byte, 0, 128)
	dataBuf = append(dataBuf, '(')

	for i, fi := range si.fields {
		if i > 0 {
			dataBuf = append(dataBuf, ',')
		}
		fv := rv.FieldByIndex(fi.index)
		dataBuf, _ = marshalFieldValue(dataBuf, fv, false)
	}
	dataBuf = append(dataBuf, ')')

	// Build schema header (recursive — includes nested struct schemas)
	buf = appendStructSchema(buf, si, typed)
	buf = append(buf, ':')

	// Append data
	buf = append(buf, dataBuf...)

	result := make([]byte, len(buf))
	copy(result, buf)
	*bp = buf
	putBuf(bp)
	return result, nil
}

// encodeSliceInner serializes a slice of structs to ASON format.
// Output: [{field1,field2,...}]:(v1,v2,...),(v3,v4,...)
func encodeSliceInner(v any, typed bool) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() != reflect.Slice {
		return nil, &MarshalError{"Encode requires a slice of structs"}
	}
	elemType := rv.Type().Elem()
	for elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}
	if elemType.Kind() != reflect.Struct {
		return nil, &MarshalError{"Encode requires a slice of structs"}
	}

	si := getStructInfo(elemType)

	bp := getBuf()
	buf := *bp

	// Schema header (recursive — includes nested struct schemas)
	buf = append(buf, '[')
	buf = appendStructSchema(buf, si, typed)
	buf = append(buf, ']', ':')

	// Data rows
	for i := 0; i < rv.Len(); i++ {
		if i > 0 {
			buf = append(buf, ',')
		}
		elem := rv.Index(i)
		if elem.Kind() == reflect.Ptr {
			elem = elem.Elem()
		}
		buf = marshalStruct(buf, elem, si, false)
	}

	result := make([]byte, len(buf))
	copy(result, buf)
	*bp = buf
	putBuf(bp)
	return result, nil
}

// ---------------------------------------------------------------------------
// Internal marshal functions
// ---------------------------------------------------------------------------

func marshalStruct(buf []byte, rv reflect.Value, si *structInfo, isTop bool) []byte {
	buf = append(buf, '(')
	for i, fi := range si.fields {
		if i > 0 {
			buf = append(buf, ',')
		}
		fv := rv.FieldByIndex(fi.index)
		buf = marshalNestedValue(buf, fv)
	}
	buf = append(buf, ')')
	return buf
}

// marshalFieldValue marshals a value and returns the type hint (for typed mode).
func marshalFieldValue(buf []byte, fv reflect.Value, typed bool) ([]byte, string) {
	hint := ""
	// Unwrap pointer
	for fv.Kind() == reflect.Ptr {
		if fv.IsNil() {
			return buf, ""
		}
		fv = fv.Elem()
	}

	switch fv.Kind() {
	case reflect.Bool:
		if typed {
			hint = "bool"
		}
		if fv.Bool() {
			buf = append(buf, "true"...)
		} else {
			buf = append(buf, "false"...)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if typed {
			hint = "int"
		}
		buf = appendI64(buf, fv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		if typed {
			hint = "int"
		}
		buf = appendU64(buf, fv.Uint())
	case reflect.Float32, reflect.Float64:
		if typed {
			hint = "float"
		}
		buf = appendFloat64(buf, fv.Float())
	case reflect.String:
		if typed {
			hint = "str"
		}
		buf = appendStr(buf, fv.String())
	case reflect.Slice:
		buf = marshalSliceValue(buf, fv)
	case reflect.Map:
		buf = marshalMapValue(buf, fv)
	case reflect.Struct:
		si := getStructInfo(fv.Type())
		buf = marshalStruct(buf, fv, si, false)
	case reflect.Interface:
		if !fv.IsNil() {
			buf = marshalNestedValue(buf, fv.Elem())
		}
	}
	return buf, hint
}

func marshalNestedValue(buf []byte, fv reflect.Value) []byte {
	for fv.Kind() == reflect.Ptr || fv.Kind() == reflect.Interface {
		if fv.IsNil() {
			return buf
		}
		fv = fv.Elem()
	}

	switch fv.Kind() {
	case reflect.Bool:
		if fv.Bool() {
			buf = append(buf, "true"...)
		} else {
			buf = append(buf, "false"...)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		buf = appendI64(buf, fv.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		buf = appendU64(buf, fv.Uint())
	case reflect.Float32, reflect.Float64:
		buf = appendFloat64(buf, fv.Float())
	case reflect.String:
		buf = appendStr(buf, fv.String())
	case reflect.Slice:
		buf = marshalSliceValue(buf, fv)
	case reflect.Map:
		buf = marshalMapValue(buf, fv)
	case reflect.Struct:
		si := getStructInfo(fv.Type())
		buf = marshalStruct(buf, fv, si, false)
	}
	return buf
}

func marshalSliceValue(buf []byte, fv reflect.Value) []byte {
	buf = append(buf, '[')
	for i := 0; i < fv.Len(); i++ {
		if i > 0 {
			buf = append(buf, ',')
		}
		buf = marshalNestedValue(buf, fv.Index(i))
	}
	buf = append(buf, ']')
	return buf
}

func marshalMapValue(buf []byte, fv reflect.Value) []byte {
	buf = append(buf, '[')
	iter := fv.MapRange()
	first := true
	for iter.Next() {
		if !first {
			buf = append(buf, ',')
		}
		first = false
		buf = append(buf, '(')
		buf = marshalNestedValue(buf, iter.Key())
		buf = append(buf, ',')
		buf = marshalNestedValue(buf, iter.Value())
		buf = append(buf, ')')
	}
	buf = append(buf, ']')
	return buf
}

// unsafeString converts a byte slice to string without copying.
func unsafeString(b []byte) string {
	if len(b) == 0 {
		return ""
	}
	return unsafe.String(&b[0], len(b))
}
