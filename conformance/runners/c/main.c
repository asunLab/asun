/*
 * Conformance runner for asun-c.
 *
 * Loads ../../cases.json (untyped decode) and ../../encode-cases.json
 * (round-trip), feeding each case through asun_value_decode / asun_value_encode.
 *
 * A small inline JSON parser is used to keep the runner self-contained
 * (matching the asun-java runner's approach) — it parses just the subset
 * of JSON used by the conformance manifests.
 */

#include "asun.h"
#include "asun_value.h"

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <stdint.h>
#include <stdbool.h>

/* ============================================================================
 * Minimal JSON parser
 * ========================================================================== */

typedef enum { J_NULL, J_BOOL, J_INT, J_DOUBLE, J_STRING, J_ARRAY, J_OBJECT } j_tag_t;

typedef struct j_val_s j_val_t;
typedef struct { char* k; j_val_t* v; } j_pair_t;

struct j_val_s {
    j_tag_t tag;
    bool b;
    int64_t i;
    double d;
    char* s;       size_t s_len;
    j_val_t** arr; size_t arr_len; size_t arr_cap;
    j_pair_t* obj; size_t obj_len; size_t obj_cap;
};

static j_val_t* j_alloc(j_tag_t tag) { j_val_t* v = calloc(1, sizeof(*v)); v->tag = tag; return v; }
static void j_free(j_val_t* v) {
    if (!v) return;
    if (v->s) free(v->s);
    for (size_t i = 0; i < v->arr_len; i++) j_free(v->arr[i]);
    if (v->arr) free(v->arr);
    for (size_t i = 0; i < v->obj_len; i++) { free(v->obj[i].k); j_free(v->obj[i].v); }
    if (v->obj) free(v->obj);
    free(v);
}
static j_val_t* j_obj_get(const j_val_t* o, const char* k) {
    for (size_t i = 0; i < o->obj_len; i++) if (strcmp(o->obj[i].k, k) == 0) return o->obj[i].v;
    return NULL;
}

typedef struct { const char* s; size_t n; size_t p; } j_parser_t;

static void j_skip(j_parser_t* P) {
    while (P->p < P->n) {
        char c = P->s[P->p];
        if (c == ' ' || c == '\t' || c == '\n' || c == '\r') { P->p++; continue; }
        break;
    }
}

static char* j_str(j_parser_t* P, size_t* out_len);
static j_val_t* j_val(j_parser_t* P);

static j_val_t* j_obj_(j_parser_t* P) {
    j_val_t* v = j_alloc(J_OBJECT);
    P->p++;
    j_skip(P);
    if (P->p < P->n && P->s[P->p] == '}') { P->p++; return v; }
    for (;;) {
        j_skip(P);
        size_t klen;
        char* k = j_str(P, &klen);
        j_skip(P);
        P->p++; /* ':' */
        j_val_t* val = j_val(P);
        if (v->obj_len >= v->obj_cap) {
            v->obj_cap = v->obj_cap ? v->obj_cap * 2 : 8;
            v->obj = realloc(v->obj, v->obj_cap * sizeof(j_pair_t));
        }
        v->obj[v->obj_len++] = (j_pair_t){k, val};
        j_skip(P);
        char c = P->s[P->p++];
        if (c == ',') continue;
        if (c == '}') return v;
        fprintf(stderr, "json: expected , or } got %c\n", c); exit(2);
    }
}
static j_val_t* j_arr_(j_parser_t* P) {
    j_val_t* v = j_alloc(J_ARRAY);
    P->p++;
    j_skip(P);
    if (P->p < P->n && P->s[P->p] == ']') { P->p++; return v; }
    for (;;) {
        j_val_t* e = j_val(P);
        if (v->arr_len >= v->arr_cap) {
            v->arr_cap = v->arr_cap ? v->arr_cap * 2 : 8;
            v->arr = realloc(v->arr, v->arr_cap * sizeof(j_val_t*));
        }
        v->arr[v->arr_len++] = e;
        j_skip(P);
        char c = P->s[P->p++];
        if (c == ',') continue;
        if (c == ']') return v;
        fprintf(stderr, "json: expected , or ] got %c\n", c); exit(2);
    }
}
static char* j_str(j_parser_t* P, size_t* out_len) {
    if (P->p >= P->n || P->s[P->p] != '"') { fprintf(stderr, "json: expected string\n"); exit(2); }
    P->p++;
    size_t cap = 16, len = 0;
    char* buf = malloc(cap);
    while (P->p < P->n) {
        char c = P->s[P->p++];
        if (c == '"') { buf[len] = 0; if (out_len) *out_len = len; return buf; }
        if (c == '\\') {
            char e = P->s[P->p++];
            char w = 0;
            int u_bytes = 0; char ub[4] = {0};
            switch (e) {
                case '"':  w = '"';  break;
                case '\\': w = '\\'; break;
                case '/':  w = '/';  break;
                case 'n':  w = '\n'; break;
                case 't':  w = '\t'; break;
                case 'r':  w = '\r'; break;
                case 'b':  w = '\b'; break;
                case 'f':  w = '\f'; break;
                case 'u': {
                    char hex[5] = { P->s[P->p], P->s[P->p+1], P->s[P->p+2], P->s[P->p+3], 0 };
                    P->p += 4;
                    unsigned long cp = strtoul(hex, NULL, 16);
                    if (cp < 0x80) { ub[0] = (char)cp; u_bytes = 1; }
                    else if (cp < 0x800) {
                        ub[0] = (char)(0xC0 | (cp >> 6));
                        ub[1] = (char)(0x80 | (cp & 0x3F));
                        u_bytes = 2;
                    } else {
                        ub[0] = (char)(0xE0 | (cp >> 12));
                        ub[1] = (char)(0x80 | ((cp >> 6) & 0x3F));
                        ub[2] = (char)(0x80 | (cp & 0x3F));
                        u_bytes = 3;
                    }
                    break;
                }
                default:   fprintf(stderr, "json: bad escape \\%c\n", e); exit(2);
            }
            if (u_bytes) {
                while (len + (size_t)u_bytes + 1 >= cap) cap *= 2;
                buf = realloc(buf, cap);
                memcpy(buf + len, ub, (size_t)u_bytes);
                len += (size_t)u_bytes;
            } else {
                if (len + 1 >= cap) { cap *= 2; buf = realloc(buf, cap); }
                buf[len++] = w;
            }
        } else {
            if (len + 1 >= cap) { cap *= 2; buf = realloc(buf, cap); }
            buf[len++] = c;
        }
    }
    fprintf(stderr, "json: unterminated string\n"); exit(2);
}
static j_val_t* j_num_(j_parser_t* P) {
    size_t start = P->p;
    if (P->p < P->n && P->s[P->p] == '-') P->p++;
    while (P->p < P->n && P->s[P->p] >= '0' && P->s[P->p] <= '9') P->p++;
    bool is_float = false;
    if (P->p < P->n && P->s[P->p] == '.') {
        is_float = true; P->p++;
        while (P->p < P->n && P->s[P->p] >= '0' && P->s[P->p] <= '9') P->p++;
    }
    if (P->p < P->n && (P->s[P->p] == 'e' || P->s[P->p] == 'E')) {
        is_float = true; P->p++;
        if (P->p < P->n && (P->s[P->p] == '+' || P->s[P->p] == '-')) P->p++;
        while (P->p < P->n && P->s[P->p] >= '0' && P->s[P->p] <= '9') P->p++;
    }
    char tmp[64];
    size_t n = P->p - start;
    if (n >= sizeof(tmp)) n = sizeof(tmp) - 1;
    memcpy(tmp, P->s + start, n); tmp[n] = 0;
    j_val_t* v;
    if (is_float) { v = j_alloc(J_DOUBLE); v->d = strtod(tmp, NULL); }
    else          { v = j_alloc(J_INT);    v->i = strtoll(tmp, NULL, 10); }
    return v;
}
static j_val_t* j_val(j_parser_t* P) {
    j_skip(P);
    char c = P->s[P->p];
    if (c == '{') return j_obj_(P);
    if (c == '[') return j_arr_(P);
    if (c == '"') {
        size_t len;
        char* s = j_str(P, &len);
        j_val_t* v = j_alloc(J_STRING);
        v->s = s; v->s_len = len;
        return v;
    }
    if (c == 't') { P->p += 4; j_val_t* v = j_alloc(J_BOOL); v->b = true;  return v; }
    if (c == 'f') { P->p += 5; j_val_t* v = j_alloc(J_BOOL); v->b = false; return v; }
    if (c == 'n') { P->p += 4; return j_alloc(J_NULL); }
    return j_num_(P);
}
static j_val_t* j_parse(const char* s, size_t n) {
    j_parser_t P = {s, n, 0};
    j_skip(&P);
    j_val_t* v = j_val(&P);
    j_skip(&P);
    return v;
}

/* Convert JSON value -> asun_value_t */
static asun_value_t* jv_to_av(const j_val_t* j) {
    switch (j->tag) {
        case J_NULL:   return asun_value_make_null();
        case J_BOOL:   return asun_value_make_bool(j->b);
        case J_INT:    return asun_value_make_int(j->i);
        case J_DOUBLE: return asun_value_make_double(j->d);
        case J_STRING: return asun_value_make_string(j->s, j->s_len);
        case J_ARRAY: {
            asun_value_t* a = asun_value_alloc(ASUN_VAL_ARRAY);
            for (size_t k = 0; k < j->arr_len; k++) {
                asun_value_t* c = jv_to_av(j->arr[k]);
                asun_value_array_push(a, c);
            }
            return a;
        }
        case J_OBJECT: return asun_value_make_null();
    }
    return asun_value_make_null();
}

/* Diagnostic printer */
static void av_print(const asun_value_t* v, asun_buf_t* out) {
    switch (v->tag) {
        case ASUN_VAL_NULL:   asun_buf_appends(out, "null"); return;
        case ASUN_VAL_BOOL:   asun_buf_appends(out, v->b ? "true" : "false"); return;
        case ASUN_VAL_INT:    asun_buf_append_i64(out, v->i); return;
        case ASUN_VAL_DOUBLE: asun_buf_append_f64(out, v->d); return;
        case ASUN_VAL_STRING:
            asun_buf_push(out, '"');
            for (size_t k = 0; k < v->s_len; k++) {
                unsigned char c = (unsigned char)v->s[k];
                if (c == '"' || c == '\\') { asun_buf_push(out, '\\'); asun_buf_push(out, (char)c); }
                else if (c == '\n') asun_buf_appends(out, "\\n");
                else if (c == '\r') asun_buf_appends(out, "\\r");
                else if (c == '\t') asun_buf_appends(out, "\\t");
                else if (c < 0x20) {
                    char hex[8]; snprintf(hex, sizeof(hex), "\\u%04x", c); asun_buf_appends(out, hex);
                } else asun_buf_push(out, (char)c);
            }
            asun_buf_push(out, '"');
            return;
        case ASUN_VAL_ARRAY:
            asun_buf_push(out, '[');
            for (size_t k = 0; k < v->arr_len; k++) {
                if (k) asun_buf_push(out, ',');
                av_print(&v->arr[k], out);
            }
            asun_buf_push(out, ']');
            return;
    }
}
static char* av_to_diag(const asun_value_t* v) {
    asun_buf_t b = asun_buf_new(64);
    av_print(v, &b);
    asun_buf_push(&b, 0);
    return b.data; /* owned; caller frees */
}

/* ============================================================================
 * Failure tracking
 * ========================================================================== */
typedef struct { char* id; char* msg; } failure_t;
typedef struct { failure_t* d; size_t n; size_t cap; } failures_t;
static void f_push(failures_t* F, const char* id, char* msg) {
    if (F->n >= 25) { free(msg); return; }
    if (F->n >= F->cap) { F->cap = F->cap ? F->cap * 2 : 8; F->d = realloc(F->d, F->cap * sizeof(failure_t)); }
    F->d[F->n].id = strdup(id);
    F->d[F->n].msg = msg;
    F->n++;
}

static char* read_file(const char* path, size_t* out_len) {
    FILE* f = fopen(path, "rb");
    if (!f) { fprintf(stderr, "cannot open %s\n", path); exit(2); }
    fseek(f, 0, SEEK_END); long n = ftell(f); fseek(f, 0, SEEK_SET);
    char* buf = malloc((size_t)n + 1);
    fread(buf, 1, (size_t)n, f);
    buf[n] = 0;
    fclose(f);
    if (out_len) *out_len = (size_t)n;
    return buf;
}

/* ============================================================================
 * Main
 * ========================================================================== */
int main(void) {
    /* ---- decode (cases.json) ---- */
    size_t n;
    char* dtext = read_file("../../cases.json", &n);
    j_val_t* d_root = j_parse(dtext, n);
    j_val_t* dcases = j_obj_get(d_root, "cases");
    printf("loaded %zu cases from conformance/cases.json\n", dcases->arr_len);

    size_t d_total = 0, d_ok_pass = 0, d_ok_fail = 0;
    size_t d_err_pass = 0, d_err_fail = 0, d_skipped = 0;
    failures_t df = {0};

    for (size_t i = 0; i < dcases->arr_len; i++) {
        d_total++;
        j_val_t* c = dcases->arr[i];
        const char* id = j_obj_get(c, "id")->s;
        j_val_t* sd = j_obj_get(c, "schemaDriven");
        if (sd && sd->tag == J_BOOL && sd->b) { d_skipped++; continue; }
        const char* input = j_obj_get(c, "input")->s;
        size_t input_len = j_obj_get(c, "input")->s_len;
        const char* kind = j_obj_get(c, "kind")->s;
        bool expect_ok = strcmp(kind, "ok") == 0;

        asun_value_t* got = asun_value_decode(input, input_len);
        if (!got) {
            if (!expect_ok) d_err_pass++;
            else {
                d_ok_fail++;
                size_t buflen = strlen(input) + 80;
                char* msg = malloc(buflen);
                snprintf(msg, buflen, "expected ok, got error\n    input: %s", input);
                f_push(&df, id, msg);
            }
            continue;
        }
        if (expect_ok) {
            asun_value_t* expected = jv_to_av(j_obj_get(c, "expected"));
            if (asun_value_eq(got, expected)) d_ok_pass++;
            else {
                d_ok_fail++;
                char* exp_s = av_to_diag(expected);
                char* got_s = av_to_diag(got);
                size_t L = strlen(input) + strlen(exp_s) + strlen(got_s) + 80;
                char* msg = malloc(L);
                snprintf(msg, L, "value mismatch\n    input:    %s\n    expected: %s\n    actual:   %s",
                         input, exp_s, got_s);
                free(exp_s); free(got_s);
                f_push(&df, id, msg);
            }
            asun_value_free(expected);
        } else {
            d_err_fail++;
            char* got_s = av_to_diag(got);
            size_t L = strlen(input) + strlen(got_s) + 80;
            char* msg = malloc(L);
            snprintf(msg, L, "expected error, got ok: %s\n    input: %s", got_s, input);
            free(got_s);
            f_push(&df, id, msg);
        }
        asun_value_free(got);
    }

    printf("\n================ ASUN-C conformance ================\n");
    printf("total                : %zu\n", d_total);
    printf("untyped ok-cases pass: %zu\n", d_ok_pass);
    printf("untyped ok-cases fail: %zu\n", d_ok_fail);
    printf("untyped err-cases pass: %zu\n", d_err_pass);
    printf("untyped err-cases fail: %zu\n", d_err_fail);
    printf("skipped (needs typed): %zu\n", d_skipped);
    size_t executed = d_total - d_skipped;
    double dpct = executed ? 100.0 * (d_ok_pass + d_err_pass) / executed : 0.0;
    printf("untyped pass rate    : %zu/%zu (%.1f%%)\n", d_ok_pass + d_err_pass, executed, dpct);
    printf("====================================================\n");
    for (size_t i = 0; i < df.n; i++) printf("\n[%s]\n    %s\n", df.d[i].id, df.d[i].msg);

    /* ---- encode (encode-cases.json) ---- */
    size_t en;
    char* etext = read_file("../../encode-cases.json", &en);
    j_val_t* e_root = j_parse(etext, en);
    j_val_t* ecases = j_obj_get(e_root, "cases");
    printf("loaded %zu encode cases from conformance/encode-cases.json\n", ecases->arr_len);

    size_t e_pass = 0, e_fail = 0;
    failures_t ef = {0};

    for (size_t i = 0; i < ecases->arr_len; i++) {
        j_val_t* c = ecases->arr[i];
        const char* id = j_obj_get(c, "id")->s;
        asun_value_t* value = jv_to_av(j_obj_get(c, "value"));

        asun_buf_t enc = asun_value_encode(value);
        asun_value_t* decoded = asun_value_decode(enc.data, enc.len);
        if (!decoded) {
            e_fail++;
            char* enc_dup = strndup(enc.data, enc.len);
            size_t L = strlen(enc_dup) + 80;
            char* msg = malloc(L);
            snprintf(msg, L, "decode error after encode\n    encoded: %s", enc_dup);
            free(enc_dup);
            f_push(&ef, id, msg);
        } else if (!asun_value_eq(decoded, value)) {
            e_fail++;
            char* exp_s = av_to_diag(value);
            char* got_s = av_to_diag(decoded);
            char* enc_dup = strndup(enc.data, enc.len);
            size_t L = strlen(exp_s) + strlen(got_s) + strlen(enc_dup) + 80;
            char* msg = malloc(L);
            snprintf(msg, L, "round-trip mismatch\n    encoded:  %s\n    expected: %s\n    actual:   %s",
                     enc_dup, exp_s, got_s);
            free(exp_s); free(got_s); free(enc_dup);
            f_push(&ef, id, msg);
        } else {
            e_pass++;
        }
        asun_buf_free(&enc);
        asun_value_free(decoded);
        asun_value_free(value);
    }

    printf("\n================ ASUN-C encode conformance ================\n");
    printf("total : %zu\n", ecases->arr_len);
    printf("pass  : %zu\n", e_pass);
    printf("fail  : %zu\n", e_fail);
    double ept = ecases->arr_len ? 100.0 * e_pass / ecases->arr_len : 0.0;
    printf("rate  : %zu/%zu (%.1f%%)\n", e_pass, ecases->arr_len, ept);
    printf("============================================================\n");
    for (size_t i = 0; i < ef.n; i++) printf("\n[%s]\n    %s\n", ef.d[i].id, ef.d[i].msg);

    j_free(d_root); j_free(e_root); free(dtext); free(etext);
    for (size_t i = 0; i < df.n; i++) { free(df.d[i].id); free(df.d[i].msg); } free(df.d);
    for (size_t i = 0; i < ef.n; i++) { free(ef.d[i].id); free(ef.d[i].msg); } free(ef.d);

    if (d_ok_fail || d_err_fail || e_fail) return 1;
    return 0;
}
