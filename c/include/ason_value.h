/**
 * @file ason_value.h
 * @brief ASON Value types and operations
 */

#ifndef ASON_VALUE_H
#define ASON_VALUE_H

#include <stddef.h>
#include <stdint.h>
#include <stdbool.h>

#ifdef __cplusplus
extern "C" {
#endif

/* Error codes */
typedef enum {
    ASON_OK = 0,
    ASON_ERROR_NULL_POINTER,
    ASON_ERROR_OUT_OF_MEMORY,
    ASON_ERROR_TYPE_MISMATCH,
    ASON_ERROR_INDEX_OUT_OF_BOUNDS,
    ASON_ERROR_KEY_NOT_FOUND,
    ASON_ERROR_PARSE_ERROR,
    ASON_ERROR_UNEXPECTED_TOKEN,
    ASON_ERROR_UNCLOSED_STRING,
    ASON_ERROR_UNCLOSED_COMMENT,
    ASON_ERROR_INVALID_NUMBER,
} ason_error;

/* Value types */
typedef enum {
    ASON_TYPE_NULL,
    ASON_TYPE_BOOL,
    ASON_TYPE_INTEGER,
    ASON_TYPE_FLOAT,
    ASON_TYPE_STRING,
    ASON_TYPE_ARRAY,
    ASON_TYPE_OBJECT,
} ason_type;

/* Forward declarations */
typedef struct ason_value ason_value;
typedef struct ason_array ason_array;
typedef struct ason_object ason_object;

/* String with length (for binary safety) */
typedef struct {
    char* data;
    size_t len;
} ason_string;

/* Array structure */
struct ason_array {
    ason_value** items;
    size_t len;
    size_t capacity;
};

/* Object key-value pair */
typedef struct {
    char* key;
    ason_value* value;
} ason_kv;

/* Object structure (simple dynamic array of key-value pairs) */
struct ason_object {
    ason_kv* pairs;
    size_t len;
    size_t capacity;
};

/* Value structure */
struct ason_value {
    ason_type type;
    union {
        bool boolean;
        int64_t integer;
        double floating;
        ason_string string;
        ason_array array;
        ason_object object;
    } data;
};

/* ============ Value Creation ============ */

/** Create a null value */
ason_value* ason_value_null(void);

/** Create a boolean value */
ason_value* ason_value_bool(bool b);

/** Create an integer value */
ason_value* ason_value_integer(int64_t n);

/** Create a float value */
ason_value* ason_value_float(double f);

/** Create a string value (copies the string) */
ason_value* ason_value_string(const char* s);

/** Create a string value with length (for binary data) */
ason_value* ason_value_string_n(const char* s, size_t len);

/** Create an empty array */
ason_value* ason_value_array(void);

/** Create an empty object */
ason_value* ason_value_object(void);

/** Free a value and all its children */
void ason_value_free(ason_value* v);

/** Deep copy a value */
ason_value* ason_value_clone(const ason_value* v);

/* ============ Type Checking ============ */

ason_type ason_value_type(const ason_value* v);
bool ason_value_is_null(const ason_value* v);
bool ason_value_is_bool(const ason_value* v);
bool ason_value_is_integer(const ason_value* v);
bool ason_value_is_float(const ason_value* v);
bool ason_value_is_string(const ason_value* v);
bool ason_value_is_array(const ason_value* v);
bool ason_value_is_object(const ason_value* v);

/* ============ Value Access ============ */

/** Get boolean value. Returns false if not a bool. */
bool ason_value_get_bool(const ason_value* v);

/** Get integer value. Returns 0 if not an integer. */
int64_t ason_value_get_integer(const ason_value* v);

/** Get float value. Converts integer to float if needed. Returns 0.0 if invalid. */
double ason_value_get_float(const ason_value* v);

/** Get string pointer. Returns NULL if not a string. */
const char* ason_value_get_string(const ason_value* v);

/** Get string length. Returns 0 if not a string. */
size_t ason_value_get_string_len(const ason_value* v);

/* ============ Array Operations ============ */

/** Get array length */
size_t ason_array_len(const ason_value* arr);

/** Get array element by index. Returns NULL if out of bounds. */
ason_value* ason_array_get(const ason_value* arr, size_t index);

/** Append a value to array (takes ownership) */
ason_error ason_array_push(ason_value* arr, ason_value* item);

/* ============ Object Operations ============ */

/** Get object size (number of key-value pairs) */
size_t ason_object_len(const ason_value* obj);

/** Get value by key. Returns NULL if not found. */
ason_value* ason_object_get(const ason_value* obj, const char* key);

/** Set a key-value pair (takes ownership of value, copies key) */
ason_error ason_object_set(ason_value* obj, const char* key, ason_value* value);

/** Get key at index (for iteration) */
const char* ason_object_key_at(const ason_value* obj, size_t index);

/** Get value at index (for iteration) */
ason_value* ason_object_value_at(const ason_value* obj, size_t index);

/* ============ Error Handling ============ */

/** Get error message for error code */
const char* ason_error_str(ason_error err);

#ifdef __cplusplus
}
#endif

#endif /* ASON_VALUE_H */

