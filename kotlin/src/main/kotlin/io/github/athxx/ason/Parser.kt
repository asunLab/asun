package io.github.athxx.ason

/**
 * ASON Parser.
 */
class Parser(input: String) {
    
    private val lexer = Lexer(input)
    private var current: Lexer.Token = lexer.nextToken()
    
    private fun advance() {
        current = lexer.nextToken()
    }
    
    private fun expect(type: Lexer.TokenType) {
        if (current.type != type) {
            throw ParseException("Expected $type, got ${current.type} at line ${current.line}, column ${current.column}")
        }
        advance()
    }
    
    fun parse(): Value {
        return if (current.type == Lexer.TokenType.LBRACE) {
            parseSchemaAndData()
        } else {
            parseValue()
        }
    }
    
    private fun parseSchemaAndData(): Value {
        val schema = parseSchema()
        expect(Lexer.TokenType.COLON)
        
        val results = mutableListOf<Value>()
        while (true) {
            results.add(parseDataWithSchema(schema))
            if (current.type != Lexer.TokenType.COMMA) break
            advance()
            if (current.type != Lexer.TokenType.LPAREN) break
        }
        
        return if (results.size == 1) results[0] else Value.Array(results)
    }
    
    private data class SchemaField(
        val name: String,
        var isArray: Boolean = false,
        var children: List<SchemaField>? = null
    )
    
    private fun parseSchema(): List<SchemaField> {
        expect(Lexer.TokenType.LBRACE)
        
        val fields = mutableListOf<SchemaField>()
        while (current.type != Lexer.TokenType.RBRACE && current.type != Lexer.TokenType.EOF) {
            fields.add(parseSchemaField())
            if (current.type == Lexer.TokenType.COMMA) advance()
        }
        expect(Lexer.TokenType.RBRACE)
        return fields
    }
    
    private fun parseSchemaField(): SchemaField {
        if (current.type != Lexer.TokenType.IDENT) {
            throw ParseException("Expected identifier at line ${current.line}")
        }
        
        val field = SchemaField(current.value)
        advance()
        
        when (current.type) {
            Lexer.TokenType.LBRACKET -> {
                advance()
                field.isArray = true
                if (current.type == Lexer.TokenType.LBRACE) {
                    field.children = parseSchema()
                }
                expect(Lexer.TokenType.RBRACKET)
            }
            Lexer.TokenType.LBRACE -> {
                field.children = parseSchema()
            }
            else -> {}
        }
        
        return field
    }
    
    private fun parseDataWithSchema(schema: List<SchemaField>): Value {
        expect(Lexer.TokenType.LPAREN)
        
        val obj = Value.Object()
        schema.forEachIndexed { i, field ->
            if (i > 0) {
                if (current.type != Lexer.TokenType.COMMA) {
                    throw ParseException("Expected comma at line ${current.line}")
                }
                advance()
            }
            obj[field.name] = parseFieldValue(field)
        }
        
        expect(Lexer.TokenType.RPAREN)
        return obj
    }
    
    private fun parseFieldValue(field: SchemaField): Value {
        return when {
            field.isArray -> parseArrayValue(field)
            field.children != null -> parseDataWithSchema(field.children!!)
            else -> parseValue()
        }
    }
    
    private fun parseArrayValue(field: SchemaField): Value {
        expect(Lexer.TokenType.LBRACKET)
        
        val arr = Value.Array()
        var first = true
        while (current.type != Lexer.TokenType.RBRACKET && current.type != Lexer.TokenType.EOF) {
            if (!first) {
                if (current.type != Lexer.TokenType.COMMA) break
                advance()
            }
            first = false
            
            val item = if (field.children != null) parseDataWithSchema(field.children!!) else parseValue()
            arr.add(item)
        }
        
        expect(Lexer.TokenType.RBRACKET)
        return arr
    }
    
    private fun parseValue(): Value {
        return when (current.type) {
            Lexer.TokenType.NULL -> { advance(); Value.Null }
            Lexer.TokenType.TRUE -> { advance(); Value.Bool(true) }
            Lexer.TokenType.FALSE -> { advance(); Value.Bool(false) }
            Lexer.TokenType.INTEGER -> {
                val v = current.value.toLong()
                advance()
                Value.Integer(v)
            }
            Lexer.TokenType.FLOAT -> {
                val v = current.value.toDouble()
                advance()
                Value.Float(v)
            }
            Lexer.TokenType.STRING, Lexer.TokenType.IDENT -> {
                val v = current.value
                advance()
                Value.Str(v)
            }
            Lexer.TokenType.LBRACKET -> parseArray()
            Lexer.TokenType.LPAREN -> parseTuple()
            else -> throw ParseException("Unexpected token ${current.type} at line ${current.line}")
        }
    }

    private fun parseArray(): Value {
        expect(Lexer.TokenType.LBRACKET)

        val arr = Value.Array()
        var first = true
        while (current.type != Lexer.TokenType.RBRACKET && current.type != Lexer.TokenType.EOF) {
            if (!first) {
                if (current.type != Lexer.TokenType.COMMA) break
                advance()
            }
            first = false
            arr.add(parseValue())
        }

        expect(Lexer.TokenType.RBRACKET)
        return arr
    }

    private fun parseTuple(): Value {
        expect(Lexer.TokenType.LPAREN)

        val arr = Value.Array()
        var first = true
        while (current.type != Lexer.TokenType.RPAREN && current.type != Lexer.TokenType.EOF) {
            if (!first) {
                if (current.type != Lexer.TokenType.COMMA) break
                advance()
            }
            first = false
            arr.add(parseValue())
        }

        expect(Lexer.TokenType.RPAREN)
        return arr
    }

    companion object {
        fun parse(input: String): Value = Parser(input).parse()
    }
}

