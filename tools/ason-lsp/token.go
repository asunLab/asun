package main

// TokenType represents the type of a lexical token.
type TokenType int

const (
	TokenLBrace   TokenType = iota // {
	TokenRBrace                    // }
	TokenLParen                    // (
	TokenRParen                    // )
	TokenLBracket                  // [
	TokenRBracket                  // ]
	TokenColon                     // :
	TokenComma                     // ,
	TokenIdent                     // field name
	TokenTypeHint                  // int, str, float, bool, etc.
	TokenMapKw                     // map keyword
	TokenString                    // quoted string "..."
	TokenNumber                    // integer or float
	TokenBool                      // true / false
	TokenPlainStr                  // unquoted string value
	TokenComment                   // /* ... */
	TokenNewline                   // \n
	TokenEOF                       // end of input
	TokenError                     // lexer error
)

var tokenNames = [...]string{
	TokenLBrace:   "LBrace",
	TokenRBrace:   "RBrace",
	TokenLParen:   "LParen",
	TokenRParen:   "RParen",
	TokenLBracket: "LBracket",
	TokenRBracket: "RBracket",
	TokenColon:    "Colon",
	TokenComma:    "Comma",
	TokenIdent:    "Ident",
	TokenTypeHint: "TypeHint",
	TokenMapKw:    "MapKw",
	TokenString:   "String",
	TokenNumber:   "Number",
	TokenBool:     "Bool",
	TokenPlainStr: "PlainStr",
	TokenComment:  "Comment",
	TokenNewline:  "Newline",
	TokenEOF:      "EOF",
	TokenError:    "Error",
}

func (t TokenType) String() string {
	if int(t) < len(tokenNames) {
		return tokenNames[t]
	}
	return "Unknown"
}

// Token is a single lexical token with positional information.
type Token struct {
	Type    TokenType
	Value   string
	Offset  int // byte offset in source
	Line    int // 0-based line number
	Col     int // 0-based column (byte offset within line)
	EndOff  int // byte offset of end (exclusive)
	EndLine int
	EndCol  int
}

// Pos returns the start position.
func (t Token) Pos() Position {
	return Position{Line: t.Line, Col: t.Col, Offset: t.Offset}
}

// End returns the end position.
func (t Token) End() Position {
	return Position{Line: t.EndLine, Col: t.EndCol, Offset: t.EndOff}
}

// Position in source.
type Position struct {
	Line   int
	Col    int
	Offset int
}
