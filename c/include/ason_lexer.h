/**
 * @file ason_lexer.h
 * @brief ASON Lexer - tokenizes input string
 */

#ifndef ASON_LEXER_H
#define ASON_LEXER_H

#include "ason_value.h"

#ifdef __cplusplus
extern "C" {
#endif

/* Token types */
typedef enum {
    ASON_TOKEN_EOF,
    ASON_TOKEN_LPAREN,      /* ( */
    ASON_TOKEN_RPAREN,      /* ) */
    ASON_TOKEN_LBRACKET,    /* [ */
    ASON_TOKEN_RBRACKET,    /* ] */
    ASON_TOKEN_LBRACE,      /* { */
    ASON_TOKEN_RBRACE,      /* } */
    ASON_TOKEN_COMMA,       /* , */
    ASON_TOKEN_COLON,       /* : */
    ASON_TOKEN_IDENT,       /* identifier or unquoted string */
    ASON_TOKEN_STRING,      /* quoted string */
    ASON_TOKEN_ERROR,       /* lexer error */
} ason_token_type;

/* Token structure */
typedef struct {
    ason_token_type type;
    const char* start;      /* pointer into original input */
    size_t len;
    size_t line;
    size_t column;
} ason_token;

/* Lexer structure */
typedef struct {
    const char* input;
    size_t pos;
    size_t len;
    size_t line;
    size_t column;
    ason_error error;
    char error_msg[256];
} ason_lexer;

/** Initialize a lexer with input string */
void ason_lexer_init(ason_lexer* lexer, const char* input);

/** Get the next token */
ason_token ason_lexer_next(ason_lexer* lexer);

/** Peek at the next token without consuming it */
ason_token ason_lexer_peek(ason_lexer* lexer);

/** Get token text as a newly allocated string (caller must free) */
char* ason_token_to_string(const ason_token* token);

/** Check if token text equals a string */
bool ason_token_equals(const ason_token* token, const char* str);

#ifdef __cplusplus
}
#endif

#endif /* ASON_LEXER_H */

