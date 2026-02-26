package ason

import (
	"math"
	"reflect"
	"testing"
)

// ============================================================================
// Dimension 1: Extra trailing fields — source has more fields than target
// ============================================================================

type FullUser struct {
	ID     int64   `ason:"id"`
	Name   string  `ason:"name"`
	Age    int     `ason:"age"`
	Active bool    `ason:"active"`
	Score  float64 `ason:"score"`
}

type MiniUser struct {
	ID   int64  `ason:"id"`
	Name string `ason:"name"`
}

func TestCrossCompat_TrailingFieldsDropped(t *testing.T) {
	src := []FullUser{
		{ID: 1, Name: "Alice", Age: 30, Active: true, Score: 95.5},
		{ID: 2, Name: "Bob", Age: 25, Active: false, Score: 87.0},
	}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst []MiniUser
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if len(dst) != 2 {
		t.Fatalf("expected 2, got %d", len(dst))
	}
	if dst[0].ID != 1 || dst[0].Name != "Alice" {
		t.Fatalf("row0 mismatch: %+v", dst[0])
	}
	if dst[1].ID != 2 || dst[1].Name != "Bob" {
		t.Fatalf("row1 mismatch: %+v", dst[1])
	}
}

// Single struct, not vec
func TestCrossCompat_TrailingFieldsSingleStruct(t *testing.T) {
	src := FullUser{ID: 99, Name: "Zara", Age: 40, Active: true, Score: 100.0}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst MiniUser
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.ID != 99 || dst.Name != "Zara" {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 2: Trailing field is a complex type (array, struct, map)
// ============================================================================

type RichProfile struct {
	ID     int64            `ason:"id"`
	Name   string           `ason:"name"`
	Tags   []string         `ason:"tags"`
	Scores []int64          `ason:"scores"`
	Meta   map[string]int64 `ason:"meta"`
}

type ThinProfile struct {
	ID   int64  `ason:"id"`
	Name string `ason:"name"`
}

func TestCrossCompat_SkipTrailingArrayAndMap(t *testing.T) {
	src := RichProfile{
		ID:     1,
		Name:   "Alice",
		Tags:   []string{"go", "rust"},
		Scores: []int64{90, 85, 92},
		Meta:   map[string]int64{"level": 5, "xp": 1200},
	}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst ThinProfile
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.ID != 1 || dst.Name != "Alice" {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 3: Nested struct — inner struct has fewer fields
// ============================================================================

type InnerFull struct {
	X int64   `ason:"x"`
	Y int64   `ason:"y"`
	Z float64 `ason:"z"`
	W bool    `ason:"w"`
}

type OuterFull struct {
	Name  string    `ason:"name"`
	Inner InnerFull `ason:"inner"`
	Flag  bool      `ason:"flag"`
}

type InnerThin struct {
	X int64 `ason:"x"`
	Y int64 `ason:"y"`
}

type OuterThin struct {
	Name  string    `ason:"name"`
	Inner InnerThin `ason:"inner"`
}

func TestCrossCompat_NestedStructFewerFields(t *testing.T) {
	src := OuterFull{Name: "test", Inner: InnerFull{X: 10, Y: 20, Z: 3.14, W: true}, Flag: true}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst OuterThin
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.Name != "test" || dst.Inner.X != 10 || dst.Inner.Y != 20 {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 4: Vec of nested structs — inner elements have extra fields
// ============================================================================

type TaskFull struct {
	Title    string  `ason:"title"`
	Done     bool    `ason:"done"`
	Priority int64   `ason:"priority"`
	Weight   float64 `ason:"weight"`
}

type ProjectFull struct {
	Name  string     `ason:"name"`
	Tasks []TaskFull `ason:"tasks"`
}

type TaskThin struct {
	Title string `ason:"title"`
	Done  bool   `ason:"done"`
}

type ProjectThin struct {
	Name  string     `ason:"name"`
	Tasks []TaskThin `ason:"tasks"`
}

func TestCrossCompat_VecNestedStructArraySkipExtra(t *testing.T) {
	src := []ProjectFull{
		{
			Name: "Alpha",
			Tasks: []TaskFull{
				{Title: "Design", Done: true, Priority: 1, Weight: 0.5},
				{Title: "Code", Done: false, Priority: 2, Weight: 0.8},
			},
		},
		{
			Name: "Beta",
			Tasks: []TaskFull{
				{Title: "Test", Done: false, Priority: 3, Weight: 1.0},
			},
		},
	}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst []ProjectThin
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if len(dst) != 2 {
		t.Fatalf("expected 2, got %d", len(dst))
	}
	if dst[0].Name != "Alpha" || len(dst[0].Tasks) != 2 {
		t.Fatalf("row0: %+v", dst[0])
	}
	if dst[0].Tasks[0].Title != "Design" || !dst[0].Tasks[0].Done {
		t.Fatalf("task0: %+v", dst[0].Tasks[0])
	}
	if dst[0].Tasks[1].Title != "Code" || dst[0].Tasks[1].Done {
		t.Fatalf("task1: %+v", dst[0].Tasks[1])
	}
	if dst[1].Name != "Beta" || len(dst[1].Tasks) != 1 {
		t.Fatalf("row1: %+v", dst[1])
	}
}

// ============================================================================
// Dimension 5: Deeply nested — 3 levels, each level drops fields
// ============================================================================

type L3Full struct {
	A int64  `ason:"a"`
	B string `ason:"b"`
	C bool   `ason:"c"`
}

type L2Full struct {
	Name string   `ason:"name"`
	Sub  L3Full   `ason:"sub"`
	Code int64    `ason:"code"`
	Tags []string `ason:"tags"`
}

type L1Full struct {
	ID    int64  `ason:"id"`
	Child L2Full `ason:"child"`
	Extra string `ason:"extra"`
}

type L3Thin struct {
	A int64 `ason:"a"`
}

type L2Thin struct {
	Name string `ason:"name"`
	Sub  L3Thin `ason:"sub"`
}

type L1Thin struct {
	ID    int64  `ason:"id"`
	Child L2Thin `ason:"child"`
}

func TestCrossCompat_DeepNesting3Levels(t *testing.T) {
	src := L1Full{
		ID:    1,
		Child: L2Full{Name: "mid", Sub: L3Full{A: 42, B: "hello", C: true}, Code: 7, Tags: []string{"x", "y"}},
		Extra: "dropped",
	}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst L1Thin
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.ID != 1 || dst.Child.Name != "mid" || dst.Child.Sub.A != 42 {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 6: Schema field reorder — source and target have different order
// ============================================================================

type OrderABC struct {
	A int64  `ason:"a"`
	B string `ason:"b"`
	C bool   `ason:"c"`
}

type OrderCAB struct {
	C bool   `ason:"c"`
	A int64  `ason:"a"`
	B string `ason:"b"`
}

func TestCrossCompat_FieldReorder(t *testing.T) {
	src := OrderABC{A: 1, B: "hi", C: true}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	// Schema is {a,b,c}, decoder maps by name
	var dst OrderCAB
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.A != 1 || dst.B != "hi" || !dst.C {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 7: Reorder + drop trailing — combined
// ============================================================================

type BigRecord struct {
	ID     int64   `ason:"id"`
	Name   string  `ason:"name"`
	Score  float64 `ason:"score"`
	Active bool    `ason:"active"`
	Level  int64   `ason:"level"`
}

type SmallReordered struct {
	Score float64 `ason:"score"`
	ID    int64   `ason:"id"`
}

func TestCrossCompat_ReorderPlusDropTrailing(t *testing.T) {
	src := []BigRecord{
		{ID: 1, Name: "A", Score: 9.5, Active: true, Level: 3},
		{ID: 2, Name: "B", Score: 8.0, Active: false, Level: 1},
	}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst []SmallReordered
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if len(dst) != 2 {
		t.Fatalf("expected 2, got %d", len(dst))
	}
	if dst[0].ID != 1 || dst[0].Score != 9.5 {
		t.Fatalf("row0: %+v", dst[0])
	}
	if dst[1].ID != 2 || dst[1].Score != 8.0 {
		t.Fatalf("row1: %+v", dst[1])
	}
}

// ============================================================================
// Dimension 8: Target has fields that don't exist in source (zero-value)
// ============================================================================

type SrcSmall struct {
	ID   int64  `ason:"id"`
	Name string `ason:"name"`
}

type DstBig struct {
	ID      int64   `ason:"id"`
	Name    string  `ason:"name"`
	Missing bool    `ason:"missing"`
	Extra   float64 `ason:"extra"`
}

func TestCrossCompat_TargetHasExtraFields(t *testing.T) {
	src := SrcSmall{ID: 42, Name: "Alice"}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst DstBig
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.ID != 42 || dst.Name != "Alice" {
		t.Fatalf("mismatch: %+v", dst)
	}
	// Missing fields should be zero-value
	if dst.Missing != false || dst.Extra != 0.0 {
		t.Fatalf("expected zero values for missing fields: %+v", dst)
	}
}

// ============================================================================
// Dimension 9: Optional (pointer) fields across compat
// ============================================================================

type SrcWithOptionals struct {
	ID    int64    `ason:"id"`
	Label *string  `ason:"label"`
	Score *float64 `ason:"score"`
	Flag  bool     `ason:"flag"`
}

type DstFewerOptionals struct {
	ID    int64   `ason:"id"`
	Label *string `ason:"label"`
}

func TestCrossCompat_OptionalFieldsSkipTrailing(t *testing.T) {
	lbl := "hello"
	sc := 95.5
	src := SrcWithOptionals{ID: 1, Label: &lbl, Score: &sc, Flag: true}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst DstFewerOptionals
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.ID != 1 || dst.Label == nil || *dst.Label != "hello" {
		t.Fatalf("mismatch: %+v", dst)
	}
}

func TestCrossCompat_OptionalNilSkipTrailing(t *testing.T) {
	src := SrcWithOptionals{ID: 2, Label: nil, Score: nil, Flag: false}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst DstFewerOptionals
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.ID != 2 || dst.Label != nil {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 10: Trailing field is a quoted string with special chars
// ============================================================================

type SrcSpecialStr struct {
	ID   int64  `ason:"id"`
	Name string `ason:"name"`
	Bio  string `ason:"bio"`
}

type DstNoStr struct {
	ID int64 `ason:"id"`
}

func TestCrossCompat_SkipQuotedStringWithSpecialChars(t *testing.T) {
	src := SrcSpecialStr{ID: 1, Name: `comma,here`, Bio: `paren(test) and "quotes"`}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst DstNoStr
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.ID != 1 {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 11: Vec of vec — nested arrays as trailing field
// ============================================================================

type SrcNestedArray struct {
	ID     int64    `ason:"id"`
	Matrix []int64  `ason:"matrix"`
	Tags   []string `ason:"tags"`
}

type DstNestedArrayThin struct {
	ID int64 `ason:"id"`
}

func TestCrossCompat_SkipTrailingArrayFields(t *testing.T) {
	src := []SrcNestedArray{
		{ID: 1, Matrix: []int64{1, 2, 3}, Tags: []string{"a", "b"}},
		{ID: 2, Matrix: []int64{4, 5}, Tags: []string{"c"}},
	}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst []DstNestedArrayThin
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if len(dst) != 2 || dst[0].ID != 1 || dst[1].ID != 2 {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 12: Type widening — int32 source read as int64 target
// ============================================================================

type SrcNarrow struct {
	ID    int32  `ason:"id"`
	Score int32  `ason:"score"`
	Name  string `ason:"name"`
}

type DstWide struct {
	ID    int64  `ason:"id"`
	Score int64  `ason:"score"`
	Name  string `ason:"name"`
}

func TestCrossCompat_IntWidening(t *testing.T) {
	src := SrcNarrow{ID: 100, Score: 999, Name: "wide"}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst DstWide
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.ID != 100 || dst.Score != 999 || dst.Name != "wide" {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 13: Float to int truncation in text (text is "95.5")
//   This tests reading "95" as float64 (no dot) — and vice versa direction
// ============================================================================

type SrcFloats struct {
	ID    int64   `ason:"id"`
	Value float64 `ason:"value"`
}

type DstFloats struct {
	ID    int64   `ason:"id"`
	Value float64 `ason:"value"`
}

func TestCrossCompat_FloatRoundtrip(t *testing.T) {
	src := SrcFloats{ID: 1, Value: 3.14159}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst DstFloats
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if math.Abs(dst.Value-3.14159) > 1e-10 {
		t.Fatalf("float mismatch: got %v", dst.Value)
	}
}

// ============================================================================
// Dimension 14: Negative numbers and edge-case numbers
// ============================================================================

type SrcNegative struct {
	A int64   `ason:"a"`
	B int64   `ason:"b"`
	C float64 `ason:"c"`
	D string  `ason:"d"`
}

type DstNegativeThin struct {
	A int64 `ason:"a"`
	B int64 `ason:"b"`
}

func TestCrossCompat_NegativeNumbersSkipTrailing(t *testing.T) {
	src := SrcNegative{A: -1, B: -999999, C: -3.14, D: "neg"}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst DstNegativeThin
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.A != -1 || dst.B != -999999 {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 15: Empty string fields
// ============================================================================

type SrcEmpty struct {
	ID   int64  `ason:"id"`
	Name string `ason:"name"`
	Bio  string `ason:"bio"`
}

type DstEmptyThin struct {
	ID int64 `ason:"id"`
}

func TestCrossCompat_EmptyStringFields(t *testing.T) {
	src := SrcEmpty{ID: 1, Name: "", Bio: ""}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst DstEmptyThin
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.ID != 1 {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 16: Map field as trailing — complex skip
// ============================================================================

type SrcWithMap struct {
	ID   int64            `ason:"id"`
	Name string           `ason:"name"`
	Meta map[string]int64 `ason:"meta"`
}

type DstNoMap struct {
	ID   int64  `ason:"id"`
	Name string `ason:"name"`
}

func TestCrossCompat_SkipTrailingMapField(t *testing.T) {
	src := SrcWithMap{
		ID:   1,
		Name: "Alice",
		Meta: map[string]int64{"age": 30, "score": 95},
	}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst DstNoMap
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.ID != 1 || dst.Name != "Alice" {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 17: Vec decode with typed schema
// ============================================================================

func TestCrossCompat_TypedSchemaVecDecode(t *testing.T) {
	src := []FullUser{
		{ID: 1, Name: "Alice", Age: 30, Active: true, Score: 95.5},
	}
	data, err := EncodeTyped(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst []MiniUser
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if len(dst) != 1 || dst[0].ID != 1 || dst[0].Name != "Alice" {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 18: Single struct typed schema
// ============================================================================

func TestCrossCompat_TypedSchemaSingleDecode(t *testing.T) {
	src := FullUser{ID: 42, Name: "Bob", Age: 25, Active: false, Score: 88.0}
	data, err := EncodeTyped(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst MiniUser
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.ID != 42 || dst.Name != "Bob" {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 19: Struct with nested struct array + trailing fields
// ============================================================================

type DetailFull struct {
	ID     int64  `ason:"id"`
	Name   string `ason:"name"`
	Age    int    `ason:"age"`
	Gender bool   `ason:"gender"`
}

type UserFull struct {
	Details []DetailFull `ason:"details"`
	Code    int64        `ason:"code"`
	Label   string       `ason:"label"`
}

type PersonThin struct {
	ID   int64  `ason:"id"`
	Name string `ason:"name"`
}

type HumanThin struct {
	Details []PersonThin `ason:"details"`
}

func TestCrossCompat_NestedVecStructPlusTrailingOuterFields(t *testing.T) {
	src := []UserFull{
		{
			Details: []DetailFull{
				{ID: 1, Name: "Alice", Age: 30, Gender: true},
				{ID: 2, Name: "Bob", Age: 25, Gender: false},
			},
			Code:  42,
			Label: "test",
		},
	}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst []HumanThin
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if len(dst) != 1 {
		t.Fatalf("expected 1, got %d", len(dst))
	}
	if len(dst[0].Details) != 2 {
		t.Fatalf("expected 2 details, got %d", len(dst[0].Details))
	}
	if dst[0].Details[0].ID != 1 || dst[0].Details[0].Name != "Alice" {
		t.Fatalf("detail0: %+v", dst[0].Details[0])
	}
	if dst[0].Details[1].ID != 2 || dst[0].Details[1].Name != "Bob" {
		t.Fatalf("detail1: %+v", dst[0].Details[1])
	}
}

// ============================================================================
// Dimension 20: All-bool trailing fields
// ============================================================================

type SrcBools struct {
	ID int64 `ason:"id"`
	A  bool  `ason:"a"`
	B  bool  `ason:"b"`
	C  bool  `ason:"c"`
}

type DstBoolsThin struct {
	ID int64 `ason:"id"`
}

func TestCrossCompat_SkipTrailingBools(t *testing.T) {
	src := []SrcBools{
		{ID: 1, A: true, B: false, C: true},
		{ID: 2, A: false, B: true, C: false},
	}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst []DstBoolsThin
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if len(dst) != 2 || dst[0].ID != 1 || dst[1].ID != 2 {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 21: Only middle fields picked via reorder
// ============================================================================

type SrcFiveFields struct {
	A int64   `ason:"a"`
	B string  `ason:"b"`
	C float64 `ason:"c"`
	D bool    `ason:"d"`
	E int64   `ason:"e"`
}

type DstMiddleOnly struct {
	C float64 `ason:"c"`
}

func TestCrossCompat_PickMiddleFieldOnly(t *testing.T) {
	src := SrcFiveFields{A: 1, B: "hi", C: 3.14, D: true, E: 99}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst DstMiddleOnly
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.C != 3.14 {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 22: Pick last field only
// ============================================================================

type DstLastOnly struct {
	E int64 `ason:"e"`
}

func TestCrossCompat_PickLastFieldOnly(t *testing.T) {
	src := SrcFiveFields{A: 1, B: "hi", C: 3.14, D: true, E: 42}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst DstLastOnly
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.E != 42 {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 23: No overlapping fields at all
// ============================================================================

type SrcAlpha struct {
	X int64  `ason:"x"`
	Y string `ason:"y"`
}

type DstBeta struct {
	P int64  `ason:"p"`
	Q string `ason:"q"`
}

func TestCrossCompat_NoOverlappingFields(t *testing.T) {
	src := SrcAlpha{X: 1, Y: "hello"}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst DstBeta
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	// All fields should be zero-valued since none match
	if dst.P != 0 || dst.Q != "" {
		t.Fatalf("expected zero values: %+v", dst)
	}
}

// ============================================================================
// Dimension 24: Nested struct with array of structs — extra nested fields
// ============================================================================

type WorkerFull struct {
	Name    string   `ason:"name"`
	Skills  []string `ason:"skills"`
	YearsXP int64    `ason:"years_xp"`
	Rating  float64  `ason:"rating"`
}

type TeamFull struct {
	Lead    string       `ason:"lead"`
	Workers []WorkerFull `ason:"workers"`
	Budget  float64      `ason:"budget"`
}

type WorkerThin struct {
	Name   string   `ason:"name"`
	Skills []string `ason:"skills"`
}

type TeamThin struct {
	Lead    string       `ason:"lead"`
	Workers []WorkerThin `ason:"workers"`
}

func TestCrossCompat_NestedArrayOfStructsWithExtraFields(t *testing.T) {
	src := TeamFull{
		Lead: "Alice",
		Workers: []WorkerFull{
			{Name: "Bob", Skills: []string{"go", "rust"}, YearsXP: 5, Rating: 4.5},
			{Name: "Carol", Skills: []string{"python"}, YearsXP: 3, Rating: 3.8},
		},
		Budget: 100000.0,
	}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst TeamThin
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.Lead != "Alice" || len(dst.Workers) != 2 {
		t.Fatalf("mismatch: %+v", dst)
	}
	if dst.Workers[0].Name != "Bob" || !reflect.DeepEqual(dst.Workers[0].Skills, []string{"go", "rust"}) {
		t.Fatalf("worker0: %+v", dst.Workers[0])
	}
	if dst.Workers[1].Name != "Carol" || !reflect.DeepEqual(dst.Workers[1].Skills, []string{"python"}) {
		t.Fatalf("worker1: %+v", dst.Workers[1])
	}
}

// ============================================================================
// Dimension 25: Encode with typed annotations, decode with struct that has
//               both missing and extra fields
// ============================================================================

type SrcTyped struct {
	A int64   `ason:"a"`
	B string  `ason:"b"`
	C float64 `ason:"c"`
	D bool    `ason:"d"`
}

type DstMixed struct {
	B     string  `ason:"b"`
	D     bool    `ason:"d"`
	Extra int64   `ason:"extra"`
	More  float64 `ason:"more"`
}

func TestCrossCompat_TypedSchemaMixedFields(t *testing.T) {
	src := SrcTyped{A: 1, B: "test", C: 2.5, D: true}
	data, err := EncodeTyped(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst DstMixed
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.B != "test" || !dst.D {
		t.Fatalf("matched fields wrong: %+v", dst)
	}
	if dst.Extra != 0 || dst.More != 0.0 {
		t.Fatalf("unmatched fields should be zero: %+v", dst)
	}
}

// ============================================================================
// Dimension 26: Large number of trailing fields (stress skip)
// ============================================================================

type SrcWide struct {
	F1  int64  `ason:"f1"`
	F2  string `ason:"f2"`
	F3  bool   `ason:"f3"`
	F4  int64  `ason:"f4"`
	F5  string `ason:"f5"`
	F6  bool   `ason:"f6"`
	F7  int64  `ason:"f7"`
	F8  string `ason:"f8"`
	F9  bool   `ason:"f9"`
	F10 int64  `ason:"f10"`
}

type DstNarrow struct {
	F1 int64 `ason:"f1"`
}

func TestCrossCompat_ManyTrailingFields(t *testing.T) {
	src := SrcWide{F1: 42, F2: "a", F3: true, F4: 4, F5: "b", F6: false, F7: 7, F8: "c", F9: true, F10: 10}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst DstNarrow
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.F1 != 42 {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 27: Vec with single row — edge case
// ============================================================================

func TestCrossCompat_VecSingleRow(t *testing.T) {
	src := []FullUser{{ID: 1, Name: "Alice", Age: 30, Active: true, Score: 95.5}}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst []MiniUser
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if len(dst) != 1 || dst[0].ID != 1 || dst[0].Name != "Alice" {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 28: Encoded string that contains ASON-like syntax
// ============================================================================

type SrcAsonLike struct {
	ID   int64  `ason:"id"`
	Data string `ason:"data"`
	Code string `ason:"code"`
}

type DstAsonLikeThin struct {
	ID int64 `ason:"id"`
}

func TestCrossCompat_SkipStringContainingAsonSyntax(t *testing.T) {
	src := SrcAsonLike{
		ID:   1,
		Data: `{a,b}:(1,2)`,
		Code: `[(x,y),(z,w)]`,
	}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst DstAsonLikeThin
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.ID != 1 {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 29: Unicode strings in trailing fields
// ============================================================================

type SrcUnicode struct {
	ID   int64  `ason:"id"`
	Name string `ason:"name"`
	Bio  string `ason:"bio"`
}

type DstUnicodeThin struct {
	ID int64 `ason:"id"`
}

func TestCrossCompat_SkipUnicodeInTrailing(t *testing.T) {
	src := SrcUnicode{ID: 1, Name: "日本語テスト", Bio: "中文描述，包含逗号"}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst DstUnicodeThin
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.ID != 1 {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 30: Roundtrip — encode as A, decode as B, encode as B, decode as A
// ============================================================================

type VersionA struct {
	ID     int64  `ason:"id"`
	Name   string `ason:"name"`
	Active bool   `ason:"active"`
}

type VersionB struct {
	ID   int64  `ason:"id"`
	Name string `ason:"name"`
}

func TestCrossCompat_RoundtripABBA(t *testing.T) {
	// A->B
	srcA := VersionA{ID: 1, Name: "test", Active: true}
	dataA, err := Encode(srcA)
	if err != nil {
		t.Fatal(err)
	}
	var dstB VersionB
	if err := Decode(dataA, &dstB); err != nil {
		t.Fatal(err)
	}
	if dstB.ID != 1 || dstB.Name != "test" {
		t.Fatalf("A->B failed: %+v", dstB)
	}

	// B->A (missing fields = zero)
	dataB, err := Encode(dstB)
	if err != nil {
		t.Fatal(err)
	}
	var dstA VersionA
	if err := Decode(dataB, &dstA); err != nil {
		t.Fatal(err)
	}
	if dstA.ID != 1 || dstA.Name != "test" || dstA.Active != false {
		t.Fatalf("B->A failed: %+v", dstA)
	}
}

// ============================================================================
// Dimension 31: Multiple rows, some empty arrays in nested
// ============================================================================

type SrcWithArr struct {
	ID    int64    `ason:"id"`
	Items []string `ason:"items"`
	Score int64    `ason:"score"`
}

type DstWithArrThin struct {
	ID    int64    `ason:"id"`
	Items []string `ason:"items"`
}

func TestCrossCompat_EmptyArrayInMiddleField(t *testing.T) {
	src := []SrcWithArr{
		{ID: 1, Items: []string{}, Score: 10},
		{ID: 2, Items: []string{"a", "b"}, Score: 20},
		{ID: 3, Items: nil, Score: 30},
	}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst []DstWithArrThin
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if len(dst) != 3 {
		t.Fatalf("expected 3, got %d", len(dst))
	}
	if dst[0].ID != 1 || len(dst[0].Items) != 0 {
		t.Fatalf("row0: %+v", dst[0])
	}
	if dst[1].ID != 2 || !reflect.DeepEqual(dst[1].Items, []string{"a", "b"}) {
		t.Fatalf("row1: %+v", dst[1])
	}
	if dst[2].ID != 3 {
		t.Fatalf("row2: %+v", dst[2])
	}
}

// ============================================================================
// Dimension 32: Nested struct read as flat — inner struct becomes unmatched
// ============================================================================

type SrcWithNested struct {
	ID    int64 `ason:"id"`
	Inner struct {
		A int64  `ason:"a"`
		B string `ason:"b"`
	} `ason:"inner"`
	Tail string `ason:"tail"`
}

type DstFlat struct {
	ID int64 `ason:"id"`
}

func TestCrossCompat_SkipNestedStructAsTuple(t *testing.T) {
	type Inner struct {
		A int64  `ason:"a"`
		B string `ason:"b"`
	}
	type Src struct {
		ID    int64  `ason:"id"`
		Inner Inner  `ason:"inner"`
		Tail  string `ason:"tail"`
	}
	src := Src{ID: 1, Inner: Inner{A: 10, B: "nested"}, Tail: "end"}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst DstFlat
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.ID != 1 {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 33: Vec with many rows — stress test skip logic
// ============================================================================

func TestCrossCompat_ManyRows(t *testing.T) {
	src := make([]FullUser, 100)
	for i := range src {
		src[i] = FullUser{ID: int64(i), Name: "user", Age: i, Active: i%2 == 0, Score: float64(i) * 0.1}
	}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst []MiniUser
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if len(dst) != 100 {
		t.Fatalf("expected 100, got %d", len(dst))
	}
	for i, d := range dst {
		if d.ID != int64(i) || d.Name != "user" {
			t.Fatalf("row %d mismatch: %+v", i, d)
		}
	}
}

// ============================================================================
// Dimension 34: Typed encode, target has subset + reorder
// ============================================================================

func TestCrossCompat_TypedEncodeSubsetReorder(t *testing.T) {
	src := []BigRecord{
		{ID: 1, Name: "A", Score: 9.5, Active: true, Level: 3},
		{ID: 2, Name: "B", Score: 8.0, Active: false, Level: 1},
	}
	data, err := EncodeTyped(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst []SmallReordered
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if len(dst) != 2 || dst[0].Score != 9.5 || dst[0].ID != 1 {
		t.Fatalf("mismatch: %+v", dst)
	}
}

// ============================================================================
// Dimension 35: Zero-value source fields
// ============================================================================

func TestCrossCompat_ZeroValueSourceFields(t *testing.T) {
	src := FullUser{ID: 0, Name: "", Age: 0, Active: false, Score: 0.0}
	data, err := Encode(src)
	if err != nil {
		t.Fatal(err)
	}
	var dst MiniUser
	if err := Decode(data, &dst); err != nil {
		t.Fatal(err)
	}
	if dst.ID != 0 || dst.Name != "" {
		t.Fatalf("mismatch: %+v", dst)
	}
}
