#include "ason.h"
#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* External test macros */
extern int tests_run, tests_passed, tests_failed;
#define TEST_ASSERT(cond, msg) do { \
    tests_run++; \
    if (!(cond)) { \
        printf("  FAIL: %s\n", msg); \
        tests_failed++; \
    } else { \
        tests_passed++; \
    } \
} while(0)

void test_lexer(void) {
    /* Test empty input */
    {
        ason_lexer lexer;
        ason_lexer_init(&lexer, "");
        ason_token tok = ason_lexer_next(&lexer);
        TEST_ASSERT(tok.type == ASON_TOKEN_EOF, "empty input gives EOF");
    }

    /* Test delimiters */
    {
        ason_lexer lexer;
        ason_lexer_init(&lexer, "()[]{},:");

        TEST_ASSERT(ason_lexer_next(&lexer).type == ASON_TOKEN_LPAREN, "lparen");
        TEST_ASSERT(ason_lexer_next(&lexer).type == ASON_TOKEN_RPAREN, "rparen");
        TEST_ASSERT(ason_lexer_next(&lexer).type == ASON_TOKEN_LBRACKET, "lbracket");
        TEST_ASSERT(ason_lexer_next(&lexer).type == ASON_TOKEN_RBRACKET, "rbracket");
        TEST_ASSERT(ason_lexer_next(&lexer).type == ASON_TOKEN_LBRACE, "lbrace");
        TEST_ASSERT(ason_lexer_next(&lexer).type == ASON_TOKEN_RBRACE, "rbrace");
        TEST_ASSERT(ason_lexer_next(&lexer).type == ASON_TOKEN_COMMA, "comma");
        TEST_ASSERT(ason_lexer_next(&lexer).type == ASON_TOKEN_COLON, "colon");
        TEST_ASSERT(ason_lexer_next(&lexer).type == ASON_TOKEN_EOF, "eof");
    }

    /* Test identifier */
    {
        ason_lexer lexer;
        ason_lexer_init(&lexer, "hello world123");

        ason_token tok1 = ason_lexer_next(&lexer);
        TEST_ASSERT(tok1.type == ASON_TOKEN_IDENT, "ident 1");
        TEST_ASSERT(ason_token_equals(&tok1, "hello"), "ident 1 value");

        ason_token tok2 = ason_lexer_next(&lexer);
        TEST_ASSERT(tok2.type == ASON_TOKEN_IDENT, "ident 2");
        TEST_ASSERT(ason_token_equals(&tok2, "world123"), "ident 2 value");
    }

    /* Test quoted string */
    {
        ason_lexer lexer;
        ason_lexer_init(&lexer, "\"hello world\"");

        ason_token tok = ason_lexer_next(&lexer);
        TEST_ASSERT(tok.type == ASON_TOKEN_STRING, "quoted string");

        char* str = ason_token_to_string(&tok);
        TEST_ASSERT(strcmp(str, "hello world") == 0, "quoted string value");
        free(str);
    }

    /* Test escaped quotes */
    {
        ason_lexer lexer;
        ason_lexer_init(&lexer, "\"hello\\\"world\"");

        ason_token tok = ason_lexer_next(&lexer);
        TEST_ASSERT(tok.type == ASON_TOKEN_STRING, "escaped string");

        char* str = ason_token_to_string(&tok);
        TEST_ASSERT(strcmp(str, "hello\"world") == 0, "escaped string value");
        free(str);
    }

    /* Test comments */
    {
        ason_lexer lexer;
        ason_lexer_init(&lexer, "/* comment */ hello");

        ason_token tok = ason_lexer_next(&lexer);
        TEST_ASSERT(tok.type == ASON_TOKEN_IDENT, "after comment");
        TEST_ASSERT(ason_token_equals(&tok, "hello"), "after comment value");
    }

    /* Test whitespace */
    {
        ason_lexer lexer;
        ason_lexer_init(&lexer, "  \t\n  hello  ");

        ason_token tok = ason_lexer_next(&lexer);
        TEST_ASSERT(tok.type == ASON_TOKEN_IDENT, "after whitespace");
        TEST_ASSERT(ason_token_equals(&tok, "hello"), "after whitespace value");
    }

    /* Test peek */
    {
        ason_lexer lexer;
        ason_lexer_init(&lexer, "hello world");

        ason_token peek = ason_lexer_peek(&lexer);
        TEST_ASSERT(ason_token_equals(&peek, "hello"), "peek 1");

        ason_token tok = ason_lexer_next(&lexer);
        TEST_ASSERT(ason_token_equals(&tok, "hello"), "after peek");

        tok = ason_lexer_next(&lexer);
        TEST_ASSERT(ason_token_equals(&tok, "world"), "next after peek");
    }

    printf("  Lexer tests completed\n");
}

