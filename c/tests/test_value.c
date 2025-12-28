#include "ason.h"
#include <stdio.h>
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

void test_value(void) {
    /* Test null */
    {
        ason_value* v = ason_value_null();
        TEST_ASSERT(v != NULL, "null value created");
        TEST_ASSERT(ason_value_is_null(v), "is_null");
        TEST_ASSERT(ason_value_type(v) == ASON_TYPE_NULL, "type is null");
        ason_value_free(v);
    }
    
    /* Test bool */
    {
        ason_value* v = ason_value_bool(true);
        TEST_ASSERT(v != NULL, "bool value created");
        TEST_ASSERT(ason_value_is_bool(v), "is_bool");
        TEST_ASSERT(ason_value_get_bool(v) == true, "bool value is true");
        ason_value_free(v);
    }
    
    /* Test integer */
    {
        ason_value* v = ason_value_integer(42);
        TEST_ASSERT(v != NULL, "integer value created");
        TEST_ASSERT(ason_value_is_integer(v), "is_integer");
        TEST_ASSERT_EQ_INT(ason_value_get_integer(v), 42, "integer value");
        ason_value_free(v);
    }
    
    /* Test float */
    {
        ason_value* v = ason_value_float(3.14);
        TEST_ASSERT(v != NULL, "float value created");
        TEST_ASSERT(ason_value_is_float(v), "is_float");
        TEST_ASSERT(ason_value_get_float(v) > 3.13 && ason_value_get_float(v) < 3.15, "float value");
        ason_value_free(v);
    }
    
    /* Test string */
    {
        ason_value* v = ason_value_string("hello");
        TEST_ASSERT(v != NULL, "string value created");
        TEST_ASSERT(ason_value_is_string(v), "is_string");
        TEST_ASSERT_EQ_STR(ason_value_get_string(v), "hello", "string value");
        TEST_ASSERT_EQ_INT(ason_value_get_string_len(v), 5, "string length");
        ason_value_free(v);
    }
    
    /* Test array */
    {
        ason_value* arr = ason_value_array();
        TEST_ASSERT(arr != NULL, "array created");
        TEST_ASSERT(ason_value_is_array(arr), "is_array");
        
        ason_array_push(arr, ason_value_integer(1));
        ason_array_push(arr, ason_value_integer(2));
        ason_array_push(arr, ason_value_integer(3));
        
        TEST_ASSERT_EQ_INT(ason_array_len(arr), 3, "array length");
        TEST_ASSERT_EQ_INT(ason_value_get_integer(ason_array_get(arr, 0)), 1, "array[0]");
        TEST_ASSERT_EQ_INT(ason_value_get_integer(ason_array_get(arr, 1)), 2, "array[1]");
        TEST_ASSERT_EQ_INT(ason_value_get_integer(ason_array_get(arr, 2)), 3, "array[2]");
        
        ason_value_free(arr);
    }
    
    /* Test object */
    {
        ason_value* obj = ason_value_object();
        TEST_ASSERT(obj != NULL, "object created");
        TEST_ASSERT(ason_value_is_object(obj), "is_object");
        
        ason_object_set(obj, "name", ason_value_string("Alice"));
        ason_object_set(obj, "age", ason_value_integer(30));
        
        TEST_ASSERT_EQ_INT(ason_object_len(obj), 2, "object length");
        TEST_ASSERT_EQ_STR(ason_value_get_string(ason_object_get(obj, "name")), "Alice", "obj.name");
        TEST_ASSERT_EQ_INT(ason_value_get_integer(ason_object_get(obj, "age")), 30, "obj.age");
        
        ason_value_free(obj);
    }
    
    /* Test clone */
    {
        ason_value* original = ason_value_string("test");
        ason_value* clone = ason_value_clone(original);
        
        TEST_ASSERT(clone != NULL, "clone created");
        TEST_ASSERT_EQ_STR(ason_value_get_string(clone), "test", "clone value");
        
        ason_value_free(original);
        ason_value_free(clone);
    }
    
    printf("  Value tests completed\n");
}

