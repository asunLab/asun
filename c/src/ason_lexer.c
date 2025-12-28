#include "ason_lexer.h"
#include <stdlib.h>
#include <string.h>
#include <ctype.h>

static char peek_char(ason_lexer* lexer) {
    if (lexer->pos >= lexer->len) return '\0';
    return lexer->input[lexer->pos];
}

static char advance(ason_lexer* lexer) {
    if (lexer->pos >= lexer->len) return '\0';
    char c = lexer->input[lexer->pos++];
    if (c == '\n') {
        lexer->line++;
        lexer->column = 1;
    } else {
        lexer->column++;
    }
    return c;
}

static void skip_whitespace(ason_lexer* lexer) {
    while (lexer->pos < lexer->len) {
        char c = peek_char(lexer);
        if (c == ' ' || c == '\t' || c == '\n' || c == '\r') {
            advance(lexer);
        } else if (c == '/' && lexer->pos + 1 < lexer->len && lexer->input[lexer->pos + 1] == '*') {
            /* Block comment */
            advance(lexer); /* / */
            advance(lexer); /* * */
            while (lexer->pos < lexer->len) {
                if (peek_char(lexer) == '*' && lexer->pos + 1 < lexer->len && 
                    lexer->input[lexer->pos + 1] == '/') {
                    advance(lexer); /* * */
                    advance(lexer); /* / */
                    break;
                }
                if (lexer->pos >= lexer->len) {
                    lexer->error = ASON_ERROR_UNCLOSED_COMMENT;
                    return;
                }
                advance(lexer);
            }
        } else {
            break;
        }
    }
}

static bool is_ident_start(char c) {
    return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') || c == '_' ||
           (unsigned char)c >= 0x80; /* UTF-8 */
}

static bool is_ident_char(char c) {
    return is_ident_start(c) || (c >= '0' && c <= '9') || c == '-' || c == '.';
}

static bool is_special_char(char c) {
    return c == '(' || c == ')' || c == '[' || c == ']' || c == '{' || c == '}' ||
           c == ',' || c == ':' || c == '"' || c == '\'' || c == ' ' || c == '\t' ||
           c == '\n' || c == '\r' || c == '\0';
}

void ason_lexer_init(ason_lexer* lexer, const char* input) {
    lexer->input = input;
    lexer->pos = 0;
    lexer->len = input ? strlen(input) : 0;
    lexer->line = 1;
    lexer->column = 1;
    lexer->error = ASON_OK;
    lexer->error_msg[0] = '\0';
}

ason_token ason_lexer_next(ason_lexer* lexer) {
    ason_token token = {ASON_TOKEN_EOF, NULL, 0, lexer->line, lexer->column};
    
    skip_whitespace(lexer);
    
    if (lexer->error != ASON_OK) {
        token.type = ASON_TOKEN_ERROR;
        return token;
    }
    
    if (lexer->pos >= lexer->len) {
        return token; /* EOF */
    }
    
    token.start = lexer->input + lexer->pos;
    token.line = lexer->line;
    token.column = lexer->column;
    
    char c = peek_char(lexer);
    
    switch (c) {
        case '(': advance(lexer); token.type = ASON_TOKEN_LPAREN; token.len = 1; return token;
        case ')': advance(lexer); token.type = ASON_TOKEN_RPAREN; token.len = 1; return token;
        case '[': advance(lexer); token.type = ASON_TOKEN_LBRACKET; token.len = 1; return token;
        case ']': advance(lexer); token.type = ASON_TOKEN_RBRACKET; token.len = 1; return token;
        case '{': advance(lexer); token.type = ASON_TOKEN_LBRACE; token.len = 1; return token;
        case '}': advance(lexer); token.type = ASON_TOKEN_RBRACE; token.len = 1; return token;
        case ',': advance(lexer); token.type = ASON_TOKEN_COMMA; token.len = 1; return token;
        case ':': advance(lexer); token.type = ASON_TOKEN_COLON; token.len = 1; return token;
        default: break;
    }
    
    /* Quoted string */
    if (c == '"' || c == '\'') {
        char quote = c;
        advance(lexer); /* skip opening quote */
        token.start = lexer->input + lexer->pos;
        size_t start_pos = lexer->pos;
        
        while (lexer->pos < lexer->len) {
            c = peek_char(lexer);
            if (c == quote) {
                token.len = lexer->pos - start_pos;
                advance(lexer); /* skip closing quote */
                token.type = ASON_TOKEN_STRING;
                return token;
            }
            if (c == '\\' && lexer->pos + 1 < lexer->len) {
                advance(lexer); /* skip backslash */
            }
            advance(lexer);
        }
        
        lexer->error = ASON_ERROR_UNCLOSED_STRING;
        token.type = ASON_TOKEN_ERROR;
        return token;
    }
    
    /* Identifier or unquoted value */
    if (!is_special_char(c)) {
        size_t start_pos = lexer->pos;
        while (lexer->pos < lexer->len && !is_special_char(peek_char(lexer))) {
            advance(lexer);
        }
        token.len = lexer->pos - start_pos;
        token.type = ASON_TOKEN_IDENT;
        return token;
    }
    
    /* Unknown character - treat as single char token */
    advance(lexer);
    token.type = ASON_TOKEN_ERROR;
    token.len = 1;
    return token;
}

ason_token ason_lexer_peek(ason_lexer* lexer) {
    size_t saved_pos = lexer->pos;
    size_t saved_line = lexer->line;
    size_t saved_col = lexer->column;

    ason_token token = ason_lexer_next(lexer);

    lexer->pos = saved_pos;
    lexer->line = saved_line;
    lexer->column = saved_col;

    return token;
}

char* ason_token_to_string(const ason_token* token) {
    if (!token || !token->start) return NULL;

    char* str = (char*)malloc(token->len + 1);
    if (!str) return NULL;

    /* Handle escape sequences for quoted strings */
    size_t j = 0;
    for (size_t i = 0; i < token->len; i++) {
        if (token->start[i] == '\\' && i + 1 < token->len) {
            i++;
            switch (token->start[i]) {
                case 'n': str[j++] = '\n'; break;
                case 't': str[j++] = '\t'; break;
                case 'r': str[j++] = '\r'; break;
                case '\\': str[j++] = '\\'; break;
                case '"': str[j++] = '"'; break;
                case '\'': str[j++] = '\''; break;
                default: str[j++] = token->start[i]; break;
            }
        } else {
            str[j++] = token->start[i];
        }
    }
    str[j] = '\0';
    return str;
}

bool ason_token_equals(const ason_token* token, const char* str) {
    if (!token || !str) return false;
    size_t slen = strlen(str);
    if (token->len != slen) return false;
    return strncmp(token->start, str, token->len) == 0;
}

