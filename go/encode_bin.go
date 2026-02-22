package ason

import (
	"encoding/binary"
	"math"
	"reflect"
)

// EncodeBinary serializes a Go value to ASON-BIN format.
func EncodeBinary(v any) ([]byte, error) {
	if v == nil {
		return nil, &MarshalError{Message: "cannot marshal nil"}
	}
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		if rv.IsNil() {
			return nil, &MarshalError{Message: "cannot marshal nil pointer"}
		}
		rv = rv.Elem()
	}

	var buf []byte
	buf = marshalBinValue(buf, rv)
	return buf, nil
}

func marshalBinValue(buf []byte, rv reflect.Value) []byte {
	switch rv.Kind() {
	case reflect.Bool:
		if rv.Bool() {
			return append(buf, 1)
		}
		return append(buf, 0)
	case reflect.Int8:
		return append(buf, byte(rv.Int()))
	case reflect.Int16:
		buf = binary.LittleEndian.AppendUint16(buf, uint16(rv.Int()))
		return buf
	case reflect.Int32:
		buf = binary.LittleEndian.AppendUint32(buf, uint32(rv.Int()))
		return buf
	case reflect.Int, reflect.Int64:
		buf = binary.LittleEndian.AppendUint64(buf, uint64(rv.Int()))
		return buf
	case reflect.Uint8:
		return append(buf, byte(rv.Uint()))
	case reflect.Uint16:
		buf = binary.LittleEndian.AppendUint16(buf, uint16(rv.Uint()))
		return buf
	case reflect.Uint32:
		buf = binary.LittleEndian.AppendUint32(buf, uint32(rv.Uint()))
		return buf
	case reflect.Uint, reflect.Uint64:
		buf = binary.LittleEndian.AppendUint64(buf, uint64(rv.Uint()))
		return buf
	case reflect.Float32:
		buf = binary.LittleEndian.AppendUint32(buf, math.Float32bits(float32(rv.Float())))
		return buf
	case reflect.Float64:
		buf = binary.LittleEndian.AppendUint64(buf, math.Float64bits(rv.Float()))
		return buf
	case reflect.String:
		s := rv.String()
		buf = binary.LittleEndian.AppendUint32(buf, uint32(len(s)))
		buf = append(buf, s...)
		return buf
	case reflect.Slice:
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			b := rv.Bytes()
			buf = binary.LittleEndian.AppendUint32(buf, uint32(len(b)))
			buf = append(buf, b...)
			return buf
		}
		n := rv.Len()
		buf = binary.LittleEndian.AppendUint32(buf, uint32(n))
		for i := 0; i < n; i++ {
			buf = marshalBinValue(buf, rv.Index(i))
		}
		return buf
	case reflect.Array:
		n := rv.Len()
		buf = binary.LittleEndian.AppendUint32(buf, uint32(n))
		for i := 0; i < n; i++ {
			buf = marshalBinValue(buf, rv.Index(i))
		}
		return buf
	case reflect.Map:
		n := rv.Len()
		buf = binary.LittleEndian.AppendUint32(buf, uint32(n))
		iter := rv.MapRange()
		for iter.Next() {
			buf = marshalBinValue(buf, iter.Key())
			buf = marshalBinValue(buf, iter.Value())
		}
		return buf
	case reflect.Struct:
		si := getStructInfo(rv.Type())
		for _, f := range si.fields {
			fv := rv.FieldByIndex(f.index)
			buf = marshalBinValue(buf, fv)
		}
		return buf
	case reflect.Ptr:
		if rv.IsNil() {
			return append(buf, 0)
		}
		buf = append(buf, 1)
		return marshalBinValue(buf, rv.Elem())
	case reflect.Interface:
		if rv.IsNil() {
			return append(buf, 0)
		}
		buf = append(buf, 1)
		return marshalBinValue(buf, rv.Elem())
	default:
		return buf
	}
}
