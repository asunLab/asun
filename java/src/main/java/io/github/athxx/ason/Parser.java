package io.github.athxx.ason;

import java.util.ArrayList;
import java.util.List;

/**
 * ASON Parser.
 */
public class Parser {
    
    private final Lexer lexer;
    private Lexer.Token current;
    
    public Parser(String input) {
        this.lexer = new Lexer(input);
        advance();
    }
    
    private void advance() {
        current = lexer.nextToken();
    }
    
    private void expect(Lexer.TokenType type) throws ParseException {
        if (current.type != type) {
            throw new ParseException("Expected " + type + ", got " + current.type +
                " at line " + current.line + ", column " + current.column);
        }
        advance();
    }
    
    public Value parse() throws ParseException {
        if (current.type == Lexer.TokenType.LBRACE) {
            return parseSchemaAndData();
        }
        return parseValue();
    }
    
    private Value parseSchemaAndData() throws ParseException {
        List<SchemaField> schema = parseSchema();
        expect(Lexer.TokenType.COLON);
        
        List<Value> results = new ArrayList<>();
        while (true) {
            Value val = parseDataWithSchema(schema);
            results.add(val);
            
            if (current.type != Lexer.TokenType.COMMA) break;
            advance();
            if (current.type != Lexer.TokenType.LPAREN) break;
        }
        
        if (results.size() == 1) {
            return results.get(0);
        }
        return Value.ofArray(results.toArray(new Value[0]));
    }
    
    private List<SchemaField> parseSchema() throws ParseException {
        expect(Lexer.TokenType.LBRACE);
        
        List<SchemaField> fields = new ArrayList<>();
        while (current.type != Lexer.TokenType.RBRACE && current.type != Lexer.TokenType.EOF) {
            fields.add(parseSchemaField());
            if (current.type == Lexer.TokenType.COMMA) {
                advance();
            }
        }
        expect(Lexer.TokenType.RBRACE);
        return fields;
    }
    
    private SchemaField parseSchemaField() throws ParseException {
        if (current.type != Lexer.TokenType.IDENT) {
            throw new ParseException("Expected identifier at line " + current.line);
        }
        
        SchemaField field = new SchemaField(current.value);
        advance();
        
        if (current.type == Lexer.TokenType.LBRACKET) {
            advance();
            field.isArray = true;
            if (current.type == Lexer.TokenType.LBRACE) {
                field.children = parseSchema();
            }
            expect(Lexer.TokenType.RBRACKET);
        } else if (current.type == Lexer.TokenType.LBRACE) {
            field.children = parseSchema();
        }
        
        return field;
    }
    
    private Value parseDataWithSchema(List<SchemaField> schema) throws ParseException {
        expect(Lexer.TokenType.LPAREN);
        
        Value obj = Value.ofObject();
        for (int i = 0; i < schema.size(); i++) {
            if (i > 0) {
                if (current.type != Lexer.TokenType.COMMA) {
                    throw new ParseException("Expected comma at line " + current.line);
                }
                advance();
            }
            
            SchemaField field = schema.get(i);
            Value val = parseFieldValue(field);
            obj.set(field.name, val);
        }
        
        expect(Lexer.TokenType.RPAREN);
        return obj;
    }
    
    private Value parseFieldValue(SchemaField field) throws ParseException {
        if (field.isArray) {
            return parseArrayValue(field);
        }
        if (field.children != null) {
            return parseDataWithSchema(field.children);
        }
        return parseValue();
    }
    
    private Value parseArrayValue(SchemaField field) throws ParseException {
        expect(Lexer.TokenType.LBRACKET);
        
        Value arr = Value.ofArray();
        boolean first = true;
        while (current.type != Lexer.TokenType.RBRACKET && current.type != Lexer.TokenType.EOF) {
            if (!first) {
                if (current.type != Lexer.TokenType.COMMA) break;
                advance();
            }
            first = false;
            
            Value val = (field.children != null) ? parseDataWithSchema(field.children) : parseValue();
            arr.push(val);
        }
        
        expect(Lexer.TokenType.RBRACKET);
        return arr;
    }
    
    private Value parseValue() throws ParseException {
        switch (current.type) {
            case NULL:
                advance();
                return Value.ofNull();
            case TRUE:
                advance();
                return Value.ofBool(true);
            case FALSE:
                advance();
                return Value.ofBool(false);
            case INTEGER:
                long intVal = Long.parseLong(current.value);
                advance();
                return Value.ofInteger(intVal);
            case FLOAT:
                double floatVal = Double.parseDouble(current.value);
                advance();
                return Value.ofFloat(floatVal);
            case STRING:
            case IDENT:
                String strVal = current.value;
                advance();
                return Value.ofString(strVal);
            case LBRACKET:
                return parseArray();
            case LPAREN:
                return parseTuple();
            default:
                throw new ParseException("Unexpected token " + current.type +
                    " at line " + current.line + ", column " + current.column);
        }
    }

    private Value parseArray() throws ParseException {
        expect(Lexer.TokenType.LBRACKET);

        Value arr = Value.ofArray();
        boolean first = true;
        while (current.type != Lexer.TokenType.RBRACKET && current.type != Lexer.TokenType.EOF) {
            if (!first) {
                if (current.type != Lexer.TokenType.COMMA) break;
                advance();
            }
            first = false;
            arr.push(parseValue());
        }

        expect(Lexer.TokenType.RBRACKET);
        return arr;
    }

    private Value parseTuple() throws ParseException {
        expect(Lexer.TokenType.LPAREN);

        Value arr = Value.ofArray();
        boolean first = true;
        while (current.type != Lexer.TokenType.RPAREN && current.type != Lexer.TokenType.EOF) {
            if (!first) {
                if (current.type != Lexer.TokenType.COMMA) break;
                advance();
            }
            first = false;
            arr.push(parseValue());
        }

        expect(Lexer.TokenType.RPAREN);
        return arr;
    }

    // Schema field helper class
    private static class SchemaField {
        String name;
        boolean isArray = false;
        List<SchemaField> children = null;

        SchemaField(String name) {
            this.name = name;
        }
    }

    // Static parse method
    public static Value parse(String input) throws ParseException {
        return new Parser(input).parse();
    }
}

