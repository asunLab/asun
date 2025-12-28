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

#define TEST_ASSERT_EQ_STR(a, b, msg) do { \
    tests_run++; \
    if (strcmp((a) ? (a) : "", (b) ? (b) : "") != 0) { \
        printf("  FAIL: %s (expected \"%s\", got \"%s\")\n", msg, (b) ? (b) : "(null)", (a) ? (a) : "(null)"); \
        tests_failed++; \
    } else { \
        tests_passed++; \
    } \
} while(0)

void test_serializer(void) {
    /* Test primitives */
    {
        char* s;
        
        ason_value* v = ason_value_bool(true);
        s = ason_to_string(v);
        TEST_ASSERT_EQ_STR(s, "true", "bool true");
        free(s);
        ason_value_free(v);
        
        v = ason_value_bool(false);
        s = ason_to_string(v);
        TEST_ASSERT_EQ_STR(s, "false", "bool false");
        free(s);
        ason_value_free(v);
        
        v = ason_value_integer(42);
        s = ason_to_string(v);
        TEST_ASSERT_EQ_STR(s, "42", "integer");
        free(s);
        ason_value_free(v);
        
        v = ason_value_float(3.14);
        s = ason_to_string(v);
        TEST_ASSERT(s != NULL, "float serialized");
        free(s);
        ason_value_free(v);
    }
    
    /* Test string no quotes */
    {
        ason_value* v = ason_value_string("hello");
        char* s = ason_to_string(v);
        TEST_ASSERT_EQ_STR(s, "hello", "string no quotes");
        free(s);
        ason_value_free(v);
    }
    
    /* Test string with quotes */
    {
        ason_value* v = ason_value_string("hello world");
        char* s = ason_to_string(v);
        TEST_ASSERT_EQ_STR(s, "\"hello world\"", "string with quotes");
        free(s);
        ason_value_free(v);
    }
    
    /* Test array */
    {
        ason_value* arr = ason_value_array();
        ason_array_push(arr, ason_value_integer(1));
        ason_array_push(arr, ason_value_integer(2));
        ason_array_push(arr, ason_value_integer(3));
        
        char* s = ason_to_string(arr);
        TEST_ASSERT_EQ_STR(s, "[1,2,3]", "array");
        free(s);
        ason_value_free(arr);
    }
    
    /* Test object */
    {
        ason_value* obj = ason_value_object();
        ason_object_set(obj, "name", ason_value_string("Alice"));
        ason_object_set(obj, "age", ason_value_integer(30));
        
        char* s = ason_to_string(obj);
        TEST_ASSERT_EQ_STR(s, "(Alice,30)", "object");
        free(s);
        ason_value_free(obj);
    }
    
    /* Test nested object */
    {
        ason_value* inner = ason_value_object();
        ason_object_set(inner, "city", ason_value_string("NYC"));
        ason_object_set(inner, "zip", ason_value_integer(10001));
        
        ason_value* outer = ason_value_object();
        ason_object_set(outer, "name", ason_value_string("Alice"));
        ason_object_set(outer, "addr", inner);
        
        char* s = ason_to_string(outer);
        TEST_ASSERT_EQ_STR(s, "(Alice,(NYC,10001))", "nested object");
        free(s);
        ason_value_free(outer);
    }
    
    /* Test special strings that need quotes */
    {
        ason_value* v = ason_value_string("123");
        char* s = ason_to_string(v);
        TEST_ASSERT_EQ_STR(s, "\"123\"", "number-like string");
        free(s);
        ason_value_free(v);
        
        v = ason_value_string("true");
        s = ason_to_string(v);
        TEST_ASSERT_EQ_STR(s, "\"true\"", "keyword-like string");
        free(s);
        ason_value_free(v);
    }
    
    printf("  Serializer tests completed\n");
}

