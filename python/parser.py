"""ASON Parser."""

from dataclasses import dataclass
from typing import Any, Dict, List, Optional, Union

from lexer import Lexer, TokenType

AsonValue = Union[None, bool, int, float, str, List[Any], Dict[str, Any]]


@dataclass
class SchemaField:
    name: str
    is_array: bool = False
    children: Optional[List["SchemaField"]] = None


class ParseError(Exception):
    """ASON parse error."""
    pass


class Parser:
    """Parses ASON input."""
    
    def __init__(self, input_str: str):
        self.lexer = Lexer(input_str)
        self.cur = self.lexer.next_token()
    
    def advance(self):
        self.cur = self.lexer.next_token()
    
    def expect(self, typ: TokenType):
        if self.cur.type != typ:
            raise ParseError(f"Expected {typ}, got {self.cur.type} at line {self.cur.line}, column {self.cur.column}")
        self.advance()
    
    def parse(self) -> AsonValue:
        """Parse the ASON input."""
        if self.cur.type == TokenType.LBRACE:
            return self.parse_schema_and_data()
        return self.parse_value()
    
    def parse_schema_and_data(self) -> AsonValue:
        schema = self.parse_schema()
        self.expect(TokenType.COLON)
        
        results = []
        while True:
            val = self.parse_data_with_schema(schema)
            results.append(val)
            
            if self.cur.type != TokenType.COMMA:
                break
            self.advance()
            
            if self.cur.type != TokenType.LPAREN:
                break
        
        return results[0] if len(results) == 1 else results
    
    def parse_schema(self) -> List[SchemaField]:
        self.expect(TokenType.LBRACE)
        
        fields = []
        while self.cur.type not in (TokenType.RBRACE, TokenType.EOF):
            field = self.parse_schema_field()
            fields.append(field)
            
            if self.cur.type == TokenType.COMMA:
                self.advance()
        
        self.expect(TokenType.RBRACE)
        return fields
    
    def parse_schema_field(self) -> SchemaField:
        if self.cur.type != TokenType.IDENT:
            raise ParseError(f"Expected identifier at line {self.cur.line}, column {self.cur.column}")
        
        field = SchemaField(name=self.cur.value)
        self.advance()
        
        if self.cur.type == TokenType.LBRACKET:
            self.advance()
            field.is_array = True
            if self.cur.type == TokenType.LBRACE:
                field.children = self.parse_schema()
            self.expect(TokenType.RBRACKET)
        elif self.cur.type == TokenType.LBRACE:
            field.children = self.parse_schema()
        
        return field
    
    def parse_data_with_schema(self, schema: List[SchemaField]) -> Dict[str, Any]:
        self.expect(TokenType.LPAREN)
        
        obj = {}
        for i, field in enumerate(schema):
            if i > 0:
                if self.cur.type != TokenType.COMMA:
                    raise ParseError(f"Expected comma at line {self.cur.line}, column {self.cur.column}")
                self.advance()
            
            val = self.parse_field_value(field)
            obj[field.name] = val
        
        self.expect(TokenType.RPAREN)
        return obj
    
    def parse_field_value(self, field: SchemaField) -> AsonValue:
        if field.is_array:
            return self.parse_array_value(field)
        if field.children:
            return self.parse_data_with_schema(field.children)
        return self.parse_value()
    
    def parse_array_value(self, field: SchemaField) -> List[Any]:
        self.expect(TokenType.LBRACKET)
        
        arr = []
        first = True
        while self.cur.type not in (TokenType.RBRACKET, TokenType.EOF):
            if not first:
                if self.cur.type != TokenType.COMMA:
                    break
                self.advance()
            first = False
            
            if field.children:
                val = self.parse_data_with_schema(field.children)
            else:
                val = self.parse_value()
            arr.append(val)
        
        self.expect(TokenType.RBRACKET)
        return arr
    
    def parse_value(self) -> AsonValue:
        t = self.cur.type
        
        if t == TokenType.NULL:
            self.advance()
            return None
        if t == TokenType.TRUE:
            self.advance()
            return True
        if t == TokenType.FALSE:
            self.advance()
            return False
        if t == TokenType.INTEGER:
            val = int(self.cur.value)
            self.advance()
            return val
        if t == TokenType.FLOAT:
            val = float(self.cur.value)
            self.advance()
            return val
        if t in (TokenType.STRING, TokenType.IDENT):
            val = self.cur.value
            self.advance()
            return val
        if t == TokenType.LBRACKET:
            return self.parse_array()
        if t == TokenType.LPAREN:
            return self.parse_tuple()

        raise ParseError(f"Unexpected token {t} at line {self.cur.line}, column {self.cur.column}")

    def parse_array(self) -> List[Any]:
        self.expect(TokenType.LBRACKET)

        arr = []
        first = True
        while self.cur.type not in (TokenType.RBRACKET, TokenType.EOF):
            if not first:
                if self.cur.type != TokenType.COMMA:
                    break
                self.advance()
            first = False
            arr.append(self.parse_value())

        self.expect(TokenType.RBRACKET)
        return arr

    def parse_tuple(self) -> List[Any]:
        self.expect(TokenType.LPAREN)

        arr = []
        first = True
        while self.cur.type not in (TokenType.RPAREN, TokenType.EOF):
            if not first:
                if self.cur.type != TokenType.COMMA:
                    break
                self.advance()
            first = False
            arr.append(self.parse_value())

        self.expect(TokenType.RPAREN)
        return arr


def parse(input_str: str) -> AsonValue:
    """Parse an ASON string and return the result."""
    p = Parser(input_str)
    return p.parse()

