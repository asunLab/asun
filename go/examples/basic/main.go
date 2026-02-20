package main

import (
	"fmt"
	"log"

	ason "github.com/example/ason"
)

type User struct {
	ID     int64  `ason:"id"`
	Name   string `ason:"name"`
	Active bool   `ason:"active"`
}

func main() {
	fmt.Println("=== ASON Basic Examples ===")
	fmt.Println()

	// 1. Serialize a single struct
	user := User{ID: 1, Name: "Alice", Active: true}
	b, err := ason.Marshal(&user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Serialize single struct:")
	fmt.Printf("  %s\n\n", b)

	// 2. Serialize with type annotations (MarshalTyped)
	typed, err := ason.MarshalTyped(&user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Serialize with type annotations:")
	fmt.Printf("  %s\n\n", typed)

	// 3. Deserialize from ASON (accepts both annotated and unannotated)
	input := []byte("{id:int,name:str,active:bool}:(1,Alice,true)")
	var u User
	if err := ason.Unmarshal(input, &u); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deserialize single struct:")
	fmt.Printf("  %+v\n\n", u)

	// 4. Serialize a vec of structs
	users := []User{
		{ID: 1, Name: "Alice", Active: true},
		{ID: 2, Name: "Bob", Active: false},
		{ID: 3, Name: "Carol Smith", Active: true},
	}
	vecBytes, err := ason.MarshalSlice(users)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Serialize vec (schema-driven):")
	fmt.Printf("  %s\n\n", vecBytes)

	// 5. Serialize vec with type annotations
	typedVec, err := ason.MarshalSliceTyped(users, []string{"int", "str", "bool"})
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Serialize vec with type annotations:")
	fmt.Printf("  %s\n\n", typedVec)

	// 6. Deserialize vec
	vecInput := []byte(`{id:int,name:str,active:bool}:(1,Alice,true),(2,Bob,false),(3,"Carol Smith",true)`)
	var parsed []User
	if err := ason.UnmarshalSlice(vecInput, &parsed); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deserialize vec:")
	for _, u := range parsed {
		fmt.Printf("  %+v\n", u)
	}

	// 7. Multiline format
	fmt.Println("\nMultiline format:")
	multiline := []byte(`{id:int, name:str, active:bool}:
  (1, Alice, true),
  (2, Bob, false),
  (3, "Carol Smith", true)`)
	var multi []User
	if err := ason.UnmarshalSlice(multiline, &multi); err != nil {
		log.Fatal(err)
	}
	for _, u := range multi {
		fmt.Printf("  %+v\n", u)
	}

	// 8. Roundtrip
	fmt.Println("\nRoundtrip test:")
	original := User{ID: 42, Name: "Test User", Active: true}
	serialized, err := ason.Marshal(&original)
	if err != nil {
		log.Fatal(err)
	}
	var deserialized User
	if err := ason.Unmarshal(serialized, &deserialized); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  original:     %+v\n", original)
	fmt.Printf("  serialized:   %s\n", serialized)
	fmt.Printf("  deserialized: %+v\n", deserialized)
	if original != deserialized {
		log.Fatal("roundtrip mismatch")
	}
	fmt.Println("  ✓ roundtrip OK")

	// 9. Optional fields
	fmt.Println("\nOptional fields:")
	type Item struct {
		ID    int64   `ason:"id"`
		Label *string `ason:"label"`
	}
	var item Item
	if err := ason.Unmarshal([]byte("{id,label}:(1,hello)"), &item); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  with value: %+v (label=%s)\n", item, *item.Label)

	var item2 Item
	if err := ason.Unmarshal([]byte("{id,label}:(2,)"), &item2); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  with null:  %+v\n", item2)

	// 10. Array fields
	fmt.Println("\nArray fields:")
	type Tagged struct {
		Name string   `ason:"name"`
		Tags []string `ason:"tags"`
	}
	var t Tagged
	if err := ason.Unmarshal([]byte("{name,tags}:(Alice,[rust,go,python])"), &t); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  %+v\n", t)

	// 11. Comments
	fmt.Println("\nWith comments:")
	var commented User
	if err := ason.Unmarshal([]byte("/* user list */ {id,name,active}:(1,Alice,true)"), &commented); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  %+v\n", commented)

	fmt.Println("\n=== All examples passed! ===")
}
