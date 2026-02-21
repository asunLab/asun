package main

import (
	"encoding/json"
	"fmt"
	"reflect"
	"time"

	ason "github.com/example/ason"
)

type User struct {
	ID     int64   `ason:"id" json:"id"`
	Name   string  `ason:"name" json:"name"`
	Email  string  `ason:"email" json:"email"`
	Age    int64   `ason:"age" json:"age"`
	Score  float64 `ason:"score" json:"score"`
	Active bool    `ason:"active" json:"active"`
	Role   string  `ason:"role" json:"role"`
	City   string  `ason:"city" json:"city"`
}

func generateUsers(count int) []User {
	users := make([]User, count)
	for i := 0; i < count; i++ {
		users[i] = User{
			ID:     int64(i),
			Name:   fmt.Sprintf("User%d", i),
			Email:  fmt.Sprintf("user%d@example.com", i),
			Age:    int64(20 + (i % 40)),
			Score:  float64(i%100) + 0.5,
			Active: i%2 == 0,
			Role:   "user",
			City:   "City",
		}
	}
	return users
}

func main() {
	fmt.Println("=== ASON-BIN Go Examples & Benchmarks ===")

	// 1. Basic Usage
	fmt.Println("\n--- 1. Basic Usage ---")
	u := User{ID: 1, Name: "Alice", Active: true, Score: 9.5}
	b, err := ason.MarshalBinary(&u)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Serialized User to %d bytes\n", len(b))

	var u2 User
	err = ason.UnmarshalBinary(b, &u2)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Deserialized User: %+v\n", u2)
	if !reflect.DeepEqual(u, u2) {
		panic("roundtrip failed")
	}

	// 2. Benchmarks
	fmt.Println("\n--- 2. Benchmarks ---")
	benchFlat(1000, 1000)
	benchFlat(5000, 200)
}

func benchFlat(count int, iterations int) {
	users := generateUsers(count)

	// JSON Serialize
	start := time.Now()
	var jsonBytes []byte
	for i := 0; i < iterations; i++ {
		jsonBytes, _ = json.Marshal(&users)
	}
	jsonSerTime := time.Since(start)

	// ASON Text Serialize
	start = time.Now()
	var asonTextBytes []byte
	for i := 0; i < iterations; i++ {
		asonTextBytes, _ = ason.MarshalSlice(&users)
	}
	asonTextSerTime := time.Since(start)

	// ASON-BIN Serialize
	start = time.Now()
	var binBytes []byte
	for i := 0; i < iterations; i++ {
		binBytes, _ = ason.MarshalBinary(&users)
	}
	binSerTime := time.Since(start)

	// JSON Deserialize
	start = time.Now()
	for i := 0; i < iterations; i++ {
		var out []User
		_ = json.Unmarshal(jsonBytes, &out)
	}
	jsonDeTime := time.Since(start)

	// ASON Text Deserialize
	start = time.Now()
	for i := 0; i < iterations; i++ {
		var out []User
		_ = ason.UnmarshalSlice(asonTextBytes, &out)
	}
	asonTextDeTime := time.Since(start)

	// ASON-BIN Deserialize
	start = time.Now()
	for i := 0; i < iterations; i++ {
		var out []User
		_ = ason.UnmarshalBinary(binBytes, &out)
	}
	binDeTime := time.Since(start)

	fmt.Printf("\nFlat struct x %d (8 fields)\n", count)
	fmt.Printf("  Serialize:\n")
	fmt.Printf("    JSON:      %8.2f ms\n", float64(jsonSerTime.Milliseconds())/float64(iterations))
	fmt.Printf("    ASON Text: %8.2f ms (%.1fx faster than JSON)\n", float64(asonTextSerTime.Milliseconds())/float64(iterations), float64(jsonSerTime)/float64(asonTextSerTime))
	fmt.Printf("    ASON-BIN:  %8.2f ms (%.1fx faster than JSON)\n", float64(binSerTime.Milliseconds())/float64(iterations), float64(jsonSerTime)/float64(binSerTime))
	
	fmt.Printf("  Deserialize:\n")
	fmt.Printf("    JSON:      %8.2f ms\n", float64(jsonDeTime.Milliseconds())/float64(iterations))
	fmt.Printf("    ASON Text: %8.2f ms (%.1fx faster than JSON)\n", float64(asonTextDeTime.Milliseconds())/float64(iterations), float64(jsonDeTime)/float64(asonTextDeTime))
	fmt.Printf("    ASON-BIN:  %8.2f ms (%.1fx faster than JSON)\n", float64(binDeTime.Milliseconds())/float64(iterations), float64(jsonDeTime)/float64(binDeTime))
	
	fmt.Printf("  Size:\n")
	fmt.Printf("    JSON:      %8d B\n", len(jsonBytes))
	fmt.Printf("    ASON Text: %8d B (%.0f%% smaller)\n", len(asonTextBytes), (1.0-float64(len(asonTextBytes))/float64(len(jsonBytes)))*100)
	fmt.Printf("    ASON-BIN:  %8d B (%.0f%% smaller)\n", len(binBytes), (1.0-float64(len(binBytes))/float64(len(jsonBytes)))*100)
}
