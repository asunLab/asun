package main

import (
	"encoding/json"
	"fmt"
	"runtime"
	"strings"
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

type AllTypes struct {
	B       bool     `ason:"b" json:"b"`
	I8v     int8     `ason:"i8v" json:"i8v"`
	I16v    int16    `ason:"i16v" json:"i16v"`
	I32v    int32    `ason:"i32v" json:"i32v"`
	I64v    int64    `ason:"i64v" json:"i64v"`
	U8v     uint8    `ason:"u8v" json:"u8v"`
	U16v    uint16   `ason:"u16v" json:"u16v"`
	U32v    uint32   `ason:"u32v" json:"u32v"`
	U64v    uint64   `ason:"u64v" json:"u64v"`
	F32v    float32  `ason:"f32v" json:"f32v"`
	F64v    float64  `ason:"f64v" json:"f64v"`
	S       string   `ason:"s" json:"s"`
	OptSome *int64   `ason:"opt_some" json:"opt_some"`
	OptNone *int64   `ason:"opt_none" json:"opt_none"`
	VecInt  []int64  `ason:"vec_int" json:"vec_int"`
	VecStr  []string `ason:"vec_str" json:"vec_str"`
}

type Task struct {
	ID       int64   `ason:"id" json:"id"`
	Title    string  `ason:"title" json:"title"`
	Priority int64   `ason:"priority" json:"priority"`
	Done     bool    `ason:"done" json:"done"`
	Hours    float64 `ason:"hours" json:"hours"`
}

type Project struct {
	Name   string  `ason:"name" json:"name"`
	Budget float64 `ason:"budget" json:"budget"`
	Active bool    `ason:"active" json:"active"`
	Tasks  []Task  `ason:"tasks" json:"tasks"`
}

type Team struct {
	Name     string    `ason:"name" json:"name"`
	Lead     string    `ason:"lead" json:"lead"`
	Size     int64     `ason:"size" json:"size"`
	Projects []Project `ason:"projects" json:"projects"`
}

type Division struct {
	Name      string `ason:"name" json:"name"`
	Location  string `ason:"location" json:"location"`
	Headcount int64  `ason:"headcount" json:"headcount"`
	Teams     []Team `ason:"teams" json:"teams"`
}

type Company struct {
	Name      string     `ason:"name" json:"name"`
	Founded   int64      `ason:"founded" json:"founded"`
	RevenueM  float64    `ason:"revenue_m" json:"revenue_m"`
	Public    bool       `ason:"public" json:"public"`
	Divisions []Division `ason:"divisions" json:"divisions"`
	Tags      []string   `ason:"tags" json:"tags"`
}

func generateUsers(n int) []User {
	names := []string{"Alice", "Bob", "Carol", "David", "Eve", "Frank", "Grace", "Hank"}
	roles := []string{"engineer", "designer", "manager", "analyst"}
	cities := []string{"NYC", "LA", "Chicago", "Houston", "Phoenix"}
	users := make([]User, n)
	for i := 0; i < n; i++ {
		users[i] = User{
			ID: int64(i), Name: names[i%len(names)],
			Email:  fmt.Sprintf("%s@example.com", strings.ToLower(names[i%len(names)])),
			Age:    int64(25 + i%40),
			Score:  50.0 + float64(i%50) + 0.5,
			Active: i%3 != 0, Role: roles[i%len(roles)], City: cities[i%len(cities)],
		}
	}
	return users
}

func i64ptr(v int64) *int64 { return &v }

func generateAllTypes(n int) []AllTypes {
	items := make([]AllTypes, n)
	for i := 0; i < n; i++ {
		var optSome *int64
		if i%2 == 0 {
			optSome = i64ptr(int64(i))
		}
		items[i] = AllTypes{
			B: i%2 == 0, I8v: int8(i % 256), I16v: -int16(i),
			I32v: int32(i) * 1000, I64v: int64(i) * 100000,
			U8v: uint8(i % 256), U16v: uint16(i % 65536),
			U32v: uint32(i) * 7919, U64v: uint64(i) * 1000000007,
			F32v: float32(i) * 1.5, F64v: float64(i)*0.25 + 0.5,
			S: fmt.Sprintf("item_%d", i), OptSome: optSome, OptNone: nil,
			VecInt: []int64{int64(i), int64(i + 1), int64(i + 2)},
			VecStr: []string{fmt.Sprintf("tag%d", i%5), fmt.Sprintf("cat%d", i%3)},
		}
	}
	return items
}

func generateCompanies(n int) []Company {
	locs := []string{"NYC", "London", "Tokyo", "Berlin"}
	leads := []string{"Alice", "Bob", "Carol", "David"}
	companies := make([]Company, n)
	for i := 0; i < n; i++ {
		divisions := make([]Division, 2)
		for d := 0; d < 2; d++ {
			teams := make([]Team, 2)
			for t := 0; t < 2; t++ {
				projects := make([]Project, 3)
				for p := 0; p < 3; p++ {
					tasks := make([]Task, 4)
					for tk := 0; tk < 4; tk++ {
						tasks[tk] = Task{
							ID: int64(i*100 + d*10 + t*5 + tk), Title: fmt.Sprintf("Task_%d", tk),
							Priority: int64(tk%3 + 1), Done: tk%2 == 0, Hours: 2.0 + float64(tk)*1.5,
						}
					}
					projects[p] = Project{
						Name: fmt.Sprintf("Proj_%d_%d", t, p), Budget: 100.0 + float64(p)*50.5,
						Active: p%2 == 0, Tasks: tasks,
					}
				}
				teams[t] = Team{
					Name: fmt.Sprintf("Team_%d_%d_%d", i, d, t), Lead: leads[t%4],
					Size: int64(5 + t*2), Projects: projects,
				}
			}
			divisions[d] = Division{
				Name: fmt.Sprintf("Div_%d_%d", i, d), Location: locs[d%4],
				Headcount: int64(50 + d*20), Teams: teams,
			}
		}
		companies[i] = Company{
			Name: fmt.Sprintf("Corp_%d", i), Founded: int64(1990 + i%35),
			RevenueM: 10.0 + float64(i)*5.5, Public: i%2 == 0, Divisions: divisions,
			Tags: []string{"enterprise", "tech", fmt.Sprintf("sector_%d", i%5)},
		}
	}
	return companies
}

func formatBytes(b int) string {
	if b >= 1048576 {
		return fmt.Sprintf("%.1f MB", float64(b)/1048576.0)
	}
	if b >= 1024 {
		return fmt.Sprintf("%.1f KB", float64(b)/1024.0)
	}
	return fmt.Sprintf("%d B", b)
}

type benchResult struct {
	name      string
	jsonSerMs float64
	asonSerMs float64
	jsonDeMs  float64
	asonDeMs  float64
	jsonBytes int
	asonBytes int
}

func (r *benchResult) print() {
	serRatio := r.jsonSerMs / r.asonSerMs
	deRatio := r.jsonDeMs / r.asonDeMs
	saving := (1.0 - float64(r.asonBytes)/float64(r.jsonBytes)) * 100.0
	serMark, deMark := "", ""
	if serRatio >= 1.0 {
		serMark = "✓ ASON faster"
	}
	if deRatio >= 1.0 {
		deMark = "✓ ASON faster"
	}
	fmt.Printf("  %s\n", r.name)
	fmt.Printf("    Serialize:   JSON %8.2fms | ASON %8.2fms | ratio %.2fx %s\n",
		r.jsonSerMs, r.asonSerMs, serRatio, serMark)
	fmt.Printf("    Deserialize: JSON %8.2fms | ASON %8.2fms | ratio %.2fx %s\n",
		r.jsonDeMs, r.asonDeMs, deRatio, deMark)
	fmt.Printf("    Size:        JSON %8d B | ASON %8d B | saving %.0f%%\n",
		r.jsonBytes, r.asonBytes, saving)
}

func benchFlat(count, iterations int) benchResult {
	users := generateUsers(count)
	var jsonStr []byte
	start := time.Now()
	for i := 0; i < iterations; i++ {
		jsonStr, _ = json.Marshal(users)
	}
	jsonSer := time.Since(start)

	var asonStr []byte
	start = time.Now()
	for i := 0; i < iterations; i++ {
		asonStr, _ = ason.Encode(users)
	}
	asonSer := time.Since(start)

	start = time.Now()
	for i := 0; i < iterations; i++ {
		var out []User
		json.Unmarshal(jsonStr, &out)
	}
	jsonDe := time.Since(start)

	start = time.Now()
	for i := 0; i < iterations; i++ {
		var out []User
		ason.Decode(asonStr, &out)
	}
	asonDe := time.Since(start)

	var decoded []User
	ason.Decode(asonStr, &decoded)
	if len(decoded) != count {
		panic(fmt.Sprintf("flat %d roundtrip failed: got %d", count, len(decoded)))
	}
	return benchResult{
		name:      fmt.Sprintf("Flat struct × %d (8 fields)", count),
		jsonSerMs: float64(jsonSer.Nanoseconds()) / 1e6, asonSerMs: float64(asonSer.Nanoseconds()) / 1e6,
		jsonDeMs: float64(jsonDe.Nanoseconds()) / 1e6, asonDeMs: float64(asonDe.Nanoseconds()) / 1e6,
		jsonBytes: len(jsonStr), asonBytes: len(asonStr),
	}
}

func benchAllTypes(count, iterations int) benchResult {
	items := generateAllTypes(count)
	var jsonStr []byte
	start := time.Now()
	for i := 0; i < iterations; i++ {
		jsonStr, _ = json.Marshal(items)
	}
	jsonSer := time.Since(start)

	var asonLines [][]byte
	start = time.Now()
	for i := 0; i < iterations; i++ {
		asonLines = make([][]byte, len(items))
		for j := range items {
			asonLines[j], _ = ason.Encode(&items[j])
		}
	}
	asonSer := time.Since(start)

	totalASON := 0
	for _, l := range asonLines {
		totalASON += len(l)
	}

	start = time.Now()
	for i := 0; i < iterations; i++ {
		var out []AllTypes
		json.Unmarshal(jsonStr, &out)
	}
	jsonDe := time.Since(start)

	start = time.Now()
	for i := 0; i < iterations; i++ {
		for _, l := range asonLines {
			var out AllTypes
			ason.Decode(l, &out)
		}
	}
	asonDe := time.Since(start)

	return benchResult{
		name:      fmt.Sprintf("All-types struct × %d (16 fields, per-struct)", count),
		jsonSerMs: float64(jsonSer.Nanoseconds()) / 1e6, asonSerMs: float64(asonSer.Nanoseconds()) / 1e6,
		jsonDeMs: float64(jsonDe.Nanoseconds()) / 1e6, asonDeMs: float64(asonDe.Nanoseconds()) / 1e6,
		jsonBytes: len(jsonStr), asonBytes: totalASON,
	}
}

func benchDeep(count, iterations int) benchResult {
	companies := generateCompanies(count)
	var jsonStr []byte
	start := time.Now()
	for i := 0; i < iterations; i++ {
		jsonStr, _ = json.Marshal(companies)
	}
	jsonSer := time.Since(start)

	asonStrs := make([][]byte, count)
	start = time.Now()
	for i := 0; i < iterations; i++ {
		for j := range companies {
			asonStrs[j], _ = ason.Encode(&companies[j])
		}
	}
	asonSer := time.Since(start)

	totalASON := 0
	for _, s := range asonStrs {
		totalASON += len(s)
	}

	start = time.Now()
	for i := 0; i < iterations; i++ {
		var out []Company
		json.Unmarshal(jsonStr, &out)
	}
	jsonDe := time.Since(start)

	start = time.Now()
	for i := 0; i < iterations; i++ {
		for _, s := range asonStrs {
			var out Company
			ason.Decode(s, &out)
		}
	}
	asonDe := time.Since(start)

	for i, s := range asonStrs {
		var c2 Company
		if err := ason.Decode(s, &c2); err != nil {
			panic(fmt.Sprintf("deep roundtrip failed at %d: %v", i, err))
		}
	}

	return benchResult{
		name:      fmt.Sprintf("5-level deep × %d (~48 nodes each)", count),
		jsonSerMs: float64(jsonSer.Nanoseconds()) / 1e6, asonSerMs: float64(asonSer.Nanoseconds()) / 1e6,
		jsonDeMs: float64(jsonDe.Nanoseconds()) / 1e6, asonDeMs: float64(asonDe.Nanoseconds()) / 1e6,
		jsonBytes: len(jsonStr), asonBytes: totalASON,
	}
}

func benchSingleRoundtrip(iterations int) (asonMs, jsonMs float64) {
	user := User{ID: 1, Name: "Alice", Email: "alice@example.com", Age: 30, Score: 95.5, Active: true, Role: "engineer", City: "NYC"}
	start := time.Now()
	for i := 0; i < iterations; i++ {
		s, _ := ason.Encode(&user)
		var out User
		ason.Decode(s, &out)
	}
	asonMs = float64(time.Since(start).Nanoseconds()) / 1e6
	start = time.Now()
	for i := 0; i < iterations; i++ {
		s, _ := json.Marshal(&user)
		var out User
		json.Unmarshal(s, &out)
	}
	jsonMs = float64(time.Since(start).Nanoseconds()) / 1e6
	return
}

func benchDeepSingleRoundtrip(iterations int) (asonMs, jsonMs float64) {
	company := Company{
		Name: "MegaCorp", Founded: 2000, RevenueM: 500.5, Public: true,
		Divisions: []Division{{
			Name: "Engineering", Location: "SF", Headcount: 200,
			Teams: []Team{{Name: "Backend", Lead: "Alice", Size: 12,
				Projects: []Project{{Name: "API v3", Budget: 250.0, Active: true,
					Tasks: []Task{
						{ID: 1, Title: "Design", Priority: 1, Done: true, Hours: 40.0},
						{ID: 2, Title: "Implement", Priority: 1, Done: false, Hours: 120.0},
						{ID: 3, Title: "Test", Priority: 2, Done: false, Hours: 30.0},
					},
				}},
			}},
		}},
		Tags: []string{"tech", "public"},
	}
	start := time.Now()
	for i := 0; i < iterations; i++ {
		s, _ := ason.Encode(&company)
		var out Company
		ason.Decode(s, &out)
	}
	asonMs = float64(time.Since(start).Nanoseconds()) / 1e6
	start = time.Now()
	for i := 0; i < iterations; i++ {
		s, _ := json.Marshal(&company)
		var out Company
		json.Unmarshal(s, &out)
	}
	jsonMs = float64(time.Since(start).Nanoseconds()) / 1e6
	return
}

func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║            ASON vs JSON Comprehensive Benchmark              ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
	fmt.Printf("\nSystem: %s %s\n", runtime.GOOS, runtime.GOARCH)
	var memBefore runtime.MemStats
	runtime.ReadMemStats(&memBefore)
	fmt.Printf("Alloc before: %s\n\n", formatBytes(int(memBefore.Alloc)))
	iterations := 100
	fmt.Printf("Iterations per test: %d\n", iterations)

	// Section 1: Flat struct
	fmt.Println("\n┌─────────────────────────────────────────────┐")
	fmt.Println("│  Section 1: Flat Struct (schema-driven vec) │")
	fmt.Println("└─────────────────────────────────────────────┘")
	for _, count := range []int{100, 500, 1000, 5000} {
		r := benchFlat(count, iterations)
		r.print()
		fmt.Println()
	}

	// Section 2: All-types struct
	fmt.Println("┌──────────────────────────────────────────────┐")
	fmt.Println("│  Section 2: All-Types Struct (16 fields)     │")
	fmt.Println("└──────────────────────────────────────────────┘")
	for _, count := range []int{100, 500} {
		r := benchAllTypes(count, iterations)
		r.print()
		fmt.Println()
	}

	// Section 3: 5-level deep nested struct
	fmt.Println("┌──────────────────────────────────────────────────────────┐")
	fmt.Println("│  Section 3: 5-Level Deep Nesting (Company hierarchy)     │")
	fmt.Println("└──────────────────────────────────────────────────────────┘")
	for _, count := range []int{10, 50, 100} {
		r := benchDeep(count, iterations)
		r.print()
		fmt.Println()
	}

	// Section 4: Single struct roundtrip
	fmt.Println("┌──────────────────────────────────────────────┐")
	fmt.Println("│  Section 4: Single Struct Roundtrip (10000x) │")
	fmt.Println("└──────────────────────────────────────────────┘")
	asonFlat, jsonFlat := benchSingleRoundtrip(10000)
	fmt.Printf("  Flat:  ASON %6.2fms | JSON %6.2fms | ratio %.2fx\n", asonFlat, jsonFlat, jsonFlat/asonFlat)
	asonDeep, jsonDeep := benchDeepSingleRoundtrip(10000)
	fmt.Printf("  Deep:  ASON %6.2fms | JSON %6.2fms | ratio %.2fx\n", asonDeep, jsonDeep, jsonDeep/asonDeep)

	// Section 5: Large payload
	fmt.Println("\n┌──────────────────────────────────────────────┐")
	fmt.Println("│  Section 5: Large Payload (10k records)      │")
	fmt.Println("└──────────────────────────────────────────────┘")
	rLarge := benchFlat(10000, 10)
	fmt.Println("  (10 iterations for large payload)")
	rLarge.print()

	// Section 6: Annotated vs Unannotated (deserialize)
	fmt.Println("\n┌──────────────────────────────────────────────────────────────┐")
	fmt.Println("│  Section 6: Annotated vs Unannotated Schema (deserialize)    │")
	fmt.Println("└──────────────────────────────────────────────────────────────┘")
	{
		users1k := generateUsers(1000)
		asonUntyped, _ := ason.Encode(users1k)
		untypedStr := string(asonUntyped)
		typedStr := strings.Replace(untypedStr,
			"[{id,name,email,age,score,active,role,city}]:",
			"[{id:int,name:str,email:str,age:int,score:float,active:bool,role:str,city:str}]:", 1)
		asonTyped := []byte(typedStr)

		deIters := 200
		start := time.Now()
		for i := 0; i < deIters; i++ {
			var out []User
			ason.Decode(asonUntyped, &out)
		}
		untypedMs := float64(time.Since(start).Nanoseconds()) / 1e6
		start = time.Now()
		for i := 0; i < deIters; i++ {
			var out []User
			ason.Decode(asonTyped, &out)
		}
		typedMs := float64(time.Since(start).Nanoseconds()) / 1e6
		fmt.Printf("  Flat struct × 1000 (%d iters, deserialize only)\n", deIters)
		fmt.Printf("    Unannotated: %8.2fms  (%d B)\n", untypedMs, len(asonUntyped))
		fmt.Printf("    Annotated:   %8.2fms  (%d B)\n", typedMs, len(asonTyped))
		fmt.Printf("    Ratio: %.3fx\n\n", untypedMs/typedMs)

		company := generateCompanies(1)[0]
		deepUntyped, _ := ason.Encode(&company)
		deepUntypedStr := string(deepUntyped)
		deepTypedStr := strings.Replace(deepUntypedStr,
			"{name,founded,revenue_m,public,divisions,tags}:",
			"{name:str,founded:int,revenue_m:float,public:bool,divisions,tags}:", 1)
		deepTyped := []byte(deepTypedStr)
		deepIters := 5000
		start = time.Now()
		for i := 0; i < deepIters; i++ {
			var out Company
			ason.Decode(deepUntyped, &out)
		}
		deepUntypedMs := float64(time.Since(start).Nanoseconds()) / 1e6
		start = time.Now()
		for i := 0; i < deepIters; i++ {
			var out Company
			ason.Decode(deepTyped, &out)
		}
		deepTypedMs := float64(time.Since(start).Nanoseconds()) / 1e6
		fmt.Printf("  5-level deep (%d iters, deserialize only)\n", deepIters)
		fmt.Printf("    Unannotated: %8.2fms  (%d B)\n", deepUntypedMs, len(deepUntyped))
		fmt.Printf("    Annotated:   %8.2fms  (%d B)\n", deepTypedMs, len(deepTyped))
		fmt.Printf("    Ratio: %.3fx\n", deepUntypedMs/deepTypedMs)
	}

	// Section 7: Annotated vs Unannotated (serialize)
	fmt.Println("\n┌──────────────────────────────────────────────────────────────┐")
	fmt.Println("│  Section 7: Annotated vs Unannotated Schema (serialize)      │")
	fmt.Println("└──────────────────────────────────────────────────────────────┘")
	{
		users1k := generateUsers(1000)
		serIters := 200
		start := time.Now()
		var untypedOut []byte
		for i := 0; i < serIters; i++ {
			untypedOut, _ = ason.Encode(users1k)
		}
		untypedMs := float64(time.Since(start).Nanoseconds()) / 1e6
		start = time.Now()
		var typedOut []byte
		for i := 0; i < serIters; i++ {
			typedOut, _ = ason.EncodeTyped(users1k)
		}
		typedMs := float64(time.Since(start).Nanoseconds()) / 1e6
		fmt.Printf("  Flat struct × 1000 vec (%d iters, serialize only)\n", serIters)
		fmt.Printf("    Unannotated: %8.2fms  (%d B)\n", untypedMs, len(untypedOut))
		fmt.Printf("    Annotated:   %8.2fms  (%d B)\n", typedMs, len(typedOut))
		fmt.Printf("    Ratio: %.3fx\n\n", untypedMs/typedMs)

		singleUser := users1k[0]
		singleIters := 10000
		start = time.Now()
		var singleUntyped []byte
		for i := 0; i < singleIters; i++ {
			singleUntyped, _ = ason.Encode(&singleUser)
		}
		singleUntypedMs := float64(time.Since(start).Nanoseconds()) / 1e6
		start = time.Now()
		var singleTyped []byte
		for i := 0; i < singleIters; i++ {
			singleTyped, _ = ason.EncodeTyped(&singleUser)
		}
		singleTypedMs := float64(time.Since(start).Nanoseconds()) / 1e6
		fmt.Printf("  Single flat struct (%d iters, serialize only)\n", singleIters)
		fmt.Printf("    Unannotated: %8.2fms  (%d B)\n", singleUntypedMs, len(singleUntyped))
		fmt.Printf("    Annotated:   %8.2fms  (%d B)\n", singleTypedMs, len(singleTyped))
		fmt.Printf("    Ratio: %.3fx\n", singleUntypedMs/singleTypedMs)
	}

	// Section 8: Throughput summary
	fmt.Println("\n┌──────────────────────────────────────────────┐")
	fmt.Println("│  Section 8: Throughput Summary               │")
	fmt.Println("└──────────────────────────────────────────────┘")
	{
		users1k := generateUsers(1000)
		jsonData, _ := json.Marshal(users1k)
		asonData, _ := ason.Encode(users1k)
		iters := 100

		start := time.Now()
		for i := 0; i < iters; i++ {
			json.Marshal(users1k)
		}
		jsonSerDur := time.Since(start)
		start = time.Now()
		for i := 0; i < iters; i++ {
			ason.Encode(users1k)
		}
		asonSerDur := time.Since(start)
		start = time.Now()
		for i := 0; i < iters; i++ {
			var out []User
			json.Unmarshal(jsonData, &out)
		}
		jsonDeDur := time.Since(start)
		start = time.Now()
		for i := 0; i < iters; i++ {
			var out []User
			ason.Decode(asonData, &out)
		}
		asonDeDur := time.Since(start)

		totalRecords := 1000.0 * float64(iters)
		jsonSerRPS := totalRecords / jsonSerDur.Seconds()
		asonSerRPS := totalRecords / asonSerDur.Seconds()
		jsonDeRPS := totalRecords / jsonDeDur.Seconds()
		asonDeRPS := totalRecords / asonDeDur.Seconds()

		serMark, deMark := "", ""
		if asonSerRPS > jsonSerRPS {
			serMark = "✓ ASON faster"
		}
		if asonDeRPS > jsonDeRPS {
			deMark = "✓ ASON faster"
		}

		fmt.Printf("  Serialize throughput (1000 records × %d iters):\n", iters)
		fmt.Printf("    JSON: %.0f records/s\n", jsonSerRPS)
		fmt.Printf("    ASON: %.0f records/s\n", asonSerRPS)
		fmt.Printf("    Speed: %.2fx %s\n", asonSerRPS/jsonSerRPS, serMark)
		fmt.Println("  Deserialize throughput:")
		fmt.Printf("    JSON: %.0f records/s\n", jsonDeRPS)
		fmt.Printf("    ASON: %.0f records/s\n", asonDeRPS)
		fmt.Printf("    Speed: %.2fx %s\n", asonDeRPS/jsonDeRPS, deMark)

		var memAfter runtime.MemStats
		runtime.ReadMemStats(&memAfter)
		fmt.Printf("\n  Memory: Alloc=%s TotalAlloc=%s Sys=%s\n",
			formatBytes(int(memAfter.Alloc)), formatBytes(int(memAfter.TotalAlloc)), formatBytes(int(memAfter.Sys)))
	}

	// Section 9: Binary Format (ASON-BIN) benchmarks
	fmt.Println("\n┌──────────────────────────────────────────────────────────────┐")
	fmt.Println("│  Section 9: Binary Format (ASON-BIN) vs ASON text vs JSON    │")
	fmt.Println("└──────────────────────────────────────────────────────────────┘")
	{
		// -- Flat struct benchmark --
		fmt.Println("\n  ── Flat struct ──")
		for _, count := range []int{100, 1000, 5000} {
			iters := 50
			if count >= 5000 {
				iters = 5
			} else if count >= 1000 {
				iters = 20
			}
			users := generateUsers(count)

			// Serialize once to get sizes
			var binTotal int
			var asonTotal int
			var jsonTotal int
			for _, u := range users {
				binData, _ := ason.EncodeBinary(&u)
				binTotal += len(binData)
			}
			asonData, _ := ason.Encode(users)
			asonTotal = len(asonData)
			jsonData, _ := json.Marshal(users)
			jsonTotal = len(jsonData)

			// Benchmark BIN serialize
			start := time.Now()
			for it := 0; it < iters; it++ {
				for i := range users {
					ason.EncodeBinary(&users[i])
				}
			}
			binSerMs := float64(time.Since(start).Nanoseconds()) / 1e6

			// Benchmark BIN deserialize
			binBlobs := make([][]byte, len(users))
			for i := range users {
				binBlobs[i], _ = ason.EncodeBinary(&users[i])
			}
			start = time.Now()
			for it := 0; it < iters; it++ {
				for i := range binBlobs {
					var u User
					ason.DecodeBinary(binBlobs[i], &u)
				}
			}
			binDeMs := float64(time.Since(start).Nanoseconds()) / 1e6

			// Benchmark ASON text
			start = time.Now()
			for it := 0; it < iters; it++ {
				ason.Encode(users)
			}
			asonSerMs := float64(time.Since(start).Nanoseconds()) / 1e6
			start = time.Now()
			for it := 0; it < iters; it++ {
				var out []User
				ason.Decode(asonData, &out)
			}
			asonDeMs := float64(time.Since(start).Nanoseconds()) / 1e6

			// Benchmark JSON
			start = time.Now()
			for it := 0; it < iters; it++ {
				json.Marshal(users)
			}
			jsonSerMs := float64(time.Since(start).Nanoseconds()) / 1e6
			start = time.Now()
			for it := 0; it < iters; it++ {
				var out []User
				json.Unmarshal(jsonData, &out)
			}
			jsonDeMs := float64(time.Since(start).Nanoseconds()) / 1e6

			fmt.Printf("  %d records × %d iters:\n", count, iters)
			fmt.Printf("    Size:  BIN %6d B | ASON %6d B | JSON %6d B\n", binTotal, asonTotal, jsonTotal)
			fmt.Printf("    Ser:   BIN %7.1fms | ASON %7.1fms | JSON %7.1fms\n", binSerMs, asonSerMs, jsonSerMs)
			fmt.Printf("    De:    BIN %7.1fms | ASON %7.1fms | JSON %7.1fms\n", binDeMs, asonDeMs, jsonDeMs)
			fmt.Printf("    vs JSON: BIN ser %.1fx | ASON ser %.1fx | BIN de %.1fx | ASON de %.1fx\n\n",
				jsonSerMs/binSerMs, jsonSerMs/asonSerMs,
				jsonDeMs/binDeMs, jsonDeMs/asonDeMs)
		}

		// -- Deep struct benchmark --
		fmt.Println("  ── Deep struct (5-level nested) ──")
		for _, count := range []int{10, 100} {
			iters := 50
			if count >= 100 {
				iters = 10
			}
			companies := generateCompanies(count)

			// Sizes
			var binTotal, asonTotal, jsonTotal int
			for i := range companies {
				b, _ := ason.EncodeBinary(&companies[i])
				binTotal += len(b)
				a, _ := ason.Encode(&companies[i])
				asonTotal += len(a)
				j, _ := json.Marshal(&companies[i])
				jsonTotal += len(j)
			}

			// BIN serialize
			start := time.Now()
			for it := 0; it < iters; it++ {
				for i := range companies {
					ason.EncodeBinary(&companies[i])
				}
			}
			binSerMs := float64(time.Since(start).Nanoseconds()) / 1e6

			// BIN deserialize
			binBlobs := make([][]byte, len(companies))
			for i := range companies {
				binBlobs[i], _ = ason.EncodeBinary(&companies[i])
			}
			start = time.Now()
			for it := 0; it < iters; it++ {
				for i := range binBlobs {
					var c Company
					ason.DecodeBinary(binBlobs[i], &c)
				}
			}
			binDeMs := float64(time.Since(start).Nanoseconds()) / 1e6

			// ASON text
			asonBlobs := make([][]byte, len(companies))
			for i := range companies {
				asonBlobs[i], _ = ason.Encode(&companies[i])
			}
			start = time.Now()
			for it := 0; it < iters; it++ {
				for i := range companies {
					ason.Encode(&companies[i])
				}
			}
			asonSerMs := float64(time.Since(start).Nanoseconds()) / 1e6
			start = time.Now()
			for it := 0; it < iters; it++ {
				for i := range asonBlobs {
					var c Company
					ason.Decode(asonBlobs[i], &c)
				}
			}
			asonDeMs := float64(time.Since(start).Nanoseconds()) / 1e6

			// JSON
			jsonBlobs := make([][]byte, len(companies))
			for i := range companies {
				jsonBlobs[i], _ = json.Marshal(&companies[i])
			}
			start = time.Now()
			for it := 0; it < iters; it++ {
				for i := range companies {
					json.Marshal(&companies[i])
				}
			}
			jsonSerMs := float64(time.Since(start).Nanoseconds()) / 1e6
			start = time.Now()
			for it := 0; it < iters; it++ {
				for i := range jsonBlobs {
					var c Company
					json.Unmarshal(jsonBlobs[i], &c)
				}
			}
			jsonDeMs := float64(time.Since(start).Nanoseconds()) / 1e6

			fmt.Printf("  %d companies × %d iters:\n", count, iters)
			fmt.Printf("    Size:  BIN %6d B | ASON %6d B | JSON %6d B\n", binTotal, asonTotal, jsonTotal)
			fmt.Printf("    Ser:   BIN %7.1fms | ASON %7.1fms | JSON %7.1fms\n", binSerMs, asonSerMs, jsonSerMs)
			fmt.Printf("    De:    BIN %7.1fms | ASON %7.1fms | JSON %7.1fms\n", binDeMs, asonDeMs, jsonDeMs)
			fmt.Printf("    vs JSON: BIN ser %.1fx | ASON ser %.1fx | BIN de %.1fx | ASON de %.1fx\n\n",
				jsonSerMs/binSerMs, jsonSerMs/asonSerMs,
				jsonDeMs/binDeMs, jsonDeMs/asonDeMs)
		}

		// -- Single struct roundtrip --
		fmt.Println("  ── Single User roundtrip ──")
		{
			user := User{ID: 42, Name: "Alice", Email: "alice@example.com", Age: 30, Score: 9.8, Active: true, Role: "admin", City: "Berlin"}
			rtIters := 100000

			start := time.Now()
			for i := 0; i < rtIters; i++ {
				b, _ := ason.EncodeBinary(&user)
				var u User
				ason.DecodeBinary(b, &u)
			}
			binNs := float64(time.Since(start).Nanoseconds()) / float64(rtIters)

			start = time.Now()
			for i := 0; i < rtIters; i++ {
				s, _ := ason.Encode(&user)
				var u User
				ason.Decode(s, &u)
			}
			asonNs := float64(time.Since(start).Nanoseconds()) / float64(rtIters)

			start = time.Now()
			for i := 0; i < rtIters; i++ {
				s, _ := json.Marshal(&user)
				var u User
				json.Unmarshal(s, &u)
			}
			jsonNs := float64(time.Since(start).Nanoseconds()) / float64(rtIters)

			fmt.Printf("    × %d: BIN %6.0fns | ASON %6.0fns | JSON %6.0fns\n",
				rtIters, binNs, asonNs, jsonNs)
			fmt.Printf("    Speedup vs JSON: BIN %.1fx faster | ASON %.1fx faster\n",
				jsonNs/binNs, jsonNs/asonNs)
		}
	}

	fmt.Println("\n╔══════════════════════════════════════════════════════════════╗")
	fmt.Println("║                    Benchmark Complete                        ║")
	fmt.Println("╚══════════════════════════════════════════════════════════════╝")
}
