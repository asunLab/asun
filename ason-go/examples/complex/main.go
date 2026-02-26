package main

import (
	"encoding/json"
	"fmt"
	"log"

	ason "github.com/example/ason"
)

type Department struct {
	Title string `ason:"title" json:"title"`
}

type Employee struct {
	ID     int64      `ason:"id" json:"id"`
	Name   string     `ason:"name" json:"name"`
	Dept   Department `ason:"dept" json:"dept"`
	Skills []string   `ason:"skills" json:"skills"`
	Active bool       `ason:"active" json:"active"`
}

type WithMap struct {
	Name  string           `ason:"name" json:"name"`
	Attrs map[string]int64 `ason:"attrs" json:"attrs"`
}

type Nested struct {
	Name string  `ason:"name" json:"name"`
	Addr Address `ason:"addr" json:"addr"`
}

type Address struct {
	City string `ason:"city" json:"city"`
	Zip  int64  `ason:"zip" json:"zip"`
}

type AllTypes struct {
	B         bool      `ason:"b" json:"b"`
	I8v       int8      `ason:"i8v" json:"i8v"`
	I16v      int16     `ason:"i16v" json:"i16v"`
	I32v      int32     `ason:"i32v" json:"i32v"`
	I64v      int64     `ason:"i64v" json:"i64v"`
	U8v       uint8     `ason:"u8v" json:"u8v"`
	U16v      uint16    `ason:"u16v" json:"u16v"`
	U32v      uint32    `ason:"u32v" json:"u32v"`
	U64v      uint64    `ason:"u64v" json:"u64v"`
	F32v      float32   `ason:"f32v" json:"f32v"`
	F64v      float64   `ason:"f64v" json:"f64v"`
	S         string    `ason:"s" json:"s"`
	OptSome   *int64    `ason:"opt_some" json:"opt_some"`
	OptNone   *int64    `ason:"opt_none" json:"opt_none"`
	VecInt    []int64   `ason:"vec_int" json:"vec_int"`
	VecStr    []string  `ason:"vec_str" json:"vec_str"`
	NestedVec [][]int64 `ason:"nested_vec" json:"nested_vec"`
}

type Building struct {
	Name        string  `ason:"name" json:"name"`
	Floors      int64   `ason:"floors" json:"floors"`
	Residential bool    `ason:"residential" json:"residential"`
	HeightM     float64 `ason:"height_m" json:"height_m"`
}

type Street struct {
	Name      string     `ason:"name" json:"name"`
	LengthKm  float64    `ason:"length_km" json:"length_km"`
	Buildings []Building `ason:"buildings" json:"buildings"`
}

type District struct {
	Name       string   `ason:"name" json:"name"`
	Population int64    `ason:"population" json:"population"`
	Streets    []Street `ason:"streets" json:"streets"`
}

type City struct {
	Name       string     `ason:"name" json:"name"`
	Population int64      `ason:"population" json:"population"`
	AreaKm2    float64    `ason:"area_km2" json:"area_km2"`
	Districts  []District `ason:"districts" json:"districts"`
}

type Region struct {
	Name   string `ason:"name" json:"name"`
	Cities []City `ason:"cities" json:"cities"`
}

type Country struct {
	Name        string   `ason:"name" json:"name"`
	Code        string   `ason:"code" json:"code"`
	Population  int64    `ason:"population" json:"population"`
	GdpTrillion float64  `ason:"gdp_trillion" json:"gdp_trillion"`
	Regions     []Region `ason:"regions" json:"regions"`
}

type State struct {
	Name       string `ason:"name" json:"name"`
	Capital    string `ason:"capital" json:"capital"`
	Population int64  `ason:"population" json:"population"`
}

type Nation struct {
	Name   string  `ason:"name" json:"name"`
	States []State `ason:"states" json:"states"`
}

type Continent struct {
	Name    string   `ason:"name" json:"name"`
	Nations []Nation `ason:"nations" json:"nations"`
}

type Planet struct {
	Name       string      `ason:"name" json:"name"`
	RadiusKm   float64     `ason:"radius_km" json:"radius_km"`
	HasLife    bool        `ason:"has_life" json:"has_life"`
	Continents []Continent `ason:"continents" json:"continents"`
}

type SolarSystem struct {
	Name     string   `ason:"name" json:"name"`
	StarType string   `ason:"star_type" json:"star_type"`
	Planets  []Planet `ason:"planets" json:"planets"`
}

type Galaxy struct {
	Name              string        `ason:"name" json:"name"`
	StarCountBillions float64       `ason:"star_count_billions" json:"star_count_billions"`
	Systems           []SolarSystem `ason:"systems" json:"systems"`
}

type Universe struct {
	Name            string   `ason:"name" json:"name"`
	AgeBillionYears float64  `ason:"age_billion_years" json:"age_billion_years"`
	Galaxies        []Galaxy `ason:"galaxies" json:"galaxies"`
}

type DbConfig struct {
	Host           string  `ason:"host" json:"host"`
	Port           int64   `ason:"port" json:"port"`
	MaxConnections int64   `ason:"max_connections" json:"max_connections"`
	SSL            bool    `ason:"ssl" json:"ssl"`
	TimeoutMs      float64 `ason:"timeout_ms" json:"timeout_ms"`
}

type CacheConfig struct {
	Enabled    bool  `ason:"enabled" json:"enabled"`
	TTLSeconds int64 `ason:"ttl_seconds" json:"ttl_seconds"`
	MaxSizeMb  int64 `ason:"max_size_mb" json:"max_size_mb"`
}

type LogConfig struct {
	Level  string  `ason:"level" json:"level"`
	File   *string `ason:"file" json:"file"`
	Rotate bool    `ason:"rotate" json:"rotate"`
}

type ServiceConfig struct {
	Name     string            `ason:"name" json:"name"`
	Version  string            `ason:"version" json:"version"`
	Db       DbConfig          `ason:"db" json:"db"`
	Cache    CacheConfig       `ason:"cache" json:"cache"`
	Log      LogConfig         `ason:"log" json:"log"`
	Features []string          `ason:"features" json:"features"`
	Env      map[string]string `ason:"env" json:"env"`
}

func i64ptr(v int64) *int64   { return &v }
func strptr(v string) *string { return &v }

func main() {
	fmt.Println("=== ASON Complex Examples ===")
	fmt.Println()

	// 1. Nested struct
	fmt.Println("1. Nested struct:")
	var emp Employee
	if err := ason.Decode([]byte("{id,name,dept:{title},skills,active}:(1,Alice,(Manager),[rust],true)"), &emp); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   %+v\n\n", emp)

	// 2. Vec with nested structs
	fmt.Println("2. Vec with nested structs:")
	input2 := []byte(`[{id:int,name:str,dept:{title:str},skills:[str],active:bool}]:
  (1, Alice, (Manager), [Rust, Go], true),
  (2, Bob, (Engineer), [Python], false),
  (3, "Carol Smith", (Director), [Leadership, Strategy], true)`)
	var employees []Employee
	if err := ason.Decode(input2, &employees); err != nil {
		log.Fatal(err)
	}
	for _, e := range employees {
		fmt.Printf("   %+v\n", e)
	}

	// 3. Map/Dict field
	fmt.Println("\n3. Map/Dict field:")
	var wm WithMap
	if err := ason.Decode([]byte("{name,attrs}:(Alice,[(age,30),(score,95)])"), &wm); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   %+v\n", wm)

	// 4. Nested struct roundtrip
	fmt.Println("\n4. Nested struct roundtrip:")
	nested := Nested{Name: "Alice", Addr: Address{City: "NYC", Zip: 10001}}
	s, err := ason.Encode(&nested)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   serialized:   %s\n", s)
	var nested2 Nested
	if err := ason.Decode(s, &nested2); err != nil {
		log.Fatal(err)
	}
	if nested != nested2 {
		log.Fatal("nested roundtrip mismatch")
	}
	fmt.Println("   ✓ roundtrip OK")

	// 5. Escaped strings
	fmt.Println("\n5. Escaped strings:")
	type Note struct {
		Text string `ason:"text"`
	}
	note := Note{Text: "say \"hi\", then (wave)\tnewline\nend"}
	s, err = ason.Encode(&note)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   serialized:   %s\n", s)
	var note2 Note
	if err := ason.Decode(s, &note2); err != nil {
		log.Fatal(err)
	}
	if note != note2 {
		log.Fatal("escape roundtrip mismatch")
	}
	fmt.Println("   ✓ escape roundtrip OK")

	// 6. Float fields
	fmt.Println("\n6. Float fields:")
	type Measurement struct {
		ID    int64   `ason:"id"`
		Value float64 `ason:"value"`
		Label string  `ason:"label"`
	}
	m := Measurement{ID: 2, Value: 95.0, Label: "score"}
	s, err = ason.Encode(&m)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   serialized: %s\n", s)
	var m2 Measurement
	if err := ason.Decode(s, &m2); err != nil {
		log.Fatal(err)
	}
	if m != m2 {
		log.Fatal("float roundtrip mismatch")
	}
	fmt.Println("   ✓ float roundtrip OK")

	// 7. Negative numbers
	fmt.Println("\n7. Negative numbers:")
	type Nums struct {
		A int64   `ason:"a"`
		B float64 `ason:"b"`
		C int64   `ason:"c"`
	}
	n := Nums{A: -42, B: -3.14, C: -9223372036854775807}
	s, err = ason.Encode(&n)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   serialized:   %s\n", s)
	var n2 Nums
	if err := ason.Decode(s, &n2); err != nil {
		log.Fatal(err)
	}
	if n != n2 {
		log.Fatal("negative roundtrip mismatch")
	}
	fmt.Println("   ✓ negative roundtrip OK")

	// 8. All types struct
	fmt.Println("\n8. All types struct:")
	optVal := int64(42)
	all := AllTypes{
		B: true, I8v: -128, I16v: -32768, I32v: -2147483648, I64v: -9223372036854775807,
		U8v: 255, U16v: 65535, U32v: 4294967295, U64v: 18446744073709551615,
		F32v: 3.15, F64v: 2.718281828459045,
		S:         "hello, world (test) [arr]",
		OptSome:   &optVal,
		OptNone:   nil,
		VecInt:    []int64{1, 2, 3, -4, 0},
		VecStr:    []string{"alpha", "beta gamma", "delta"},
		NestedVec: [][]int64{{1, 2}, {3, 4, 5}},
	}
	s, err = ason.Encode(&all)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   serialized (%d bytes):\n", len(s))
	fmt.Printf("   %s\n", s)
	var all2 AllTypes
	if err := ason.Decode(s, &all2); err != nil {
		log.Fatal(err)
	}
	if all.B != all2.B || all.I64v != all2.I64v || all.U64v != all2.U64v || all.S != all2.S {
		log.Fatal("all-types roundtrip mismatch")
	}
	if all2.OptSome == nil || *all2.OptSome != 42 {
		log.Fatal("all-types opt_some mismatch")
	}
	if all2.OptNone != nil {
		log.Fatal("all-types opt_none should be nil")
	}
	fmt.Println("   ✓ all-types roundtrip OK")

	// 9. 5-level deep: Country > Region > City > District > Street > Building
	fmt.Println("\n9. Five-level nesting (Country>Region>City>District>Street>Building):")
	country := Country{
		Name: "Rustland", Code: "RL", Population: 50000000, GdpTrillion: 1.5,
		Regions: []Region{
			{Name: "Northern", Cities: []City{
				{Name: "Ferriton", Population: 2000000, AreaKm2: 350.5, Districts: []District{
					{Name: "Downtown", Population: 500000, Streets: []Street{
						{Name: "Main St", LengthKm: 2.5, Buildings: []Building{
							{Name: "Tower A", Floors: 50, Residential: false, HeightM: 200.0},
							{Name: "Apt Block 1", Floors: 12, Residential: true, HeightM: 40.5},
						}},
						{Name: "Oak Ave", LengthKm: 1.2, Buildings: []Building{
							{Name: "Library", Floors: 3, Residential: false, HeightM: 15.0},
						}},
					}},
					{Name: "Harbor", Population: 150000, Streets: []Street{
						{Name: "Dock Rd", LengthKm: 0.8, Buildings: []Building{
							{Name: "Warehouse 7", Floors: 1, Residential: false, HeightM: 8.0},
						}},
					}},
				}},
			}},
			{Name: "Southern", Cities: []City{
				{Name: "Crabville", Population: 800000, AreaKm2: 120.0, Districts: []District{
					{Name: "Old Town", Population: 200000, Streets: []Street{
						{Name: "Heritage Ln", LengthKm: 0.5, Buildings: []Building{
							{Name: "Museum", Floors: 2, Residential: false, HeightM: 12.0},
							{Name: "Town Hall", Floors: 4, Residential: false, HeightM: 20.0},
						}},
					}},
				}},
			}},
		},
	}
	s, err = ason.Encode(&country)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   serialized (%d bytes)\n", len(s))
	preview := string(s)
	if len(preview) > 200 {
		preview = preview[:200] + "..."
	}
	fmt.Printf("   first 200 chars: %s\n", preview)
	var country2 Country
	if err := ason.Decode(s, &country2); err != nil {
		log.Fatal(err)
	}
	if country.Name != country2.Name || country.Population != country2.Population {
		log.Fatal("5-level roundtrip mismatch")
	}
	fmt.Println("   ✓ 5-level ASON-text roundtrip OK")

	// ASON binary roundtrip
	binBytes, err := ason.EncodeBinary(&country)
	if err != nil {
		log.Fatal(err)
	}
	var country3 Country
	if err := ason.DecodeBinary(binBytes, &country3); err != nil {
		log.Fatal(err)
	}
	if country.Name != country3.Name || country.Population != country3.Population {
		log.Fatal("5-level binary roundtrip mismatch")
	}
	fmt.Println("   ✓ 5-level ASON-bin roundtrip OK")

	jsonBytes, _ := json.Marshal(&country)
	fmt.Printf("   ASON text: %d B | ASON bin: %d B | JSON: %d B\n",
		len(s), len(binBytes), len(jsonBytes))
	fmt.Printf("   BIN vs JSON: %.0f%% smaller | TEXT vs JSON: %.0f%% smaller\n",
		(1.0-float64(len(binBytes))/float64(len(jsonBytes)))*100.0,
		(1.0-float64(len(s))/float64(len(jsonBytes)))*100.0)

	// 10. 7-level deep
	fmt.Println("\n10. Seven-level nesting (Universe>Galaxy>SolarSystem>Planet>Continent>Nation>State):")
	universe := Universe{
		Name: "Observable", AgeBillionYears: 13.8,
		Galaxies: []Galaxy{{
			Name: "Milky Way", StarCountBillions: 250.0,
			Systems: []SolarSystem{{
				Name: "Sol", StarType: "G2V",
				Planets: []Planet{
					{Name: "Earth", RadiusKm: 6371.0, HasLife: true, Continents: []Continent{
						{Name: "Asia", Nations: []Nation{
							{Name: "Japan", States: []State{
								{Name: "Tokyo", Capital: "Shinjuku", Population: 14000000},
								{Name: "Osaka", Capital: "Osaka City", Population: 8800000},
							}},
							{Name: "China", States: []State{
								{Name: "Beijing", Capital: "Beijing", Population: 21500000},
							}},
						}},
						{Name: "Europe", Nations: []Nation{
							{Name: "Germany", States: []State{
								{Name: "Bavaria", Capital: "Munich", Population: 13000000},
								{Name: "Berlin", Capital: "Berlin", Population: 3600000},
							}},
						}},
					}},
					{Name: "Mars", RadiusKm: 3389.5, HasLife: false, Continents: []Continent{}},
				},
			}},
		}},
	}
	s, err = ason.Encode(&universe)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   serialized (%d bytes)\n", len(s))
	var universe2 Universe
	if err := ason.Decode(s, &universe2); err != nil {
		log.Fatal(err)
	}
	if universe.Name != universe2.Name {
		log.Fatal("7-level roundtrip mismatch")
	}
	fmt.Println("   ✓ 7-level ASON-text roundtrip OK")

	// ASON binary roundtrip
	uniBin, err := ason.EncodeBinary(&universe)
	if err != nil {
		log.Fatal(err)
	}
	var universe3 Universe
	if err := ason.DecodeBinary(uniBin, &universe3); err != nil {
		log.Fatal(err)
	}
	if universe.Name != universe3.Name {
		log.Fatal("7-level binary roundtrip mismatch")
	}
	fmt.Println("   ✓ 7-level ASON-bin roundtrip OK")

	jsonBytes, _ = json.Marshal(&universe)
	fmt.Printf("   ASON text: %d B | ASON bin: %d B | JSON: %d B\n",
		len(s), len(uniBin), len(jsonBytes))
	fmt.Printf("   BIN vs JSON: %.0f%% smaller | TEXT vs JSON: %.0f%% smaller\n",
		(1.0-float64(len(uniBin))/float64(len(jsonBytes)))*100.0,
		(1.0-float64(len(s))/float64(len(jsonBytes)))*100.0)

	// 11. Service config
	fmt.Println("\n11. Complex config struct (nested + map + optional):")
	config := ServiceConfig{
		Name: "my-service", Version: "2.1.0",
		Db:       DbConfig{Host: "db.example.com", Port: 5432, MaxConnections: 100, SSL: true, TimeoutMs: 3000.5},
		Cache:    CacheConfig{Enabled: true, TTLSeconds: 3600, MaxSizeMb: 512},
		Log:      LogConfig{Level: "info", File: strptr("/var/log/app.log"), Rotate: true},
		Features: []string{"auth", "rate-limit", "websocket"},
		Env: map[string]string{
			"RUST_LOG":     "debug",
			"DATABASE_URL": "postgres://localhost:5432/mydb",
			"SECRET_KEY":   "abc123!@#",
		},
	}
	s, err = ason.Encode(&config)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   serialized (%d bytes):\n", len(s))
	fmt.Printf("   %s\n", s)
	var config2 ServiceConfig
	if err := ason.Decode(s, &config2); err != nil {
		log.Fatal(err)
	}
	if config.Name != config2.Name || config.Db.Port != config2.Db.Port {
		log.Fatal("config roundtrip mismatch")
	}
	fmt.Println("   ✓ config ASON-text roundtrip OK")
	jsonBytes, _ = json.Marshal(&config)
	fmt.Printf("   ASON text: %d B | JSON: %d B | TEXT vs JSON: %.0f%% smaller\n",
		len(s), len(jsonBytes), (1.0-float64(len(s))/float64(len(jsonBytes)))*100.0)
	// Binary roundtrip
	cfgBin, err := ason.EncodeBinary(&config)
	if err != nil {
		log.Fatal(err)
	}
	var config3 ServiceConfig
	if err := ason.DecodeBinary(cfgBin, &config3); err != nil {
		log.Fatal(err)
	}
	if config.Name != config3.Name || config.Db.Port != config3.Db.Port {
		log.Fatal("config binary roundtrip mismatch")
	}
	fmt.Println("   ✓ config ASON-bin roundtrip OK")
	fmt.Printf("   ASON bin: %d B | BIN vs JSON: %.0f%% smaller\n",
		len(cfgBin), (1.0-float64(len(cfgBin))/float64(len(jsonBytes)))*100.0)

	// 12. Large structure — 100 countries
	fmt.Println("\n12. Large structure (100 countries × nested regions):")
	countries := make([]Country, 100)
	for i := range countries {
		regions := make([]Region, 3)
		for r := 0; r < 3; r++ {
			cities := make([]City, 2)
			for c := 0; c < 2; c++ {
				cities[c] = City{
					Name: fmt.Sprintf("City_%d_%d_%d", i, r, c), Population: int64(100000 + c*50000),
					AreaKm2: 50.0 + float64(c)*25.5,
					Districts: []District{{
						Name: fmt.Sprintf("Dist_%d", c), Population: int64(50000 + c*10000),
						Streets: []Street{{
							Name: fmt.Sprintf("St_%d", c), LengthKm: 1.0 + float64(c)*0.5,
							Buildings: []Building{
								{Name: fmt.Sprintf("Bldg_%d_0", c), Floors: 5, Residential: true, HeightM: 15.0},
								{Name: fmt.Sprintf("Bldg_%d_1", c), Floors: 8, Residential: false, HeightM: 25.5},
							},
						}},
					}},
				}
			}
			regions[r] = Region{Name: fmt.Sprintf("Region_%d_%d", i, r), Cities: cities}
		}
		countries[i] = Country{
			Name: fmt.Sprintf("Country_%d", i), Code: fmt.Sprintf("C%02d", i%100),
			Population: int64(1000000 + i*500000), GdpTrillion: float64(i) * 0.5, Regions: regions,
		}
	}
	totalASON, totalJSON, totalBIN := 0, 0, 0
	for i := range countries {
		as, _ := ason.Encode(&countries[i])
		js, _ := json.Marshal(&countries[i])
		bs, _ := ason.EncodeBinary(&countries[i])
		// Verify text roundtrip
		var c2 Country
		if err := ason.Decode(as, &c2); err != nil {
			log.Fatalf("country %d roundtrip failed: %v", i, err)
		}
		if countries[i].Name != c2.Name {
			log.Fatalf("country %d name mismatch", i)
		}
		// Verify binary roundtrip
		var c3 Country
		if err := ason.DecodeBinary(bs, &c3); err != nil {
			log.Fatalf("country %d binary roundtrip failed: %v", i, err)
		}
		if countries[i].Name != c3.Name {
			log.Fatalf("country %d binary name mismatch", i)
		}
		totalASON += len(as)
		totalJSON += len(js)
		totalBIN += len(bs)
	}
	fmt.Println("   100 countries with 5-level nesting:")
	fmt.Printf("   Total ASON text: %d bytes (%.1f KB)\n", totalASON, float64(totalASON)/1024.0)
	fmt.Printf("   Total ASON bin:  %d bytes (%.1f KB)\n", totalBIN, float64(totalBIN)/1024.0)
	fmt.Printf("   Total JSON:      %d bytes (%.1f KB)\n", totalJSON, float64(totalJSON)/1024.0)
	fmt.Printf("   TEXT vs JSON: %.0f%% smaller | BIN vs JSON: %.0f%% smaller\n",
		(1.0-float64(totalASON)/float64(totalJSON))*100.0,
		(1.0-float64(totalBIN)/float64(totalJSON))*100.0)
	fmt.Println("   ✓ all 100 countries roundtrip OK (text + bin)")

	// 13. Deserialize with nested schema type hints
	fmt.Println("\n13. Deserialize with nested schema type hints:")
	deepInput := []byte("{name:str,code:str,population:int,gdp_trillion:float,regions:[{name:str,cities:[{name:str,population:int,area_km2:float,districts:[{name:str,population:int,streets:[{name:str,length_km:float,buildings:[{name:str,floors:int,residential:bool,height_m:float}]}]}]}]}]}:(TestLand,TL,1000000,0.5,[(TestRegion,[(TestCity,500000,100.0,[(Central,250000,[(Main St,2.5,[(HQ,10,false,45.0)])])])])])")
	var dc Country
	if err := ason.Decode(deepInput, &dc); err != nil {
		log.Fatal(err)
	}
	if dc.Name != "TestLand" {
		log.Fatal("deep schema parse name mismatch")
	}
	bld := dc.Regions[0].Cities[0].Districts[0].Streets[0].Buildings[0]
	if bld.Name != "HQ" {
		log.Fatal("deep schema parse building name mismatch")
	}
	fmt.Println("   ✓ deep schema type-hint parse OK")
	fmt.Printf("   Building at depth 6: %+v\n", bld)

	// 14. Typed serialization
	fmt.Println("\n14. Typed serialization (MarshalTyped):")
	empForTyped := Employee{ID: 1, Name: "Alice", Dept: Department{Title: "Engineering"}, Skills: []string{"Rust", "Go"}, Active: true}
	typedBytes, err := ason.EncodeTyped(&empForTyped)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   nested struct: %s\n", typedBytes)
	var empBack Employee
	if err := ason.Decode(typedBytes, &empBack); err != nil {
		log.Fatal(err)
	}
	if empBack.Name != "Alice" {
		log.Fatal("typed nested struct roundtrip mismatch")
	}
	fmt.Println("   ✓ typed nested struct roundtrip OK")

	allTyped, err := ason.EncodeTyped(&all)
	if err != nil {
		log.Fatal(err)
	}
	allTypedStr := string(allTyped)
	p := allTypedStr
	if len(p) > 80 {
		p = p[:80]
	}
	fmt.Printf("   all-types (%d bytes): %s...\n", len(allTyped), p)
	var allBack AllTypes
	if err := ason.Decode(allTyped, &allBack); err != nil {
		log.Fatal(err)
	}
	if allBack.S != all.S || allBack.B != all.B {
		log.Fatal("typed all-types roundtrip mismatch")
	}
	fmt.Println("   ✓ typed all-types roundtrip OK")

	configTyped, err := ason.EncodeTyped(&config)
	if err != nil {
		log.Fatal(err)
	}
	configTypedStr := string(configTyped)
	p = configTypedStr
	if len(p) > 100 {
		p = p[:100]
	}
	fmt.Printf("   config (%d bytes): %s...\n", len(configTyped), p)
	var configBack ServiceConfig
	if err := ason.Decode(configTyped, &configBack); err != nil {
		log.Fatal(err)
	}
	if configBack.Name != config.Name {
		log.Fatal("typed config roundtrip mismatch")
	}
	fmt.Println("   ✓ typed config roundtrip OK")
	untyped, _ := ason.Encode(&config)
	fmt.Printf("   untyped: %d bytes | typed: %d bytes | overhead: %d bytes\n",
		len(untyped), len(configTyped), len(configTyped)-len(untyped))

	// 15. Edge cases
	fmt.Println("\n15. Edge cases:")
	type WithVec struct {
		Items []int64 `ason:"items"`
	}
	wv := WithVec{Items: []int64{}}
	s, _ = ason.Encode(&wv)
	fmt.Printf("   empty vec: %s\n", s)
	var wv2 WithVec
	if err := ason.Decode(s, &wv2); err != nil {
		log.Fatal(err)
	}

	type Special struct {
		Val string `ason:"val"`
	}
	sp := Special{Val: "tabs\there, newlines\nhere, quotes\"and\\backslash"}
	s, _ = ason.Encode(&sp)
	fmt.Printf("   special chars: %s\n", s)
	var sp2 Special
	if err := ason.Decode(s, &sp2); err != nil {
		log.Fatal(err)
	}
	if sp != sp2 {
		log.Fatal("special chars roundtrip mismatch")
	}

	sp3 := Special{Val: "true"}
	s, _ = ason.Encode(&sp3)
	fmt.Printf("   bool-like string: %s\n", s)
	var sp4 Special
	ason.Decode(s, &sp4)
	if sp3 != sp4 {
		log.Fatal("bool-like string roundtrip mismatch")
	}

	sp5 := Special{Val: "12345"}
	s, _ = ason.Encode(&sp5)
	fmt.Printf("   number-like string: %s\n", s)
	var sp6 Special
	ason.Decode(s, &sp6)
	if sp5 != sp6 {
		log.Fatal("number-like string roundtrip mismatch")
	}
	fmt.Println("   ✓ all edge cases OK")

	// 16. Triple-nested arrays
	fmt.Println("\n16. Triple-nested arrays:")
	type Matrix3D struct {
		Data [][][]int64 `ason:"data"`
	}
	m3 := Matrix3D{Data: [][][]int64{{{1, 2}, {3, 4}}, {{5, 6, 7}, {8}}}}
	s, _ = ason.Encode(&m3)
	fmt.Printf("   %s\n", s)
	var m3b Matrix3D
	if err := ason.Decode(s, &m3b); err != nil {
		log.Fatal(err)
	}
	if len(m3b.Data) != 2 || len(m3b.Data[0]) != 2 || m3b.Data[0][0][0] != 1 {
		log.Fatal("triple-nested array roundtrip mismatch")
	}
	fmt.Println("   ✓ triple-nested array roundtrip OK")

	// 17. Comments
	fmt.Println("\n17. Comments:")
	var empComment Employee
	if err := ason.Decode([]byte("{id,name,dept:{title},skills,active}:/* inline */ (1,Alice,(HR),[rust],true)"), &empComment); err != nil {
		log.Fatal(err)
	}
	fmt.Printf("   with inline comment: %+v\n", empComment)
	fmt.Println("   ✓ comment parsing OK")

	fmt.Printf("\n=== All %d complex examples passed! ===\n", 17)
}
