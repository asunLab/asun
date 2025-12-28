package ason

import (
	"testing"
)

func TestValueTypes(t *testing.T) {
	// Null
	n := Null()
	if !n.IsNull() {
		t.Error("Expected null")
	}

	// Bool
	b := Bool(true)
	if !b.IsBool() || !b.AsBool() {
		t.Error("Expected true")
	}

	// Integer
	i := Integer(42)
	if !i.IsInteger() || i.AsInteger() != 42 {
		t.Error("Expected 42")
	}

	// Float
	f := Float(3.14)
	if !f.IsFloat() || f.AsFloat() != 3.14 {
		t.Error("Expected 3.14")
	}

	// String
	s := String("hello")
	if !s.IsString() || s.AsString() != "hello" {
		t.Error("Expected hello")
	}

	// Array
	arr := Array(Integer(1), Integer(2), Integer(3))
	if !arr.IsArray() || arr.Len() != 3 {
		t.Error("Expected array with 3 elements")
	}
	if arr.Get(1).AsInteger() != 2 {
		t.Error("Expected arr[1] = 2")
	}

	// Object
	obj := Object()
	obj.Set("name", String("Alice"))
	obj.Set("age", Integer(30))
	if !obj.IsObject() || obj.Len() != 2 {
		t.Error("Expected object with 2 fields")
	}
	if obj.Field("name").AsString() != "Alice" {
		t.Error("Expected name = Alice")
	}
}

func TestParseSimpleObject(t *testing.T) {
	v, err := Parse("{name,age}:(Alice,30)")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if !v.IsObject() {
		t.Fatal("Expected object")
	}
	if v.Field("name").AsString() != "Alice" {
		t.Errorf("Expected name=Alice, got %s", v.Field("name").AsString())
	}
	if v.Field("age").AsInteger() != 30 {
		t.Errorf("Expected age=30, got %d", v.Field("age").AsInteger())
	}
}

func TestParseMultipleObjects(t *testing.T) {
	v, err := Parse("{name,age}:(Alice,30),(Bob,25)")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if !v.IsArray() || v.Len() != 2 {
		t.Fatal("Expected array with 2 objects")
	}
	if v.Get(0).Field("name").AsString() != "Alice" {
		t.Error("Expected first name = Alice")
	}
	if v.Get(1).Field("name").AsString() != "Bob" {
		t.Error("Expected second name = Bob")
	}
}

func TestParseNestedObject(t *testing.T) {
	v, err := Parse("{name,addr{city,zip}}:(Alice,(NYC,10001))")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	if v.Field("name").AsString() != "Alice" {
		t.Error("Expected name = Alice")
	}
	addr := v.Field("addr")
	if !addr.IsObject() {
		t.Fatal("Expected addr to be object")
	}
	if addr.Field("city").AsString() != "NYC" {
		t.Error("Expected city = NYC")
	}
	if addr.Field("zip").AsInteger() != 10001 {
		t.Error("Expected zip = 10001")
	}
}

func TestParseArrayField(t *testing.T) {
	v, err := Parse("{name,scores[]}:(Alice,[90,85,95])")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	scores := v.Field("scores")
	if !scores.IsArray() || scores.Len() != 3 {
		t.Fatal("Expected scores array with 3 elements")
	}
	if scores.Get(0).AsInteger() != 90 {
		t.Error("Expected scores[0] = 90")
	}
}

func TestParseObjectArray(t *testing.T) {
	v, err := Parse("{users[{id,name}]}:([(1,Alice),(2,Bob)])")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	users := v.Field("users")
	if !users.IsArray() || users.Len() != 2 {
		t.Fatal("Expected users array with 2 elements")
	}
	if users.Get(0).Field("id").AsInteger() != 1 {
		t.Error("Expected users[0].id = 1")
	}
	if users.Get(1).Field("name").AsString() != "Bob" {
		t.Error("Expected users[1].name = Bob")
	}
}

func TestSerialize(t *testing.T) {
	obj := Object()
	obj.Set("name", String("Alice"))
	obj.Set("age", Integer(30))
	
	s := Serialize(obj)
	if s != "(Alice,30)" {
		t.Errorf("Expected (Alice,30), got %s", s)
	}
}

func TestSerializeWithSchema(t *testing.T) {
	obj := Object()
	obj.Set("name", String("Alice"))
	obj.Set("age", Integer(30))
	
	s := SerializeWithSchema(obj)
	if s != "{name,age}:(Alice,30)" {
		t.Errorf("Expected {name,age}:(Alice,30), got %s", s)
	}
}

