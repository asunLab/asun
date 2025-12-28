/**
 * @file test_main.c
 * @brief Simple test framework for ASON C library
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

/* Test framework - global variables */
int tests_run = 0;
int tests_passed = 0;
int tests_failed = 0;

#define TEST_ASSERT(cond, msg) do { \
    tests_run++; \
    if (!(cond)) { \
        printf("  FAIL: %s\n", msg); \
        tests_failed++; \
    } else { \
        tests_passed++; \
    } \
} while(0)

#define TEST_ASSERT_EQ_INT(a, b, msg) do { \
    tests_run++; \
    if ((a) != (b)) { \
        printf("  FAIL: %s (expected %lld, got %lld)\n", msg, (long long)(b), (long long)(a)); \
        tests_failed++; \
    } else { \
        tests_passed++; \
    } \
} while(0)

#define TEST_ASSERT_EQ_STR(a, b, msg) do { \
    tests_run++; \
    if (strcmp((a) ? (a) : "", (b) ? (b) : "") != 0) { \
        printf("  FAIL: %s (expected \"%s\", got \"%s\")\n", msg, (b) ? (b) : "(null)", (a) ? (a) : "(null)"); \
        tests_failed++; \
    } else { \
        tests_passed++; \
    } \
} while(0)

/* Test function declarations */
void test_value(void);
void test_lexer(void);
void test_parser(void);
void test_serializer(void);

int main(void) {
    printf("=== ASON C Library Tests ===\n\n");
    
    printf("[Value Tests]\n");
    test_value();
    printf("\n");
    
    printf("[Lexer Tests]\n");
    test_lexer();
    printf("\n");
    
    printf("[Parser Tests]\n");
    test_parser();
    printf("\n");
    
    printf("[Serializer Tests]\n");
    test_serializer();
    printf("\n");
    
    printf("=== Results ===\n");
    printf("Tests run: %d\n", tests_run);
    printf("Passed: %d\n", tests_passed);
    printf("Failed: %d\n", tests_failed);
    
    return tests_failed > 0 ? 1 : 0;
}

