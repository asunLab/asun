#include "ason_serializer.h"
#include <stdlib.h>
#include <string.h>
#include <stdio.h>
#include <ctype.h>

/* Dynamic string buffer */
typedef struct {
    char* data;
    size_t len;
    size_t capacity;
} string_buf;

static void buf_init(string_buf* buf) {
    buf->data = NULL;
    buf->len = 0;
    buf->capacity = 0;
}

static void buf_free(string_buf* buf) {
    free(buf->data);
    buf->data = NULL;
    buf->len = 0;
    buf->capacity = 0;
}

static bool buf_grow(string_buf* buf, size_t needed) {
    if (buf->len + needed < buf->capacity) return true;
    
    size_t new_cap = buf->capacity == 0 ? 64 : buf->capacity * 2;
    while (new_cap < buf->len + needed) new_cap *= 2;
    
    char* new_data = (char*)realloc(buf->data, new_cap);
    if (!new_data) return false;
    
    buf->data = new_data;
    buf->capacity = new_cap;
    return true;
}

static bool buf_append(string_buf* buf, const char* s, size_t len) {
    if (!buf_grow(buf, len + 1)) return false;
    memcpy(buf->data + buf->len, s, len);
    buf->len += len;
    buf->data[buf->len] = '\0';
    return true;
}

static bool buf_append_str(string_buf* buf, const char* s) {
    return buf_append(buf, s, strlen(s));
}

static bool buf_append_char(string_buf* buf, char c) {
    return buf_append(buf, &c, 1);
}

/* Check if string needs quoting */
static bool needs_quotes(const char* s, size_t len) {
    if (len == 0) return true;
    
    /* Check for special characters */
    for (size_t i = 0; i < len; i++) {
        char c = s[i];
        if (c == '(' || c == ')' || c == '[' || c == ']' || 
            c == '{' || c == '}' || c == ',' || c == ':' ||
            c == '"' || c == '\'' || c == ' ' || c == '\t' ||
            c == '\n' || c == '\r') {
            return true;
        }
    }
    
    /* Check if it looks like a number */
    bool is_number = true;
    size_t i = 0;
    if (s[0] == '-') i++;
    
    for (; i < len; i++) {
        if (!isdigit((unsigned char)s[i]) && s[i] != '.') {
            is_number = false;
            break;
        }
    }
    if (is_number && i > 0) return true;
    
    /* Check for keywords */
    if (len == 4 && strncmp(s, "true", 4) == 0) return true;
    if (len == 5 && strncmp(s, "false", 5) == 0) return true;
    if (len == 4 && strncmp(s, "null", 4) == 0) return true;
    
    return false;
}

/* Serialize a string */
static bool serialize_string(string_buf* buf, const char* s, size_t len) {
    if (!needs_quotes(s, len)) {
        return buf_append(buf, s, len);
    }
    
    /* Use quotes */
    if (!buf_append_char(buf, '"')) return false;
    
    for (size_t i = 0; i < len; i++) {
        char c = s[i];
        switch (c) {
            case '"':  if (!buf_append_str(buf, "\\\"")) return false; break;
            case '\\': if (!buf_append_str(buf, "\\\\")) return false; break;
            case '\n': if (!buf_append_str(buf, "\\n")) return false; break;
            case '\r': if (!buf_append_str(buf, "\\r")) return false; break;
            case '\t': if (!buf_append_str(buf, "\\t")) return false; break;
            default:   if (!buf_append_char(buf, c)) return false; break;
        }
    }
    
    return buf_append_char(buf, '"');
}

/* Forward declaration */
static bool serialize_value(string_buf* buf, const ason_value* v, int indent, bool pretty);

static bool serialize_array(string_buf* buf, const ason_value* arr, int indent, bool pretty) {
    if (!buf_append_char(buf, '[')) return false;
    
    size_t len = ason_array_len(arr);
    for (size_t i = 0; i < len; i++) {
        if (i > 0) {
            if (!buf_append_char(buf, ',')) return false;
        }
        if (!serialize_value(buf, ason_array_get(arr, i), indent, pretty)) return false;
    }
    
    return buf_append_char(buf, ']');
}

static bool serialize_object(string_buf* buf, const ason_value* obj, int indent, bool pretty) {
    if (!buf_append_char(buf, '(')) return false;
    
    size_t len = ason_object_len(obj);
    for (size_t i = 0; i < len; i++) {
        if (i > 0) {
            if (!buf_append_char(buf, ',')) return false;
        }
        
        ason_value* val = ason_object_value_at(obj, i);
        if (!serialize_value(buf, val, indent, pretty)) return false;
    }
    
    return buf_append_char(buf, ')');
}

static bool serialize_value(string_buf* buf, const ason_value* v, int indent, bool pretty) {
    (void)indent; (void)pretty; /* TODO: implement pretty printing */
    
    if (!v) return buf_append_str(buf, "null");
    
    switch (v->type) {
        case ASON_TYPE_NULL:
            /* Empty for null in tuple context */
            return true;
            
        case ASON_TYPE_BOOL:
            return buf_append_str(buf, v->data.boolean ? "true" : "false");
            
        case ASON_TYPE_INTEGER: {
            char num[32];
            snprintf(num, sizeof(num), "%lld", (long long)v->data.integer);
            return buf_append_str(buf, num);
        }

        case ASON_TYPE_FLOAT: {
            char num[64];
            snprintf(num, sizeof(num), "%g", v->data.floating);
            return buf_append_str(buf, num);
        }

        case ASON_TYPE_STRING:
            return serialize_string(buf, v->data.string.data, v->data.string.len);

        case ASON_TYPE_ARRAY:
            return serialize_array(buf, v, indent, pretty);

        case ASON_TYPE_OBJECT:
            return serialize_object(buf, v, indent, pretty);
    }

    return false;
}

char* ason_to_string(const ason_value* value) {
    if (!value) return NULL;

    string_buf buf;
    buf_init(&buf);

    if (!serialize_value(&buf, value, 0, false)) {
        buf_free(&buf);
        return NULL;
    }

    return buf.data;
}

char* ason_to_string_pretty(const ason_value* value) {
    if (!value) return NULL;

    string_buf buf;
    buf_init(&buf);

    if (!serialize_value(&buf, value, 0, true)) {
        buf_free(&buf);
        return NULL;
    }

    return buf.data;
}

