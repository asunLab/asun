package io.github.athxx.ason

/**
 * ASON Lexer - tokenizes input string.
 */
class Lexer(private val input: String) {
    
    enum class TokenType {
        EOF, ERROR,
        LPAREN, RPAREN, LBRACKET, RBRACKET, LBRACE, RBRACE,
        COMMA, COLON,
        NULL, TRUE, FALSE,
        INTEGER, FLOAT, STRING, IDENT
    }
    
    data class Token(
        val type: TokenType,
        val value: String,
        val line: Int,
        val column: Int
    )
    
    private var pos = 0
    private var line = 1
    private var column = 1
    
    private fun peek(): Char = if (pos < input.length) input[pos] else '\u0000'
    private fun peekN(n: Int): Char = if (pos + n < input.length) input[pos + n] else '\u0000'
    
    private fun advance(): Char {
        if (pos >= input.length) return '\u0000'
        val ch = input[pos++]
        if (ch == '\n') {
            line++
            column = 1
        } else {
            column++
        }
        return ch
    }
    
    private fun skipWhitespace() {
        while (true) {
            val ch = peek()
            if (ch == '\u0000') break
            when {
                ch in " \t\n\r" -> advance()
                ch == '/' && peekN(1) == '*' -> skipBlockComment()
                else -> break
            }
        }
    }
    
    private fun skipBlockComment() {
        advance() // /
        advance() // *
        while (true) {
            val ch = advance()
            if (ch == '\u0000') break
            if (ch == '*' && peek() == '/') {
                advance()
                break
            }
        }
    }
    
    fun nextToken(): Token {
        skipWhitespace()
        
        val ln = line
        val col = column
        val ch = peek()
        
        if (ch == '\u0000') return Token(TokenType.EOF, "", ln, col)
        
        return when (ch) {
            '(' -> { advance(); Token(TokenType.LPAREN, "(", ln, col) }
            ')' -> { advance(); Token(TokenType.RPAREN, ")", ln, col) }
            '[' -> { advance(); Token(TokenType.LBRACKET, "[", ln, col) }
            ']' -> { advance(); Token(TokenType.RBRACKET, "]", ln, col) }
            '{' -> { advance(); Token(TokenType.LBRACE, "{", ln, col) }
            '}' -> { advance(); Token(TokenType.RBRACE, "}", ln, col) }
            ',' -> { advance(); Token(TokenType.COMMA, ",", ln, col) }
            ':' -> { advance(); Token(TokenType.COLON, ":", ln, col) }
            '"' -> scanQuotedString()
            in "-+", in '0'..'9' -> scanNumber()
            in 'a'..'z', in 'A'..'Z', '_' -> scanIdentOrKeyword()
            else -> scanUnquotedString()
        }
    }
    
    private fun scanQuotedString(): Token {
        val ln = line
        val col = column
        advance() // opening quote
        
        val sb = StringBuilder()
        while (true) {
            val ch = peek()
            if (ch == '\u0000') return Token(TokenType.ERROR, "unterminated string", ln, col)
            if (ch == '"') {
                advance()
                break
            }
            if (ch == '\\') {
                advance()
                when (val escaped = advance()) {
                    'n' -> sb.append('\n')
                    't' -> sb.append('\t')
                    'r' -> sb.append('\r')
                    '\\' -> sb.append('\\')
                    '"' -> sb.append('"')
                    else -> sb.append(escaped)
                }
            } else {
                sb.append(advance())
            }
        }
        return Token(TokenType.STRING, sb.toString(), ln, col)
    }
    
    private fun scanNumber(): Token {
        val ln = line
        val col = column
        val start = pos
        var isFloat = false
        
        if (peek() in "+-") advance()
        while (peek().isDigit()) advance()
        
        if (peek() == '.' && peekN(1).isDigit()) {
            isFloat = true
            advance()
            while (peek().isDigit()) advance()
        }
        
        if (peek() in "eE") {
            isFloat = true
            advance()
            if (peek() in "+-") advance()
            while (peek().isDigit()) advance()
        }
        
        val value = input.substring(start, pos)
        return Token(if (isFloat) TokenType.FLOAT else TokenType.INTEGER, value, ln, col)
    }
    
    private fun scanIdentOrKeyword(): Token {
        val ln = line
        val col = column
        val start = pos
        
        while (peek().isLetterOrDigit() || peek() == '_') advance()
        
        val value = input.substring(start, pos)
        return when (value) {
            "null" -> Token(TokenType.NULL, value, ln, col)
            "true" -> Token(TokenType.TRUE, value, ln, col)
            "false" -> Token(TokenType.FALSE, value, ln, col)
            else -> Token(TokenType.IDENT, value, ln, col)
        }
    }
    
    private fun scanUnquotedString(): Token {
        val ln = line
        val col = column
        val sb = StringBuilder()
        
        while (true) {
            val ch = peek()
            if (ch == '\u0000' || ch in "()[]{},:\"" || ch.isWhitespace()) break
            sb.append(advance())
        }
        
        return Token(TokenType.STRING, sb.toString(), ln, col)
    }
}

