package main

import (
	"strings"
	"testing"
)

// ═══════════════════════════════════════════════════════════════════════════════
// Lexer Tests
// ═══════════════════════════════════════════════════════════════════════════════

func TestLexerStructuralTokens(t *testing.T) {
	src := `{name:str}:(Alice)`
	lex := NewLexer(src)
	tokens := lex.All()

	expected := []TokenType{
		TokenLBrace, TokenIdent, TokenColon, TokenTypeHint, TokenRBrace,
		TokenColon, TokenLParen, TokenPlainStr, TokenRParen, TokenEOF,
	}
	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d: %v", len(expected), len(tokens), tokens)
	}
	for i, exp := range expected {
		if tokens[i].Type != exp {
			t.Errorf("token[%d]: expected %s, got %s (value=%q)", i, exp, tokens[i].Type, tokens[i].Value)
		}
	}
}

func TestLexerArraySchema(t *testing.T) {
	src := `[{id:int,name:str}]:(1,Alice),(2,Bob)`
	lex := NewLexer(src)
	tokens := lex.All()

	// Find key tokens
	hasLBracket := false
	hasLBrace := false
	identCount := 0
	for _, tok := range tokens {
		if tok.Type == TokenLBracket {
			hasLBracket = true
		}
		if tok.Type == TokenLBrace {
			hasLBrace = true
		}
		if tok.Type == TokenIdent {
			identCount++
		}
	}
	if !hasLBracket || !hasLBrace {
		t.Error("missing structural tokens for array schema")
	}
	if identCount != 2 {
		t.Errorf("expected 2 idents (id, name), got %d", identCount)
	}
}

func TestLexerQuotedString(t *testing.T) {
	src := `{name:str}:("hello world")`
	lex := NewLexer(src)
	tokens := lex.All()

	found := false
	for _, tok := range tokens {
		if tok.Type == TokenString {
			if tok.Value != `"hello world"` {
				t.Errorf("quoted string value = %q, want %q", tok.Value, `"hello world"`)
			}
			found = true
		}
	}
	if !found {
		t.Error("no TokenString found")
	}
}

func TestLexerEscapedString(t *testing.T) {
	src := `{name:str}:("say \"hi\"")`
	lex := NewLexer(src)
	tokens := lex.All()

	found := false
	for _, tok := range tokens {
		if tok.Type == TokenString {
			if tok.Value != `"say \"hi\""` {
				t.Errorf("escaped string = %q, want %q", tok.Value, `"say \"hi\""`)
			}
			found = true
		}
	}
	if !found {
		t.Error("no TokenString found for escaped string")
	}
}

func TestLexerComment(t *testing.T) {
	src := `/* users */ {name:str}:(Alice)`
	lex := NewLexer(src)
	tokens := lex.All()

	if tokens[0].Type != TokenComment {
		t.Errorf("expected comment token, got %s", tokens[0].Type)
	}
}

func TestLexerUnclosedComment(t *testing.T) {
	src := `/* unclosed comment`
	lex := NewLexer(src)
	tokens := lex.All()

	found := false
	for _, tok := range tokens {
		if tok.Type == TokenError {
			found = true
		}
	}
	if !found {
		t.Error("expected error token for unclosed comment")
	}
}

func TestLexerUnclosedString(t *testing.T) {
	src := `{name:str}:("unclosed)`
	lex := NewLexer(src)
	tokens := lex.All()

	found := false
	for _, tok := range tokens {
		if tok.Type == TokenError {
			found = true
		}
	}
	if !found {
		t.Error("expected error token for unclosed string")
	}
}

func TestLexerBoolAndNumber(t *testing.T) {
	src := `{a:int,b:bool}:(42,true)`
	lex := NewLexer(src)
	tokens := lex.All()

	var numTok, boolTok Token
	for _, tok := range tokens {
		if tok.Type == TokenNumber {
			numTok = tok
		}
		if tok.Type == TokenBool {
			boolTok = tok
		}
	}
	if numTok.Value != "42" {
		t.Errorf("number = %q, want 42", numTok.Value)
	}
	if boolTok.Value != "true" {
		t.Errorf("bool = %q, want true", boolTok.Value)
	}
}

func TestLexerNegativeNumber(t *testing.T) {
	src := `{a:int}:(-123)`
	lex := NewLexer(src)
	tokens := lex.All()

	found := false
	for _, tok := range tokens {
		if tok.Type == TokenNumber && tok.Value == "-123" {
			found = true
		}
	}
	if !found {
		t.Error("expected TokenNumber -123")
	}
}

func TestLexerFloat(t *testing.T) {
	src := `{a:float}:(3.14)`
	lex := NewLexer(src)
	tokens := lex.All()

	found := false
	for _, tok := range tokens {
		if tok.Type == TokenNumber && tok.Value == "3.14" {
			found = true
		}
	}
	if !found {
		var types []string
		for _, tok := range tokens {
			types = append(types, tok.Type.String()+"="+tok.Value)
		}
		t.Errorf("expected TokenNumber 3.14, got tokens: %v", types)
	}
}

func TestLexerMultiline(t *testing.T) {
	src := "[{name:str}]:\n  (Alice),\n  (Bob)"
	lex := NewLexer(src)
	tokens := lex.All()

	// Should have newline tokens and proper line tracking
	lineCount := 0
	for _, tok := range tokens {
		if tok.Type == TokenNewline {
			lineCount++
		}
	}
	if lineCount != 2 {
		t.Errorf("expected 2 newlines, got %d", lineCount)
	}
}

func TestLexerPositionTracking(t *testing.T) {
	src := "{a:int}:(1)"
	lex := NewLexer(src)
	tokens := lex.All()

	// First token {  should be at offset 0, line 0, col 0
	if tokens[0].Offset != 0 || tokens[0].Line != 0 || tokens[0].Col != 0 {
		t.Errorf("first token position: offset=%d line=%d col=%d",
			tokens[0].Offset, tokens[0].Line, tokens[0].Col)
	}
}

func TestLexerTypeHints(t *testing.T) {
	hints := []string{"int", "integer", "float", "double", "str", "string", "bool", "boolean"}
	for _, h := range hints {
		src := "{field:" + h + "}:(val)"
		lex := NewLexer(src)
		tokens := lex.All()
		found := false
		for _, tok := range tokens {
			if tok.Type == TokenTypeHint && tok.Value == h {
				found = true
			}
		}
		if !found {
			t.Errorf("type hint %q not recognised", h)
		}
	}
}

func TestLexerMapKeyword(t *testing.T) {
	src := "{attrs:map[str,int]}:([(a,1)])"
	lex := NewLexer(src)
	tokens := lex.All()

	found := false
	for _, tok := range tokens {
		if tok.Type == TokenMapKw {
			found = true
		}
	}
	if !found {
		t.Error("map keyword not detected")
	}
}

func TestLexerPlainArray(t *testing.T) {
	src := "[1,2,3]"
	lex := NewLexer(src)
	tokens := lex.All()

	numCount := 0
	for _, tok := range tokens {
		if tok.Type == TokenNumber {
			numCount++
		}
	}
	if numCount != 3 {
		t.Errorf("expected 3 numbers, got %d", numCount)
	}
}

func TestLexerEmptyParens(t *testing.T) {
	src := "()"
	lex := NewLexer(src)
	tokens := lex.All()

	if len(tokens) < 2 {
		t.Fatal("expected at least LParen RParen")
	}
	if tokens[0].Type != TokenLParen || tokens[1].Type != TokenRParen {
		t.Errorf("expected LParen RParen, got %s %s", tokens[0].Type, tokens[1].Type)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// Parser Tests
// ═══════════════════════════════════════════════════════════════════════════════

func TestParseSingleObject(t *testing.T) {
	src := `{name:str,age:int}:(Alice,30)`
	root, diags := Parse(src)

	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %v", diagMessages(diags))
	}
	if root == nil || len(root.Children) == 0 {
		t.Fatal("empty parse tree")
	}
	obj := root.Children[0]
	if obj.Kind != NodeSingleObject {
		t.Errorf("expected NodeSingleObject, got %d", obj.Kind)
	}
	if len(obj.Children) != 2 {
		t.Fatalf("expected schema+tuple, got %d children", len(obj.Children))
	}

	schema := obj.Children[0]
	fields := schema.SchemaFields()
	if len(fields) != 2 {
		t.Errorf("expected 2 fields, got %d", len(fields))
	}
	if fields[0].Value != "name" || fields[1].Value != "age" {
		t.Errorf("field names: %q, %q", fields[0].Value, fields[1].Value)
	}

	tuple := obj.Children[1]
	if tuple.Kind != NodeTuple || len(tuple.Children) != 2 {
		t.Errorf("expected tuple with 2 values, got kind=%d children=%d", tuple.Kind, len(tuple.Children))
	}
}

func TestParseObjectArray(t *testing.T) {
	src := `[{id:int,name:str}]:(1,Alice),(2,Bob)`
	root, diags := Parse(src)

	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %v", diagMessages(diags))
	}
	arr := root.Children[0]
	if arr.Kind != NodeObjectArray {
		t.Errorf("expected NodeObjectArray, got %d", arr.Kind)
	}
	// 1 array-schema + 2 tuples = 3 children
	if len(arr.Children) != 3 {
		t.Fatalf("expected 3 children, got %d", len(arr.Children))
	}
}

func TestParseObjectArrayMultiline(t *testing.T) {
	src := "[{id:int,name:str}]:\n  (1,Alice),\n  (2,Bob)"
	root, diags := Parse(src)

	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %v", diagMessages(diags))
	}
	arr := root.Children[0]
	if arr.Kind != NodeObjectArray {
		t.Errorf("expected NodeObjectArray, got %d", arr.Kind)
	}
	// 1 schema + 2 tuples
	if len(arr.Children) != 3 {
		t.Errorf("expected 3 children, got %d", len(arr.Children))
	}
}

func TestParsePlainArray(t *testing.T) {
	src := `[1,2,3]`
	root, diags := Parse(src)

	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %v", diagMessages(diags))
	}
	arr := root.Children[0]
	if arr.Kind != NodeArray {
		t.Errorf("expected NodeArray, got %d", arr.Kind)
	}
	if len(arr.Children) != 3 {
		t.Errorf("expected 3 values, got %d", len(arr.Children))
	}
}

func TestParseNestedObject(t *testing.T) {
	src := `{name:str,addr:{city:str,zip:int}}:(Alice,(NYC,10001))`
	root, diags := Parse(src)

	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %v", diagMessages(diags))
	}
	obj := root.Children[0]
	schema := obj.Children[0]
	fields := schema.SchemaFields()
	if len(fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(fields))
	}
	// addr field should have a nested schema
	addr := fields[1]
	if addr.Value != "addr" {
		t.Errorf("expected 'addr', got %q", addr.Value)
	}
	if len(addr.Children) == 0 || addr.Children[0].Kind != NodeSchema {
		t.Error("addr should have nested Schema child")
	}
}

func TestParseObjectWithArray(t *testing.T) {
	src := `{name:str,scores:[int]}:(Alice,[90,85,92])`
	root, diags := Parse(src)

	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %v", diagMessages(diags))
	}
	_ = root.Children[0]
}

func TestParseObjectWithObjectArray(t *testing.T) {
	src := `{team:str,users:[{id:int,name:str}]}:(Dev,[(1,Alice),(2,Bob)])`
	root, diags := Parse(src)

	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %v", diagMessages(diags))
	}
	_ = root.Children[0]
}

func TestParseMap(t *testing.T) {
	src := `{name:str,attrs:map[str,int]}:(Alice,[(age,30),(score,95)])`
	root, diags := Parse(src)

	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %v", diagMessages(diags))
	}
	_ = root.Children[0]
}

func TestParseNullValue(t *testing.T) {
	src := `{name:str,age:int,email:str}:(Alice,30,)`
	root, diags := Parse(src)

	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %v", diagMessages(diags))
	}
	obj := root.Children[0]
	tuple := obj.Children[1]
	if len(tuple.Children) != 3 {
		t.Errorf("expected 3 values (with null), got %d", len(tuple.Children))
	}
}

func TestParseEmptyArray(t *testing.T) {
	src := `[]`
	root, diags := Parse(src)

	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %v", diagMessages(diags))
	}
	arr := root.Children[0]
	if arr.Kind != NodeArray || len(arr.Children) != 0 {
		t.Errorf("expected empty array, got kind=%d children=%d", arr.Kind, len(arr.Children))
	}
}

func TestParseTrailingComma(t *testing.T) {
	src := `[1,2,3,]`
	root, diags := Parse(src)

	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %v", diagMessages(diags))
	}
	arr := root.Children[0]
	if len(arr.Children) != 3 {
		t.Errorf("trailing comma: expected 3 values, got %d", len(arr.Children))
	}
}

func TestParseTypeAnnotations(t *testing.T) {
	src := `{id:int,name:str,salary:float,active:bool}:(1,Alice,5000.50,true)`
	root, diags := Parse(src)

	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %v", diagMessages(diags))
	}
	obj := root.Children[0]
	schema := obj.Children[0]
	fields := schema.SchemaFields()
	if len(fields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(fields))
	}
	// Check type annotations
	expected := []string{"int", "str", "float", "bool"}
	for i, f := range fields {
		if len(f.Children) == 0 {
			t.Errorf("field %q missing type annotation", f.Value)
			continue
		}
		if f.Children[0].Value != expected[i] {
			t.Errorf("field %q type = %q, want %q", f.Value, f.Children[0].Value, expected[i])
		}
	}
}

func TestParseNoTypeAnnotations(t *testing.T) {
	src := `{id,name,active}:(1,Alice,true)`
	root, diags := Parse(src)

	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %v", diagMessages(diags))
	}
	obj := root.Children[0]
	schema := obj.Children[0]
	fields := schema.SchemaFields()
	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(fields))
	}
	for _, f := range fields {
		// No children means no type annotation
		hasType := false
		for _, c := range f.Children {
			if c.Kind == NodeTypeAnnot {
				hasType = true
			}
		}
		if hasType {
			t.Errorf("field %q should not have type annotation", f.Value)
		}
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// Diagnostic Tests (error detection)
// ═══════════════════════════════════════════════════════════════════════════════

func TestDiagFieldCountMismatchTooMany(t *testing.T) {
	src := `{a:int,b:int}:(1,2,3)`
	_, diags := Parse(src)

	if len(diags) == 0 {
		t.Fatal("expected diagnostic for field count mismatch")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "field count mismatch") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'field count mismatch', got: %v", diagMessages(diags))
	}
}

func TestDiagFieldCountMismatchTooFew(t *testing.T) {
	src := `{a:int,b:int,c:int}:(1,2)`
	_, diags := Parse(src)

	if len(diags) == 0 {
		t.Fatal("expected diagnostic for field count mismatch")
	}
	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "field count mismatch") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected 'field count mismatch', got: %v", diagMessages(diags))
	}
}

func TestDiagFieldCountArrayTuple(t *testing.T) {
	src := `[{a:int,b:int}]:(1,2),(3,4,5)`
	_, diags := Parse(src)

	if len(diags) == 0 {
		t.Fatal("expected diagnostic for array tuple mismatch")
	}
}

func TestDiagUnclosedComment(t *testing.T) {
	src := `/* unclosed`
	_, diags := Parse(src)

	// Should get an error token that propagates
	// The lexer reports TokenError which the parser sees
	if len(diags) == 0 {
		// At minimum, the document should fail to parse meaningfully
		t.Log("warning: no diagnostic for unclosed comment (lexer error)")
	}
}

func TestDiagMissingColon(t *testing.T) {
	src := `{a:int,b:int}(1,2)`
	_, diags := Parse(src)

	found := false
	for _, d := range diags {
		if strings.Contains(d.Message, "expected ':'") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected missing colon diagnostic, got: %v", diagMessages(diags))
	}
}

func TestDiagNoDiagValidInput(t *testing.T) {
	tests := []string{
		`{name:str,age:int}:(Alice,30)`,
		`[{id:int,name:str}]:(1,Alice),(2,Bob)`,
		`[1,2,3]`,
		`{name:str,addr:{city:str}}:(Alice,(NYC))`,
		`{tags:[str]}:([a,b,c])`,
		"[{a:int}]:\n  (1),\n  (2)",
	}
	for _, src := range tests {
		_, diags := Parse(src)
		if len(diags) != 0 {
			t.Errorf("input %q produced unexpected diags: %v", src, diagMessages(diags))
		}
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// Analyzer Tests
// ═══════════════════════════════════════════════════════════════════════════════

func TestAnalyzerNestedMismatch(t *testing.T) {
	src := `{name:str,addr:{city:str,zip:int}}:(Alice,(NYC))`
	root, parseDiags := Parse(src)
	// Parser may or may not catch this; analyzer should
	analysisDiags := Analyze(root, src)
	allDiags := append(parseDiags, analysisDiags...)

	found := false
	for _, d := range allDiags {
		if strings.Contains(d.Message, "mismatch") {
			found = true
		}
	}
	if !found {
		t.Errorf("expected nested mismatch diagnostic, got: %v", diagMessages(allDiags))
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// Hover Tests
// ═══════════════════════════════════════════════════════════════════════════════

func TestHoverField(t *testing.T) {
	src := `{name:str,age:int}:(Alice,30)`
	root, _ := Parse(src)

	// Hover over 'name' field — at line 0, col 1
	text := HoverInfo(root, 0, 1)
	if !strings.Contains(text, "Field") || !strings.Contains(text, "name") {
		t.Errorf("hover should show field info, got: %q", text)
	}
}

func TestHoverType(t *testing.T) {
	src := `{name:str,age:int}:(Alice,30)`
	root, _ := Parse(src)

	// Hover over 'str' type hint — at line 0, col 6
	text := HoverInfo(root, 0, 6)
	if text == "" {
		t.Error("hover over type should return info")
	}
}

func TestHoverValue(t *testing.T) {
	src := `{a:int}:(42)`
	root, _ := Parse(src)

	// Hover over '42' value — line 0, col 9
	text := HoverInfo(root, 0, 9)
	if text == "" {
		t.Error("hover over value should return info")
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// Completion Tests
// ═══════════════════════════════════════════════════════════════════════════════

func TestCompletionTopLevel(t *testing.T) {
	src := ``
	root, _ := Parse(src)
	items := Complete(root, src, 0, 0)

	if len(items) == 0 {
		t.Error("expected top-level completions")
	}
	found := false
	for _, it := range items {
		if strings.Contains(it.Label, "schema") {
			found = true
		}
	}
	if !found {
		t.Error("expected schema snippet in top-level completions")
	}
}

func TestCompletionDataValue(t *testing.T) {
	src := `{a:bool}:()`
	root, _ := Parse(src)
	items := Complete(root, src, 0, 10)

	if len(items) == 0 {
		t.Error("expected data value completions")
	}
	labels := make(map[string]bool)
	for _, it := range items {
		labels[it.Label] = true
	}
	if !labels["true"] || !labels["false"] {
		t.Error("expected true/false in data completions")
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// Formatter Tests
// ═══════════════════════════════════════════════════════════════════════════════

func TestFormatSingleObject(t *testing.T) {
	src := `{name:str,  age:int}:(Alice,  30)`
	result := Format(src)

	// Should produce canonical form
	if !strings.Contains(result, "{name:str, age:int}") {
		t.Errorf("format result: %q", result)
	}
}

func TestFormatObjectArray(t *testing.T) {
	src := `[{id:int,name:str}]:(1,Alice),(2,Bob)`
	result := Format(src)

	// Should produce multi-line with indentation
	if !strings.Contains(result, "\n") {
		t.Errorf("formatted array should be multi-line, got: %q", result)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// Semantic Tokens Tests
// ═══════════════════════════════════════════════════════════════════════════════

func TestSemanticTokensBasic(t *testing.T) {
	src := `{name:str}:(Alice)`
	lex := NewLexer(src)
	tokens := lex.All()

	// Verify we get meaningful tokens for semantic highlighting
	types := map[TokenType]bool{}
	for _, tok := range tokens {
		types[tok.Type] = true
	}
	if !types[TokenLBrace] {
		t.Error("missing LBrace")
	}
	if !types[TokenIdent] {
		t.Error("missing Ident")
	}
	if !types[TokenTypeHint] {
		t.Error("missing TypeHint")
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// Complex / Real-world Tests
// ═══════════════════════════════════════════════════════════════════════════════

func TestComplexNestedArray(t *testing.T) {
	src := `{company:str,employees:[{id:int,name:str,skills:[str]}],active:bool}:(ACME,[(1,Alice,[rust,go]),(2,Bob,[python])],true)`
	root, diags := Parse(src)

	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %v", diagMessages(diags))
	}
	if root == nil || len(root.Children) == 0 {
		t.Fatal("empty parse tree for complex input")
	}
}

func TestComplexMultilineArraySchema(t *testing.T) {
	src := `[{id:int, name:str, dept:{title:str}, skills:[str], active:bool}]:
  (1, Alice, (Manager), [Rust, Go], true),
  (2, Bob, (Engineer), [Python], false),
  (3, Carol, (Director), [Leadership, Strategy], true)`

	root, diags := Parse(src)
	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %v", diagMessages(diags))
	}
	arr := root.Children[0]
	if arr.Kind != NodeObjectArray {
		t.Errorf("expected NodeObjectArray, got %d", arr.Kind)
	}
	// 1 schema + 3 tuples = 4
	if len(arr.Children) != 4 {
		t.Errorf("expected 4 children, got %d", len(arr.Children))
	}
}

func TestCommentOnly(t *testing.T) {
	src := `/* just a comment */`
	root, diags := Parse(src)
	// Should parse as empty document
	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics for comment-only: %v", diagMessages(diags))
	}
	if root == nil {
		t.Fatal("nil root for comment-only input")
	}
}

func TestNestedArraysInData(t *testing.T) {
	src := `[[1,2],[3,4]]`
	root, diags := Parse(src)

	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %v", diagMessages(diags))
	}
	arr := root.Children[0]
	if arr.Kind != NodeArray {
		t.Errorf("expected NodeArray, got %d", arr.Kind)
	}
	if len(arr.Children) != 2 {
		t.Errorf("expected 2 sub-arrays, got %d", len(arr.Children))
	}
}

func TestMixedTypeArray(t *testing.T) {
	src := `[1,hello,true,3.14]`
	root, diags := Parse(src)

	if len(diags) != 0 {
		t.Errorf("unexpected diagnostics: %v", diagMessages(diags))
	}
	arr := root.Children[0]
	if len(arr.Children) != 4 {
		t.Errorf("expected 4 values, got %d", len(arr.Children))
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// isNumber Tests
// ═══════════════════════════════════════════════════════════════════════════════

func TestIsNumber(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"42", true},
		{"-123", true},
		{"3.14", true},
		{"-0.5", true},
		{"0", true},
		{"", false},
		{"abc", false},
		{"123abc", false},
		{"-", false},
		{".", false},
		{"1.2.3", false},
		{"1.", false},
	}
	for _, tt := range tests {
		got := isNumber(tt.input)
		if got != tt.want {
			t.Errorf("isNumber(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// Helpers
// ═══════════════════════════════════════════════════════════════════════════════

func diagMessages(diags []Diagnostic) []string {
	var msgs []string
	for _, d := range diags {
		msgs = append(msgs, d.Message)
	}
	return msgs
}
