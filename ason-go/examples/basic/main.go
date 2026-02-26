package main

import (
	"encoding/json"
	"fmt"
	"log"

	ason "github.com/example/ason"
)

type User struct {
	ID     int64  `ason:"id" json:"id"`
	Name   string `ason:"name" json:"name"`
	Active bool   `ason:"active" json:"active"`
}

func main() {
	fmt.Println("=== ASON Basic Examples ===")
	fmt.Println()

	// 1. Serialize a single struct
	user := User{ID: 1, Name: "Alice", Active: true}
	b, err := ason.Encode(&user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Serialize single struct:")
	fmt.Printf("  %s\n\n", b)

	// 2. Serialize with type annotations (MarshalTyped)
	typed, err := ason.EncodeTyped(&user)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Serialize with type annotations:")
	fmt.Printf("  %s\n\n", typed)

	// 3. Deserialize from ASON (accepts both annotated and unannotated)
	input := []byte("{id:int,name:str,active:bool}:(1,Alice,true)")
	var u User
	if err := ason.Decode(input, &u); err != nil {
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
	vecBytes, err := ason.Encode(users)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Serialize vec (schema-driven):")
	fmt.Printf("  %s\n\n", vecBytes)

	// 5. Serialize vec with type annotations
	typedVec, err := ason.EncodeTyped(users)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Serialize vec with type annotations:")
	fmt.Printf("  %s\n\n", typedVec)

	// 6. Deserialize vec
	vecInput := []byte(`[{id:int,name:str,active:bool}]:(1,Alice,true),(2,Bob,false),(3,"Carol Smith",true)`)
	var parsed []User
	if err := ason.Decode(vecInput, &parsed); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Deserialize vec:")
	for _, u := range parsed {
		fmt.Printf("  %+v\n", u)
	}

	// 7. Multiline format
	fmt.Println("\nMultiline format:")
	multiline := []byte(`[{id:int, name:str, active:bool}]:
  (1, Alice, true),
  (2, Bob, false),
  (3, "Carol Smith", true)`)
	var multi []User
	if err := ason.Decode(multiline, &multi); err != nil {
		log.Fatal(err)
	}
	for _, u := range multi {
		fmt.Printf("  %+v\n", u)
	}

	// 8. Roundtrip (ASON-text vs ASON-bin vs JSON)
	fmt.Println("\n8. Roundtrip (ASON-text vs ASON-bin vs JSON):")
	original := User{ID: 42, Name: "Test User", Active: true}
	// ASON text
	asonText, err := ason.Encode(&original)
	if err != nil {
		log.Fatal(err)
	}
	var fromAson User
	if err := ason.Decode(asonText, &fromAson); err != nil {
		log.Fatal(err)
	}
	if original != fromAson {
		log.Fatal("ASON text roundtrip mismatch")
	}
	// ASON binary
	asonBin, err := ason.EncodeBinary(&original)
	if err != nil {
		log.Fatal(err)
	}
	var fromBin User
	if err := ason.DecodeBinary(asonBin, &fromBin); err != nil {
		log.Fatal(err)
	}
	if original != fromBin {
		log.Fatal("ASON binary roundtrip mismatch")
	}
	// JSON
	jsonData, err := json.Marshal(&original)
	if err != nil {
		log.Fatal(err)
	}
	var fromJSON User
	if err := json.Unmarshal(jsonData, &fromJSON); err != nil {
		log.Fatal(err)
	}
	if original != fromJSON {
		log.Fatal("JSON roundtrip mismatch")
	}
	fmt.Printf("  original:     %+v\n", original)
	fmt.Printf("  ASON text:    %s (%d B)\n", asonText, len(asonText))
	fmt.Printf("  ASON binary:  %d B\n", len(asonBin))
	fmt.Printf("  JSON:         %s (%d B)\n", jsonData, len(jsonData))
	fmt.Println("  ✓ all 3 formats roundtrip OK")

	// 9. Optional fields
	fmt.Println("\n9. Optional fields:")
	type Item struct {
		ID    int64   `ason:"id" json:"id"`
		Label *string `ason:"label" json:"label"`
	}
	var item Item
	if err := ason.Decode([]byte("{id,label}:(1,hello)"), &item); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  with value: %+v (label=%s)\n", item, *item.Label)

	var item2 Item
	if err := ason.Decode([]byte("{id,label}:(2,)"), &item2); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  with null:  %+v\n", item2)

	// 10. Array fields
	fmt.Println("\n10. Array fields:")
	type Tagged struct {
		Name string   `ason:"name" json:"name"`
		Tags []string `ason:"tags" json:"tags"`
	}
	var t Tagged
	if err := ason.Decode([]byte("{name,tags}:(Alice,[rust,go,python])"), &t); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  %+v\n", t)

	// 11. Comments
	fmt.Println("\n11. With comments:")
	var commented User
	if err := ason.Decode([]byte("/* user list */ {id,name,active}:(1,Alice,true)"), &commented); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("  %+v\n", commented)

	fmt.Println("\n=== All examples passed! ===")
}
