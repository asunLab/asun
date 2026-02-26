package ason

import (
"reflect"
"testing"
)

type BinUser struct {
	ID     int64
	Name   string
	Active bool
	Tags   []string
	Score  float64
}

func TestBinaryRoundtrip(t *testing.T) {
	u := BinUser{
		ID:     42,
		Name:   "Alice",
		Active: true,
		Tags:   []string{"rust", "go"},
		Score:  9.8,
	}
	b, err := EncodeBinary(&u)
	if err != nil {
		t.Fatal(err)
	}
	var u2 BinUser
	err = DecodeBinary(b, &u2)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(u, u2) {
		t.Fatalf("expected %+v, got %+v", u, u2)
	}
}
