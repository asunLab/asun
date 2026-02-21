/*
 * ASON - A Schema-Oriented Notation (C Implementation)
 * Field dump/load function implementations.
 */

#include "ason.h"

/* ============================================================================
 * Dump functions
 * ============================================================================ */

void ason_dump_bool(ason_buf_t* buf, const void* base, size_t offset) {
    bool v = *(const bool*)((const char*)base + offset);
    ason_buf_appends(buf, v ? "true" : "false");
}

void ason_dump_i8(ason_buf_t* buf, const void* base, size_t offset) {
    ason_buf_append_i64(buf, (int64_t)*(const int8_t*)((const char*)base + offset));
}

void ason_dump_i16(ason_buf_t* buf, const void* base, size_t offset) {
    ason_buf_append_i64(buf, (int64_t)*(const int16_t*)((const char*)base + offset));
}

void ason_dump_i32(ason_buf_t* buf, const void* base, size_t offset) {
    ason_buf_append_i64(buf, (int64_t)*(const int32_t*)((const char*)base + offset));
}

void ason_dump_i64(ason_buf_t* buf, const void* base, size_t offset) {
    ason_buf_append_i64(buf, *(const int64_t*)((const char*)base + offset));
}

void ason_dump_u8(ason_buf_t* buf, const void* base, size_t offset) {
    ason_buf_append_u64(buf, (uint64_t)*(const uint8_t*)((const char*)base + offset));
}

void ason_dump_u16(ason_buf_t* buf, const void* base, size_t offset) {
    ason_buf_append_u64(buf, (uint64_t)*(const uint16_t*)((const char*)base + offset));
}

void ason_dump_u32(ason_buf_t* buf, const void* base, size_t offset) {
    ason_buf_append_u64(buf, (uint64_t)*(const uint32_t*)((const char*)base + offset));
}

void ason_dump_u64(ason_buf_t* buf, const void* base, size_t offset) {
    ason_buf_append_u64(buf, *(const uint64_t*)((const char*)base + offset));
}

void ason_dump_f32(ason_buf_t* buf, const void* base, size_t offset) {
    ason_buf_append_f64(buf, (double)*(const float*)((const char*)base + offset));
}

void ason_dump_f64(ason_buf_t* buf, const void* base, size_t offset) {
    ason_buf_append_f64(buf, *(const double*)((const char*)base + offset));
}

void ason_dump_char(ason_buf_t* buf, const void* base, size_t offset) {
    char c = *(const char*)((const char*)base + offset);
    if (c == '\0') { return; }
    char s[2] = {c, '\0'};
    ason_buf_append_str(buf, s, 1);
}

void ason_dump_str(ason_buf_t* buf, const void* base, size_t offset) {
    const ason_string_t* s = (const ason_string_t*)((const char*)base + offset);
    if (!s->data || s->len == 0) { return; }
    ason_buf_append_str(buf, s->data, s->len);
}

void ason_dump_opt_i64(ason_buf_t* buf, const void* base, size_t offset) {
    const ason_opt_i64* opt = (const ason_opt_i64*)((const char*)base + offset);
    if (opt->has_value) {
        ason_buf_append_i64(buf, opt->value);
    }
    /* Empty = no output (will show as ,, in tuple) */
}

void ason_dump_opt_str(ason_buf_t* buf, const void* base, size_t offset) {
    const ason_opt_str* opt = (const ason_opt_str*)((const char*)base + offset);
    if (opt->has_value && opt->value.data) {
        ason_buf_append_str(buf, opt->value.data, opt->value.len);
    }
}

void ason_dump_opt_f64(ason_buf_t* buf, const void* base, size_t offset) {
    const ason_opt_f64* opt = (const ason_opt_f64*)((const char*)base + offset);
    if (opt->has_value) {
        ason_buf_append_f64(buf, opt->value);
    }
}

void ason_dump_vec_i64(ason_buf_t* buf, const void* base, size_t offset) {
    const ason_vec_i64* v = (const ason_vec_i64*)((const char*)base + offset);
    ason_buf_push(buf, '[');
    for (size_t i = 0; i < v->len; i++) {
        if (i > 0) ason_buf_push(buf, ',');
        ason_buf_append_i64(buf, v->data[i]);
    }
    ason_buf_push(buf, ']');
}

void ason_dump_vec_u64(ason_buf_t* buf, const void* base, size_t offset) {
    const ason_vec_u64* v = (const ason_vec_u64*)((const char*)base + offset);
    ason_buf_push(buf, '[');
    for (size_t i = 0; i < v->len; i++) {
        if (i > 0) ason_buf_push(buf, ',');
        ason_buf_append_u64(buf, v->data[i]);
    }
    ason_buf_push(buf, ']');
}

void ason_dump_vec_f64(ason_buf_t* buf, const void* base, size_t offset) {
    const ason_vec_f64* v = (const ason_vec_f64*)((const char*)base + offset);
    ason_buf_push(buf, '[');
    for (size_t i = 0; i < v->len; i++) {
        if (i > 0) ason_buf_push(buf, ',');
        ason_buf_append_f64(buf, v->data[i]);
    }
    ason_buf_push(buf, ']');
}

void ason_dump_vec_str(ason_buf_t* buf, const void* base, size_t offset) {
    const ason_vec_str* v = (const ason_vec_str*)((const char*)base + offset);
    ason_buf_push(buf, '[');
    for (size_t i = 0; i < v->len; i++) {
        if (i > 0) ason_buf_push(buf, ',');
        ason_buf_append_str(buf, v->data[i].data, v->data[i].len);
    }
    ason_buf_push(buf, ']');
}

void ason_dump_vec_bool(ason_buf_t* buf, const void* base, size_t offset) {
    const ason_vec_bool* v = (const ason_vec_bool*)((const char*)base + offset);
    ason_buf_push(buf, '[');
    for (size_t i = 0; i < v->len; i++) {
        if (i > 0) ason_buf_push(buf, ',');
        ason_buf_appends(buf, v->data[i] ? "true" : "false");
    }
    ason_buf_push(buf, ']');
}

void ason_dump_vec_vec_i64(ason_buf_t* buf, const void* base, size_t offset) {
    const ason_vec_vec_i64* v = (const ason_vec_vec_i64*)((const char*)base + offset);
    ason_buf_push(buf, '[');
    for (size_t i = 0; i < v->len; i++) {
        if (i > 0) ason_buf_push(buf, ',');
        ason_buf_push(buf, '[');
        for (size_t j = 0; j < v->data[i].len; j++) {
            if (j > 0) ason_buf_push(buf, ',');
            ason_buf_append_i64(buf, v->data[i].data[j]);
        }
        ason_buf_push(buf, ']');
    }
    ason_buf_push(buf, ']');
}

void ason_dump_map_si(ason_buf_t* buf, const void* base, size_t offset) {
    const ason_map_si* m = (const ason_map_si*)((const char*)base + offset);
    ason_buf_push(buf, '[');
    for (size_t i = 0; i < m->len; i++) {
        if (i > 0) ason_buf_push(buf, ',');
        ason_buf_push(buf, '(');
        ason_buf_append_str(buf, m->data[i].key.data, m->data[i].key.len);
        ason_buf_push(buf, ',');
        ason_buf_append_i64(buf, m->data[i].val);
        ason_buf_push(buf, ')');
    }
    ason_buf_push(buf, ']');
}

void ason_dump_map_ss(ason_buf_t* buf, const void* base, size_t offset) {
    const ason_map_ss* m = (const ason_map_ss*)((const char*)base + offset);
    ason_buf_push(buf, '[');
    for (size_t i = 0; i < m->len; i++) {
        if (i > 0) ason_buf_push(buf, ',');
        ason_buf_push(buf, '(');
        ason_buf_append_str(buf, m->data[i].key.data, m->data[i].key.len);
        ason_buf_push(buf, ',');
        ason_buf_append_str(buf, m->data[i].val.data, m->data[i].val.len);
        ason_buf_push(buf, ')');
    }
    ason_buf_push(buf, ']');
}

/* ============================================================================
 * Load functions
 * ============================================================================ */

static ason_err_t load_i64_raw(const char** pos, const char* end, int64_t* out) {
    ason_skip_ws(pos, end);
    bool neg = false;
    if (*pos < end && **pos == '-') { neg = true; (*pos)++; }
    uint64_t val = 0;
    int digits = 0;
    while (*pos < end && **pos >= '0' && **pos <= '9') {
        val = val * 10 + (**pos - '0');
        (*pos)++; digits++;
    }
    if (digits == 0) return ASON_ERR_INVALID_NUMBER;
    *out = neg ? -(int64_t)val : (int64_t)val;
    return ASON_OK;
}

static ason_err_t load_u64_raw(const char** pos, const char* end, uint64_t* out) {
    ason_skip_ws(pos, end);
    uint64_t val = 0;
    int digits = 0;
    while (*pos < end && **pos >= '0' && **pos <= '9') {
        val = val * 10 + (**pos - '0');
        (*pos)++; digits++;
    }
    if (digits == 0) return ASON_ERR_INVALID_NUMBER;
    *out = val;
    return ASON_OK;
}

static ason_err_t load_f64_raw(const char** pos, const char* end, double* out) {
    ason_skip_ws(pos, end);
    if (*pos >= end) return ASON_ERR_INVALID_NUMBER;
    char* endptr = NULL;
    *out = strtod(*pos, &endptr);
    if (endptr == *pos) return ASON_ERR_INVALID_NUMBER;
    *pos = endptr;
    return ASON_OK;
}

ason_err_t ason_load_bool(const char** pos, const char* end, void* base, size_t offset) {
    ason_skip_ws(pos, end);
    bool* out = (bool*)((char*)base + offset);
    if (*pos + 4 <= end && memcmp(*pos, "true", 4) == 0) {
        *out = true; *pos += 4; return ASON_OK;
    }
    if (*pos + 5 <= end && memcmp(*pos, "false", 5) == 0) {
        *out = false; *pos += 5; return ASON_OK;
    }
    return ASON_ERR_SYNTAX;
}

ason_err_t ason_load_i8(const char** pos, const char* end, void* base, size_t offset) {
    int64_t v; ason_err_t e = load_i64_raw(pos, end, &v);
    if (e == ASON_OK) *(int8_t*)((char*)base + offset) = (int8_t)v;
    return e;
}

ason_err_t ason_load_i16(const char** pos, const char* end, void* base, size_t offset) {
    int64_t v; ason_err_t e = load_i64_raw(pos, end, &v);
    if (e == ASON_OK) *(int16_t*)((char*)base + offset) = (int16_t)v;
    return e;
}

ason_err_t ason_load_i32(const char** pos, const char* end, void* base, size_t offset) {
    int64_t v; ason_err_t e = load_i64_raw(pos, end, &v);
    if (e == ASON_OK) *(int32_t*)((char*)base + offset) = (int32_t)v;
    return e;
}

ason_err_t ason_load_i64(const char** pos, const char* end, void* base, size_t offset) {
    return load_i64_raw(pos, end, (int64_t*)((char*)base + offset));
}

ason_err_t ason_load_u8(const char** pos, const char* end, void* base, size_t offset) {
    uint64_t v; ason_err_t e = load_u64_raw(pos, end, &v);
    if (e == ASON_OK) *(uint8_t*)((char*)base + offset) = (uint8_t)v;
    return e;
}

ason_err_t ason_load_u16(const char** pos, const char* end, void* base, size_t offset) {
    uint64_t v; ason_err_t e = load_u64_raw(pos, end, &v);
    if (e == ASON_OK) *(uint16_t*)((char*)base + offset) = (uint16_t)v;
    return e;
}

ason_err_t ason_load_u32(const char** pos, const char* end, void* base, size_t offset) {
    uint64_t v; ason_err_t e = load_u64_raw(pos, end, &v);
    if (e == ASON_OK) *(uint32_t*)((char*)base + offset) = (uint32_t)v;
    return e;
}

ason_err_t ason_load_u64(const char** pos, const char* end, void* base, size_t offset) {
    return load_u64_raw(pos, end, (uint64_t*)((char*)base + offset));
}

ason_err_t ason_load_f32(const char** pos, const char* end, void* base, size_t offset) {
    double v; ason_err_t e = load_f64_raw(pos, end, &v);
    if (e == ASON_OK) *(float*)((char*)base + offset) = (float)v;
    return e;
}

ason_err_t ason_load_f64(const char** pos, const char* end, void* base, size_t offset) {
    return load_f64_raw(pos, end, (double*)((char*)base + offset));
}

ason_err_t ason_load_char(const char** pos, const char* end, void* base, size_t offset) {
    char* out_str = NULL;
    size_t out_len = 0;
    bool allocated = false;
    ason_err_t err = ason_parse_string_value(pos, end, &out_str, &out_len, &allocated);
    if (err != ASON_OK) return err;
    *(char*)((char*)base + offset) = (out_len > 0) ? out_str[0] : '\0';
    if (allocated) free(out_str);
    return ASON_OK;
}

ason_err_t ason_load_str(const char** pos, const char* end, void* base, size_t offset) {
    ason_skip_ws(pos, end);
    ason_string_t* s = (ason_string_t*)((char*)base + offset);
    if (ason_at_value_end(*pos, end)) {
        s->data = NULL; s->len = 0;
        return ASON_OK;
    }
    char* out_str = NULL;
    size_t out_len = 0;
    bool allocated = false;
    ason_err_t err = ason_parse_string_value(pos, end, &out_str, &out_len, &allocated);
    if (err != ASON_OK) return err;
    if (allocated) {
        /* Take ownership of already-allocated buffer */
        s->data = out_str;
        s->len = out_len;
    } else {
        /* Make an owned copy of zero-copy result */
        s->data = (char*)malloc(out_len + 1);
        memcpy(s->data, out_str, out_len);
        s->data[out_len] = '\0';
        s->len = out_len;
    }
    return ASON_OK;
}

ason_err_t ason_load_opt_i64(const char** pos, const char* end, void* base, size_t offset) {
    ason_skip_ws(pos, end);
    ason_opt_i64* opt = (ason_opt_i64*)((char*)base + offset);
    if (ason_at_value_end(*pos, end)) {
        opt->has_value = false;
        return ASON_OK;
    }
    opt->has_value = true;
    return load_i64_raw(pos, end, &opt->value);
}

ason_err_t ason_load_opt_str(const char** pos, const char* end, void* base, size_t offset) {
    ason_skip_ws(pos, end);
    ason_opt_str* opt = (ason_opt_str*)((char*)base + offset);
    if (ason_at_value_end(*pos, end)) {
        opt->has_value = false;
        opt->value.data = NULL; opt->value.len = 0;
        return ASON_OK;
    }
    opt->has_value = true;
    char* out_str = NULL; size_t out_len = 0; bool allocated = false;
    ason_err_t err = ason_parse_string_value(pos, end, &out_str, &out_len, &allocated);
    if (err != ASON_OK) return err;
    if (allocated) {
        opt->value.data = out_str;
        opt->value.len = out_len;
    } else {
        opt->value.data = (char*)malloc(out_len + 1);
        memcpy(opt->value.data, out_str, out_len);
        opt->value.data[out_len] = '\0';
        opt->value.len = out_len;
    }
    return ASON_OK;
}

ason_err_t ason_load_opt_f64(const char** pos, const char* end, void* base, size_t offset) {
    ason_skip_ws(pos, end);
    ason_opt_f64* opt = (ason_opt_f64*)((char*)base + offset);
    if (ason_at_value_end(*pos, end)) {
        opt->has_value = false;
        return ASON_OK;
    }
    opt->has_value = true;
    return load_f64_raw(pos, end, &opt->value);
}

ason_err_t ason_load_vec_i64(const char** pos, const char* end, void* base, size_t offset) {
    ason_skip_ws(pos, end);
    if (*pos >= end || **pos != '[') return ASON_ERR_SYNTAX;
    (*pos)++;
    ason_vec_i64* v = (ason_vec_i64*)((char*)base + offset);
    *v = ason_vec_i64_new();
    bool first = true;
    while (1) {
        ason_skip_ws(pos, end);
        if (*pos >= end || **pos == ']') { (*pos)++; break; }
        if (!first) {
            if (**pos == ',') { (*pos)++; ason_skip_ws(pos, end); if (*pos < end && **pos == ']') { (*pos)++; break; } }
            else break;
        }
        first = false;
        int64_t val;
        ason_err_t e = load_i64_raw(pos, end, &val);
        if (e != ASON_OK) return e;
        ason_vec_i64_push(v, val);
    }
    return ASON_OK;
}

ason_err_t ason_load_vec_u64(const char** pos, const char* end, void* base, size_t offset) {
    ason_skip_ws(pos, end);
    if (*pos >= end || **pos != '[') return ASON_ERR_SYNTAX;
    (*pos)++;
    ason_vec_u64* v = (ason_vec_u64*)((char*)base + offset);
    *v = ason_vec_u64_new();
    bool first = true;
    while (1) {
        ason_skip_ws(pos, end);
        if (*pos >= end || **pos == ']') { (*pos)++; break; }
        if (!first) {
            if (**pos == ',') { (*pos)++; ason_skip_ws(pos, end); if (*pos < end && **pos == ']') { (*pos)++; break; } }
            else break;
        }
        first = false;
        uint64_t val;
        ason_err_t e = load_u64_raw(pos, end, &val);
        if (e != ASON_OK) return e;
        ason_vec_u64_push(v, val);
    }
    return ASON_OK;
}

ason_err_t ason_load_vec_f64(const char** pos, const char* end, void* base, size_t offset) {
    ason_skip_ws(pos, end);
    if (*pos >= end || **pos != '[') return ASON_ERR_SYNTAX;
    (*pos)++;
    ason_vec_f64* v = (ason_vec_f64*)((char*)base + offset);
    *v = ason_vec_f64_new();
    bool first = true;
    while (1) {
        ason_skip_ws(pos, end);
        if (*pos >= end || **pos == ']') { (*pos)++; break; }
        if (!first) {
            if (**pos == ',') { (*pos)++; ason_skip_ws(pos, end); if (*pos < end && **pos == ']') { (*pos)++; break; } }
            else break;
        }
        first = false;
        double val;
        ason_err_t e = load_f64_raw(pos, end, &val);
        if (e != ASON_OK) return e;
        ason_vec_f64_push(v, val);
    }
    return ASON_OK;
}

ason_err_t ason_load_vec_str(const char** pos, const char* end, void* base, size_t offset) {
    ason_skip_ws(pos, end);
    if (*pos >= end || **pos != '[') return ASON_ERR_SYNTAX;
    (*pos)++;
    ason_vec_str* v = (ason_vec_str*)((char*)base + offset);
    *v = ason_vec_str_new();
    bool first = true;
    while (1) {
        ason_skip_ws(pos, end);
        if (*pos >= end || **pos == ']') { (*pos)++; break; }
        if (!first) {
            if (**pos == ',') { (*pos)++; ason_skip_ws(pos, end); if (*pos < end && **pos == ']') { (*pos)++; break; } }
            else break;
        }
        first = false;
        char* out_str = NULL; size_t out_len = 0; bool allocated = false;
        ason_err_t e = ason_parse_string_value(pos, end, &out_str, &out_len, &allocated);
        if (e != ASON_OK) return e;
        ason_string_t s;
        if (allocated) {
            s.data = out_str;
            s.len = out_len;
        } else {
            s.data = (char*)malloc(out_len + 1);
            memcpy(s.data, out_str, out_len);
            s.data[out_len] = '\0';
            s.len = out_len;
        }
        ason_vec_str_push(v, s);
    }
    return ASON_OK;
}

ason_err_t ason_load_vec_bool(const char** pos, const char* end, void* base, size_t offset) {
    ason_skip_ws(pos, end);
    if (*pos >= end || **pos != '[') return ASON_ERR_SYNTAX;
    (*pos)++;
    ason_vec_bool* v = (ason_vec_bool*)((char*)base + offset);
    *v = ason_vec_bool_new();
    bool first = true;
    while (1) {
        ason_skip_ws(pos, end);
        if (*pos >= end || **pos == ']') { (*pos)++; break; }
        if (!first) {
            if (**pos == ',') { (*pos)++; ason_skip_ws(pos, end); if (*pos < end && **pos == ']') { (*pos)++; break; } }
            else break;
        }
        first = false;
        if (*pos + 4 <= end && memcmp(*pos, "true", 4) == 0) {
            ason_vec_bool_push(v, true); *pos += 4;
        } else if (*pos + 5 <= end && memcmp(*pos, "false", 5) == 0) {
            ason_vec_bool_push(v, false); *pos += 5;
        } else return ASON_ERR_SYNTAX;
    }
    return ASON_OK;
}

ason_err_t ason_load_vec_vec_i64(const char** pos, const char* end, void* base, size_t offset) {
    ason_skip_ws(pos, end);
    if (*pos >= end || **pos != '[') return ASON_ERR_SYNTAX;
    (*pos)++;
    ason_vec_vec_i64* v = (ason_vec_vec_i64*)((char*)base + offset);
    *v = ason_vec_vec_i64_new();
    bool first = true;
    while (1) {
        ason_skip_ws(pos, end);
        if (*pos >= end || **pos == ']') { (*pos)++; break; }
        if (!first) {
            if (**pos == ',') { (*pos)++; ason_skip_ws(pos, end); if (*pos < end && **pos == ']') { (*pos)++; break; } }
            else break;
        }
        first = false;
        if (*pos >= end || **pos != '[') return ASON_ERR_SYNTAX;
        (*pos)++;
        ason_vec_i64 inner = ason_vec_i64_new();
        bool ifirst = true;
        while (1) {
            ason_skip_ws(pos, end);
            if (*pos >= end || **pos == ']') { (*pos)++; break; }
            if (!ifirst) {
                if (**pos == ',') { (*pos)++; ason_skip_ws(pos, end); if (*pos < end && **pos == ']') { (*pos)++; break; } }
                else break;
            }
            ifirst = false;
            int64_t val;
            ason_err_t e = load_i64_raw(pos, end, &val);
            if (e != ASON_OK) return e;
            ason_vec_i64_push(&inner, val);
        }
        ason_vec_vec_i64_push(v, inner);
    }
    return ASON_OK;
}

ason_err_t ason_load_map_si(const char** pos, const char* end, void* base, size_t offset) {
    ason_skip_ws(pos, end);
    if (*pos >= end || **pos != '[') return ASON_ERR_SYNTAX;
    (*pos)++;
    ason_map_si* m = (ason_map_si*)((char*)base + offset);
    *m = ason_map_si_new();
    bool first = true;
    while (1) {
        ason_skip_ws(pos, end);
        if (*pos >= end || **pos == ']') { (*pos)++; break; }
        if (!first) {
            if (**pos == ',') { (*pos)++; ason_skip_ws(pos, end); if (*pos < end && **pos == ']') { (*pos)++; break; } }
            else break;
        }
        first = false;
        if (*pos >= end || **pos != '(') return ASON_ERR_SYNTAX;
        (*pos)++;
        /* key */
        char* kstr = NULL; size_t klen = 0; bool kall = false;
        ason_err_t e = ason_parse_string_value(pos, end, &kstr, &klen, &kall);
        if (e != ASON_OK) return e;
        ason_skip_ws(pos, end);
        if (*pos < end && **pos == ',') (*pos)++;
        /* value */
        int64_t val;
        e = load_i64_raw(pos, end, &val);
        if (e != ASON_OK) return e;
        ason_skip_ws(pos, end);
        if (*pos < end && **pos == ')') (*pos)++;
        ason_map_si_entry_t entry;
        if (kall) {
            entry.key.data = kstr;
            entry.key.len = klen;
        } else {
            entry.key.data = (char*)malloc(klen + 1);
            memcpy(entry.key.data, kstr, klen);
            entry.key.data[klen] = '\0';
            entry.key.len = klen;
        }
        entry.val = val;
        ason_map_si_push(m, entry);
    }
    return ASON_OK;
}

ason_err_t ason_load_map_ss(const char** pos, const char* end, void* base, size_t offset) {
    ason_skip_ws(pos, end);
    if (*pos >= end || **pos != '[') return ASON_ERR_SYNTAX;
    (*pos)++;
    ason_map_ss* m = (ason_map_ss*)((char*)base + offset);
    *m = ason_map_ss_new();
    bool first = true;
    while (1) {
        ason_skip_ws(pos, end);
        if (*pos >= end || **pos == ']') { (*pos)++; break; }
        if (!first) {
            if (**pos == ',') { (*pos)++; ason_skip_ws(pos, end); if (*pos < end && **pos == ']') { (*pos)++; break; } }
            else break;
        }
        first = false;
        if (*pos >= end || **pos != '(') return ASON_ERR_SYNTAX;
        (*pos)++;
        /* key */
        char* kstr = NULL; size_t klen = 0; bool kall = false;
        ason_err_t e = ason_parse_string_value(pos, end, &kstr, &klen, &kall);
        if (e != ASON_OK) return e;
        ason_skip_ws(pos, end);
        if (*pos < end && **pos == ',') (*pos)++;
        /* value */
        char* vstr = NULL; size_t vlen = 0; bool vall = false;
        e = ason_parse_string_value(pos, end, &vstr, &vlen, &vall);
        if (e != ASON_OK) return e;
        ason_skip_ws(pos, end);
        if (*pos < end && **pos == ')') (*pos)++;
        ason_map_ss_entry_t entry;
        if (kall) {
            entry.key.data = kstr;
            entry.key.len = klen;
        } else {
            entry.key.data = (char*)malloc(klen + 1);
            memcpy(entry.key.data, kstr, klen); entry.key.data[klen] = '\0'; entry.key.len = klen;
        }
        if (vall) {
            entry.val.data = vstr;
            entry.val.len = vlen;
        } else {
            entry.val.data = (char*)malloc(vlen + 1);
            memcpy(entry.val.data, vstr, vlen); entry.val.data[vlen] = '\0'; entry.val.len = vlen;
        }
        ason_map_ss_push(m, entry);
    }
    return ASON_OK;
}

/* ============================================================================
 * Generic struct dump/load via descriptor
 * ============================================================================ */

void ason_dump_struct(ason_buf_t* buf, const void* obj, const ason_desc_t* desc) {
    ason_buf_push(buf, '(');
    for (int i = 0; i < desc->field_count; i++) {
        if (i > 0) ason_buf_push(buf, ',');
        const ason_field_t* f = &desc->fields[i];
        if (f->type == ASON_STRUCT && f->sub_desc) {
            if (f->dump_fn) {
                /* vec-of-struct or custom dump */
                f->dump_fn(buf, obj, f->offset);
            } else {
                ason_dump_struct(buf, (const char*)obj + f->offset,
                                (const ason_desc_t*)f->sub_desc);
            }
        } else {
            f->dump_fn(buf, obj, f->offset);
        }
    }
    ason_buf_push(buf, ')');
}

ason_err_t ason_load_struct(const char** pos, const char* end, void* obj, const ason_desc_t* desc) {
    ason_skip_ws(pos, end);

    /* If starts with '{', it has an inline schema */
    if (*pos < end && **pos == '{') {
        ason_schema_field_t schema[64];
        int schema_count = 0;
        ason_err_t err = ason_parse_schema(pos, end, schema, &schema_count, 64);
        if (err != ASON_OK) return err;
        ason_skip_ws(pos, end);
        if (*pos >= end || **pos != ':') return ASON_ERR_SYNTAX;
        (*pos)++;
        ason_skip_ws(pos, end);
        /* Build field map */
        int field_map[64];
        for (int i = 0; i < schema_count; i++) {
            field_map[i] = -1;
            for (int j = 0; j < desc->field_count; j++) {
                if (schema[i].len == desc->fields[j].name_len &&
                    memcmp(schema[i].name, desc->fields[j].name, schema[i].len) == 0) {
                    field_map[i] = j; break;
                }
            }
        }
        if (*pos >= end || **pos != '(') return ASON_ERR_SYNTAX;
        (*pos)++;
        for (int i = 0; i < schema_count; i++) {
            ason_skip_ws(pos, end);
            if (*pos < end && **pos == ')') break;
            if (i > 0) {
                if (**pos == ',') { (*pos)++; ason_skip_ws(pos, end); if (*pos < end && **pos == ')') break; }
                else if (**pos == ')') break;
                else return ASON_ERR_SYNTAX;
            }
            if (field_map[i] >= 0) {
                const ason_field_t* f = &desc->fields[field_map[i]];
                if (f->type == ASON_STRUCT && f->sub_desc) {
                    if (f->load_fn) {
                        err = f->load_fn(pos, end, obj, f->offset);
                    } else {
                        err = ason_load_struct(pos, end, (char*)obj + f->offset,
                                               (const ason_desc_t*)f->sub_desc);
                    }
                } else {
                    err = f->load_fn(pos, end, obj, f->offset);
                }
                if (err != ASON_OK) return err;
            } else {
                ason_skip_value(pos, end);
            }
        }
        ason_skip_ws(pos, end);
        if (*pos < end && **pos == ')') (*pos)++;
        return ASON_OK;
    }

    /* Positional tuple: (val1,val2,...) */
    if (*pos < end && **pos == '(') {
        (*pos)++;
        for (int i = 0; i < desc->field_count; i++) {
            ason_skip_ws(pos, end);
            if (*pos < end && **pos == ')') break;
            if (i > 0) {
                if (**pos == ',') { (*pos)++; ason_skip_ws(pos, end); if (*pos < end && **pos == ')') break; }
                else if (**pos == ')') break;
                else return ASON_ERR_SYNTAX;
            }
            const ason_field_t* f = &desc->fields[i];
            ason_err_t err;
            if (f->type == ASON_STRUCT && f->sub_desc) {
                if (f->load_fn) {
                    err = f->load_fn(pos, end, obj, f->offset);
                } else {
                    err = ason_load_struct(pos, end, (char*)obj + f->offset,
                                           (const ason_desc_t*)f->sub_desc);
                }
            } else {
                err = f->load_fn(pos, end, obj, f->offset);
            }
            if (err != ASON_OK) return err;
        }
        ason_skip_ws(pos, end);
        if (*pos < end && **pos == ')') (*pos)++;
        return ASON_OK;
    }

    return ASON_ERR_SYNTAX;
}

/* ===========================================================================
 * ASON Binary Format Implementation
 * Wire format: little-endian fixed-width, no field names, positional.
 * Strings: u32 length prefix + bytes (no null terminator).
 * Arrays: u32 count prefix + elements.
 * Boolean: 1 byte (0 or 1).
 * =========================================================================== */

/* ---- scalar dump ---- */
void ason_bin_dump_bool(ason_buf_t* buf, const void* base, size_t off) {
    bool v; memcpy(&v, (const char*)base + off, sizeof(v));
    ason_bin_write_u8(buf, v ? 1 : 0);
}
void ason_bin_dump_i8(ason_buf_t* buf, const void* base, size_t off) {
    int8_t v; memcpy(&v, (const char*)base + off, 1);
    ason_bin_write_u8(buf, (uint8_t)v);
}
void ason_bin_dump_i16(ason_buf_t* buf, const void* base, size_t off) {
    int16_t v; memcpy(&v, (const char*)base + off, 2);
    ason_bin_write_u16(buf, (uint16_t)v);
}
void ason_bin_dump_i32(ason_buf_t* buf, const void* base, size_t off) {
    int32_t v; memcpy(&v, (const char*)base + off, 4);
    ason_bin_write_u32(buf, (uint32_t)v);
}
void ason_bin_dump_i64(ason_buf_t* buf, const void* base, size_t off) {
    int64_t v; memcpy(&v, (const char*)base + off, 8);
    ason_bin_write_u64(buf, (uint64_t)v);
}
void ason_bin_dump_u8(ason_buf_t* buf, const void* base, size_t off) {
    uint8_t v; memcpy(&v, (const char*)base + off, 1);
    ason_bin_write_u8(buf, v);
}
void ason_bin_dump_u16(ason_buf_t* buf, const void* base, size_t off) {
    uint16_t v; memcpy(&v, (const char*)base + off, 2);
    ason_bin_write_u16(buf, v);
}
void ason_bin_dump_u32(ason_buf_t* buf, const void* base, size_t off) {
    uint32_t v; memcpy(&v, (const char*)base + off, 4);
    ason_bin_write_u32(buf, v);
}
void ason_bin_dump_u64(ason_buf_t* buf, const void* base, size_t off) {
    uint64_t v; memcpy(&v, (const char*)base + off, 8);
    ason_bin_write_u64(buf, v);
}
void ason_bin_dump_f32(ason_buf_t* buf, const void* base, size_t off) {
    float v; memcpy(&v, (const char*)base + off, 4);
    ason_bin_write_f32(buf, v);
}
void ason_bin_dump_f64(ason_buf_t* buf, const void* base, size_t off) {
    double v; memcpy(&v, (const char*)base + off, 8);
    ason_bin_write_f64(buf, v);
}
void ason_bin_dump_str(ason_buf_t* buf, const void* base, size_t off) {
    ason_string_t s; memcpy(&s, (const char*)base + off, sizeof(s));
    ason_bin_write_ason_string(buf, &s);
}

/* ---- vector dump ---- */
void ason_bin_dump_vec_i64(ason_buf_t* buf, const void* base, size_t off) {
    ason_vec_i64 v; memcpy(&v, (const char*)base + off, sizeof(v));
    ason_bin_write_u32(buf, (uint32_t)v.len);
    if (v.len) ason_buf_append(buf, (const char*)v.data, v.len * sizeof(int64_t));
}
void ason_bin_dump_vec_u64(ason_buf_t* buf, const void* base, size_t off) {
    ason_vec_u64 v; memcpy(&v, (const char*)base + off, sizeof(v));
    ason_bin_write_u32(buf, (uint32_t)v.len);
    if (v.len) ason_buf_append(buf, (const char*)v.data, v.len * sizeof(uint64_t));
}
void ason_bin_dump_vec_f64(ason_buf_t* buf, const void* base, size_t off) {
    ason_vec_f64 v; memcpy(&v, (const char*)base + off, sizeof(v));
    ason_bin_write_u32(buf, (uint32_t)v.len);
    if (v.len) ason_buf_append(buf, (const char*)v.data, v.len * sizeof(double));
}
void ason_bin_dump_vec_str(ason_buf_t* buf, const void* base, size_t off) {
    ason_vec_str v; memcpy(&v, (const char*)base + off, sizeof(v));
    ason_bin_write_u32(buf, (uint32_t)v.len);
    for (size_t i = 0; i < v.len; i++) ason_bin_write_ason_string(buf, &v.data[i]);
}
void ason_bin_dump_vec_bool(ason_buf_t* buf, const void* base, size_t off) {
    ason_vec_bool v; memcpy(&v, (const char*)base + off, sizeof(v));
    ason_bin_write_u32(buf, (uint32_t)v.len);
    for (size_t i = 0; i < v.len; i++) ason_bin_write_u8(buf, v.data[i] ? 1 : 0);
}

/* ---- scalar load ---- */
ason_err_t ason_bin_load_bool(const char** pos, const char* end, void* base, size_t off) {
    uint8_t v; ason_err_t e = ason_bin_read_u8(pos, end, &v);
    if (e) return e;
    bool b = v != 0; memcpy((char*)base + off, &b, sizeof(b)); return ASON_OK;
}
ason_err_t ason_bin_load_i8(const char** pos, const char* end, void* base, size_t off) {
    uint8_t v; ason_err_t e = ason_bin_read_u8(pos, end, &v);
    if (e) return e; memcpy((char*)base + off, &v, 1); return ASON_OK;
}
ason_err_t ason_bin_load_i16(const char** pos, const char* end, void* base, size_t off) {
    uint16_t v; ason_err_t e = ason_bin_read_u16(pos, end, &v);
    if (e) return e; memcpy((char*)base + off, &v, 2); return ASON_OK;
}
ason_err_t ason_bin_load_i32(const char** pos, const char* end, void* base, size_t off) {
    uint32_t v; ason_err_t e = ason_bin_read_u32(pos, end, &v);
    if (e) return e; memcpy((char*)base + off, &v, 4); return ASON_OK;
}
ason_err_t ason_bin_load_i64(const char** pos, const char* end, void* base, size_t off) {
    uint64_t v; ason_err_t e = ason_bin_read_u64(pos, end, &v);
    if (e) return e; memcpy((char*)base + off, &v, 8); return ASON_OK;
}
ason_err_t ason_bin_load_u8(const char** pos, const char* end, void* base, size_t off) {
    uint8_t v; ason_err_t e = ason_bin_read_u8(pos, end, &v);
    if (e) return e; memcpy((char*)base + off, &v, 1); return ASON_OK;
}
ason_err_t ason_bin_load_u16(const char** pos, const char* end, void* base, size_t off) {
    uint16_t v; ason_err_t e = ason_bin_read_u16(pos, end, &v);
    if (e) return e; memcpy((char*)base + off, &v, 2); return ASON_OK;
}
ason_err_t ason_bin_load_u32(const char** pos, const char* end, void* base, size_t off) {
    uint32_t v; ason_err_t e = ason_bin_read_u32(pos, end, &v);
    if (e) return e; memcpy((char*)base + off, &v, 4); return ASON_OK;
}
ason_err_t ason_bin_load_u64(const char** pos, const char* end, void* base, size_t off) {
    uint64_t v; ason_err_t e = ason_bin_read_u64(pos, end, &v);
    if (e) return e; memcpy((char*)base + off, &v, 8); return ASON_OK;
}
ason_err_t ason_bin_load_f32(const char** pos, const char* end, void* base, size_t off) {
    float v; ason_err_t e = ason_bin_read_f32(pos, end, &v);
    if (e) return e; memcpy((char*)base + off, &v, 4); return ASON_OK;
}
ason_err_t ason_bin_load_f64(const char** pos, const char* end, void* base, size_t off) {
    double v; ason_err_t e = ason_bin_read_f64(pos, end, &v);
    if (e) return e; memcpy((char*)base + off, &v, 8); return ASON_OK;
}
ason_err_t ason_bin_load_str(const char** pos, const char* end, void* base, size_t off) {
    uint32_t len;
    ason_err_t e = ason_bin_read_u32(pos, end, &len);
    if (e) return e;
    if ((size_t)(end - *pos) < len) return ASON_ERR_BUFFER_OVERFLOW;
    char* buf = (char*)malloc(len + 1);
    if (!buf) return ASON_ERR_ALLOC;
    if (len) memcpy(buf, *pos, len);
    buf[len] = '\0';
    *pos += len;
    ason_string_t s = {buf, len};
    memcpy((char*)base + off, &s, sizeof(s));
    return ASON_OK;
}

/* ---- vector load ---- */
ason_err_t ason_bin_load_vec_i64(const char** pos, const char* end, void* base, size_t off) {
    uint32_t n; ason_err_t e = ason_bin_read_u32(pos, end, &n); if (e) return e;
    if ((size_t)(end - *pos) < (size_t)n * 8) return ASON_ERR_BUFFER_OVERFLOW;
    int64_t* arr = (int64_t*)malloc((size_t)n * sizeof(int64_t));
    if (!arr && n) return ASON_ERR_ALLOC;
    memcpy(arr, *pos, (size_t)n * 8); *pos += (size_t)n * 8;
    ason_vec_i64 v = {arr, n, n}; memcpy((char*)base + off, &v, sizeof(v)); return ASON_OK;
}
ason_err_t ason_bin_load_vec_u64(const char** pos, const char* end, void* base, size_t off) {
    uint32_t n; ason_err_t e = ason_bin_read_u32(pos, end, &n); if (e) return e;
    if ((size_t)(end - *pos) < (size_t)n * 8) return ASON_ERR_BUFFER_OVERFLOW;
    uint64_t* arr = (uint64_t*)malloc((size_t)n * sizeof(uint64_t));
    if (!arr && n) return ASON_ERR_ALLOC;
    memcpy(arr, *pos, (size_t)n * 8); *pos += (size_t)n * 8;
    ason_vec_u64 v = {arr, n, n}; memcpy((char*)base + off, &v, sizeof(v)); return ASON_OK;
}
ason_err_t ason_bin_load_vec_f64(const char** pos, const char* end, void* base, size_t off) {
    uint32_t n; ason_err_t e = ason_bin_read_u32(pos, end, &n); if (e) return e;
    if ((size_t)(end - *pos) < (size_t)n * 8) return ASON_ERR_BUFFER_OVERFLOW;
    double* arr = (double*)malloc((size_t)n * sizeof(double));
    if (!arr && n) return ASON_ERR_ALLOC;
    memcpy(arr, *pos, (size_t)n * 8); *pos += (size_t)n * 8;
    ason_vec_f64 v = {arr, n, n}; memcpy((char*)base + off, &v, sizeof(v)); return ASON_OK;
}
ason_err_t ason_bin_load_vec_str(const char** pos, const char* end, void* base, size_t off) {
    uint32_t n; ason_err_t e = ason_bin_read_u32(pos, end, &n); if (e) return e;
    ason_string_t* arr = (ason_string_t*)malloc((size_t)n * sizeof(ason_string_t));
    if (!arr && n) return ASON_ERR_ALLOC;
    for (uint32_t i = 0; i < n; i++) {
        uint32_t slen;
        e = ason_bin_read_u32(pos, end, &slen); if (e) { free(arr); return e; }
        if ((size_t)(end - *pos) < slen) { free(arr); return ASON_ERR_BUFFER_OVERFLOW; }
        char* sbuf = (char*)malloc(slen + 1);
        if (!sbuf && slen) { free(arr); return ASON_ERR_ALLOC; }
        if (slen) memcpy(sbuf, *pos, slen);
        if (sbuf) sbuf[slen] = '\0';
        *pos += slen;
        arr[i].data = sbuf ? sbuf : (char*)"";
        arr[i].len  = slen;
    }
    ason_vec_str v = {arr, n, n}; memcpy((char*)base + off, &v, sizeof(v)); return ASON_OK;
}
ason_err_t ason_bin_load_vec_bool(const char** pos, const char* end, void* base, size_t off) {
    uint32_t n; ason_err_t e = ason_bin_read_u32(pos, end, &n); if (e) return e;
    bool* arr = (bool*)malloc((size_t)n * sizeof(bool));
    if (!arr && n) return ASON_ERR_ALLOC;
    for (uint32_t i = 0; i < n; i++) {
        uint8_t b; e = ason_bin_read_u8(pos, end, &b); if (e) { free(arr); return e; }
        arr[i] = b != 0;
    }
    ason_vec_bool v = {arr, n, n}; memcpy((char*)base + off, &v, sizeof(v)); return ASON_OK;
}

/* ---- type dispatch ---- */
static void ason_bin_select_dump(ason_buf_t* buf, const ason_field_t* f, const void* base) {
    switch (f->type) {
        case ASON_BOOL:       ason_bin_dump_bool    (buf, base, f->offset); break;
        case ASON_I8:         ason_bin_dump_i8      (buf, base, f->offset); break;
        case ASON_I16:        ason_bin_dump_i16     (buf, base, f->offset); break;
        case ASON_I32:        ason_bin_dump_i32     (buf, base, f->offset); break;
        case ASON_I64:        ason_bin_dump_i64     (buf, base, f->offset); break;
        case ASON_U8:         ason_bin_dump_u8      (buf, base, f->offset); break;
        case ASON_U16:        ason_bin_dump_u16     (buf, base, f->offset); break;
        case ASON_U32:        ason_bin_dump_u32     (buf, base, f->offset); break;
        case ASON_U64:        ason_bin_dump_u64     (buf, base, f->offset); break;
        case ASON_F32:        ason_bin_dump_f32     (buf, base, f->offset); break;
        case ASON_F64:        ason_bin_dump_f64     (buf, base, f->offset); break;
        case ASON_STR:        ason_bin_dump_str     (buf, base, f->offset); break;
        case ASON_VEC_I64:    ason_bin_dump_vec_i64 (buf, base, f->offset); break;
        case ASON_VEC_U64:    ason_bin_dump_vec_u64 (buf, base, f->offset); break;
        case ASON_VEC_F64:    ason_bin_dump_vec_f64 (buf, base, f->offset); break;
        case ASON_VEC_STR:    ason_bin_dump_vec_str (buf, base, f->offset); break;
        case ASON_VEC_BOOL:   ason_bin_dump_vec_bool(buf, base, f->offset); break;
        case ASON_STRUCT:
            if (f->dump_fn) f->dump_fn(buf, base, f->offset);
            else ason_bin_dump_struct(buf, (const char*)base + f->offset,
                                      (const ason_desc_t*)f->sub_desc);
            break;
        default: break;
    }
}

static ason_err_t ason_bin_select_load(const char** pos, const char* end,
                                        const ason_field_t* f, void* base) {
    switch (f->type) {
        case ASON_BOOL:       return ason_bin_load_bool    (pos, end, base, f->offset);
        case ASON_I8:         return ason_bin_load_i8      (pos, end, base, f->offset);
        case ASON_I16:        return ason_bin_load_i16     (pos, end, base, f->offset);
        case ASON_I32:        return ason_bin_load_i32     (pos, end, base, f->offset);
        case ASON_I64:        return ason_bin_load_i64     (pos, end, base, f->offset);
        case ASON_U8:         return ason_bin_load_u8      (pos, end, base, f->offset);
        case ASON_U16:        return ason_bin_load_u16     (pos, end, base, f->offset);
        case ASON_U32:        return ason_bin_load_u32     (pos, end, base, f->offset);
        case ASON_U64:        return ason_bin_load_u64     (pos, end, base, f->offset);
        case ASON_F32:        return ason_bin_load_f32     (pos, end, base, f->offset);
        case ASON_F64:        return ason_bin_load_f64     (pos, end, base, f->offset);
        case ASON_STR:        return ason_bin_load_str     (pos, end, base, f->offset);
        case ASON_VEC_I64:    return ason_bin_load_vec_i64 (pos, end, base, f->offset);
        case ASON_VEC_U64:    return ason_bin_load_vec_u64 (pos, end, base, f->offset);
        case ASON_VEC_F64:    return ason_bin_load_vec_f64 (pos, end, base, f->offset);
        case ASON_VEC_STR:    return ason_bin_load_vec_str (pos, end, base, f->offset);
        case ASON_VEC_BOOL:   return ason_bin_load_vec_bool(pos, end, base, f->offset);
        case ASON_STRUCT:
            if (f->load_fn) return f->load_fn(pos, end, base, f->offset);
            return ason_bin_load_struct(pos, end, (char*)base + f->offset,
                                        (const ason_desc_t*)f->sub_desc);
        default: return ASON_OK;
    }
}

/* ---- struct dump / load ---- */
void ason_bin_dump_struct(ason_buf_t* buf, const void* obj, const ason_desc_t* desc) {
    for (int i = 0; i < desc->field_count; i++)
        ason_bin_select_dump(buf, &desc->fields[i], obj);
}

ason_err_t ason_bin_load_struct(const char** pos, const char* end,
                                  void* obj, const ason_desc_t* desc) {
    for (int i = 0; i < desc->field_count; i++) {
        ason_err_t e = ason_bin_select_load(pos, end, &desc->fields[i], obj);
        if (e != ASON_OK) return e;
    }
    return ASON_OK;
}
