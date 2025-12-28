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

void test_parser(void) {
    /* Test simple object */
    {
        ason_parse_result r = ason_parse("{name,age}:(Alice,30)");
        TEST_ASSERT(r.error == ASON_OK, "simple object parse");
        TEST_ASSERT(r.value != NULL, "simple object value");
        TEST_ASSERT(ason_value_is_object(r.value), "is object");
        TEST_ASSERT_EQ_STR(ason_value_get_string(ason_object_get(r.value, "name")), "Alice", "name");
        TEST_ASSERT_EQ_INT(ason_value_get_integer(ason_object_get(r.value, "age")), 30, "age");
        ason_parse_result_free(&r);
    }
    
    /* Test multiple objects */
    {
        ason_parse_result r = ason_parse("{name,age}:(Alice,30),(Bob,25)");
        TEST_ASSERT(r.error == ASON_OK, "multiple objects parse");
        TEST_ASSERT(ason_value_is_array(r.value), "is array");
        TEST_ASSERT_EQ_INT(ason_array_len(r.value), 2, "array length");
        
        ason_value* first = ason_array_get(r.value, 0);
        TEST_ASSERT_EQ_STR(ason_value_get_string(ason_object_get(first, "name")), "Alice", "first name");
        
        ason_value* second = ason_array_get(r.value, 1);
        TEST_ASSERT_EQ_STR(ason_value_get_string(ason_object_get(second, "name")), "Bob", "second name");
        
        ason_parse_result_free(&r);
    }
    
    /* Test nested object */
    {
        ason_parse_result r = ason_parse("{name,addr{city,zip}}:(Alice,(NYC,10001))");
        TEST_ASSERT(r.error == ASON_OK, "nested object parse");
        
        ason_value* addr = ason_object_get(r.value, "addr");
        TEST_ASSERT(addr != NULL, "addr exists");
        TEST_ASSERT(ason_value_is_object(addr), "addr is object");
        TEST_ASSERT_EQ_STR(ason_value_get_string(ason_object_get(addr, "city")), "NYC", "city");
        TEST_ASSERT_EQ_INT(ason_value_get_integer(ason_object_get(addr, "zip")), 10001, "zip");
        
        ason_parse_result_free(&r);
    }
    
    /* Test array field */
    {
        ason_parse_result r = ason_parse("{name,scores[]}:(Alice,[90,85,95])");
        TEST_ASSERT(r.error == ASON_OK, "array field parse");
        
        ason_value* scores = ason_object_get(r.value, "scores");
        TEST_ASSERT(scores != NULL, "scores exists");
        TEST_ASSERT(ason_value_is_array(scores), "scores is array");
        TEST_ASSERT_EQ_INT(ason_array_len(scores), 3, "scores length");
        TEST_ASSERT_EQ_INT(ason_value_get_integer(ason_array_get(scores, 0)), 90, "score 0");
        
        ason_parse_result_free(&r);
    }
    
    /* Test standalone array */
    {
        ason_parse_result r = ason_parse("[1,2,3]");
        TEST_ASSERT(r.error == ASON_OK, "standalone array parse");
        TEST_ASSERT(ason_value_is_array(r.value), "is array");
        TEST_ASSERT_EQ_INT(ason_array_len(r.value), 3, "array length");
        TEST_ASSERT_EQ_INT(ason_value_get_integer(ason_array_get(r.value, 0)), 1, "arr[0]");
        ason_parse_result_free(&r);
    }
    
    /* Test booleans */
    {
        ason_parse_result r = ason_parse("{active,admin}:(true,false)");
        TEST_ASSERT(r.error == ASON_OK, "booleans parse");
        TEST_ASSERT(ason_value_get_bool(ason_object_get(r.value, "active")) == true, "active");
        TEST_ASSERT(ason_value_get_bool(ason_object_get(r.value, "admin")) == false, "admin");
        ason_parse_result_free(&r);
    }
    
    /* Test floats */
    {
        ason_parse_result r = ason_parse("{price,tax}:(19.99,0.08)");
        TEST_ASSERT(r.error == ASON_OK, "floats parse");
        double price = ason_value_get_float(ason_object_get(r.value, "price"));
        TEST_ASSERT(price > 19.98 && price < 20.0, "price");
        ason_parse_result_free(&r);
    }
    
    /* Test null values */
    {
        ason_parse_result r = ason_parse("{name,age}:(Alice,)");
        TEST_ASSERT(r.error == ASON_OK, "null value parse");
        TEST_ASSERT(ason_value_is_null(ason_object_get(r.value, "age")), "age is null");
        ason_parse_result_free(&r);
    }
    
    /* Test negative numbers */
    {
        ason_parse_result r = ason_parse("{temp}:(-10)");
        TEST_ASSERT(r.error == ASON_OK, "negative number parse");
        TEST_ASSERT_EQ_INT(ason_value_get_integer(ason_object_get(r.value, "temp")), -10, "temp");
        ason_parse_result_free(&r);
    }
    
    /* Test Unicode */
    {
        ason_parse_result r = ason_parse("{name,city}:(小明,北京)");
        TEST_ASSERT(r.error == ASON_OK, "unicode parse");
        TEST_ASSERT_EQ_STR(ason_value_get_string(ason_object_get(r.value, "name")), "小明", "name");
        TEST_ASSERT_EQ_STR(ason_value_get_string(ason_object_get(r.value, "city")), "北京", "city");
        ason_parse_result_free(&r);
    }
    
    printf("  Parser tests completed\n");
}

