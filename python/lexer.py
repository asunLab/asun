"""ASON Lexer."""

from enum import Enum, auto
from dataclasses import dataclass


class TokenType(Enum):
    EOF = auto()
    ERROR = auto()
    LPAREN = auto()    # (
    RPAREN = auto()    # )
    LBRACKET = auto()  # [
    RBRACKET = auto()  # ]
    LBRACE = auto()    # {
    RBRACE = auto()    # }
    COMMA = auto()     # ,
    COLON = auto()     # :
    NULL = auto()
    TRUE = auto()
    FALSE = auto()
    INTEGER = auto()
    FLOAT = auto()
    STRING = auto()
    IDENT = auto()


@dataclass
class Token:
    type: TokenType
    value: str
    line: int
    column: int


class Lexer:
    """Tokenizes ASON input."""

    def __init__(self, input_str: str):
        self.input = input_str
        self.pos = 0
        self.line = 1
        self.column = 1

    def peek(self) -> str:
        if self.pos >= len(self.input):
            return ""
        return self.input[self.pos]

    def peek_n(self, n: int) -> str:
        pos = self.pos + n
        if pos >= len(self.input):
            return ""
        return self.input[pos]

    def advance(self) -> str:
        if self.pos >= len(self.input):
            return ""
        ch = self.input[self.pos]
        self.pos += 1
        if ch == "\n":
            self.line += 1
            self.column = 1
        else:
            self.column += 1
        return ch
    
    def skip_whitespace(self):
        while True:
            ch = self.peek()
            if not ch:
                break
            if ch in " \t\n\r":
                self.advance()
            elif ch == "/" and self.peek_n(1) == "*":
                self.skip_block_comment()
            else:
                break
    
    def skip_block_comment(self):
        self.advance()  # /
        self.advance()  # *
        while True:
            ch = self.advance()
            if not ch:
                break
            if ch == "*" and self.peek() == "/":
                self.advance()
                break
    
    def next_token(self) -> Token:
        self.skip_whitespace()
        
        line, col = self.line, self.column
        ch = self.peek()
        
        if not ch:
            return Token(TokenType.EOF, "", line, col)
        
        # Single character tokens
        single_chars = {
            "(": TokenType.LPAREN,
            ")": TokenType.RPAREN,
            "[": TokenType.LBRACKET,
            "]": TokenType.RBRACKET,
            "{": TokenType.LBRACE,
            "}": TokenType.RBRACE,
            ",": TokenType.COMMA,
            ":": TokenType.COLON,
        }
        
        if ch in single_chars:
            self.advance()
            return Token(single_chars[ch], ch, line, col)
        
        if ch == '"':
            return self.scan_quoted_string()
        
        if ch == "-" or ch == "+" or ch.isdigit():
            return self.scan_number()
        
        if ch.isalpha() or ch == "_":
            return self.scan_ident_or_keyword()
        
        return self.scan_unquoted_string()
    
    def scan_quoted_string(self) -> Token:
        line, col = self.line, self.column
        self.advance()  # opening quote
        
        result = []
        while True:
            ch = self.peek()
            if not ch:
                return Token(TokenType.ERROR, "unterminated string", line, col)
            if ch == '"':
                self.advance()
                break
            if ch == "\\":
                self.advance()
                escaped = self.advance()
                escape_map = {"n": "\n", "t": "\t", "r": "\r", "\\": "\\", '"': '"'}
                result.append(escape_map.get(escaped, escaped))
            else:
                result.append(self.advance())
        
        return Token(TokenType.STRING, "".join(result), line, col)
    
    def scan_number(self) -> Token:
        line, col = self.line, self.column
        start = self.pos
        is_float = False

        # Sign
        if self.peek() in "+-":
            self.advance()

        # Integer part
        while self.peek() and self.peek().isdigit():
            self.advance()

        # Decimal part
        if self.peek() == "." and self.peek_n(1) and self.peek_n(1).isdigit():
            is_float = True
            self.advance()
            while self.peek() and self.peek().isdigit():
                self.advance()

        # Exponent
        if self.peek() in "eE":
            is_float = True
            self.advance()
            if self.peek() in "+-":
                self.advance()
            while self.peek() and self.peek().isdigit():
                self.advance()

        value = self.input[start:self.pos]
        token_type = TokenType.FLOAT if is_float else TokenType.INTEGER
        return Token(token_type, value, line, col)

    def scan_ident_or_keyword(self) -> Token:
        line, col = self.line, self.column
        start = self.pos

        while self.peek() and (self.peek().isalnum() or self.peek() == "_"):
            self.advance()

        value = self.input[start:self.pos]
        keywords = {"null": TokenType.NULL, "true": TokenType.TRUE, "false": TokenType.FALSE}
        token_type = keywords.get(value, TokenType.IDENT)
        return Token(token_type, value, line, col)

    def scan_unquoted_string(self) -> Token:
        line, col = self.line, self.column
        result = []

        while True:
            ch = self.peek()
            if not ch or ch in "()[]{},:\"" or ch.isspace():
                break
            result.append(self.advance())

        return Token(TokenType.STRING, "".join(result), line, col)

    def peek_token(self) -> Token:
        pos, line, col = self.pos, self.line, self.column
        tok = self.next_token()
        self.pos, self.line, self.column = pos, line, col
        return tok

