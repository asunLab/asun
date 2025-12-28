// Package ason implements ASON (Array-Schema Object Notation) format.
package ason

import (
	"fmt"
	"strconv"
)

// ValueType represents the type of an ASON value.
type ValueType int

const (
	TypeNull ValueType = iota
	TypeBool
	TypeInteger
	TypeFloat
	TypeString
	TypeArray
	TypeObject
)

// String returns the string representation of the value type.
func (t ValueType) String() string {
	switch t {
	case TypeNull:
		return "null"
	case TypeBool:
		return "bool"
	case TypeInteger:
		return "integer"
	case TypeFloat:
		return "float"
	case TypeString:
		return "string"
	case TypeArray:
		return "array"
	case TypeObject:
		return "object"
	default:
		return "unknown"
	}
}

// Value represents an ASON value.
type Value struct {
	typ      ValueType
	boolVal  bool
	intVal   int64
	floatVal float64
	strVal   string
	arrVal   []*Value
	objKeys  []string
	objVals  map[string]*Value
}

// Null creates a null value.
func Null() *Value {
	return &Value{typ: TypeNull}
}

// Bool creates a boolean value.
func Bool(v bool) *Value {
	return &Value{typ: TypeBool, boolVal: v}
}

// Integer creates an integer value.
func Integer(v int64) *Value {
	return &Value{typ: TypeInteger, intVal: v}
}

// Float creates a float value.
func Float(v float64) *Value {
	return &Value{typ: TypeFloat, floatVal: v}
}

// String creates a string value.
func String(v string) *Value {
	return &Value{typ: TypeString, strVal: v}
}

// Array creates an empty array value.
func Array(items ...*Value) *Value {
	arr := &Value{typ: TypeArray, arrVal: make([]*Value, 0)}
	for _, item := range items {
		arr.arrVal = append(arr.arrVal, item)
	}
	return arr
}

// Object creates an empty object value.
func Object() *Value {
	return &Value{
		typ:     TypeObject,
		objKeys: make([]string, 0),
		objVals: make(map[string]*Value),
	}
}

// Type returns the type of the value.
func (v *Value) Type() ValueType {
	if v == nil {
		return TypeNull
	}
	return v.typ
}

// IsNull returns true if the value is null.
func (v *Value) IsNull() bool { return v == nil || v.typ == TypeNull }

// IsBool returns true if the value is a boolean.
func (v *Value) IsBool() bool { return v != nil && v.typ == TypeBool }

// IsInteger returns true if the value is an integer.
func (v *Value) IsInteger() bool { return v != nil && v.typ == TypeInteger }

// IsFloat returns true if the value is a float.
func (v *Value) IsFloat() bool { return v != nil && v.typ == TypeFloat }

// IsNumber returns true if the value is a number (integer or float).
func (v *Value) IsNumber() bool { return v.IsInteger() || v.IsFloat() }

// IsString returns true if the value is a string.
func (v *Value) IsString() bool { return v != nil && v.typ == TypeString }

// IsArray returns true if the value is an array.
func (v *Value) IsArray() bool { return v != nil && v.typ == TypeArray }

// IsObject returns true if the value is an object.
func (v *Value) IsObject() bool { return v != nil && v.typ == TypeObject }

// AsBool returns the boolean value.
func (v *Value) AsBool() bool {
	if v == nil || v.typ != TypeBool {
		return false
	}
	return v.boolVal
}

// AsInteger returns the integer value.
func (v *Value) AsInteger() int64 {
	if v == nil {
		return 0
	}
	switch v.typ {
	case TypeInteger:
		return v.intVal
	case TypeFloat:
		return int64(v.floatVal)
	default:
		return 0
	}
}

// AsFloat returns the float value.
func (v *Value) AsFloat() float64 {
	if v == nil {
		return 0
	}
	switch v.typ {
	case TypeFloat:
		return v.floatVal
	case TypeInteger:
		return float64(v.intVal)
	default:
		return 0
	}
}

// AsString returns the string value.
func (v *Value) AsString() string {
	if v == nil || v.typ != TypeString {
		return ""
	}
	return v.strVal
}

// Len returns the length of an array or object.
func (v *Value) Len() int {
	if v == nil {
		return 0
	}
	switch v.typ {
	case TypeArray:
		return len(v.arrVal)
	case TypeObject:
		return len(v.objKeys)
	default:
		return 0
	}
}

// Get returns an element from an array by index.
func (v *Value) Get(index int) *Value {
	if v == nil || v.typ != TypeArray || index < 0 || index >= len(v.arrVal) {
		return nil
	}
	return v.arrVal[index]
}

// Push appends a value to an array.
func (v *Value) Push(item *Value) {
	if v == nil || v.typ != TypeArray {
		return
	}
	v.arrVal = append(v.arrVal, item)
}

// Items returns all items in an array.
func (v *Value) Items() []*Value {
	if v == nil || v.typ != TypeArray {
		return nil
	}
	return v.arrVal
}

// Field returns a field from an object by key.
func (v *Value) Field(key string) *Value {
	if v == nil || v.typ != TypeObject {
		return nil
	}
	return v.objVals[key]
}

// Set sets a field in an object.
func (v *Value) Set(key string, val *Value) {
	if v == nil || v.typ != TypeObject {
		return
	}
	if _, exists := v.objVals[key]; !exists {
		v.objKeys = append(v.objKeys, key)
	}
	v.objVals[key] = val
}

// Keys returns all keys in an object (in insertion order).
func (v *Value) Keys() []string {
	if v == nil || v.typ != TypeObject {
		return nil
	}
	return v.objKeys
}

// Clone creates a deep copy of the value.
func (v *Value) Clone() *Value {
	if v == nil {
		return nil
	}
	switch v.typ {
	case TypeNull:
		return Null()
	case TypeBool:
		return Bool(v.boolVal)
	case TypeInteger:
		return Integer(v.intVal)
	case TypeFloat:
		return Float(v.floatVal)
	case TypeString:
		return String(v.strVal)
	case TypeArray:
		arr := Array()
		for _, item := range v.arrVal {
			arr.Push(item.Clone())
		}
		return arr
	case TypeObject:
		obj := Object()
		for _, key := range v.objKeys {
			obj.Set(key, v.objVals[key].Clone())
		}
		return obj
	default:
		return nil
	}
}

// GoString implements fmt.GoStringer for debugging.
func (v *Value) GoString() string {
	if v == nil {
		return "nil"
	}
	switch v.typ {
	case TypeNull:
		return "null"
	case TypeBool:
		return strconv.FormatBool(v.boolVal)
	case TypeInteger:
		return strconv.FormatInt(v.intVal, 10)
	case TypeFloat:
		return strconv.FormatFloat(v.floatVal, 'g', -1, 64)
	case TypeString:
		return fmt.Sprintf("%q", v.strVal)
	case TypeArray:
		return fmt.Sprintf("Array(%d)", len(v.arrVal))
	case TypeObject:
		return fmt.Sprintf("Object(%d)", len(v.objKeys))
	default:
		return "unknown"
	}
}
