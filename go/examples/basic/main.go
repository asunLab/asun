package main

import (
	"fmt"

	ason "github.com/athxx/ason/go"
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

func main() {
	fmt.Println("=== ASON Go Examples ===\n")

	// 1. Simple struct - Marshal
	fmt.Println("1. Simple struct (Marshal)")
	user := User{Name: "Alice", Age: 30}
	s, _ := ason.Marshal(user)
	fmt.Println("   Input:", user)
	fmt.Println("   Output:", s)

	// 2. Simple struct - Unmarshal
	fmt.Println("\n2. Simple struct (Unmarshal)")
	input := "{name,age}:(Bob,25)"
	var parsed User
	ason.Unmarshal(input, &parsed)
	fmt.Println("   Input:", input)
	fmt.Printf("   Output: %+v\n", parsed)

	// 3. Nested struct - Marshal
	fmt.Println("\n3. Nested struct (Marshal)")
	person := Person{Name: "Alice", Addr: Address{City: "NYC", Zip: 10001}}
	s, _ = ason.Marshal(person)
	fmt.Println("   Input:", person)
	fmt.Println("   Output:", s)

	// 4. Nested struct - Unmarshal
	fmt.Println("\n4. Nested struct (Unmarshal)")
	input = "{name,addr{city,zip}}:(Charlie,(LA,90001))"
	var parsedPerson Person
	ason.Unmarshal(input, &parsedPerson)
	fmt.Println("   Input:", input)
	fmt.Printf("   Output: %+v\n", parsedPerson)

	// 5. Array of structs - Marshal
	fmt.Println("\n5. Array of structs (Marshal)")
	users := []User{
		{Name: "Alice", Age: 30},
		{Name: "Bob", Age: 25},
	}
	s, _ = ason.Marshal(users)
	fmt.Println("   Input:", users)
	fmt.Println("   Output:", s)

	// 6. Array of structs - Unmarshal
	fmt.Println("\n6. Array of structs (Unmarshal)")
	input = "{name,age}:(David,40),(Eve,35),(Frank,28)"
	var parsedUsers []User
	ason.Unmarshal(input, &parsedUsers)
	fmt.Println("   Input:", input)
	fmt.Println("   Output:")
	for _, u := range parsedUsers {
		fmt.Printf("     - %s is %d years old\n", u.Name, u.Age)
	}

	// 7. Data only (no schema)
	fmt.Println("\n7. Data only (no schema)")
	s, _ = ason.MarshalData(user)
	fmt.Println("   Input:", user)
	fmt.Println("   Output:", s)

	// 8. Pretty print with indent
	fmt.Println("\n8. Pretty print (MarshalIndent)")
	s, _ = ason.MarshalIndent(user, "", "  ")
	fmt.Println("   Output:")
	fmt.Println(s)

	// 9. Pretty print with indent
	fmt.Println("\n9. Pretty print (MarshalIndent) Complex")
	persons := []Person{
		{Name: "Alice", Addr: Address{City: "NYC", Zip: 10001}},
		{Name: "Bob", Addr: Address{City: "LA", Zip: 90001}},
		{Name: "Carol", Addr: Address{City: "SF", Zip: 94101}},
		{Name: "Dave", Addr: Address{City: "Boston", Zip: 2121}},
		{Name: "Eve", Addr: Address{City: "Chicago", Zip: 60601}},
		{Name: "Frank", Addr: Address{City: "Houston", Zip: 77001}},
	}
	s, _ = ason.MarshalIndent(persons, "", "  ")
	fmt.Println("   Output:")
	fmt.Println(s)

	// 10. Round-trip test
	fmt.Println("\n10. Round-trip test")
	original := User{Name: "Test", Age: 99}
	s, _ = ason.Marshal(original)
	var roundTrip User
	ason.Unmarshal(s, &roundTrip)
	fmt.Printf("   Original:   %+v\n", original)
	fmt.Printf("   Serialized: %s\n", s)
	fmt.Printf("   Parsed:     %+v\n", roundTrip)
	fmt.Printf("   Match:      %v\n", original == roundTrip)

	fmt.Println("\n=== Examples Complete ===")
}
