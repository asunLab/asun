package main

import "fmt"

// ──────────────────────────────────────────────────────────────────────────────
// AST Node Types
// ──────────────────────────────────────────────────────────────────────────────

// NodeKind labels AST nodes.
type NodeKind int

const (
	NodeDocument     NodeKind = iota // top-level
	NodeSchema                       // {field1, field2, ...}
	NodeField                        // field[:type]
	NodeTypeAnnot                    // :int, :str, ...
	NodeArraySchema                  // [{schema}]
	NodeMapType                      // map[K,V]
	NodeSingleObject                 // {schema}:(...)
	NodeObjectArray                  // [{schema}]:(...),...
	NodeTuple                        // (val, val, ...)
	NodeArray                        // [val, val, ...]
	NodeValue                        // literal value
)

// Node is a position-aware AST node.
type Node struct {
	Kind     NodeKind
	Token    Token   // primary token (open-brace, paren, etc.)
	EndToken Token   // closing token
	Value    string  // for leaf nodes
	Children []*Node // sub-nodes
}

func (n *Node) StartPos() Position { return n.Token.Pos() }
func (n *Node) EndPos() Position   { return n.EndToken.End() }

// SchemaFields returns the Field children of a Schema node.
func (n *Node) SchemaFields() []*Node {
	if n == nil || (n.Kind != NodeSchema && n.Kind != NodeArraySchema) {
		return nil
	}
	if n.Kind == NodeArraySchema && len(n.Children) > 0 {
		return n.Children[0].SchemaFields()
	}
	var fields []*Node
	for _, c := range n.Children {
		if c.Kind == NodeField {
			fields = append(fields, c)
		}
	}
	return fields
}

// ──────────────────────────────────────────────────────────────────────────────
// Diagnostic
// ──────────────────────────────────────────────────────────────────────────────

// DiagSeverity matches LSP severity.
type DiagSeverity int

const (
	SeverityError   DiagSeverity = 1
	SeverityWarning DiagSeverity = 2
	SeverityInfo    DiagSeverity = 3
	SeverityHint    DiagSeverity = 4
)

// Diagnostic is a parser/analysis error or warning.
type Diagnostic struct {
	StartLine int
	StartCol  int
	EndLine   int
	EndCol    int
	Message   string
	Severity  DiagSeverity
}

// ──────────────────────────────────────────────────────────────────────────────
// Parser
// ──────────────────────────────────────────────────────────────────────────────

// Parser produces an AST and diagnostics from tokens.
type Parser struct {
	tokens []Token
	pos    int
	diags  []Diagnostic
	src    string
}

// Parse tokenizes and parses the given ASON source.
func Parse(src string) (*Node, []Diagnostic) {
	lex := NewLexer(src)
	tokens := lex.All()
	p := &Parser{tokens: tokens, src: src}
	doc := p.parseDocument()
	return doc, p.diags
}

func (p *Parser) peek() Token {
	for p.pos < len(p.tokens) {
		t := p.tokens[p.pos]
		if t.Type == TokenNewline || t.Type == TokenComment {
			p.pos++
			continue
		}
		return t
	}
	return Token{Type: TokenEOF}
}

func (p *Parser) next() Token {
	for p.pos < len(p.tokens) {
		t := p.tokens[p.pos]
		p.pos++
		if t.Type == TokenNewline || t.Type == TokenComment {
			continue
		}
		return t
	}
	return Token{Type: TokenEOF}
}

func (p *Parser) expect(typ TokenType) Token {
	t := p.next()
	if t.Type != typ {
		p.addDiag(t, fmt.Sprintf("expected %s, got %s", typ, t.Type))
	}
	return t
}

func (p *Parser) addDiag(t Token, msg string) {
	p.diags = append(p.diags, Diagnostic{
		StartLine: t.Line,
		StartCol:  t.Col,
		EndLine:   t.EndLine,
		EndCol:    t.EndCol,
		Message:   msg,
		Severity:  SeverityError,
	})
}

func (p *Parser) addDiagRange(startLine, startCol, endLine, endCol int, msg string, sev DiagSeverity) {
	p.diags = append(p.diags, Diagnostic{
		StartLine: startLine,
		StartCol:  startCol,
		EndLine:   endLine,
		EndCol:    endCol,
		Message:   msg,
		Severity:  sev,
	})
}

// ──────────────────────────────────────────────────────────────────────────────
// Parse rules
// ──────────────────────────────────────────────────────────────────────────────

func (p *Parser) parseDocument() *Node {
	doc := &Node{Kind: NodeDocument}
	t := p.peek()

	switch t.Type {
	case TokenLBrace:
		// Single object: {schema}:(data)
		doc.Children = append(doc.Children, p.parseSingleObject())
	case TokenLBracket:
		// Could be [{schema}]:(...) or plain array [...]
		if p.looksLikeArraySchema() {
			doc.Children = append(doc.Children, p.parseObjectArray())
		} else {
			doc.Children = append(doc.Children, p.parseArray())
		}
	case TokenEOF:
		// empty doc
	default:
		// Plain value / error
		doc.Children = append(doc.Children, p.parseValue())
	}

	// Check for trailing content
	if end := p.peek(); end.Type != TokenEOF {
		p.addDiag(end, "unexpected content after top-level expression")
	}
	return doc
}

// looksLikeArraySchema peeks ahead to see if we have [{...}]
func (p *Parser) looksLikeArraySchema() bool {
	saved := p.pos
	defer func() { p.pos = saved }()
	t := p.next() // [
	if t.Type != TokenLBracket {
		return false
	}
	t2 := p.peek()
	return t2.Type == TokenLBrace
}

// parseSingleObject: {schema}:(data)
func (p *Parser) parseSingleObject() *Node {
	node := &Node{Kind: NodeSingleObject}
	schema := p.parseSchema()
	node.Children = append(node.Children, schema)
	node.Token = schema.Token

	// Expect ':'
	t := p.peek()
	if t.Type == TokenColon {
		p.next()
	} else {
		p.addDiag(t, "expected ':' after schema")
	}

	// Expect '(' data tuple
	t = p.peek()
	if t.Type == TokenLParen {
		tuple := p.parseTuple()
		node.Children = append(node.Children, tuple)
		node.EndToken = tuple.EndToken

		// Validate field count
		fields := schema.SchemaFields()
		dataCount := len(tuple.Children)
		if len(fields) > 0 && dataCount != len(fields) {
			p.addDiagRange(
				tuple.Token.Line, tuple.Token.Col,
				tuple.EndToken.EndLine, tuple.EndToken.EndCol,
				fmt.Sprintf("field count mismatch: schema has %d fields, data has %d values", len(fields), dataCount),
				SeverityError,
			)
		}
	} else {
		p.addDiag(t, "expected '(' for data after schema")
	}
	return node
}

// parseObjectArray: [{schema}]:(data),(data),...
func (p *Parser) parseObjectArray() *Node {
	node := &Node{Kind: NodeObjectArray}
	openBracket := p.next() // [
	node.Token = openBracket

	schema := p.parseSchema()
	arrSchema := &Node{
		Kind:     NodeArraySchema,
		Token:    openBracket,
		Children: []*Node{schema},
	}

	closeBracket := p.expect(TokenRBracket)
	arrSchema.EndToken = closeBracket
	node.Children = append(node.Children, arrSchema)

	// Expect ':'
	t := p.peek()
	if t.Type == TokenColon {
		p.next()
	} else {
		p.addDiag(t, "expected ':' after [schema]")
	}

	// Parse tuple list
	fields := schema.SchemaFields()
	for {
		t = p.peek()
		if t.Type == TokenLParen {
			tuple := p.parseTuple()
			node.Children = append(node.Children, tuple)
			node.EndToken = tuple.EndToken

			// Validate field count per tuple
			dataCount := len(tuple.Children)
			if len(fields) > 0 && dataCount != len(fields) {
				p.addDiagRange(
					tuple.Token.Line, tuple.Token.Col,
					tuple.EndToken.EndLine, tuple.EndToken.EndCol,
					fmt.Sprintf("field count mismatch: schema has %d fields, tuple has %d values", len(fields), dataCount),
					SeverityError,
				)
			}

			// Consume trailing comma
			if p.peek().Type == TokenComma {
				p.next()
			} else {
				break
			}
		} else {
			break
		}
	}
	return node
}

// parseSchema: {field1[:type], field2[:type], ...}
func (p *Parser) parseSchema() *Node {
	openBrace := p.expect(TokenLBrace)
	node := &Node{Kind: NodeSchema, Token: openBrace}

	for {
		t := p.peek()
		if t.Type == TokenRBrace || t.Type == TokenEOF {
			break
		}
		field := p.parseField()
		node.Children = append(node.Children, field)

		if p.peek().Type == TokenComma {
			p.next()
		}
	}

	closeBrace := p.expect(TokenRBrace)
	node.EndToken = closeBrace
	return node
}

// parseField: ident[:typeExpr] or ident:{nestedSchema} or ident:[typeExpr]
func (p *Parser) parseField() *Node {
	nameToken := p.next()
	if nameToken.Type != TokenIdent && nameToken.Type != TokenTypeHint {
		p.addDiag(nameToken, fmt.Sprintf("expected field name, got %s", nameToken.Type))
	}
	field := &Node{Kind: NodeField, Token: nameToken, Value: nameToken.Value, EndToken: nameToken}

	t := p.peek()
	if t.Type == TokenColon {
		p.next() // consume ':'
		nx := p.peek()
		switch nx.Type {
		case TokenLBrace:
			// Nested schema: field:{subschema}
			sub := p.parseSchema()
			field.Children = append(field.Children, sub)
			field.EndToken = sub.EndToken
		case TokenLBracket:
			// Array type: field:[type] or field:[{schema}]
			arr := p.parseFieldArrayType()
			field.Children = append(field.Children, arr)
			field.EndToken = arr.EndToken
		case TokenMapKw:
			mapNode := p.parseMapType()
			field.Children = append(field.Children, mapNode)
			field.EndToken = mapNode.EndToken
		case TokenTypeHint:
			hint := p.next()
			annot := &Node{Kind: NodeTypeAnnot, Token: hint, Value: hint.Value, EndToken: hint}
			field.Children = append(field.Children, annot)
			field.EndToken = hint
		default:
			p.addDiag(nx, "expected type annotation or nested schema after ':'")
		}
	}
	return field
}

// parseFieldArrayType: [type] or [{schema}]
func (p *Parser) parseFieldArrayType() *Node {
	openBracket := p.next() // [
	t := p.peek()

	if t.Type == TokenLBrace {
		// [{nested schema}]
		sub := p.parseSchema()
		closeBracket := p.expect(TokenRBracket)
		return &Node{
			Kind:     NodeArraySchema,
			Token:    openBracket,
			EndToken: closeBracket,
			Children: []*Node{sub},
		}
	}

	// [typeHint]
	var inner *Node
	if t.Type == TokenTypeHint || t.Type == TokenIdent {
		hint := p.next()
		inner = &Node{Kind: NodeTypeAnnot, Token: hint, Value: hint.Value, EndToken: hint}
	}
	closeBracket := p.expect(TokenRBracket)
	node := &Node{Kind: NodeTypeAnnot, Token: openBracket, EndToken: closeBracket, Value: "array"}
	if inner != nil {
		node.Children = append(node.Children, inner)
	}
	return node
}

// parseMapType: map[K,V]
func (p *Parser) parseMapType() *Node {
	kw := p.next() // map
	node := &Node{Kind: NodeMapType, Token: kw}
	p.expect(TokenLBracket)
	// key type
	if p.peek().Type == TokenTypeHint || p.peek().Type == TokenIdent {
		p.next()
	}
	if p.peek().Type == TokenComma {
		p.next()
	}
	// value type — could be typeHint or {schema}
	if p.peek().Type == TokenTypeHint || p.peek().Type == TokenIdent {
		p.next()
	} else if p.peek().Type == TokenLBrace {
		p.parseSchema()
	}
	closeBracket := p.expect(TokenRBracket)
	node.EndToken = closeBracket
	return node
}

// parseTuple: (val, val, ...)
func (p *Parser) parseTuple() *Node {
	openParen := p.next() // (
	node := &Node{Kind: NodeTuple, Token: openParen}

	first := true
	for {
		t := p.peek()
		if t.Type == TokenRParen || t.Type == TokenEOF {
			break
		}
		if !first {
			if p.peek().Type == TokenComma {
				p.next()
				// Check for trailing comma before ')'
				if p.peek().Type == TokenRParen {
					// Trailing comma → null value
					node.Children = append(node.Children, &Node{
						Kind:     NodeValue,
						Token:    p.tokens[p.pos-1], // the comma
						EndToken: p.tokens[p.pos-1],
						Value:    "",
					})
					break
				}
				// Check for consecutive commas → null
				if p.peek().Type == TokenComma {
					// null value
					node.Children = append(node.Children, &Node{
						Kind:  NodeValue,
						Token: p.tokens[p.pos],
						Value: "",
					})
					continue
				}
			} else if p.peek().Type != TokenRParen {
				p.addDiag(p.peek(), "expected ',' or ')' in tuple")
				break
			}
		}
		first = false

		nx := p.peek()
		switch nx.Type {
		case TokenLParen:
			// Nested tuple
			node.Children = append(node.Children, p.parseTuple())
		case TokenLBracket:
			// Array value
			node.Children = append(node.Children, p.parseArray())
		case TokenRParen:
			// empty value (null) — only when double comma was read
			continue
		case TokenComma:
			// null value (empty between commas)
			node.Children = append(node.Children, &Node{
				Kind:  NodeValue,
				Token: nx,
				Value: "",
			})
		default:
			node.Children = append(node.Children, p.parseValue())
		}
	}

	closeParen := p.expect(TokenRParen)
	node.EndToken = closeParen
	return node
}

// parseArray: [val, val, ...]
func (p *Parser) parseArray() *Node {
	openBracket := p.next() // [
	node := &Node{Kind: NodeArray, Token: openBracket}

	first := true
	for {
		t := p.peek()
		if t.Type == TokenRBracket || t.Type == TokenEOF {
			break
		}
		if !first {
			if p.peek().Type == TokenComma {
				p.next()
				// trailing comma
				if p.peek().Type == TokenRBracket {
					break
				}
			} else {
				break
			}
		}
		first = false

		nx := p.peek()
		switch nx.Type {
		case TokenLParen:
			node.Children = append(node.Children, p.parseTuple())
		case TokenLBracket:
			node.Children = append(node.Children, p.parseArray())
		default:
			node.Children = append(node.Children, p.parseValue())
		}
	}

	closeBracket := p.expect(TokenRBracket)
	node.EndToken = closeBracket
	return node
}

// parseValue: literal token → Value node
func (p *Parser) parseValue() *Node {
	t := p.next()
	return &Node{Kind: NodeValue, Token: t, EndToken: t, Value: t.Value}
}
