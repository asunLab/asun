/**
 * @file basic.c
 * @brief Basic example of using the ASON C library
 */

#include <stdio.h>
#include <stdlib.h>
#include "ason.h"

int main(void) {
    printf("=== ASON C Library Demo ===\n");
    printf("Version: %s\n\n", ASON_VERSION);
    
    /* 1. Parse simple object */
    printf("1. Parse simple object:\n");
    {
        ason_parse_result r = ason_parse("{name,age}:(Alice,30)");
        if (r.error == ASON_OK) {
            printf("   name = %s\n", ason_value_get_string(ason_object_get(r.value, "name")));
            printf("   age = %lld\n", (long long)ason_value_get_integer(ason_object_get(r.value, "age")));
        } else {
            printf("   Error: %s\n", r.error_msg);
        }
        ason_parse_result_free(&r);
    }
    printf("\n");
    
    /* 2. Parse multiple objects */
    printf("2. Parse multiple objects:\n");
    {
        ason_parse_result r = ason_parse("{name,age}:(Alice,30),(Bob,25)");
        if (r.error == ASON_OK) {
            for (size_t i = 0; i < ason_array_len(r.value); i++) {
                ason_value* obj = ason_array_get(r.value, i);
                printf("   [%zu] name = %s, age = %lld\n", i,
                       ason_value_get_string(ason_object_get(obj, "name")),
                       (long long)ason_value_get_integer(ason_object_get(obj, "age")));
            }
        }
        ason_parse_result_free(&r);
    }
    printf("\n");
    
    /* 3. Parse nested object */
    printf("3. Parse nested object:\n");
    {
        ason_parse_result r = ason_parse("{name,addr{city,zip}}:(Alice,(NYC,10001))");
        if (r.error == ASON_OK) {
            printf("   name = %s\n", ason_value_get_string(ason_object_get(r.value, "name")));
            ason_value* addr = ason_object_get(r.value, "addr");
            printf("   addr.city = %s\n", ason_value_get_string(ason_object_get(addr, "city")));
            printf("   addr.zip = %lld\n", (long long)ason_value_get_integer(ason_object_get(addr, "zip")));
        }
        ason_parse_result_free(&r);
    }
    printf("\n");
    
    /* 4. Parse array field */
    printf("4. Parse array field:\n");
    {
        ason_parse_result r = ason_parse("{name,scores[]}:(Alice,[90,85,95])");
        if (r.error == ASON_OK) {
            printf("   name = %s\n", ason_value_get_string(ason_object_get(r.value, "name")));
            ason_value* scores = ason_object_get(r.value, "scores");
            printf("   scores = [");
            for (size_t i = 0; i < ason_array_len(scores); i++) {
                if (i > 0) printf(", ");
                printf("%lld", (long long)ason_value_get_integer(ason_array_get(scores, i)));
            }
            printf("]\n");
        }
        ason_parse_result_free(&r);
    }
    printf("\n");
    
    /* 5. Build and serialize */
    printf("5. Build and serialize:\n");
    {
        ason_value* obj = ason_value_object();
        ason_object_set(obj, "name", ason_value_string("Bob"));
        ason_object_set(obj, "active", ason_value_bool(true));
        ason_object_set(obj, "score", ason_value_float(95.5));
        
        char* str = ason_to_string(obj);
        printf("   %s\n", str);
        
        free(str);
        ason_value_free(obj);
    }
    printf("\n");
    
    /* 6. Unicode support */
    printf("6. Unicode support:\n");
    {
        ason_parse_result r = ason_parse("{name,city}:(小明,北京)");
        if (r.error == ASON_OK) {
            printf("   name = %s\n", ason_value_get_string(ason_object_get(r.value, "name")));
            printf("   city = %s\n", ason_value_get_string(ason_object_get(r.value, "city")));
        }
        ason_parse_result_free(&r);
    }
    printf("\n");
    
    printf("=== Demo Complete ===\n");
    return 0;
}

