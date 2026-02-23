package ason

import (
	"reflect"
	"testing"
)

type User struct {
	ID     int64  `ason:"id"`
	Name   string `ason:"name"`
	Active bool   `ason:"active"`
}

type Score struct {
	ID    int64   `ason:"id"`
	Value float64 `ason:"value"`
}

type Item struct {
	ID    int64   `ason:"id"`
	Label *string `ason:"label"`
}

type Tagged struct {
	Name string   `ason:"name"`
	Tags []string `ason:"tags"`
}

type Note struct {
	Text string `ason:"text"`
}

type Row struct {
	ID   int64  `ason:"id"`
	Name string `ason:"name"`
}

func TestSerializeStruct(t *testing.T) {
	u := User{ID: 1, Name: "Alice", Active: true}
	got, err := Encode(u)
	if err != nil {
		t.Fatal(err)
	}
	expect := "{id,name,active}:(1,Alice,true)"
	if string(got) != expect {
		t.Fatalf("got %q, want %q", got, expect)
	}
}

func TestDeserializeStructWithSchema(t *testing.T) {
	input := "{id,name,active}:(1,Alice,true)"
	var u User
	if err := Decode([]byte(input), &u); err != nil {
		t.Fatal(err)
	}
	if u.ID != 1 || u.Name != "Alice" || !u.Active {
		t.Fatalf("unexpected: %+v", u)
	}
}

func TestDeserializeStructWithTypedSchema(t *testing.T) {
	input := "{id:int,name:str,active:bool}:(1,Alice,true)"
	var u User
	if err := Decode([]byte(input), &u); err != nil {
		t.Fatal(err)
	}
	if u.ID != 1 || u.Name != "Alice" || !u.Active {
		t.Fatalf("unexpected: %+v", u)
	}
}

func TestRoundtrip(t *testing.T) {
	u := User{ID: 42, Name: "Bob", Active: false}
	data, err := Encode(u)
	if err != nil {
		t.Fatal(err)
	}
	var u2 User
	if err := Decode(data, &u2); err != nil {
		t.Fatal(err)
	}
	if u != u2 {
		t.Fatalf("roundtrip failed: %+v != %+v", u, u2)
	}
}

func TestVecDeserialize(t *testing.T) {
	input := "[{id:int,name:str,active:bool}]:(1,Alice,true),(2,Bob,false)"
	var users []User
	if err := Decode([]byte(input), &users); err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2, got %d", len(users))
	}
	if users[0].Name != "Alice" || users[1].Name != "Bob" {
		t.Fatalf("unexpected: %+v", users)
	}
}

func TestMultiline(t *testing.T) {
	input := "[{id:int,name:str,active:bool}]:\n  (1, Alice, true),\n  (2, Bob, false)"
	var users []User
	if err := Decode([]byte(input), &users); err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2, got %d", len(users))
	}
}

func TestQuotedString(t *testing.T) {
	input := `{id,name,active}:(1,"Carol Smith",true)`
	var u User
	if err := Decode([]byte(input), &u); err != nil {
		t.Fatal(err)
	}
	if u.Name != "Carol Smith" {
		t.Fatalf("got %q", u.Name)
	}
}

func TestOptionalField(t *testing.T) {
	input := "{id,label}:(1,)"
	var item Item
	if err := Decode([]byte(input), &item); err != nil {
		t.Fatal(err)
	}
	if item.ID != 1 || item.Label != nil {
		t.Fatalf("unexpected: %+v", item)
	}
}

func TestArrayField(t *testing.T) {
	input := "{name,tags}:(Alice,[rust,go])"
	var tg Tagged
	if err := Decode([]byte(input), &tg); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(tg.Tags, []string{"rust", "go"}) {
		t.Fatalf("unexpected tags: %v", tg.Tags)
	}
}

func TestNestedStruct(t *testing.T) {
	type Dept struct {
		Title string `ason:"title"`
	}
	type Employee struct {
		Name string `ason:"name"`
		Dept Dept   `ason:"dept"`
	}
	input := "{name,dept:{title}}:(Alice,(Manager))"
	var e Employee
	if err := Decode([]byte(input), &e); err != nil {
		t.Fatal(err)
	}
	if e.Name != "Alice" || e.Dept.Title != "Manager" {
		t.Fatalf("unexpected: %+v", e)
	}
}

func TestSerializeVec(t *testing.T) {
	rows := []Row{{ID: 1, Name: "Alice"}, {ID: 2, Name: "Bob"}}
	got, err := Encode(rows)
	if err != nil {
		t.Fatal(err)
	}
	expect := "[{id,name}]:(1,Alice),(2,Bob)"
	if string(got) != expect {
		t.Fatalf("got %q, want %q", got, expect)
	}
}

func TestEscapeRoundtrip(t *testing.T) {
	note := Note{Text: "hello, world (test)"}
	data, err := Encode(note)
	if err != nil {
		t.Fatal(err)
	}
	var note2 Note
	if err := Decode(data, &note2); err != nil {
		t.Fatal(err)
	}
	if note != note2 {
		t.Fatalf("roundtrip failed: %+v != %+v", note, note2)
	}
}

func TestTrailingComma(t *testing.T) {
	input := "[{id:int,name:str,active:bool}]:(1,Alice,true),(2,Bob,false),"
	var users []User
	if err := Decode([]byte(input), &users); err != nil {
		t.Fatal(err)
	}
	if len(users) != 2 {
		t.Fatalf("expected 2, got %d", len(users))
	}
}

func TestCommentStripping(t *testing.T) {
	input := "/* users */ {id,name,active}:(1,Alice,true)"
	var u User
	if err := Decode([]byte(input), &u); err != nil {
		t.Fatal(err)
	}
	if u.ID != 1 {
		t.Fatalf("unexpected id: %d", u.ID)
	}
}

func TestFloatField(t *testing.T) {
	input := "{id,value}:(1,95.5)"
	var s Score
	if err := Decode([]byte(input), &s); err != nil {
		t.Fatal(err)
	}
	if s.Value != 95.5 {
		t.Fatalf("got %f", s.Value)
	}
}

func TestMapField(t *testing.T) {
	type MapItem struct {
		Name  string           `ason:"name"`
		Attrs map[string]int64 `ason:"attrs"`
	}
	input := "{name,attrs}:(Alice,[(age,30),(score,95)])"
	var item MapItem
	if err := Decode([]byte(input), &item); err != nil {
		t.Fatal(err)
	}
	if item.Attrs["age"] != 30 || item.Attrs["score"] != 95 {
		t.Fatalf("unexpected: %+v", item.Attrs)
	}
}

func TestAnnotatedSimpleStruct(t *testing.T) {
	typed := "{id:int,name:str,active:bool}:(42,Bob,false)"
	untyped := "{id,name,active}:(42,Bob,false)"
	var u1, u2 User
	if err := Decode([]byte(typed), &u1); err != nil {
		t.Fatal(err)
	}
	if err := Decode([]byte(untyped), &u2); err != nil {
		t.Fatal(err)
	}
	if u1 != u2 {
		t.Fatalf("mismatch")
	}
	if u1.ID != 42 || u1.Name != "Bob" || u1.Active {
		t.Fatalf("unexpected: %+v", u1)
	}
}

func TestAnnotatedVec(t *testing.T) {
	typed := "[{id:int,name:str,active:bool}]:(1,Alice,true),(2,Bob,false)"
	untyped := "[{id,name,active}]:(1,Alice,true),(2,Bob,false)"
	var v1, v2 []User
	if err := Decode([]byte(typed), &v1); err != nil {
		t.Fatal(err)
	}
	if err := Decode([]byte(untyped), &v2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(v1, v2) || len(v1) != 2 {
		t.Fatalf("mismatch")
	}
}

func TestAnnotatedNestedStruct(t *testing.T) {
	type Dept struct {
		Title  string  `ason:"title"`
		Budget float64 `ason:"budget"`
	}
	type Employee struct {
		Name   string `ason:"name"`
		Age    int64  `ason:"age"`
		Dept   Dept   `ason:"dept"`
		Active bool   `ason:"active"`
	}
	typed := "{name:str,age:int,dept:{title:str,budget:float},active:bool}:(Alice,30,(Engineering,50000.5),true)"
	untyped := "{name,age,dept:{title,budget},active}:(Alice,30,(Engineering,50000.5),true)"
	var e1, e2 Employee
	if err := Decode([]byte(typed), &e1); err != nil {
		t.Fatal(err)
	}
	if err := Decode([]byte(untyped), &e2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(e1, e2) {
		t.Fatalf("mismatch")
	}
	if e1.Name != "Alice" || e1.Dept.Title != "Engineering" || e1.Dept.Budget != 50000.5 {
		t.Fatalf("unexpected: %+v", e1)
	}
}

func TestAnnotatedWithArrays(t *testing.T) {
	type Profile struct {
		Name   string   `ason:"name"`
		Scores []int64  `ason:"scores"`
		Tags   []string `ason:"tags"`
	}
	typed := "{name:str,scores:[int],tags:[str]}:(Alice,[90,85,92],[rust,go])"
	untyped := "{name,scores,tags}:(Alice,[90,85,92],[rust,go])"
	var p1, p2 Profile
	if err := Decode([]byte(typed), &p1); err != nil {
		t.Fatal(err)
	}
	if err := Decode([]byte(untyped), &p2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(p1, p2) {
		t.Fatalf("mismatch")
	}
	if !reflect.DeepEqual(p1.Scores, []int64{90, 85, 92}) {
		t.Fatalf("unexpected scores: %v", p1.Scores)
	}
}

func TestAnnotatedWithMap(t *testing.T) {
	type Config struct {
		Name  string           `ason:"name"`
		Attrs map[string]int64 `ason:"attrs"`
	}
	typed := "{name:str,attrs:map[str,int]}:(server,[(port,8080),(timeout,30)])"
	untyped := "{name,attrs}:(server,[(port,8080),(timeout,30)])"
	var c1, c2 Config
	if err := Decode([]byte(typed), &c1); err != nil {
		t.Fatal(err)
	}
	if err := Decode([]byte(untyped), &c2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c1, c2) {
		t.Fatalf("mismatch")
	}
	if c1.Attrs["port"] != 8080 {
		t.Fatalf("unexpected port")
	}
}

func TestAnnotatedWithOptional(t *testing.T) {
	type Record struct {
		ID    int64    `ason:"id"`
		Label *string  `ason:"label"`
		Score *float64 `ason:"score"`
	}
	typed := "{id:int,label:str,score:float}:(1,hello,95.5)"
	untyped := "{id,label,score}:(1,hello,95.5)"
	var r1, r2 Record
	if err := Decode([]byte(typed), &r1); err != nil {
		t.Fatal(err)
	}
	if err := Decode([]byte(untyped), &r2); err != nil {
		t.Fatal(err)
	}
	if *r1.Label != *r2.Label || *r1.Score != *r2.Score {
		t.Fatalf("mismatch")
	}
	typedNone := "{id:int,label:str,score:float}:(2,,)"
	untypedNone := "{id,label,score}:(2,,)"
	var r3, r4 Record
	if err := Decode([]byte(typedNone), &r3); err != nil {
		t.Fatal(err)
	}
	if err := Decode([]byte(untypedNone), &r4); err != nil {
		t.Fatal(err)
	}
	if r3.Label != nil || r3.Score != nil {
		t.Fatalf("expected nil")
	}
}

func TestAnnotatedDeepNesting(t *testing.T) {
	type Task struct {
		Title string `ason:"title"`
		Done  bool   `ason:"done"`
	}
	type Project struct {
		Name  string `ason:"name"`
		Tasks []Task `ason:"tasks"`
	}
	type Team struct {
		Lead     string    `ason:"lead"`
		Projects []Project `ason:"projects"`
	}
	type Company struct {
		Name    string  `ason:"name"`
		Revenue float64 `ason:"revenue"`
		Team    Team    `ason:"team"`
	}
	typed := "{name:str,revenue:float,team:{lead:str,projects:[{name:str,tasks:[{title:str,done:bool}]}]}}:(Acme,500.5,(Alice,[(API,[(Design,true),(Code,false)])]))"
	untyped := "{name,revenue,team:{lead,projects:[{name,tasks:[{title,done}]}]}}:(Acme,500.5,(Alice,[(API,[(Design,true),(Code,false)])]))"
	var c1, c2 Company
	if err := Decode([]byte(typed), &c1); err != nil {
		t.Fatal(err)
	}
	if err := Decode([]byte(untyped), &c2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(c1, c2) {
		t.Fatalf("mismatch")
	}
	if c1.Name != "Acme" || c1.Team.Lead != "Alice" {
		t.Fatalf("unexpected")
	}
	if c1.Team.Projects[0].Tasks[0].Title != "Design" || !c1.Team.Projects[0].Tasks[0].Done {
		t.Fatalf("unexpected task")
	}
	if c1.Team.Projects[0].Tasks[1].Done {
		t.Fatalf("expected not done")
	}
}

func TestAnnotatedMixedPartial(t *testing.T) {
	type Mixed struct {
		ID     int64   `ason:"id"`
		Name   string  `ason:"name"`
		Score  float64 `ason:"score"`
		Active bool    `ason:"active"`
	}
	partial := "{id:int,name,score:float,active}:(1,Alice,95.5,true)"
	full := "{id:int,name:str,score:float,active:bool}:(1,Alice,95.5,true)"
	none := "{id,name,score,active}:(1,Alice,95.5,true)"
	var m1, m2, m3 Mixed
	if err := Decode([]byte(partial), &m1); err != nil {
		t.Fatal(err)
	}
	if err := Decode([]byte(full), &m2); err != nil {
		t.Fatal(err)
	}
	if err := Decode([]byte(none), &m3); err != nil {
		t.Fatal(err)
	}
	if m1 != m2 || m2 != m3 {
		t.Fatalf("mismatch")
	}
}

func TestSerializerOutputIsUnannotated(t *testing.T) {
	u := User{ID: 1, Name: "Alice", Active: true}
	data, err := Encode(u)
	if err != nil {
		t.Fatal(err)
	}
	s := string(data)
	if s != "{id,name,active}:(1,Alice,true)" {
		t.Fatalf("got %q", s)
	}
}

func TestEncodeTypedSimple(t *testing.T) {
	u := User{ID: 1, Name: "Alice", Active: true}
	got, err := EncodeTyped(u)
	if err != nil {
		t.Fatal(err)
	}
	expect := "{id:int,name:str,active:bool}:(1,Alice,true)"
	if string(got) != expect {
		t.Fatalf("got %q, want %q", got, expect)
	}
}

func TestEncodeTypedRoundtrip(t *testing.T) {
	u := User{ID: 42, Name: "Bob", Active: false}
	data, err := EncodeTyped(u)
	if err != nil {
		t.Fatal(err)
	}
	var u2 User
	if err := Decode(data, &u2); err != nil {
		t.Fatal(err)
	}
	if u != u2 {
		t.Fatalf("roundtrip failed")
	}
}

func TestEncodeTypedFloats(t *testing.T) {
	type ScoreF struct {
		ID    int64   `ason:"id"`
		Value float64 `ason:"value"`
		Label string  `ason:"label"`
	}
	s := ScoreF{ID: 1, Value: 95.5, Label: "good"}
	got, err := EncodeTyped(s)
	if err != nil {
		t.Fatal(err)
	}
	expect := "{id:int,value:float,label:str}:(1,95.5,good)"
	if string(got) != expect {
		t.Fatalf("got %q, want %q", got, expect)
	}
}

func TestEncodeTypedAllPrimitives(t *testing.T) {
	type All struct {
		B bool    `ason:"b"`
		I int64   `ason:"i"`
		U uint32  `ason:"u"`
		F float64 `ason:"f"`
		S string  `ason:"s"`
	}
	val := All{B: true, I: -42, U: 100, F: 3.14, S: "hello"}
	got, err := EncodeTyped(val)
	if err != nil {
		t.Fatal(err)
	}
	expect := "{b:bool,i:int,u:int,f:float,s:str}:(true,-42,100,3.14,hello)"
	if string(got) != expect {
		t.Fatalf("got %q, want %q", got, expect)
	}
}

func TestEncodeTypedOptional(t *testing.T) {
	type Opt struct {
		ID    int64    `ason:"id"`
		Label *string  `ason:"label"`
		Score *float64 `ason:"score"`
	}
	hello := "hello"
	sc := 95.5
	v1 := Opt{ID: 1, Label: &hello, Score: &sc}
	out1, err := EncodeTyped(v1)
	if err != nil {
		t.Fatal(err)
	}
	if string(out1) != "{id:int,label:str,score:float}:(1,hello,95.5)" {
		t.Fatalf("got %q", out1)
	}
	v2 := Opt{ID: 2, Label: nil, Score: nil}
	out2, err := EncodeTyped(v2)
	if err != nil {
		t.Fatal(err)
	}
	if string(out2) != "{id:int,label:str,score:float}:(2,,)" {
		t.Fatalf("got %q", out2)
	}
}

func TestEncodeTypedNestedStruct(t *testing.T) {
	type Dept struct {
		Title string `ason:"title"`
	}
	type Employee struct {
		Name   string `ason:"name"`
		Dept   Dept   `ason:"dept"`
		Active bool   `ason:"active"`
	}
	e := Employee{Name: "Alice", Dept: Dept{Title: "Engineering"}, Active: true}
	got, err := EncodeTyped(e)
	if err != nil {
		t.Fatal(err)
	}
	expect := "{name:str,dept:{title:str},active:bool}:(Alice,(Engineering),true)"
	if string(got) != expect {
		t.Fatalf("got %q, want %q", got, expect)
	}
}

func TestEncodeTypedVec(t *testing.T) {
	type RowT struct {
		ID    int64   `ason:"id"`
		Name  string  `ason:"name"`
		Score float64 `ason:"score"`
	}
	rows := []RowT{
		{ID: 1, Name: "Alice", Score: 95.5},
		{ID: 2, Name: "Bob", Score: 87.0},
	}
	untyped, err := Encode(rows)
	if err != nil {
		t.Fatal(err)
	}
	if string(untyped) != "[{id,name,score}]:(1,Alice,95.5),(2,Bob,87.0)" {
		t.Fatalf("got %q", untyped)
	}
	typed, err := EncodeTyped(rows)
	if err != nil {
		t.Fatal(err)
	}
	if string(typed) != "[{id:int,name:str,score:float}]:(1,Alice,95.5),(2,Bob,87.0)" {
		t.Fatalf("got %q", typed)
	}
	var r1, r2 []RowT
	if err := Decode(untyped, &r1); err != nil {
		t.Fatal(err)
	}
	if err := Decode(typed, &r2); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(r1, r2) {
		t.Fatalf("mismatch")
	}
}
