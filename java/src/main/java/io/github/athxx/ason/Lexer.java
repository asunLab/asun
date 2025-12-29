package io.github.athxx.ason;

/**
 * ASON Lexer - tokenizes input string.
 */
public class Lexer {
    
    public enum TokenType {
        EOF, ERROR,
        LPAREN, RPAREN, LBRACKET, RBRACKET, LBRACE, RBRACE,
        COMMA, COLON,
        NULL, TRUE, FALSE,
        INTEGER, FLOAT, STRING, IDENT
    }
    
    public static class Token {
        public final TokenType type;
        public final String value;
        public final int line;
        public final int column;
        
        public Token(TokenType type, String value, int line, int column) {
            this.type = type;
            this.value = value;
            this.line = line;
            this.column = column;
        }
    }
    
    private final String input;
    private int pos = 0;
    private int line = 1;
    private int column = 1;
    
    public Lexer(String input) {
        this.input = input;
    }
    
    private char peek() {
        return pos < input.length() ? input.charAt(pos) : 0;
    }
    
    private char peekN(int n) {
        int p = pos + n;
        return p < input.length() ? input.charAt(p) : 0;
    }
    
    private char advance() {
        if (pos >= input.length()) return 0;
        char ch = input.charAt(pos++);
        if (ch == '\n') {
            line++;
            column = 1;
        } else {
            column++;
        }
        return ch;
    }
    
    private void skipWhitespace() {
        while (true) {
            char ch = peek();
            if (ch == 0) break;
            if (ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r') {
                advance();
            } else if (ch == '/' && peekN(1) == '*') {
                skipBlockComment();
            } else {
                break;
            }
        }
    }
    
    private void skipBlockComment() {
        advance(); // /
        advance(); // *
        while (true) {
            char ch = advance();
            if (ch == 0) break;
            if (ch == '*' && peek() == '/') {
                advance();
                break;
            }
        }
    }
    
    public Token nextToken() {
        skipWhitespace();
        
        int ln = line, col = column;
        char ch = peek();
        
        if (ch == 0) return new Token(TokenType.EOF, "", ln, col);
        
        switch (ch) {
            case '(': advance(); return new Token(TokenType.LPAREN, "(", ln, col);
            case ')': advance(); return new Token(TokenType.RPAREN, ")", ln, col);
            case '[': advance(); return new Token(TokenType.LBRACKET, "[", ln, col);
            case ']': advance(); return new Token(TokenType.RBRACKET, "]", ln, col);
            case '{': advance(); return new Token(TokenType.LBRACE, "{", ln, col);
            case '}': advance(); return new Token(TokenType.RBRACE, "}", ln, col);
            case ',': advance(); return new Token(TokenType.COMMA, ",", ln, col);
            case ':': advance(); return new Token(TokenType.COLON, ":", ln, col);
            case '"': return scanQuotedString();
        }
        
        if (ch == '-' || ch == '+' || Character.isDigit(ch)) {
            return scanNumber();
        }
        
        if (Character.isLetter(ch) || ch == '_') {
            return scanIdentOrKeyword();
        }
        
        return scanUnquotedString();
    }
    
    private Token scanQuotedString() {
        int ln = line, col = column;
        advance(); // opening quote
        
        StringBuilder sb = new StringBuilder();
        while (true) {
            char ch = peek();
            if (ch == 0) return new Token(TokenType.ERROR, "unterminated string", ln, col);
            if (ch == '"') {
                advance();
                break;
            }
            if (ch == '\\') {
                advance();
                char escaped = advance();
                switch (escaped) {
                    case 'n': sb.append('\n'); break;
                    case 't': sb.append('\t'); break;
                    case 'r': sb.append('\r'); break;
                    case '\\': sb.append('\\'); break;
                    case '"': sb.append('"'); break;
                    default: sb.append(escaped);
                }
            } else {
                sb.append(advance());
            }
        }
        return new Token(TokenType.STRING, sb.toString(), ln, col);
    }
    
    private Token scanNumber() {
        int ln = line, col = column;
        int start = pos;
        boolean isFloat = false;
        
        if (peek() == '-' || peek() == '+') advance();
        while (Character.isDigit(peek())) advance();
        
        if (peek() == '.' && Character.isDigit(peekN(1))) {
            isFloat = true;
            advance();
            while (Character.isDigit(peek())) advance();
        }
        
        if (peek() == 'e' || peek() == 'E') {
            isFloat = true;
            advance();
            if (peek() == '+' || peek() == '-') advance();
            while (Character.isDigit(peek())) advance();
        }

        String value = input.substring(start, pos);
        return new Token(isFloat ? TokenType.FLOAT : TokenType.INTEGER, value, ln, col);
    }

    private Token scanIdentOrKeyword() {
        int ln = line, col = column;
        int start = pos;

        while (Character.isLetterOrDigit(peek()) || peek() == '_') {
            advance();
        }

        String value = input.substring(start, pos);
        switch (value) {
            case "null": return new Token(TokenType.NULL, value, ln, col);
            case "true": return new Token(TokenType.TRUE, value, ln, col);
            case "false": return new Token(TokenType.FALSE, value, ln, col);
            default: return new Token(TokenType.IDENT, value, ln, col);
        }
    }

    private Token scanUnquotedString() {
        int ln = line, col = column;
        StringBuilder sb = new StringBuilder();

        while (true) {
            char ch = peek();
            if (ch == 0 || "()[]{},:\"".indexOf(ch) >= 0 || Character.isWhitespace(ch)) {
                break;
            }
            sb.append(advance());
        }

        return new Token(TokenType.STRING, sb.toString(), ln, col);
    }
}

