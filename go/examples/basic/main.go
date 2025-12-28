package main

import (
	"fmt"

	ason "github.com/athxx/ason/go"
)

func main() {
	fmt.Println("=== ASON Go Library Demo ===")
	fmt.Println()

	// 1. Parse simple object
	fmt.Println("1. Parse simple object:")
	v, err := ason.Parse("{name,age}:(Alice,30)")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   name = %s\n", v.Field("name").AsString())
		fmt.Printf("   age = %d\n", v.Field("age").AsInteger())
	}
	fmt.Println()

	// 2. Parse multiple objects
	fmt.Println("2. Parse multiple objects:")
	v, err = ason.Parse("{name,age}:(Alice,30),(Bob,25)")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		for i := 0; i < v.Len(); i++ {
			obj := v.Get(i)
			fmt.Printf("   [%d] name = %s, age = %d\n",
				i, obj.Field("name").AsString(), obj.Field("age").AsInteger())
		}
	}
	fmt.Println()

	// 3. Parse nested object
	fmt.Println("3. Parse nested object:")
	v, err = ason.Parse("{name,addr{city,zip}}:(Alice,(NYC,10001))")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   name = %s\n", v.Field("name").AsString())
		addr := v.Field("addr")
		fmt.Printf("   addr.city = %s\n", addr.Field("city").AsString())
		fmt.Printf("   addr.zip = %d\n", addr.Field("zip").AsInteger())
	}
	fmt.Println()

	// 4. Parse array field
	fmt.Println("4. Parse array field:")
	v, err = ason.Parse("{name,scores[]}:(Alice,[90,85,95])")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   name = %s\n", v.Field("name").AsString())
		scores := v.Field("scores")
		fmt.Print("   scores = [")
		for i := 0; i < scores.Len(); i++ {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Print(scores.Get(i).AsInteger())
		}
		fmt.Println("]")
	}
	fmt.Println()

	// 5. Build and serialize
	fmt.Println("5. Build and serialize:")
	obj := ason.Object()
	obj.Set("name", ason.String("Bob"))
	obj.Set("active", ason.Bool(true))
	obj.Set("score", ason.Float(95.5))
	fmt.Printf("   Data: %s\n", ason.Serialize(obj))
	fmt.Printf("   With schema: %s\n", ason.SerializeWithSchema(obj))
	fmt.Println()

	// 6. Unicode support
	fmt.Println("6. Unicode support:")
	v, err = ason.Parse("{name,city}:(小明,北京)")
	if err != nil {
		fmt.Printf("   Error: %v\n", err)
	} else {
		fmt.Printf("   name = %s\n", v.Field("name").AsString())
		fmt.Printf("   city = %s\n", v.Field("city").AsString())
	}
	fmt.Println()

	fmt.Println("=== Demo Complete ===")
}
