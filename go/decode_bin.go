package ason

import (
	"encoding/binary"
	"math"
	"reflect"
	"unsafe"
)

// DecodeBinary deserializes ASON-BIN format into a Go value.
// It uses zero-copy for strings where possible.
func DecodeBinary(data []byte, v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr || rv.IsNil() {
		return &UnmarshalError{Message: "unmarshal target must be a non-nil pointer"}
	}
	_, err := unmarshalBinValue(data, rv.Elem())
	return err
}

func unmarshalBinValue(data []byte, rv reflect.Value) ([]byte, error) {
	switch rv.Kind() {
	case reflect.Bool:
		if len(data) < 1 {
			return data, &UnmarshalError{Pos: 0, Message: "unexpected EOF reading bool"}
		}
		rv.SetBool(data[0] != 0)
		return data[1:], nil
	case reflect.Int8:
		if len(data) < 1 {
			return data, &UnmarshalError{Pos: 0, Message: "unexpected EOF reading int8"}
		}
		rv.SetInt(int64(int8(data[0])))
		return data[1:], nil
	case reflect.Int16:
		if len(data) < 2 {
			return data, &UnmarshalError{Pos: 0, Message: "unexpected EOF reading int16"}
		}
		rv.SetInt(int64(int16(binary.LittleEndian.Uint16(data))))
		return data[2:], nil
	case reflect.Int32:
		if len(data) < 4 {
			return data, &UnmarshalError{Pos: 0, Message: "unexpected EOF reading int32"}
		}
		rv.SetInt(int64(int32(binary.LittleEndian.Uint32(data))))
		return data[4:], nil
	case reflect.Int, reflect.Int64:
		if len(data) < 8 {
			return data, &UnmarshalError{Pos: 0, Message: "unexpected EOF reading int64"}
		}
		rv.SetInt(int64(binary.LittleEndian.Uint64(data)))
		return data[8:], nil
	case reflect.Uint8:
		if len(data) < 1 {
			return data, &UnmarshalError{Pos: 0, Message: "unexpected EOF reading uint8"}
		}
		rv.SetUint(uint64(data[0]))
		return data[1:], nil
	case reflect.Uint16:
		if len(data) < 2 {
			return data, &UnmarshalError{Pos: 0, Message: "unexpected EOF reading uint16"}
		}
		rv.SetUint(uint64(binary.LittleEndian.Uint16(data)))
		return data[2:], nil
	case reflect.Uint32:
		if len(data) < 4 {
			return data, &UnmarshalError{Pos: 0, Message: "unexpected EOF reading uint32"}
		}
		rv.SetUint(uint64(binary.LittleEndian.Uint32(data)))
		return data[4:], nil
	case reflect.Uint, reflect.Uint64:
		if len(data) < 8 {
			return data, &UnmarshalError{Pos: 0, Message: "unexpected EOF reading uint64"}
		}
		rv.SetUint(binary.LittleEndian.Uint64(data))
		return data[8:], nil
	case reflect.Float32:
		if len(data) < 4 {
			return data, &UnmarshalError{Pos: 0, Message: "unexpected EOF reading float32"}
		}
		rv.SetFloat(float64(math.Float32frombits(binary.LittleEndian.Uint32(data))))
		return data[4:], nil
	case reflect.Float64:
		if len(data) < 8 {
			return data, &UnmarshalError{Pos: 0, Message: "unexpected EOF reading float64"}
		}
		rv.SetFloat(math.Float64frombits(binary.LittleEndian.Uint64(data)))
		return data[8:], nil
	case reflect.String:
		if len(data) < 4 {
			return data, &UnmarshalError{Pos: 0, Message: "unexpected EOF reading string length"}
		}
		n := int(binary.LittleEndian.Uint32(data))
		data = data[4:]
		if len(data) < n {
			return data, &UnmarshalError{Pos: 0, Message: "unexpected EOF reading string data"}
		}
		// Zero-copy string creation
		if n == 0 {
			rv.SetString("")
		} else {
			b := data[:n]
			s := unsafe.String(unsafe.SliceData(b), len(b))
			rv.SetString(s)
		}
		return data[n:], nil
	case reflect.Slice:
		if len(data) < 4 {
			return data, &UnmarshalError{Pos: 0, Message: "unexpected EOF reading slice length"}
		}
		n := int(binary.LittleEndian.Uint32(data))
		data = data[4:]
		
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			if len(data) < n {
				return data, &UnmarshalError{Pos: 0, Message: "unexpected EOF reading bytes"}
			}
			// Zero-copy byte slice
			rv.SetBytes(data[:n:n])
			return data[n:], nil
		}
		
		if rv.IsNil() || rv.Cap() < n {
			rv.Set(reflect.MakeSlice(rv.Type(), n, n))
		} else {
			rv.SetLen(n)
		}
		var err error
		for i := 0; i < n; i++ {
			data, err = unmarshalBinValue(data, rv.Index(i))
			if err != nil {
				return data, err
			}
		}
		return data, nil
	case reflect.Array:
		if len(data) < 4 {
			return data, &UnmarshalError{Pos: 0, Message: "unexpected EOF reading array length"}
		}
		n := int(binary.LittleEndian.Uint32(data))
		data = data[4:]
		
		limit := rv.Len()
		var err error
		for i := 0; i < n; i++ {
			if i < limit {
				data, err = unmarshalBinValue(data, rv.Index(i))
			} else {
				// Skip extra elements
				// We need a dummy value to skip, but we don't know the type size easily without parsing.
				// For now, we just parse into a new value of the element type.
				dummy := reflect.New(rv.Type().Elem()).Elem()
				data, err = unmarshalBinValue(data, dummy)
			}
			if err != nil {
				return data, err
			}
		}
		return data, nil
	case reflect.Map:
		if len(data) < 4 {
			return data, &UnmarshalError{Pos: 0, Message: "unexpected EOF reading map length"}
		}
		n := int(binary.LittleEndian.Uint32(data))
		data = data[4:]
		
		if rv.IsNil() {
			rv.Set(reflect.MakeMapWithSize(rv.Type(), n))
		}
		
		keyType := rv.Type().Key()
		valType := rv.Type().Elem()
		var err error
		for i := 0; i < n; i++ {
			k := reflect.New(keyType).Elem()
			data, err = unmarshalBinValue(data, k)
			if err != nil {
				return data, err
			}
			v := reflect.New(valType).Elem()
			data, err = unmarshalBinValue(data, v)
			if err != nil {
				return data, err
			}
			rv.SetMapIndex(k, v)
		}
		return data, nil
	case reflect.Struct:
		si := getStructInfo(rv.Type())
		var err error
		for _, f := range si.fields {
			fv := rv.FieldByIndex(f.index)
			data, err = unmarshalBinValue(data, fv)
			if err != nil {
				return data, err
			}
		}
		return data, nil
	case reflect.Ptr:
		if len(data) < 1 {
			return data, &UnmarshalError{Pos: 0, Message: "unexpected EOF reading ptr tag"}
		}
		tag := data[0]
		data = data[1:]
		if tag == 0 {
			rv.SetZero()
			return data, nil
		}
		if rv.IsNil() {
			rv.Set(reflect.New(rv.Type().Elem()))
		}
		return unmarshalBinValue(data, rv.Elem())
	case reflect.Interface:
		// Interface decoding is tricky without type info.
		// In ASON-BIN, we don't encode type info for interfaces.
		// If the interface is nil, we can't decode it.
		// If it's not nil, we decode into the existing value.
		if len(data) < 1 {
			return data, &UnmarshalError{Pos: 0, Message: "unexpected EOF reading interface tag"}
		}
		tag := data[0]
		data = data[1:]
		if tag == 0 {
			rv.SetZero()
			return data, nil
		}
		if rv.IsNil() {
			return data, &UnmarshalError{Message: "cannot unmarshal into nil interface"}
		}
		// We need to decode into the concrete type
		elem := rv.Elem()
		if elem.Kind() == reflect.Ptr && !elem.IsNil() {
			return unmarshalBinValue(data, elem)
		}
		// If it's not a pointer, we can't set it directly. We need to create a new pointer,
		// decode into it, and set the interface.
		ptr := reflect.New(elem.Type())
		ptr.Elem().Set(elem)
		data, err := unmarshalBinValue(data, ptr.Elem())
		if err != nil {
			return data, err
		}
		rv.Set(ptr.Elem())
		return data, nil
	default:
		return data, &UnmarshalError{Message: "unsupported type"}
	}
}
