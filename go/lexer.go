package ason

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// TokenType represents the type of a lexer token.
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenError
	TokenLParen   // (
	TokenRParen   // )
	TokenLBracket // [
	TokenRBracket // ]
	TokenLBrace   // {
	TokenRBrace   // }
	TokenComma    // ,
	TokenColon    // :
	TokenNull     // null
	TokenTrue     // true
	TokenFalse    // false
	TokenInteger  // 123
	TokenFloat    // 3.14
	TokenString   // "hello" or hello
	TokenIdent    // identifier
)

// Token represents a lexer token.
type Token struct {
	Type   TokenType
	Value  string
	Line   int
	Column int
}

// Lexer tokenizes ASON input.
type Lexer struct {
	input  string
	pos    int
	line   int
	column int
}

// NewLexer creates a new lexer for the given input.
func NewLexer(input string) *Lexer {
	return &Lexer{
		input:  input,
		pos:    0,
		line:   1,
		column: 1,
	}
}

func (l *Lexer) peek() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[l.pos:])
	return r
}

func (l *Lexer) peekN(n int) rune {
	pos := l.pos
	for i := 0; i < n; i++ {
		if pos >= len(l.input) {
			return 0
		}
		_, size := utf8.DecodeRuneInString(l.input[pos:])
		pos += size
	}
	if pos >= len(l.input) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.input[pos:])
	return r
}

func (l *Lexer) advance() rune {
	if l.pos >= len(l.input) {
		return 0
	}
	r, size := utf8.DecodeRuneInString(l.input[l.pos:])
	l.pos += size
	if r == '\n' {
		l.line++
		l.column = 1
	} else {
		l.column++
	}
	return r
}

func (l *Lexer) skipWhitespace() {
	for {
		r := l.peek()
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			l.advance()
		} else if r == '/' && l.peekN(1) == '*' {
			l.skipBlockComment()
		} else {
			break
		}
	}
}

func (l *Lexer) skipBlockComment() {
	l.advance() // /
	l.advance() // *
	for {
		r := l.advance()
		if r == 0 {
			break
		}
		if r == '*' && l.peek() == '/' {
			l.advance()
			break
		}
	}
}

func (l *Lexer) makeToken(typ TokenType, value string, line, col int) Token {
	return Token{Type: typ, Value: value, Line: line, Column: col}
}

func (l *Lexer) errorToken(msg string) Token {
	return Token{Type: TokenError, Value: msg, Line: l.line, Column: l.column}
}

// NextToken returns the next token from the input.
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	line, col := l.line, l.column
	r := l.peek()

	if r == 0 {
		return l.makeToken(TokenEOF, "", line, col)
	}

	// Single character tokens
	switch r {
	case '(':
		l.advance()
		return l.makeToken(TokenLParen, "(", line, col)
	case ')':
		l.advance()
		return l.makeToken(TokenRParen, ")", line, col)
	case '[':
		l.advance()
		return l.makeToken(TokenLBracket, "[", line, col)
	case ']':
		l.advance()
		return l.makeToken(TokenRBracket, "]", line, col)
	case '{':
		l.advance()
		return l.makeToken(TokenLBrace, "{", line, col)
	case '}':
		l.advance()
		return l.makeToken(TokenRBrace, "}", line, col)
	case ',':
		l.advance()
		return l.makeToken(TokenComma, ",", line, col)
	case ':':
		l.advance()
		return l.makeToken(TokenColon, ":", line, col)
	case '"':
		return l.scanQuotedString()
	}

	// Numbers
	if r == '-' || r == '+' || (r >= '0' && r <= '9') {
		return l.scanNumber()
	}

	// Identifiers and keywords
	if l.isIdentStart(r) {
		return l.scanIdentOrKeyword()
	}

	// Unquoted string (starts with other characters)
	return l.scanUnquotedString()
}

func (l *Lexer) isIdentStart(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_'
}

func (l *Lexer) isIdentChar(r rune) bool {
	return l.isIdentStart(r) || (r >= '0' && r <= '9')
}

func (l *Lexer) scanQuotedString() Token {
	line, col := l.line, l.column
	l.advance() // consume opening quote

	var sb strings.Builder
	for {
		r := l.peek()
		if r == 0 {
			return l.errorToken("unterminated string")
		}
		if r == '"' {
			l.advance()
			break
		}
		if r == '\\' {
			l.advance()
			escaped := l.advance()
			switch escaped {
			case 'n':
				sb.WriteRune('\n')
			case 't':
				sb.WriteRune('\t')
			case 'r':
				sb.WriteRune('\r')
			case '\\':
				sb.WriteRune('\\')
			case '"':
				sb.WriteRune('"')
			case 'u':
				// Unicode escape \uXXXX
				hex := ""
				for i := 0; i < 4; i++ {
					hex += string(l.advance())
				}
				var code rune
				fmt.Sscanf(hex, "%x", &code)
				sb.WriteRune(code)
			default:
				sb.WriteRune(escaped)
			}
		} else {
			sb.WriteRune(l.advance())
		}
	}
	return l.makeToken(TokenString, sb.String(), line, col)
}

func (l *Lexer) scanNumber() Token {
	line, col := l.line, l.column
	start := l.pos
	isFloat := false

	// Sign
	if l.peek() == '-' || l.peek() == '+' {
		l.advance()
	}

	// Integer part
	for l.peek() >= '0' && l.peek() <= '9' {
		l.advance()
	}

	// Decimal part
	if l.peek() == '.' && l.peekN(1) >= '0' && l.peekN(1) <= '9' {
		isFloat = true
		l.advance() // .
		for l.peek() >= '0' && l.peek() <= '9' {
			l.advance()
		}
	}

	// Exponent
	if l.peek() == 'e' || l.peek() == 'E' {
		isFloat = true
		l.advance()
		if l.peek() == '+' || l.peek() == '-' {
			l.advance()
		}
		for l.peek() >= '0' && l.peek() <= '9' {
			l.advance()
		}
	}

	value := l.input[start:l.pos]
	if isFloat {
		return l.makeToken(TokenFloat, value, line, col)
	}
	return l.makeToken(TokenInteger, value, line, col)
}

func (l *Lexer) scanIdentOrKeyword() Token {
	line, col := l.line, l.column
	start := l.pos

	for l.isIdentChar(l.peek()) {
		l.advance()
	}

	value := l.input[start:l.pos]
	switch value {
	case "null":
		return l.makeToken(TokenNull, value, line, col)
	case "true":
		return l.makeToken(TokenTrue, value, line, col)
	case "false":
		return l.makeToken(TokenFalse, value, line, col)
	default:
		return l.makeToken(TokenIdent, value, line, col)
	}
}

func (l *Lexer) scanUnquotedString() Token {
	line, col := l.line, l.column
	var sb strings.Builder

	for {
		r := l.peek()
		if r == 0 || r == '(' || r == ')' || r == '[' || r == ']' ||
			r == '{' || r == '}' || r == ',' || r == ':' ||
			unicode.IsSpace(r) {
			break
		}
		sb.WriteRune(l.advance())
	}

	return l.makeToken(TokenString, sb.String(), line, col)
}

// PeekToken returns the next token without consuming it.
func (l *Lexer) PeekToken() Token {
	pos, line, col := l.pos, l.line, l.column
	tok := l.NextToken()
	l.pos, l.line, l.column = pos, line, col
	return tok
}
