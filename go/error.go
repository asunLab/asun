package ason

import "fmt"

// MarshalError is returned when serialization fails.
type MarshalError struct {
	Message string
}

func (e *MarshalError) Error() string {
	return "ason: marshal error: " + e.Message
}

// UnmarshalError is returned when deserialization fails.
type UnmarshalError struct {
	Pos     int
	Message string
}

func (e *UnmarshalError) Error() string {
	return fmt.Sprintf("ason: unmarshal error at pos %d: %s", e.Pos, e.Message)
}
