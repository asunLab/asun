package ason

import (
	"fmt"
	"strconv"
)

// SchemaField represents a field in the schema.
type SchemaField struct {
	Name     string
	IsArray  bool
	Children []*SchemaField // for nested objects
}

// Parser parses ASON input.
type Parser struct {
	lexer *Lexer
	cur   Token
}

// NewParser creates a new parser for the given input.
func NewParser(input string) *Parser {
	p := &Parser{
		lexer: NewLexer(input),
	}
	p.advance()
	return p
}

func (p *Parser) advance() {
	p.cur = p.lexer.NextToken()
}

func (p *Parser) expect(typ TokenType) error {
	if p.cur.Type != typ {
		return fmt.Errorf("expected %v, got %v at line %d, column %d",
			typ, p.cur.Type, p.cur.Line, p.cur.Column)
	}
	p.advance()
	return nil
}

// Parse parses the ASON input and returns the result.
func (p *Parser) Parse() (*Value, error) {
	// Check if this is schema:data format or just data
	if p.cur.Type == TokenLBrace {
		return p.parseSchemaAndData()
	}
	// Pure data format
	return p.parseValue()
}

func (p *Parser) parseSchemaAndData() (*Value, error) {
	// Parse schema
	schema, err := p.parseSchema()
	if err != nil {
		return nil, err
	}

	// Expect colon
	if err := p.expect(TokenColon); err != nil {
		return nil, err
	}

	// Parse data tuples
	var results []*Value
	for {
		val, err := p.parseDataWithSchema(schema)
		if err != nil {
			return nil, err
		}
		results = append(results, val)

		if p.cur.Type != TokenComma {
			break
		}
		p.advance() // consume comma

		// Check if next is another tuple or end
		if p.cur.Type != TokenLParen {
			break
		}
	}

	if len(results) == 1 {
		return results[0], nil
	}
	return Array(results...), nil
}

func (p *Parser) parseSchema() ([]*SchemaField, error) {
	if err := p.expect(TokenLBrace); err != nil {
		return nil, err
	}

	var fields []*SchemaField
	for p.cur.Type != TokenRBrace && p.cur.Type != TokenEOF {
		field, err := p.parseSchemaField()
		if err != nil {
			return nil, err
		}
		fields = append(fields, field)

		if p.cur.Type == TokenComma {
			p.advance()
		}
	}

	if err := p.expect(TokenRBrace); err != nil {
		return nil, err
	}

	return fields, nil
}

func (p *Parser) parseSchemaField() (*SchemaField, error) {
	if p.cur.Type != TokenIdent {
		return nil, fmt.Errorf("expected identifier at line %d, column %d",
			p.cur.Line, p.cur.Column)
	}

	field := &SchemaField{Name: p.cur.Value}
	p.advance()

	// Check for array marker []
	if p.cur.Type == TokenLBracket {
		p.advance()
		field.IsArray = true
		// Check for nested object schema in array
		if p.cur.Type == TokenLBrace {
			children, err := p.parseSchema()
			if err != nil {
				return nil, err
			}
			field.Children = children
		}
		if err := p.expect(TokenRBracket); err != nil {
			return nil, err
		}
	} else if p.cur.Type == TokenLBrace {
		// Nested object
		children, err := p.parseSchema()
		if err != nil {
			return nil, err
		}
		field.Children = children
	}

	return field, nil
}

func (p *Parser) parseDataWithSchema(schema []*SchemaField) (*Value, error) {
	if err := p.expect(TokenLParen); err != nil {
		return nil, err
	}

	obj := Object()
	for i, field := range schema {
		if i > 0 {
			if p.cur.Type != TokenComma {
				return nil, fmt.Errorf("expected comma at line %d, column %d",
					p.cur.Line, p.cur.Column)
			}
			p.advance()
		}

		val, err := p.parseFieldValue(field)
		if err != nil {
			return nil, err
		}
		obj.Set(field.Name, val)
	}

	if err := p.expect(TokenRParen); err != nil {
		return nil, err
	}

	return obj, nil
}

func (p *Parser) parseFieldValue(field *SchemaField) (*Value, error) {
	if field.IsArray {
		return p.parseArrayValue(field)
	}
	if field.Children != nil {
		return p.parseDataWithSchema(field.Children)
	}
	return p.parseValue()
}

func (p *Parser) parseArrayValue(field *SchemaField) (*Value, error) {
	if err := p.expect(TokenLBracket); err != nil {
		return nil, err
	}

	arr := Array()
	first := true
	for p.cur.Type != TokenRBracket && p.cur.Type != TokenEOF {
		if !first {
			if p.cur.Type != TokenComma {
				break
			}
			p.advance()
		}
		first = false

		var val *Value
		var err error
		if field.Children != nil {
			val, err = p.parseDataWithSchema(field.Children)
		} else {
			val, err = p.parseValue()
		}
		if err != nil {
			return nil, err
		}
		arr.Push(val)
	}

	if err := p.expect(TokenRBracket); err != nil {
		return nil, err
	}

	return arr, nil
}

func (p *Parser) parseValue() (*Value, error) {
	switch p.cur.Type {
	case TokenNull:
		p.advance()
		return Null(), nil
	case TokenTrue:
		p.advance()
		return Bool(true), nil
	case TokenFalse:
		p.advance()
		return Bool(false), nil
	case TokenInteger:
		val, err := strconv.ParseInt(p.cur.Value, 10, 64)
		if err != nil {
			return nil, err
		}
		p.advance()
		return Integer(val), nil
	case TokenFloat:
		val, err := strconv.ParseFloat(p.cur.Value, 64)
		if err != nil {
			return nil, err
		}
		p.advance()
		return Float(val), nil
	case TokenString, TokenIdent:
		val := p.cur.Value
		p.advance()
		return String(val), nil
	case TokenLBracket:
		return p.parseArray()
	case TokenLParen:
		return p.parseTuple()
	default:
		return nil, fmt.Errorf("unexpected token %v at line %d, column %d",
			p.cur.Type, p.cur.Line, p.cur.Column)
	}
}

func (p *Parser) parseArray() (*Value, error) {
	if err := p.expect(TokenLBracket); err != nil {
		return nil, err
	}

	arr := Array()
	first := true
	for p.cur.Type != TokenRBracket && p.cur.Type != TokenEOF {
		if !first {
			if p.cur.Type != TokenComma {
				break
			}
			p.advance()
		}
		first = false

		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		arr.Push(val)
	}

	if err := p.expect(TokenRBracket); err != nil {
		return nil, err
	}

	return arr, nil
}

func (p *Parser) parseTuple() (*Value, error) {
	if err := p.expect(TokenLParen); err != nil {
		return nil, err
	}

	arr := Array()
	first := true
	for p.cur.Type != TokenRParen && p.cur.Type != TokenEOF {
		if !first {
			if p.cur.Type != TokenComma {
				break
			}
			p.advance()
		}
		first = false

		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		arr.Push(val)
	}

	if err := p.expect(TokenRParen); err != nil {
		return nil, err
	}

	return arr, nil
}

// Parse is the main entry point for parsing ASON.
func Parse(input string) (*Value, error) {
	p := NewParser(input)
	return p.Parse()
}
