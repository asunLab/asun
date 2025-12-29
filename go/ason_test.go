package ason

import (
	"strings"
	"testing"
)

type User struct {
	Name string `json:"name"`
	Age  int    `json:"age"`
}

type Address struct {
	City string `json:"city"`
	Zip  int    `json:"zip"`
}

type Person struct {
	Name string  `json:"name"`
	Addr Address `json:"addr"`
}

type Order struct {
	ID    string `json:"id"`
	Items []Item `json:"items"`
}

type Item struct {
	Name string `json:"name"`
	Qty  int    `json:"qty"`
}

func TestMarshalSimple(t *testing.T) {
	user := User{Name: "Alice", Age: 30}
	s, err := Marshal(user)
	if err != nil {
		t.Fatal(err)
	}
	expected := "{name,age}:(Alice,30)"
	if s != expected {
		t.Errorf("got %q, want %q", s, expected)
	}
}

func TestMarshalNested(t *testing.T) {
	person := Person{Name: "Alice", Addr: Address{City: "NYC", Zip: 10001}}
	s, err := Marshal(person)
	if err != nil {
		t.Fatal(err)
	}
	expected := "{name,addr{city,zip}}:(Alice,(NYC,10001))"
	if s != expected {
		t.Errorf("got %q, want %q", s, expected)
	}
}

func TestMarshalSlice(t *testing.T) {
	users := []User{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
	}
	s, err := Marshal(users)
	if err != nil {
		t.Fatal(err)
	}
	expected := "{name,age}:(Alice,30),(Bob,25)"
	if s != expected {
		t.Errorf("got %q, want %q", s, expected)
	}
}

func TestUnmarshalSimple(t *testing.T) {
	var user User
	err := Unmarshal("{name,age}:(Alice,30)", &user)
	if err != nil {
		t.Fatal(err)
	}
	if user.Name != "Alice" || user.Age != 30 {
		t.Errorf("got %+v", user)
	}
}

func TestUnmarshalNested(t *testing.T) {
	var person Person
	err := Unmarshal("{name,addr{city,zip}}:(Alice,(NYC,10001))", &person)
	if err != nil {
		t.Fatal(err)
	}
	if person.Name != "Alice" || person.Addr.City != "NYC" || person.Addr.Zip != 10001 {
		t.Errorf("got %+v", person)
	}
}

func TestUnmarshalSlice(t *testing.T) {
	var users []User
	err := Unmarshal("{name,age}:(Alice,30),(Bob,25)", &users)
	if err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 {
		t.Fatalf("got %d users, want 2", len(users))
	}
	if users[0].Name != "Alice" || users[0].Age != 30 {
		t.Errorf("users[0] = %+v", users[0])
	}
	if users[1].Name != "Bob" || users[1].Age != 25 {
		t.Errorf("users[1] = %+v", users[1])
	}
}

func TestRoundTrip(t *testing.T) {
	tests := []interface{}{
		User{Name: "Alice", Age: 30},
		Person{Name: "Bob", Addr: Address{City: "LA", Zip: 90001}},
	}

	for _, orig := range tests {
		s, err := Marshal(orig)
		if err != nil {
			t.Fatal(err)
		}

		switch v := orig.(type) {
		case User:
			var parsed User
			if err := Unmarshal(s, &parsed); err != nil {
				t.Fatal(err)
			}
			if parsed != v {
				t.Errorf("round-trip failed: got %+v, want %+v", parsed, v)
			}
		case Person:
			var parsed Person
			if err := Unmarshal(s, &parsed); err != nil {
				t.Fatal(err)
			}
			if parsed != v {
				t.Errorf("round-trip failed: got %+v, want %+v", parsed, v)
			}
		}
	}
}

func TestMarshalTypes(t *testing.T) {
	type AllTypes struct {
		B bool    `json:"b"`
		I int     `json:"i"`
		F float64 `json:"f"`
		S string  `json:"s"`
	}

	v := AllTypes{B: true, I: 42, F: 3.14, S: "hello"}
	s, err := Marshal(v)
	if err != nil {
		t.Fatal(err)
	}
	expected := "{b,i,f,s}:(true,42,3.14,hello)"
	if s != expected {
		t.Errorf("got %q, want %q", s, expected)
	}
}

func TestMarshalIndent(t *testing.T) {
	user := User{Name: "Alice", Age: 30}
	s, err := MarshalIndent(user, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	// Just check it contains newlines and indent
	if !strings.Contains(s, "\n") {
		t.Errorf("expected newlines in output: %q", s)
	}
	if !strings.Contains(s, "  ") {
		t.Errorf("expected indentation in output: %q", s)
	}
}
