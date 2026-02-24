package main

import (
	"strings"
	"unicode/utf8"
)

// Lexer tokenises an ASON source string.
type Lexer struct {
	src  string
	pos  int // current byte offset
	line int
	col  int

	// Mode switches between schema and data context.
	// In schema mode we emit Ident/TypeHint tokens.
	// In data mode we emit value tokens (Number/Bool/PlainStr).
	InSchema bool
}

// NewLexer creates a lexer for src.
func NewLexer(src string) *Lexer {
	return &Lexer{src: src}
}

// peek returns the current byte without advancing.
func (l *Lexer) peek() byte {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

// advance moves forward n bytes, updating line/col.
func (l *Lexer) advance(n int) {
	for i := 0; i < n && l.pos < len(l.src); i++ {
		if l.src[l.pos] == '\n' {
			l.line++
			l.col = 0
		} else {
			l.col++
		}
		l.pos++
	}
}

// skipWhitespaceNoNewline skips spaces and tabs (not newlines).
func (l *Lexer) skipWhitespaceNoNewline() {
	for l.pos < len(l.src) {
		b := l.src[l.pos]
		if b == ' ' || b == '\t' || b == '\r' {
			l.advance(1)
		} else {
			break
		}
	}
}

// All returns all tokens from the source.
func (l *Lexer) All() []Token {
	var tokens []Token
	for {
		tok := l.Next()
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF || tok.Type == TokenError {
			break
		}
	}
	return tokens
}

// Next returns the next token.
func (l *Lexer) Next() Token {
	l.skipWhitespaceNoNewline()
	if l.pos >= len(l.src) {
		return l.makeToken(TokenEOF, l.pos, l.pos)
	}

	start := l.pos
	sLine := l.line
	sCol := l.col
	b := l.src[l.pos]

	// Newlines
	if b == '\n' {
		l.advance(1)
		return l.tok(TokenNewline, start, l.pos, sLine, sCol)
	}

	// Comment /* ... */
	if b == '/' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '*' {
		return l.lexComment(start, sLine, sCol)
	}

	// Structural tokens
	switch b {
	case '{':
		l.advance(1)
		l.InSchema = true
		return l.tok(TokenLBrace, start, l.pos, sLine, sCol)
	case '}':
		l.advance(1)
		return l.tok(TokenRBrace, start, l.pos, sLine, sCol)
	case '(':
		l.advance(1)
		l.InSchema = false
		return l.tok(TokenLParen, start, l.pos, sLine, sCol)
	case ')':
		l.advance(1)
		return l.tok(TokenRParen, start, l.pos, sLine, sCol)
	case '[':
		l.advance(1)
		return l.tok(TokenLBracket, start, l.pos, sLine, sCol)
	case ']':
		l.advance(1)
		return l.tok(TokenRBracket, start, l.pos, sLine, sCol)
	case ':':
		l.advance(1)
		return l.tok(TokenColon, start, l.pos, sLine, sCol)
	case ',':
		l.advance(1)
		return l.tok(TokenComma, start, l.pos, sLine, sCol)
	}

	// Quoted string
	if b == '"' {
		return l.lexQuotedString(start, sLine, sCol)
	}

	// In schema mode: identifiers and type keywords
	if l.InSchema {
		return l.lexSchemaWord(start, sLine, sCol)
	}

	// Data mode: numbers, bools, plain strings
	return l.lexDataValue(start, sLine, sCol)
}

// lexComment consumes a /* ... */ comment.
func (l *Lexer) lexComment(start, sLine, sCol int) Token {
	l.advance(2) // skip /*
	end := strings.Index(l.src[l.pos:], "*/")
	if end < 0 {
		// Unclosed comment — consume to end of file
		endPos := len(l.src)
		for l.pos < endPos {
			l.advance(1)
		}
		return l.tok(TokenError, start, l.pos, sLine, sCol)
	}
	l.advance(end + 2) // skip past */
	return l.tok(TokenComment, start, l.pos, sLine, sCol)
}

// lexQuotedString consumes "..." with escape handling.
func (l *Lexer) lexQuotedString(start, sLine, sCol int) Token {
	l.advance(1) // skip opening "
	for l.pos < len(l.src) {
		b := l.src[l.pos]
		if b == '\\' {
			l.advance(2) // skip escape pair
			continue
		}
		if b == '"' {
			l.advance(1) // skip closing "
			return l.tok(TokenString, start, l.pos, sLine, sCol)
		}
		if b < 0x80 {
			l.advance(1)
		} else {
			_, sz := utf8.DecodeRuneInString(l.src[l.pos:])
			l.advance(sz)
		}
	}
	// Unclosed string
	return l.tok(TokenError, start, l.pos, sLine, sCol)
}

// isIdentChar returns whether b is valid in an identifier/field name.
func isIdentChar(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') ||
		(b >= '0' && b <= '9') || b == '_'
}

// lexSchemaWord consumes an identifier or type keyword in schema context.
func (l *Lexer) lexSchemaWord(start, sLine, sCol int) Token {
	for l.pos < len(l.src) && isIdentChar(l.src[l.pos]) {
		l.advance(1)
	}
	if l.pos == start {
		// Unknown character
		l.advance(1)
		return l.tok(TokenError, start, l.pos, sLine, sCol)
	}
	word := l.src[start:l.pos]
	switch word {
	case "int", "integer", "float", "double", "str", "string", "bool", "boolean":
		return l.tok(TokenTypeHint, start, l.pos, sLine, sCol)
	case "map":
		return l.tok(TokenMapKw, start, l.pos, sLine, sCol)
	default:
		return l.tok(TokenIdent, start, l.pos, sLine, sCol)
	}
}

// lexDataValue consumes a value in data context:
// numbers, booleans, or plain strings.
func (l *Lexer) lexDataValue(start, sLine, sCol int) Token {
	// Consume until delimiter
	for l.pos < len(l.src) {
		b := l.src[l.pos]
		if b == ',' || b == ')' || b == ']' || b == '(' || b == '[' ||
			b == '\n' || b == '\r' {
			break
		}
		// Handle escapes
		if b == '\\' && l.pos+1 < len(l.src) {
			l.advance(2)
			continue
		}
		// Comment start breaks value
		if b == '/' && l.pos+1 < len(l.src) && l.src[l.pos+1] == '*' {
			break
		}
		if b < 0x80 {
			l.advance(1)
		} else {
			_, sz := utf8.DecodeRuneInString(l.src[l.pos:])
			l.advance(sz)
		}
	}

	raw := l.src[start:l.pos]
	trimmed := strings.TrimRight(raw, " \t")
	if trimmed == "" {
		// Will be handled as null value by the parser
		return l.tok(TokenPlainStr, start, l.pos, sLine, sCol)
	}
	// Classify
	if trimmed == "true" || trimmed == "false" {
		return l.tok(TokenBool, start, start+len(trimmed), sLine, sCol)
	}
	if isNumber(trimmed) {
		return l.tok(TokenNumber, start, start+len(trimmed), sLine, sCol)
	}
	return l.tok(TokenPlainStr, start, l.pos, sLine, sCol)
}

// isNumber checks if s matches -?[0-9]+(\.[0-9]+)?
func isNumber(s string) bool {
	i := 0
	if i < len(s) && s[i] == '-' {
		i++
	}
	if i >= len(s) || s[i] < '0' || s[i] > '9' {
		return false
	}
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		i++
	}
	if i < len(s) && s[i] == '.' {
		i++
		if i >= len(s) || s[i] < '0' || s[i] > '9' {
			return false
		}
		for i < len(s) && s[i] >= '0' && s[i] <= '9' {
			i++
		}
	}
	return i == len(s)
}

func (l *Lexer) makeToken(typ TokenType, start, end int) Token {
	return Token{
		Type:    typ,
		Value:   l.src[start:end],
		Offset:  start,
		Line:    l.line,
		Col:     l.col,
		EndOff:  end,
		EndLine: l.line,
		EndCol:  l.col,
	}
}

func (l *Lexer) tok(typ TokenType, start, end, sLine, sCol int) Token {
	return Token{
		Type:    typ,
		Value:   l.src[start:end],
		Offset:  start,
		Line:    sLine,
		Col:     sCol,
		EndOff:  end,
		EndLine: l.line,
		EndCol:  l.col,
	}
}
