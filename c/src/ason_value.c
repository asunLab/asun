#include "ason_value.h"
#include <stdlib.h>
#include <string.h>

/* ============ Value Creation ============ */

ason_value* ason_value_null(void) {
    ason_value* v = (ason_value*)calloc(1, sizeof(ason_value));
    if (v) v->type = ASON_TYPE_NULL;
    return v;
}

ason_value* ason_value_bool(bool b) {
    ason_value* v = (ason_value*)calloc(1, sizeof(ason_value));
    if (v) {
        v->type = ASON_TYPE_BOOL;
        v->data.boolean = b;
    }
    return v;
}

ason_value* ason_value_integer(int64_t n) {
    ason_value* v = (ason_value*)calloc(1, sizeof(ason_value));
    if (v) {
        v->type = ASON_TYPE_INTEGER;
        v->data.integer = n;
    }
    return v;
}

ason_value* ason_value_float(double f) {
    ason_value* v = (ason_value*)calloc(1, sizeof(ason_value));
    if (v) {
        v->type = ASON_TYPE_FLOAT;
        v->data.floating = f;
    }
    return v;
}

ason_value* ason_value_string(const char* s) {
    if (!s) return NULL;
    return ason_value_string_n(s, strlen(s));
}

ason_value* ason_value_string_n(const char* s, size_t len) {
    if (!s) return NULL;
    ason_value* v = (ason_value*)calloc(1, sizeof(ason_value));
    if (!v) return NULL;
    
    v->type = ASON_TYPE_STRING;
    v->data.string.data = (char*)malloc(len + 1);
    if (!v->data.string.data) {
        free(v);
        return NULL;
    }
    memcpy(v->data.string.data, s, len);
    v->data.string.data[len] = '\0';
    v->data.string.len = len;
    return v;
}

ason_value* ason_value_array(void) {
    ason_value* v = (ason_value*)calloc(1, sizeof(ason_value));
    if (v) {
        v->type = ASON_TYPE_ARRAY;
        v->data.array.items = NULL;
        v->data.array.len = 0;
        v->data.array.capacity = 0;
    }
    return v;
}

ason_value* ason_value_object(void) {
    ason_value* v = (ason_value*)calloc(1, sizeof(ason_value));
    if (v) {
        v->type = ASON_TYPE_OBJECT;
        v->data.object.pairs = NULL;
        v->data.object.len = 0;
        v->data.object.capacity = 0;
    }
    return v;
}

void ason_value_free(ason_value* v) {
    if (!v) return;
    
    switch (v->type) {
        case ASON_TYPE_STRING:
            free(v->data.string.data);
            break;
        case ASON_TYPE_ARRAY:
            for (size_t i = 0; i < v->data.array.len; i++) {
                ason_value_free(v->data.array.items[i]);
            }
            free(v->data.array.items);
            break;
        case ASON_TYPE_OBJECT:
            for (size_t i = 0; i < v->data.object.len; i++) {
                free(v->data.object.pairs[i].key);
                ason_value_free(v->data.object.pairs[i].value);
            }
            free(v->data.object.pairs);
            break;
        default:
            break;
    }
    free(v);
}

ason_value* ason_value_clone(const ason_value* v) {
    if (!v) return NULL;
    
    switch (v->type) {
        case ASON_TYPE_NULL:
            return ason_value_null();
        case ASON_TYPE_BOOL:
            return ason_value_bool(v->data.boolean);
        case ASON_TYPE_INTEGER:
            return ason_value_integer(v->data.integer);
        case ASON_TYPE_FLOAT:
            return ason_value_float(v->data.floating);
        case ASON_TYPE_STRING:
            return ason_value_string_n(v->data.string.data, v->data.string.len);
        case ASON_TYPE_ARRAY: {
            ason_value* arr = ason_value_array();
            if (!arr) return NULL;
            for (size_t i = 0; i < v->data.array.len; i++) {
                ason_value* item = ason_value_clone(v->data.array.items[i]);
                if (!item || ason_array_push(arr, item) != ASON_OK) {
                    ason_value_free(arr);
                    return NULL;
                }
            }
            return arr;
        }
        case ASON_TYPE_OBJECT: {
            ason_value* obj = ason_value_object();
            if (!obj) return NULL;
            for (size_t i = 0; i < v->data.object.len; i++) {
                ason_value* val = ason_value_clone(v->data.object.pairs[i].value);
                if (!val || ason_object_set(obj, v->data.object.pairs[i].key, val) != ASON_OK) {
                    ason_value_free(obj);
                    return NULL;
                }
            }
            return obj;
        }
    }
    return NULL;
}

/* ============ Type Checking ============ */

ason_type ason_value_type(const ason_value* v) {
    return v ? v->type : ASON_TYPE_NULL;
}

bool ason_value_is_null(const ason_value* v) {
    return v && v->type == ASON_TYPE_NULL;
}

bool ason_value_is_bool(const ason_value* v) {
    return v && v->type == ASON_TYPE_BOOL;
}

bool ason_value_is_integer(const ason_value* v) {
    return v && v->type == ASON_TYPE_INTEGER;
}

bool ason_value_is_float(const ason_value* v) {
    return v && v->type == ASON_TYPE_FLOAT;
}

bool ason_value_is_string(const ason_value* v) {
    return v && v->type == ASON_TYPE_STRING;
}

bool ason_value_is_array(const ason_value* v) {
    return v && v->type == ASON_TYPE_ARRAY;
}

bool ason_value_is_object(const ason_value* v) {
    return v && v->type == ASON_TYPE_OBJECT;
}

/* ============ Value Access ============ */

bool ason_value_get_bool(const ason_value* v) {
    if (v && v->type == ASON_TYPE_BOOL) return v->data.boolean;
    return false;
}

int64_t ason_value_get_integer(const ason_value* v) {
    if (v && v->type == ASON_TYPE_INTEGER) return v->data.integer;
    return 0;
}

double ason_value_get_float(const ason_value* v) {
    if (!v) return 0.0;
    if (v->type == ASON_TYPE_FLOAT) return v->data.floating;
    if (v->type == ASON_TYPE_INTEGER) return (double)v->data.integer;
    return 0.0;
}

const char* ason_value_get_string(const ason_value* v) {
    if (v && v->type == ASON_TYPE_STRING) return v->data.string.data;
    return NULL;
}

size_t ason_value_get_string_len(const ason_value* v) {
    if (v && v->type == ASON_TYPE_STRING) return v->data.string.len;
    return 0;
}

/* ============ Array Operations ============ */

size_t ason_array_len(const ason_value* arr) {
    if (arr && arr->type == ASON_TYPE_ARRAY) return arr->data.array.len;
    return 0;
}

ason_value* ason_array_get(const ason_value* arr, size_t index) {
    if (!arr || arr->type != ASON_TYPE_ARRAY) return NULL;
    if (index >= arr->data.array.len) return NULL;
    return arr->data.array.items[index];
}

ason_error ason_array_push(ason_value* arr, ason_value* item) {
    if (!arr || arr->type != ASON_TYPE_ARRAY) return ASON_ERROR_TYPE_MISMATCH;
    if (!item) return ASON_ERROR_NULL_POINTER;

    ason_array* a = &arr->data.array;
    if (a->len >= a->capacity) {
        size_t new_cap = a->capacity == 0 ? 4 : a->capacity * 2;
        ason_value** new_items = (ason_value**)realloc(a->items, new_cap * sizeof(ason_value*));
        if (!new_items) return ASON_ERROR_OUT_OF_MEMORY;
        a->items = new_items;
        a->capacity = new_cap;
    }
    a->items[a->len++] = item;
    return ASON_OK;
}

/* ============ Object Operations ============ */

size_t ason_object_len(const ason_value* obj) {
    if (obj && obj->type == ASON_TYPE_OBJECT) return obj->data.object.len;
    return 0;
}

ason_value* ason_object_get(const ason_value* obj, const char* key) {
    if (!obj || obj->type != ASON_TYPE_OBJECT || !key) return NULL;
    for (size_t i = 0; i < obj->data.object.len; i++) {
        if (strcmp(obj->data.object.pairs[i].key, key) == 0) {
            return obj->data.object.pairs[i].value;
        }
    }
    return NULL;
}

ason_error ason_object_set(ason_value* obj, const char* key, ason_value* value) {
    if (!obj || obj->type != ASON_TYPE_OBJECT) return ASON_ERROR_TYPE_MISMATCH;
    if (!key || !value) return ASON_ERROR_NULL_POINTER;

    ason_object* o = &obj->data.object;

    /* Check if key exists */
    for (size_t i = 0; i < o->len; i++) {
        if (strcmp(o->pairs[i].key, key) == 0) {
            ason_value_free(o->pairs[i].value);
            o->pairs[i].value = value;
            return ASON_OK;
        }
    }

    /* Add new key */
    if (o->len >= o->capacity) {
        size_t new_cap = o->capacity == 0 ? 4 : o->capacity * 2;
        ason_kv* new_pairs = (ason_kv*)realloc(o->pairs, new_cap * sizeof(ason_kv));
        if (!new_pairs) return ASON_ERROR_OUT_OF_MEMORY;
        o->pairs = new_pairs;
        o->capacity = new_cap;
    }

    o->pairs[o->len].key = strdup(key);
    if (!o->pairs[o->len].key) return ASON_ERROR_OUT_OF_MEMORY;
    o->pairs[o->len].value = value;
    o->len++;
    return ASON_OK;
}

const char* ason_object_key_at(const ason_value* obj, size_t index) {
    if (!obj || obj->type != ASON_TYPE_OBJECT) return NULL;
    if (index >= obj->data.object.len) return NULL;
    return obj->data.object.pairs[index].key;
}

ason_value* ason_object_value_at(const ason_value* obj, size_t index) {
    if (!obj || obj->type != ASON_TYPE_OBJECT) return NULL;
    if (index >= obj->data.object.len) return NULL;
    return obj->data.object.pairs[index].value;
}

/* ============ Error Handling ============ */

const char* ason_error_str(ason_error err) {
    switch (err) {
        case ASON_OK: return "OK";
        case ASON_ERROR_NULL_POINTER: return "Null pointer";
        case ASON_ERROR_OUT_OF_MEMORY: return "Out of memory";
        case ASON_ERROR_TYPE_MISMATCH: return "Type mismatch";
        case ASON_ERROR_INDEX_OUT_OF_BOUNDS: return "Index out of bounds";
        case ASON_ERROR_KEY_NOT_FOUND: return "Key not found";
        case ASON_ERROR_PARSE_ERROR: return "Parse error";
        case ASON_ERROR_UNEXPECTED_TOKEN: return "Unexpected token";
        case ASON_ERROR_UNCLOSED_STRING: return "Unclosed string";
        case ASON_ERROR_UNCLOSED_COMMENT: return "Unclosed comment";
        case ASON_ERROR_INVALID_NUMBER: return "Invalid number";
        default: return "Unknown error";
    }
}

