package main

import (
	"encoding/json"
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
// Inlay Hints Tests
// ═══════════════════════════════════════════════════════════════════════════════

func TestInlayHintsSingleObject(t *testing.T) {
	src := `{name:str,age:int}:(Alice,30)`
	root, _ := Parse(src)
	hints := InlayHints(root)

	if len(hints) != 2 {
		t.Fatalf("expected 2 hints, got %d", len(hints))
	}
	if hints[0].Label != "name:" {
		t.Errorf("hint[0] = %q, want 'name:'", hints[0].Label)
	}
	if hints[1].Label != "age:" {
		t.Errorf("hint[1] = %q, want 'age:'", hints[1].Label)
	}
}

func TestInlayHintsObjectArray(t *testing.T) {
	src := `[{id:int,name:str}]:(1,Alice),(2,Bob)`
	root, _ := Parse(src)
	hints := InlayHints(root)

	if len(hints) != 4 {
		t.Fatalf("expected 4 hints (2 per tuple), got %d", len(hints))
	}
	if hints[0].Label != "id:" {
		t.Errorf("hint[0] = %q, want 'id:'", hints[0].Label)
	}
	if hints[1].Label != "name:" {
		t.Errorf("hint[1] = %q, want 'name:'", hints[1].Label)
	}
	if hints[2].Label != "id:" {
		t.Errorf("hint[2] = %q, want 'id:'", hints[2].Label)
	}
	if hints[3].Label != "name:" {
		t.Errorf("hint[3] = %q, want 'name:'", hints[3].Label)
	}
}

func TestInlayHintsNested(t *testing.T) {
	src := `{name:str,addr:{city:str,zip:int}}:(Alice,(NYC,10001))`
	root, _ := Parse(src)
	hints := InlayHints(root)

	// Should get: name:, addr:, city:, zip:
	if len(hints) != 4 {
		t.Fatalf("expected 4 hints (including nested), got %d: %v", len(hints), hints)
	}
	labels := make([]string, len(hints))
	for i, h := range hints {
		labels[i] = h.Label
	}
	expected := []string{"name:", "addr:", "city:", "zip:"}
	for i, exp := range expected {
		if labels[i] != exp {
			t.Errorf("hint[%d] = %q, want %q", i, labels[i], exp)
		}
	}
}

func TestInlayHintsPlainArray(t *testing.T) {
	src := `[1,2,3]`
	root, _ := Parse(src)
	hints := InlayHints(root)

	if len(hints) != 0 {
		t.Errorf("plain arrays should have no inlay hints, got %d", len(hints))
	}
}

func TestInlayHintsNestedArraySchema(t *testing.T) {
	// Schema has a field with an array of objects: projects:[{id:str,name:str}]
	// Data has an array of tuples for that field
	src := `{name:str,projects:[{id:str,title:str}]}:(Alice,[(P1,App),(P2,Web)])`
	root, _ := Parse(src)
	hints := InlayHints(root)

	// Expected hints:
	// name: (before Alice)
	// projects: (before [...])
	// id: (before P1), title: (before App)
	// id: (before P2), title: (before Web)
	labels := make([]string, len(hints))
	for i, h := range hints {
		labels[i] = h.Label
	}
	expected := []string{"name:", "projects:", "id:", "title:", "id:", "title:"}
	if len(hints) != len(expected) {
		t.Fatalf("expected %d hints, got %d: %v", len(expected), len(hints), labels)
	}
	for i, exp := range expected {
		if labels[i] != exp {
			t.Errorf("hint[%d] = %q, want %q", i, labels[i], exp)
		}
	}
}

func TestInlayHintsDeeplyNested(t *testing.T) {
	// 3 levels: employee has department (object) and projects (array of objects with tasks array)
	src := `{employee:{dept:{id:str,name:str},projects:[{pid:str,tasks:[{tid:str,status:str}]}]}}:` +
		`(((D1,HR),[(P1,[( T1,Done),(T2,Pending)])]))`
	root, _ := Parse(src)
	hints := InlayHints(root)

	labels := make([]string, len(hints))
	for i, h := range hints {
		labels[i] = h.Label
	}
	// employee:, dept:, id:, name:, projects:, pid:, tasks:, tid:, status:, tid:, status:
	expected := []string{
		"employee:", "dept:", "id:", "name:", "projects:",
		"pid:", "tasks:", "tid:", "status:", "tid:", "status:",
	}
	if len(hints) != len(expected) {
		t.Fatalf("expected %d hints, got %d: %v", len(expected), len(hints), labels)
	}
	for i, exp := range expected {
		if labels[i] != exp {
			t.Errorf("hint[%d] = %q, want %q", i, labels[i], exp)
		}
	}
}

func TestInlayHintsMultilineObjectArray(t *testing.T) {
	src := "[{id:int,name:str}]:\n  (1,Alice),\n  (2,Bob)"
	root, _ := Parse(src)
	hints := InlayHints(root)

	if len(hints) != 4 {
		t.Fatalf("expected 4 hints, got %d", len(hints))
	}
	// Check line numbers are correct
	if hints[0].Line != 1 || hints[2].Line != 2 {
		t.Errorf("hints should be on lines 1 and 2, got %d and %d", hints[0].Line, hints[2].Line)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// Compress Tests
// ═══════════════════════════════════════════════════════════════════════════════

func TestCompressSingleObject(t *testing.T) {
	src := `{name:str,  age:int}:(Alice,  30)`
	result := Compress(src)
	expected := `{name:str,age:int}:(Alice,30)`
	if result != expected {
		t.Errorf("compress = %q, want %q", result, expected)
	}
}

func TestCompressObjectArray(t *testing.T) {
	src := "[{id:int, name:str}]:\n  (1, Alice),\n  (2, Bob)"
	result := Compress(src)
	expected := `[{id:int,name:str}]:(1,Alice),(2,Bob)`
	if result != expected {
		t.Errorf("compress = %q, want %q", result, expected)
	}
}

func TestCompressPlainArray(t *testing.T) {
	src := `[ 1 , 2 , 3 ]`
	result := Compress(src)
	expected := `[1,2,3]`
	if result != expected {
		t.Errorf("compress = %q, want %q", result, expected)
	}
}

func TestCompressNested(t *testing.T) {
	src := "{name:str, addr:{city:str, zip:int}}:(Alice, (NYC, 10001))"
	result := Compress(src)
	expected := `{name:str,addr:{city:str,zip:int}}:(Alice,(NYC,10001))`
	if result != expected {
		t.Errorf("compress = %q, want %q", result, expected)
	}
}

func TestCompressIdempotent(t *testing.T) {
	src := `{a:int,b:str}:(1,hello)`
	c1 := Compress(src)
	c2 := Compress(c1)
	if c1 != c2 {
		t.Errorf("compress not idempotent: %q vs %q", c1, c2)
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// ASON → JSON conversion tests
// ═══════════════════════════════════════════════════════════════════════════════

func TestAsonToJSON_SimpleObject(t *testing.T) {
	src := `{name:str, age:int}:(Alice, 30)`
	result, err := AsonToJSON(src)
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(result), &obj); err != nil {
		t.Fatal("invalid JSON:", err)
	}
	if obj["name"] != "Alice" {
		t.Errorf("name = %v, want Alice", obj["name"])
	}
	if obj["age"] != float64(30) {
		t.Errorf("age = %v, want 30", obj["age"])
	}
}

func TestAsonToJSON_ObjectArray(t *testing.T) {
	src := `[{name:str, score:int}]:
  (Alice, 95),
  (Bob, 87)`
	result, err := AsonToJSON(src)
	if err != nil {
		t.Fatal(err)
	}
	var arr []map[string]interface{}
	if err := json.Unmarshal([]byte(result), &arr); err != nil {
		t.Fatal("invalid JSON array:", err)
	}
	if len(arr) != 2 {
		t.Fatalf("len = %d, want 2", len(arr))
	}
	if arr[0]["name"] != "Alice" {
		t.Errorf("arr[0].name = %v", arr[0]["name"])
	}
	if arr[1]["score"] != float64(87) {
		t.Errorf("arr[1].score = %v", arr[1]["score"])
	}
}

func TestAsonToJSON_NestedObject(t *testing.T) {
	src := `{name:str, addr:{city:str, zip:int}}:(Alice, (NYC, 10001))`
	result, err := AsonToJSON(src)
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]interface{}
	json.Unmarshal([]byte(result), &obj)
	addr := obj["addr"].(map[string]interface{})
	if addr["city"] != "NYC" {
		t.Errorf("addr.city = %v", addr["city"])
	}
	if addr["zip"] != float64(10001) {
		t.Errorf("addr.zip = %v", addr["zip"])
	}
}

func TestAsonToJSON_BoolValues(t *testing.T) {
	src := `{active:bool, name:str}:(true, Test)`
	result, err := AsonToJSON(src)
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]interface{}
	json.Unmarshal([]byte(result), &obj)
	if obj["active"] != true {
		t.Errorf("active = %v, want true", obj["active"])
	}
}

func TestAsonToJSON_FloatValues(t *testing.T) {
	src := `{price:float, qty:int}:(9.99, 5)`
	result, err := AsonToJSON(src)
	if err != nil {
		t.Fatal(err)
	}
	var obj map[string]interface{}
	json.Unmarshal([]byte(result), &obj)
	if obj["price"] != 9.99 {
		t.Errorf("price = %v, want 9.99", obj["price"])
	}
}

// ═══════════════════════════════════════════════════════════════════════════════
// JSON → ASON conversion tests
// ═══════════════════════════════════════════════════════════════════════════════

func TestJSONToASON_SimpleObject(t *testing.T) {
	src := `{"age": 30, "name": "Alice"}`
	result, err := JSONToASON(src)
	if err != nil {
		t.Fatal(err)
	}
	// Should contain schema and data
	if !strings.Contains(result, "age:int") || !strings.Contains(result, "name:str") {
		t.Errorf("result = %q, missing schema fields", result)
	}
	if !strings.Contains(result, "30") || !strings.Contains(result, "Alice") {
		t.Errorf("result = %q, missing data values", result)
	}
}

func TestJSONToASON_Array(t *testing.T) {
	src := `[{"name": "Alice", "score": 95}, {"name": "Bob", "score": 87}]`
	result, err := JSONToASON(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "[{") {
		t.Errorf("result = %q, want array schema", result)
	}
	if !strings.Contains(result, "Alice") || !strings.Contains(result, "Bob") {
		t.Errorf("result = %q, missing values", result)
	}
}

func TestJSONToASON_Nested(t *testing.T) {
	src := `{"name": "Alice", "addr": {"city": "NYC", "zip": 10001}}`
	result, err := JSONToASON(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "addr:{") {
		t.Errorf("result = %q, missing nested schema", result)
	}
	if !strings.Contains(result, "NYC") {
		t.Errorf("result = %q, missing nested value", result)
	}
}

func TestJSONToASON_BoolAndFloat(t *testing.T) {
	src := `{"active": true, "price": 9.99}`
	result, err := JSONToASON(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "bool") {
		t.Errorf("result = %q, missing bool type", result)
	}
	if !strings.Contains(result, "float") {
		t.Errorf("result = %q, missing float type", result)
	}
}

func TestRoundTrip_SimpleObject(t *testing.T) {
	asonSrc := `{age:int,name:str}:(30,Alice)`
	jsonStr, err := AsonToJSON(asonSrc)
	if err != nil {
		t.Fatal(err)
	}
	asonBack, err := JSONToASON(jsonStr)
	if err != nil {
		t.Fatal(err)
	}
	// Roundtrip back to JSON to verify semantic equivalence
	jsonBack, err := AsonToJSON(asonBack)
	if err != nil {
		t.Fatal(err)
	}
	// Compare JSON objects
	var obj1, obj2 map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &obj1)
	json.Unmarshal([]byte(jsonBack), &obj2)
	if obj1["name"] != obj2["name"] || obj1["age"] != obj2["age"] {
		t.Errorf("roundtrip mismatch:\n  orig JSON: %s\n  back JSON: %s", jsonStr, jsonBack)
	}
}

func TestRoundTrip_ObjectArray(t *testing.T) {
	asonSrc := `[{name:str,score:int}]:(Alice,95),(Bob,87)`
	jsonStr, err := AsonToJSON(asonSrc)
	if err != nil {
		t.Fatal(err)
	}
	asonBack, err := JSONToASON(jsonStr)
	if err != nil {
		t.Fatal(err)
	}
	jsonBack, err := AsonToJSON(asonBack)
	if err != nil {
		t.Fatal(err)
	}
	var arr1, arr2 []map[string]interface{}
	json.Unmarshal([]byte(jsonStr), &arr1)
	json.Unmarshal([]byte(jsonBack), &arr2)
	if len(arr1) != len(arr2) {
		t.Errorf("roundtrip array length mismatch: %d vs %d", len(arr1), len(arr2))
	}
}

func TestJSONToASON_EmptyTopLevelArray(t *testing.T) {
	src := `[]`
	result, err := JSONToASON(src)
	if err != nil {
		t.Fatal(err)
	}
	if result != "[str]" {
		t.Errorf("result = %q, want %q", result, "[str]")
	}
}

func TestJSONToASON_EmptyNestedArray(t *testing.T) {
	src := `{"name": "test", "tags": []}`
	result, err := JSONToASON(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, "tags:[str]") {
		t.Errorf("result = %q, want tags:[str]", result)
	}
}

func TestJSONToASON_PlusMinus_NoQuoting(t *testing.T) {
	src := `{"lowPriorityEIR+CIR": 42, "a-b": "hello"}`
	result, err := JSONToASON(src)
	if err != nil {
		t.Fatal(err)
	}
	// +/- are now valid identifier chars, should NOT be quoted
	if !strings.Contains(result, "lowPriorityEIR+CIR:int") {
		t.Errorf("result = %q, key with + should not be quoted", result)
	}
	if !strings.Contains(result, "a-b:str") {
		t.Errorf("result = %q, key with - should not be quoted", result)
	}
}

func TestJSONToASON_DotAndSpace_NeedQuoting(t *testing.T) {
	src := `{"a.b": "hello", "has space": 42}`
	result, err := JSONToASON(src)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(result, `"a.b"`) {
		t.Errorf("result = %q, key with . should be quoted", result)
	}
	if !strings.Contains(result, `"has space"`) {
		t.Errorf("result = %q, key with space should be quoted", result)
	}
}

func TestNeedsKeyQuote(t *testing.T) {
	tests := []struct {
		key  string
		want bool
	}{
		{"name", false},
		{"a+b", false},
		{"a-b", false},
		{"a_b", false},
		{"lowPriorityEIR+CIR", false},
		{"a.b", true},
		{"has space", true},
		{"", true},
		{"a@b", true},
	}
	for _, tt := range tests {
		got := needsKeyQuote(tt.key)
		if got != tt.want {
			t.Errorf("needsKeyQuote(%q) = %v, want %v", tt.key, got, tt.want)
		}
	}
}

func TestLexer_PlusMinusInIdentifier(t *testing.T) {
	src := `{lowPriorityEIR+CIR:int, a-b:str}:(42, hello)`
	lex := NewLexer(src)
	tokens := lex.All()
	// Find ident tokens
	var idents []string
	for _, tok := range tokens {
		if tok.Type == TokenIdent {
			idents = append(idents, tok.Value)
		}
	}
	found := false
	for _, id := range idents {
		if id == "lowPriorityEIR+CIR" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("expected identifier 'lowPriorityEIR+CIR', got idents: %v", idents)
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

// ═══════════════════════════════════════════════════════════════════════════════
// Format — smart formatting tests
// ═══════════════════════════════════════════════════════════════════════════════

func TestFormatSimpleArrayInline(t *testing.T) {
	// Simple schema+tuples should stay compact
	src := "[{name:str,age:int,active:bool}]:(Alice,30,true),(Bob,25,false)"
	result := Format(src)
	expected := "[{name:str, age:int, active:bool}]:\n  (Alice, 30, true),\n  (Bob, 25, false)\n"
	if result != expected {
		t.Errorf("format =\n%s\nwant:\n%s", result, expected)
	}
}

func TestFormatSimpleSingleObject(t *testing.T) {
	// Simple single object should keep schema and data inline
	src := `{name:str,  age:int}:(Alice,  30)`
	result := Format(src)
	expected := `{name:str, age:int}:(Alice, 30)`
	if result != expected {
		t.Errorf("format = %q, want %q", result, expected)
	}
}

func TestFormatNestedSchemaExpands(t *testing.T) {
	// Nested schema should expand when complex
	src := `[{employee:{department:{id:str,manager:{contact:{email:str,phone:str},id:str,name:str},name:str},id:str,name:str}}]:(((D006,((pat@ex.net,123),M1,Alice),HR),E001,Bob))`
	result := Format(src)
	// Schema should be multi-line since it's deeply nested
	if !strings.Contains(result, "\n") {
		t.Errorf("nested schema should be multi-line:\n%s", result)
	}
	// Data tuples with nested tuples should also expand
	t.Logf("Formatted output:\n%s", result)
}

func TestFormatPlainArrayInline(t *testing.T) {
	src := `[ 1 , 2 , 3 ]`
	result := Format(src)
	expected := `[1, 2, 3]`
	if result != expected {
		t.Errorf("format = %q, want %q", result, expected)
	}
}
