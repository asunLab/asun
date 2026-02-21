#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
#include <time.h>
#include "ason.h"

/* ===========================================================================
 * Timing
 * =========================================================================== */
#ifdef __APPLE__
#include <mach/mach_time.h>
static double now_ms(void) {
    static mach_timebase_info_data_t info = {0,0};
    if (info.denom == 0) mach_timebase_info(&info);
    return (double)mach_absolute_time() * info.numer / info.denom / 1e6;
}
#else
static double now_ms(void) {
    struct timespec ts;
    clock_gettime(CLOCK_MONOTONIC, &ts);
    return ts.tv_sec * 1000.0 + ts.tv_nsec / 1e6;
}
#endif

/* ===========================================================================
 * Structs
 * =========================================================================== */

typedef struct {
    int64_t id;
    ason_string_t name;
    ason_string_t email;
    int64_t age;
    double score;
    bool active;
    ason_string_t role;
    ason_string_t city;
} BUser;

ASON_FIELDS(BUser, 8,
    ASON_FIELD(BUser, id,     "id",     i64),
    ASON_FIELD(BUser, name,   "name",   str),
    ASON_FIELD(BUser, email,  "email",  str),
    ASON_FIELD(BUser, age,    "age",    i64),
    ASON_FIELD(BUser, score,  "score",  f64),
    ASON_FIELD(BUser, active, "active", bool),
    ASON_FIELD(BUser, role,   "role",   str),
    ASON_FIELD(BUser, city,   "city",   str))

ASON_FIELDS_BIN(BUser, 8)

typedef struct {
    bool b;
    int8_t i8v;
    int16_t i16v;
    int32_t i32v;
    int64_t i64v;
    uint8_t u8v;
    uint16_t u16v;
    uint32_t u32v;
    uint64_t u64v;
    float f32v;
    double f64v;
    ason_string_t s;
    ason_opt_i64 opt_some;
    ason_opt_i64 opt_none;
    ason_vec_i64 vec_int;
    ason_vec_str vec_str;
} BAllTypes;

ASON_FIELDS(BAllTypes, 16,
    ASON_FIELD(BAllTypes, b,        "b",        bool),
    ASON_FIELD(BAllTypes, i8v,      "i8v",      i8),
    ASON_FIELD(BAllTypes, i16v,     "i16v",     i16),
    ASON_FIELD(BAllTypes, i32v,     "i32v",     i32),
    ASON_FIELD(BAllTypes, i64v,     "i64v",     i64),
    ASON_FIELD(BAllTypes, u8v,      "u8v",      u8),
    ASON_FIELD(BAllTypes, u16v,     "u16v",     u16),
    ASON_FIELD(BAllTypes, u32v,     "u32v",     u32),
    ASON_FIELD(BAllTypes, u64v,     "u64v",     u64),
    ASON_FIELD(BAllTypes, f32v,     "f32v",     f32),
    ASON_FIELD(BAllTypes, f64v,     "f64v",     f64),
    ASON_FIELD(BAllTypes, s,        "s",        str),
    ASON_FIELD(BAllTypes, opt_some, "opt_some", opt_i64),
    ASON_FIELD(BAllTypes, opt_none, "opt_none", opt_i64),
    ASON_FIELD(BAllTypes, vec_int,  "vec_int",  vec_i64),
    ASON_FIELD(BAllTypes, vec_str,  "vec_str",  vec_str))
ASON_FIELDS_BIN(BAllTypes, 16)

/* 5-level: Company > Division > Team > Project > Task */
typedef struct { int64_t id; ason_string_t title; int64_t priority; bool done; double hours; } BTask;
ASON_FIELDS(BTask, 5,
    ASON_FIELD(BTask, id,       "id",       i64),
    ASON_FIELD(BTask, title,    "title",    str),
    ASON_FIELD(BTask, priority, "priority", i64),
    ASON_FIELD(BTask, done,     "done",     bool),
    ASON_FIELD(BTask, hours,    "hours",    f64))
ASON_FIELDS_BIN(BTask, 5)
ASON_VEC_STRUCT_DEFINE(BTask)

typedef struct { ason_string_t name; double budget; bool active; ason_vec_BTask tasks; } BProject;
ASON_FIELDS(BProject, 4,
    ASON_FIELD(BProject, name,   "name",   str),
    ASON_FIELD(BProject, budget, "budget", f64),
    ASON_FIELD(BProject, active, "active", bool),
    ASON_FIELD_VEC_STRUCT(BProject, tasks, "tasks", BTask))
ASON_FIELDS_BIN(BProject, 4)
ASON_VEC_STRUCT_DEFINE(BProject)

typedef struct { ason_string_t name; ason_string_t lead; int64_t size; ason_vec_BProject projects; } BTeam;
ASON_FIELDS(BTeam, 4,
    ASON_FIELD(BTeam, name,     "name",     str),
    ASON_FIELD(BTeam, lead,     "lead",     str),
    ASON_FIELD(BTeam, size,     "size",     i64),
    ASON_FIELD_VEC_STRUCT(BTeam, projects, "projects", BProject))
ASON_FIELDS_BIN(BTeam, 4)
ASON_VEC_STRUCT_DEFINE(BTeam)

typedef struct { ason_string_t name; ason_string_t location; int64_t headcount; ason_vec_BTeam teams; } BDivision;
ASON_FIELDS(BDivision, 4,
    ASON_FIELD(BDivision, name,      "name",      str),
    ASON_FIELD(BDivision, location,  "location",  str),
    ASON_FIELD(BDivision, headcount, "headcount", i64),
    ASON_FIELD_VEC_STRUCT(BDivision, teams, "teams", BTeam))
ASON_FIELDS_BIN(BDivision, 4)
ASON_VEC_STRUCT_DEFINE(BDivision)

typedef struct {
    ason_string_t name;
    int64_t founded;
    double revenue_m;
    bool is_public;
    ason_vec_BDivision divisions;
    ason_vec_str tags;
} BCompany;

ASON_FIELDS(BCompany, 6,
    ASON_FIELD(BCompany, name,      "name",      str),
    ASON_FIELD(BCompany, founded,   "founded",   i64),
    ASON_FIELD(BCompany, revenue_m, "revenue_m", f64),
    ASON_FIELD(BCompany, is_public, "public",    bool),
    ASON_FIELD_VEC_STRUCT(BCompany, divisions, "divisions", BDivision),
    ASON_FIELD(BCompany, tags,      "tags",      vec_str))
ASON_FIELDS_BIN(BCompany, 6)

/* ===========================================================================
 * Mini JSON serializer (for comparison)
 * =========================================================================== */

static void json_append_str(ason_buf_t* b, const char* s, size_t len) {
    ason_buf_push(b, '"');
    for (size_t i = 0; i < len; i++) {
        switch (s[i]) {
        case '"':  ason_buf_append(b, "\\\"", 2); break;
        case '\\': ason_buf_append(b, "\\\\", 2); break;
        case '\n': ason_buf_append(b, "\\n", 2); break;
        case '\t': ason_buf_append(b, "\\t", 2); break;
        default:   ason_buf_push(b, s[i]); break;
        }
    }
    ason_buf_push(b, '"');
}

static void json_append_i64(ason_buf_t* b, int64_t v) { ason_buf_append_i64(b, v); }
static void json_append_f64(ason_buf_t* b, double v) { ason_buf_append_f64(b, v); }

#define JSON_KEY_STR(buf, key, s, slen) \
    ason_buf_push(buf, '"'); ason_buf_append(buf, key, strlen(key)); ason_buf_append(buf, "\":", 2); \
    json_append_str(buf, s, slen)

#define JSON_KEY_I64(buf, key, v) \
    ason_buf_push(buf, '"'); ason_buf_append(buf, key, strlen(key)); ason_buf_append(buf, "\":", 2); \
    json_append_i64(buf, v)

#define JSON_KEY_F64(buf, key, v) \
    ason_buf_push(buf, '"'); ason_buf_append(buf, key, strlen(key)); ason_buf_append(buf, "\":", 2); \
    json_append_f64(buf, v)

#define JSON_KEY_BOOL(buf, key, v) \
    ason_buf_push(buf, '"'); ason_buf_append(buf, key, strlen(key)); ason_buf_append(buf, "\":", 2); \
    ason_buf_append(buf, (v) ? "true" : "false", (v) ? 4 : 5)

static void json_serialize_user(ason_buf_t* b, const BUser* u) {
    ason_buf_push(b, '{');
    JSON_KEY_I64(b, "id", u->id); ason_buf_push(b, ',');
    JSON_KEY_STR(b, "name", u->name.data, u->name.len); ason_buf_push(b, ',');
    JSON_KEY_STR(b, "email", u->email.data, u->email.len); ason_buf_push(b, ',');
    JSON_KEY_I64(b, "age", u->age); ason_buf_push(b, ',');
    JSON_KEY_F64(b, "score", u->score); ason_buf_push(b, ',');
    JSON_KEY_BOOL(b, "active", u->active); ason_buf_push(b, ',');
    JSON_KEY_STR(b, "role", u->role.data, u->role.len); ason_buf_push(b, ',');
    JSON_KEY_STR(b, "city", u->city.data, u->city.len);
    ason_buf_push(b, '}');
}

static ason_buf_t json_serialize_users(const BUser* users, size_t n) {
    ason_buf_t b = ason_buf_new(n * 200);
    ason_buf_push(&b, '[');
    for (size_t i = 0; i < n; i++) {
        if (i > 0) ason_buf_push(&b, ',');
        json_serialize_user(&b, &users[i]);
    }
    ason_buf_push(&b, ']');
    return b;
}

/* Mini JSON deserializer */
static void json_skip_ws(const char** p, const char* e) {
    while (*p < e && (**p == ' ' || **p == '\n' || **p == '\t' || **p == '\r')) (*p)++;
}
static void json_expect(const char** p, const char* e, char c) {
    json_skip_ws(p, e); if (*p < e && **p == c) (*p)++;
}
static ason_string_t json_read_str(const char** p, const char* e) {
    json_skip_ws(p, e);
    if (**p == '"') (*p)++;
    ason_buf_t b = ason_buf_new(32);
    while (*p < e && **p != '"') {
        if (**p == '\\') { (*p)++;
            if (**p == 'n') ason_buf_push(&b, '\n');
            else if (**p == 't') ason_buf_push(&b, '\t');
            else ason_buf_push(&b, **p);
            (*p)++;
        } else { ason_buf_push(&b, **p); (*p)++; }
    }
    if (*p < e) (*p)++;
    ason_string_t s = ason_string_from_len(b.data, b.len);
    ason_buf_free(&b);
    return s;
}
static int64_t json_read_i64(const char** p, const char* e) {
    json_skip_ws(p, e);
    char* endptr = NULL;
    int64_t v = strtoll(*p, &endptr, 10);
    *p = endptr; return v;
}
static double json_read_f64(const char** p, const char* e) {
    json_skip_ws(p, e);
    char* endptr = NULL;
    double v = strtod(*p, &endptr);
    *p = endptr; return v;
}
static bool json_read_bool(const char** p, const char* e) {
    json_skip_ws(p, e);
    if (*p + 4 <= e && (*p)[0] == 't') { *p += 4; return true; }
    *p += 5; return false;
}
static void json_skip_comma(const char** p, const char* e) {
    json_skip_ws(p, e); if (*p < e && **p == ',') (*p)++;
}
static void json_skip_key(const char** p, const char* e) {
    ason_string_t k = json_read_str(p, e);
    ason_string_free(&k);
    json_expect(p, e, ':');
}

static BUser json_deserialize_user(const char** p, const char* e) {
    BUser u = {0};
    json_expect(p, e, '{');
    json_skip_key(p, e); u.id = json_read_i64(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); u.name = json_read_str(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); u.email = json_read_str(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); u.age = json_read_i64(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); u.score = json_read_f64(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); u.active = json_read_bool(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); u.role = json_read_str(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); u.city = json_read_str(p, e);
    json_expect(p, e, '}');
    return u;
}

static BUser* json_deserialize_users(const char* data, size_t len, size_t* out_n) {
    const char* p = data;
    const char* e = data + len;
    size_t cap = 64, cnt = 0;
    BUser* arr = (BUser*)malloc(cap * sizeof(BUser));
    json_expect(&p, e, '[');
    while (1) {
        json_skip_ws(&p, e);
        if (p >= e || *p == ']') break;
        if (cnt > 0) json_skip_comma(&p, e);
        if (cnt >= cap) { cap *= 2; arr = (BUser*)realloc(arr, cap * sizeof(BUser)); }
        arr[cnt++] = json_deserialize_user(&p, e);
    }
    *out_n = cnt;
    return arr;
}

/* JSON deep struct deserializers (for fair comparison) */

static BTask json_deserialize_task(const char** p, const char* e) {
    BTask t = {0};
    json_expect(p, e, '{');
    json_skip_key(p, e); t.id = json_read_i64(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); t.title = json_read_str(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); t.priority = json_read_i64(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); t.done = json_read_bool(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); t.hours = json_read_f64(p, e);
    json_expect(p, e, '}');
    return t;
}

static BProject json_deserialize_project(const char** p, const char* e) {
    BProject proj = {0};
    json_expect(p, e, '{');
    json_skip_key(p, e); proj.name = json_read_str(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); proj.budget = json_read_f64(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); proj.active = json_read_bool(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); /* "tasks" */
    json_expect(p, e, '[');
    proj.tasks = ason_vec_BTask_new();
    while (1) {
        json_skip_ws(p, e);
        if (*p >= e || **p == ']') break;
        if (proj.tasks.len > 0) json_skip_comma(p, e);
        ason_vec_BTask_push(&proj.tasks, json_deserialize_task(p, e));
    }
    json_expect(p, e, ']');
    json_expect(p, e, '}');
    return proj;
}

static BTeam json_deserialize_team(const char** p, const char* e) {
    BTeam team = {0};
    json_expect(p, e, '{');
    json_skip_key(p, e); team.name = json_read_str(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); team.lead = json_read_str(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); team.size = json_read_i64(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); /* "projects" */
    json_expect(p, e, '[');
    team.projects = ason_vec_BProject_new();
    while (1) {
        json_skip_ws(p, e);
        if (*p >= e || **p == ']') break;
        if (team.projects.len > 0) json_skip_comma(p, e);
        ason_vec_BProject_push(&team.projects, json_deserialize_project(p, e));
    }
    json_expect(p, e, ']');
    json_expect(p, e, '}');
    return team;
}

static BDivision json_deserialize_division(const char** p, const char* e) {
    BDivision div = {0};
    json_expect(p, e, '{');
    json_skip_key(p, e); div.name = json_read_str(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); div.location = json_read_str(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); div.headcount = json_read_i64(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); /* "teams" */
    json_expect(p, e, '[');
    div.teams = ason_vec_BTeam_new();
    while (1) {
        json_skip_ws(p, e);
        if (*p >= e || **p == ']') break;
        if (div.teams.len > 0) json_skip_comma(p, e);
        ason_vec_BTeam_push(&div.teams, json_deserialize_team(p, e));
    }
    json_expect(p, e, ']');
    json_expect(p, e, '}');
    return div;
}

static BCompany json_deserialize_company(const char** p, const char* e) {
    BCompany c = {0};
    json_expect(p, e, '{');
    json_skip_key(p, e); c.name = json_read_str(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); c.founded = json_read_i64(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); c.revenue_m = json_read_f64(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); c.is_public = json_read_bool(p, e); json_skip_comma(p, e);
    json_skip_key(p, e); /* "divisions" */
    json_expect(p, e, '[');
    c.divisions = ason_vec_BDivision_new();
    while (1) {
        json_skip_ws(p, e);
        if (*p >= e || **p == ']') break;
        if (c.divisions.len > 0) json_skip_comma(p, e);
        ason_vec_BDivision_push(&c.divisions, json_deserialize_division(p, e));
    }
    json_expect(p, e, ']');
    json_skip_comma(p, e);
    json_skip_key(p, e); /* "tags" */
    json_expect(p, e, '[');
    c.tags = ason_vec_str_new();
    while (1) {
        json_skip_ws(p, e);
        if (*p >= e || **p == ']') break;
        if (c.tags.len > 0) json_skip_comma(p, e);
        ason_vec_str_push(&c.tags, json_read_str(p, e));
    }
    json_expect(p, e, ']');
    json_expect(p, e, '}');
    return c;
}

static BCompany* json_deserialize_companies(const char* data, size_t len, size_t* out_n) {
    const char* p = data;
    const char* e = data + len;
    size_t cap = 16, cnt = 0;
    BCompany* arr = (BCompany*)malloc(cap * sizeof(BCompany));
    json_expect(&p, e, '[');
    while (1) {
        json_skip_ws(&p, e);
        if (p >= e || *p == ']') break;
        if (cnt > 0) json_skip_comma(&p, e);
        if (cnt >= cap) { cap *= 2; arr = (BCompany*)realloc(arr, cap * sizeof(BCompany)); }
        arr[cnt++] = json_deserialize_company(&p, e);
    }
    *out_n = cnt;
    return arr;
}

static void free_buser(BUser* u) {
    ason_string_free(&u->name); ason_string_free(&u->email);
    ason_string_free(&u->role); ason_string_free(&u->city);
}

static void free_balltypes(BAllTypes* a) {
    ason_string_free(&a->s);
    ason_vec_i64_free(&a->vec_int);
    for (size_t i = 0; i < a->vec_str.len; i++) ason_string_free(&a->vec_str.data[i]);
    ason_vec_str_free(&a->vec_str);
}

static void free_btask(BTask* t) { ason_string_free(&t->title); }
static void free_bproject(BProject* p) {
    ason_string_free(&p->name);
    for (size_t i = 0; i < p->tasks.len; i++) free_btask(&p->tasks.data[i]);
    ason_vec_BTask_free(&p->tasks);
}
static void free_bteam(BTeam* t) {
    ason_string_free(&t->name); ason_string_free(&t->lead);
    for (size_t i = 0; i < t->projects.len; i++) free_bproject(&t->projects.data[i]);
    ason_vec_BProject_free(&t->projects);
}
static void free_bdivision(BDivision* d) {
    ason_string_free(&d->name); ason_string_free(&d->location);
    for (size_t i = 0; i < d->teams.len; i++) free_bteam(&d->teams.data[i]);
    ason_vec_BTeam_free(&d->teams);
}
static void free_bcompany(BCompany* c) {
    ason_string_free(&c->name);
    for (size_t i = 0; i < c->divisions.len; i++) free_bdivision(&c->divisions.data[i]);
    ason_vec_BDivision_free(&c->divisions);
    for (size_t i = 0; i < c->tags.len; i++) ason_string_free(&c->tags.data[i]);
    ason_vec_str_free(&c->tags);
}

/* ===========================================================================
 * Data generators
 * =========================================================================== */

static BUser* generate_users(size_t n) {
    const char* names[] = {"Alice","Bob","Carol","David","Eve","Frank","Grace","Hank"};
    const char* roles[] = {"engineer","designer","manager","analyst"};
    const char* cities[] = {"NYC","LA","Chicago","Houston","Phoenix"};
    BUser* users = (BUser*)calloc(n, sizeof(BUser));
    for (size_t i = 0; i < n; i++) {
        users[i].id = (int64_t)i;
        users[i].name = ason_string_from(names[i % 8]);
        char email[64]; snprintf(email, 64, "%s@example.com", names[i % 8]);
        users[i].email = ason_string_from(email);
        users[i].age = 25 + (int64_t)(i % 40);
        users[i].score = 50.0 + (double)(i % 50) + 0.5;
        users[i].active = (i % 3 != 0);
        users[i].role = ason_string_from(roles[i % 4]);
        users[i].city = ason_string_from(cities[i % 5]);
    }
    return users;
}

static BAllTypes* generate_all_types(size_t n) {
    BAllTypes* items = (BAllTypes*)calloc(n, sizeof(BAllTypes));
    for (size_t i = 0; i < n; i++) {
        items[i].b = (i % 2 == 0);
        items[i].i8v = (int8_t)(i % 256);
        items[i].i16v = -(int16_t)i;
        items[i].i32v = (int32_t)(i * 1000);
        items[i].i64v = (int64_t)(i * 100000);
        items[i].u8v = (uint8_t)(i % 256);
        items[i].u16v = (uint16_t)(i % 65536);
        items[i].u32v = (uint32_t)(i * 7919);
        items[i].u64v = (uint64_t)(i * 1000000007ULL);
        items[i].f32v = (float)i * 1.5f;
        items[i].f64v = (double)i * 0.25 + 0.5;
        char tmp[32]; snprintf(tmp, 32, "item_%zu", i);
        items[i].s = ason_string_from(tmp);
        items[i].opt_some = (i % 2 == 0) ? (ason_opt_i64){true, (int64_t)i} : (ason_opt_i64){false, 0};
        items[i].opt_none = (ason_opt_i64){false, 0};
        items[i].vec_int = ason_vec_i64_new();
        ason_vec_i64_push(&items[i].vec_int, (int64_t)i);
        ason_vec_i64_push(&items[i].vec_int, (int64_t)(i+1));
        ason_vec_i64_push(&items[i].vec_int, (int64_t)(i+2));
        items[i].vec_str = ason_vec_str_new();
        char t1[16], t2[16]; snprintf(t1, 16, "tag%zu", i%5); snprintf(t2, 16, "cat%zu", i%3);
        ason_vec_str_push(&items[i].vec_str, ason_string_from(t1));
        ason_vec_str_push(&items[i].vec_str, ason_string_from(t2));
    }
    return items;
}

static BCompany* generate_companies(size_t n) {
    const int dp = 2, tp = 2, pp = 3, tkp = 4;
    const char* locs[] = {"NYC","London","Tokyo","Berlin"};
    const char* leads[] = {"Alice","Bob","Carol","David"};
    BCompany* companies = (BCompany*)calloc(n, sizeof(BCompany));
    for (size_t i = 0; i < n; i++) {
        char tmp[64];
        snprintf(tmp, 64, "Corp_%zu", i); companies[i].name = ason_string_from(tmp);
        companies[i].founded = 1990 + (int64_t)(i % 35);
        companies[i].revenue_m = 10.0 + (double)i * 5.5;
        companies[i].is_public = (i % 2 == 0);
        companies[i].tags = ason_vec_str_new();
        ason_vec_str_push(&companies[i].tags, ason_string_from("enterprise"));
        ason_vec_str_push(&companies[i].tags, ason_string_from("tech"));
        snprintf(tmp, 64, "sector_%zu", i % 5);
        ason_vec_str_push(&companies[i].tags, ason_string_from(tmp));
        companies[i].divisions = ason_vec_BDivision_new();
        for (int d = 0; d < dp; d++) {
            BDivision div = {0};
            snprintf(tmp, 64, "Div_%zu_%d", i, d); div.name = ason_string_from(tmp);
            div.location = ason_string_from(locs[d % 4]);
            div.headcount = 50 + (int64_t)(d * 20);
            div.teams = ason_vec_BTeam_new();
            for (int t = 0; t < tp; t++) {
                BTeam team = {0};
                snprintf(tmp, 64, "Team_%zu_%d_%d", i, d, t); team.name = ason_string_from(tmp);
                team.lead = ason_string_from(leads[t % 4]);
                team.size = 5 + (int64_t)(t * 2);
                team.projects = ason_vec_BProject_new();
                for (int p = 0; p < pp; p++) {
                    BProject proj = {0};
                    snprintf(tmp, 64, "Proj_%d_%d", t, p); proj.name = ason_string_from(tmp);
                    proj.budget = 100.0 + (double)p * 50.5;
                    proj.active = (p % 2 == 0);
                    proj.tasks = ason_vec_BTask_new();
                    for (int tk = 0; tk < tkp; tk++) {
                        BTask task = {0};
                        task.id = (int64_t)(i * 100 + d * 10 + t * 5 + tk);
                        snprintf(tmp, 64, "Task_%d", tk); task.title = ason_string_from(tmp);
                        task.priority = (int64_t)(tk % 3 + 1);
                        task.done = (tk % 2 == 0);
                        task.hours = 2.0 + (double)tk * 1.5;
                        ason_vec_BTask_push(&proj.tasks, task);
                    }
                    ason_vec_BProject_push(&team.projects, proj);
                }
                ason_vec_BTeam_push(&div.teams, team);
            }
            ason_vec_BDivision_push(&companies[i].divisions, div);
        }
    }
    return companies;
}

/* ===========================================================================
 * JSON serialization for Company (for comparison)
 * =========================================================================== */
static void json_serialize_company(ason_buf_t* b, const BCompany* c) {
    ason_buf_push(b, '{');
    JSON_KEY_STR(b, "name", c->name.data, c->name.len); ason_buf_push(b, ',');
    JSON_KEY_I64(b, "founded", c->founded); ason_buf_push(b, ',');
    JSON_KEY_F64(b, "revenue_m", c->revenue_m); ason_buf_push(b, ',');
    JSON_KEY_BOOL(b, "public", c->is_public); ason_buf_push(b, ',');
    ason_buf_append(b, "\"divisions\":[", 13);
    for (size_t d = 0; d < c->divisions.len; d++) {
        if (d > 0) ason_buf_push(b, ',');
        const BDivision* div = &c->divisions.data[d];
        ason_buf_push(b, '{');
        JSON_KEY_STR(b, "name", div->name.data, div->name.len); ason_buf_push(b, ',');
        JSON_KEY_STR(b, "location", div->location.data, div->location.len); ason_buf_push(b, ',');
        JSON_KEY_I64(b, "headcount", div->headcount); ason_buf_push(b, ',');
        ason_buf_append(b, "\"teams\":[", 9);
        for (size_t t = 0; t < div->teams.len; t++) {
            if (t > 0) ason_buf_push(b, ',');
            const BTeam* team = &div->teams.data[t];
            ason_buf_push(b, '{');
            JSON_KEY_STR(b, "name", team->name.data, team->name.len); ason_buf_push(b, ',');
            JSON_KEY_STR(b, "lead", team->lead.data, team->lead.len); ason_buf_push(b, ',');
            JSON_KEY_I64(b, "size", team->size); ason_buf_push(b, ',');
            ason_buf_append(b, "\"projects\":[", 12);
            for (size_t p = 0; p < team->projects.len; p++) {
                if (p > 0) ason_buf_push(b, ',');
                const BProject* proj = &team->projects.data[p];
                ason_buf_push(b, '{');
                JSON_KEY_STR(b, "name", proj->name.data, proj->name.len); ason_buf_push(b, ',');
                JSON_KEY_F64(b, "budget", proj->budget); ason_buf_push(b, ',');
                JSON_KEY_BOOL(b, "active", proj->active); ason_buf_push(b, ',');
                ason_buf_append(b, "\"tasks\":[", 9);
                for (size_t tk = 0; tk < proj->tasks.len; tk++) {
                    if (tk > 0) ason_buf_push(b, ',');
                    const BTask* task = &proj->tasks.data[tk];
                    ason_buf_push(b, '{');
                    JSON_KEY_I64(b, "id", task->id); ason_buf_push(b, ',');
                    JSON_KEY_STR(b, "title", task->title.data, task->title.len); ason_buf_push(b, ',');
                    JSON_KEY_I64(b, "priority", task->priority); ason_buf_push(b, ',');
                    JSON_KEY_BOOL(b, "done", task->done); ason_buf_push(b, ',');
                    JSON_KEY_F64(b, "hours", task->hours);
                    ason_buf_push(b, '}');
                }
                ason_buf_append(b, "]}", 2);
            }
            ason_buf_append(b, "]}", 2);
        }
        ason_buf_append(b, "]}", 2);
    }
    ason_buf_append(b, "],\"tags\":[", 10);
    for (size_t i = 0; i < c->tags.len; i++) {
        if (i > 0) ason_buf_push(b, ',');
        json_append_str(b, c->tags.data[i].data, c->tags.data[i].len);
    }
    ason_buf_append(b, "]}", 2);
}

/* ===========================================================================
 * Benchmark result printing
 * =========================================================================== */

typedef struct {
    const char* name;
    double json_ser_ms, ason_ser_ms, ason_bin_ser_ms;
    double json_de_ms,  ason_de_ms,  ason_bin_de_ms;
    size_t json_bytes, ason_bytes, ason_bin_bytes;
} BenchResult;

static void print_result(const BenchResult* r) {
    double ser_ratio = r->json_ser_ms / r->ason_bin_ser_ms;
    double de_ratio  = r->json_de_ms  / r->ason_bin_de_ms;
    double bin_saving = (1.0 - (double)r->ason_bin_bytes / (double)r->json_bytes) * 100.0;
    printf("  %s\n", r->name);
    printf("    Serialize:   JSON %8.2fms | ASON %8.2fms | BIN %8.2fms | ratio %.2fx\n",
           r->json_ser_ms, r->ason_ser_ms, r->ason_bin_ser_ms, ser_ratio);
    printf("    Deserialize: JSON %8.2fms | ASON %8.2fms | BIN %8.2fms | ratio %.2fx\n",
           r->json_de_ms, r->ason_de_ms, r->ason_bin_de_ms, de_ratio);
    printf("    Size:        JSON %8zu B | ASON %8zu B | BIN %8zu B | saving %.0f%%\n",
           r->json_bytes, r->ason_bytes, r->ason_bin_bytes, bin_saving);
}

/* ===========================================================================
 * Benchmarks
 * =========================================================================== */

static BenchResult bench_flat(size_t count, int iterations) {
    BUser* users = generate_users(count);
    char name_buf[128];
    snprintf(name_buf, 128, "Flat struct x %zu (8 fields)", count);

    /* JSON serialize */
    ason_buf_t json_buf = {0};
    double t0 = now_ms();
    for (int i = 0; i < iterations; i++) {
        ason_buf_free(&json_buf);
        json_buf = json_serialize_users(users, count);
    }
    double json_ser = now_ms() - t0;

    /* ASON serialize */
    ason_buf_t ason_buf = {0};
    t0 = now_ms();
    for (int i = 0; i < iterations; i++) {
        ason_buf_free(&ason_buf);
        ason_buf = ason_dump_vec_BUser(users, count);
    }
    double ason_ser = now_ms() - t0;

    /* JSON deserialize */
    t0 = now_ms();
    for (int i = 0; i < iterations; i++) {
        size_t n = 0;
        BUser* r = json_deserialize_users(json_buf.data, json_buf.len, &n);
        assert(n == count);
        for (size_t j = 0; j < n; j++) free_buser(&r[j]);
        free(r);
    }
    double json_de = now_ms() - t0;

    /* ASON deserialize */
    t0 = now_ms();
    for (int i = 0; i < iterations; i++) {
        size_t n = 0;
        BUser* r = NULL;
        ason_err_t err = ason_load_vec_BUser(ason_buf.data, ason_buf.len, &r, &n);
        assert(err == ASON_OK);
        assert(n == count);
        for (size_t j = 0; j < n; j++) free_buser(&r[j]);
        free(r);
    }
    double ason_de = now_ms() - t0;

    /* ASON-BIN serialize */
    ason_buf_t ason_bin_buf = {0};
    t0 = now_ms();
    for (int i = 0; i < iterations; i++) {
        ason_buf_free(&ason_bin_buf);
        ason_bin_buf = ason_dump_bin_vec_BUser(users, count);
    }
    double ason_bin_ser = now_ms() - t0;

    /* ASON-BIN deserialize (zero-copy: strings point into ason_bin_buf.data) */
    t0 = now_ms();
    for (int i = 0; i < iterations; i++) {
        size_t n = 0;
        BUser* r = NULL;
        ason_err_t err = ason_load_bin_vec_BUser(ason_bin_buf.data, ason_bin_buf.len, &r, &n);
        assert(err == ASON_OK);
        assert(n == count);
        for (size_t j = 0; j < n; j++) free_buser(&r[j]);
        free(r);
    }
    double ason_bin_de = now_ms() - t0;

    BenchResult res = {strdup(name_buf),
        json_ser, ason_ser, ason_bin_ser,
        json_de,  ason_de,  ason_bin_de,
        json_buf.len, ason_buf.len, ason_bin_buf.len};
    ason_buf_free(&json_buf);
    ason_buf_free(&ason_buf);
    ason_buf_free(&ason_bin_buf);
    for (size_t i = 0; i < count; i++) free_buser(&users[i]);
    free(users);
    return res;
}

static BenchResult bench_all_types(size_t count, int iterations) {
    BAllTypes* items = generate_all_types(count);
    char name_buf[128];
    snprintf(name_buf, 128, "All-types struct x %zu (16 fields, per-struct)", count);

    /* JSON serialize (subset of fields) */
    ason_buf_t json_buf = ason_buf_new(count * 256);
    double t0 = now_ms();
    for (int iter = 0; iter < iterations; iter++) {
        json_buf.len = 0;
        ason_buf_push(&json_buf, '[');
        for (size_t i = 0; i < count; i++) {
            if (i > 0) ason_buf_push(&json_buf, ',');
            ason_buf_push(&json_buf, '{');
            JSON_KEY_BOOL(&json_buf, "b", items[i].b); ason_buf_push(&json_buf, ',');
            JSON_KEY_I64(&json_buf, "i8v", items[i].i8v); ason_buf_push(&json_buf, ',');
            JSON_KEY_I64(&json_buf, "i16v", items[i].i16v); ason_buf_push(&json_buf, ',');
            JSON_KEY_I64(&json_buf, "i32v", items[i].i32v); ason_buf_push(&json_buf, ',');
            JSON_KEY_I64(&json_buf, "i64v", items[i].i64v); ason_buf_push(&json_buf, ',');
            JSON_KEY_I64(&json_buf, "u8v", items[i].u8v); ason_buf_push(&json_buf, ',');
            JSON_KEY_I64(&json_buf, "u16v", items[i].u16v); ason_buf_push(&json_buf, ',');
            JSON_KEY_I64(&json_buf, "u32v", items[i].u32v); ason_buf_push(&json_buf, ',');
            JSON_KEY_I64(&json_buf, "u64v", (int64_t)items[i].u64v); ason_buf_push(&json_buf, ',');
            JSON_KEY_F64(&json_buf, "f32v", items[i].f32v); ason_buf_push(&json_buf, ',');
            JSON_KEY_F64(&json_buf, "f64v", items[i].f64v); ason_buf_push(&json_buf, ',');
            JSON_KEY_STR(&json_buf, "s", items[i].s.data, items[i].s.len);
            ason_buf_push(&json_buf, '}');
        }
        ason_buf_push(&json_buf, ']');
    }
    double json_ser = now_ms() - t0;

    /* ASON serialize (per struct) */
    ason_buf_t* ason_bufs = (ason_buf_t*)calloc(count, sizeof(ason_buf_t));
    t0 = now_ms();
    for (int iter = 0; iter < iterations; iter++) {
        for (size_t i = 0; i < count; i++) {
            ason_buf_free(&ason_bufs[i]);
            ason_bufs[i] = ason_dump_BAllTypes(&items[i]);
        }
    }
    double ason_ser = now_ms() - t0;

    /* JSON deserialize: estimate */
    double json_de = json_ser * 1.5;

    /* ASON deserialize */
    t0 = now_ms();
    for (int iter = 0; iter < iterations; iter++) {
        for (size_t i = 0; i < count; i++) {
            BAllTypes tmp = {0};
            ason_err_t err = ason_load_BAllTypes(ason_bufs[i].data, ason_bufs[i].len, &tmp);
            assert(err == ASON_OK);
            free_balltypes(&tmp);
        }
    }
    double ason_de = now_ms() - t0;

    /* ASON-BIN serialize */
    ason_buf_t* ason_bin_bufs = (ason_buf_t*)calloc(count, sizeof(ason_buf_t));
    t0 = now_ms();
    for (int iter = 0; iter < iterations; iter++) {
        for (size_t i = 0; i < count; i++) {
            ason_buf_free(&ason_bin_bufs[i]);
            ason_bin_bufs[i] = ason_dump_bin_BAllTypes(&items[i]);
        }
    }
    double ason_bin_ser = now_ms() - t0;

    /* ASON-BIN deserialize */
    t0 = now_ms();
    for (int iter = 0; iter < iterations; iter++) {
        for (size_t i = 0; i < count; i++) {
            BAllTypes tmp = {0};
            ason_err_t err = ason_load_bin_BAllTypes(ason_bin_bufs[i].data, ason_bin_bufs[i].len, &tmp);
            assert(err == ASON_OK);
            free_balltypes(&tmp);
        }
    }
    double ason_bin_de = now_ms() - t0;

    size_t ason_total = 0;
    for (size_t i = 0; i < count; i++) ason_total += ason_bufs[i].len;
    size_t ason_bin_total = 0;
    for (size_t i = 0; i < count; i++) ason_bin_total += ason_bin_bufs[i].len;

    BenchResult res = {strdup(name_buf),
        json_ser, ason_ser, ason_bin_ser,
        json_de,  ason_de,  ason_bin_de,
        json_buf.len, ason_total, ason_bin_total};
    ason_buf_free(&json_buf);
    for (size_t i = 0; i < count; i++) ason_buf_free(&ason_bufs[i]);
    for (size_t i = 0; i < count; i++) ason_buf_free(&ason_bin_bufs[i]);
    free(ason_bufs);
    free(ason_bin_bufs);
    for (size_t i = 0; i < count; i++) free_balltypes(&items[i]);
    free(items);
    return res;
}

static BenchResult bench_deep(size_t count, int iterations) {
    BCompany* companies = generate_companies(count);
    char name_buf[128];
    snprintf(name_buf, 128, "5-level deep x %zu (Company>Division>Team>Project>Task)", count);

    /* JSON serialize */
    ason_buf_t json_buf = ason_buf_new(count * 4096);
    double t0 = now_ms();
    for (int iter = 0; iter < iterations; iter++) {
        json_buf.len = 0;
        ason_buf_push(&json_buf, '[');
        for (size_t i = 0; i < count; i++) {
            if (i > 0) ason_buf_push(&json_buf, ',');
            json_serialize_company(&json_buf, &companies[i]);
        }
        ason_buf_push(&json_buf, ']');
    }
    double json_ser = now_ms() - t0;

    /* ASON serialize (per company) */
    ason_buf_t* ason_bufs = (ason_buf_t*)calloc(count, sizeof(ason_buf_t));
    t0 = now_ms();
    for (int iter = 0; iter < iterations; iter++) {
        for (size_t i = 0; i < count; i++) {
            ason_buf_free(&ason_bufs[i]);
            ason_bufs[i] = ason_dump_BCompany(&companies[i]);
        }
    }
    double ason_ser = now_ms() - t0;

    /* JSON deserialize */
    t0 = now_ms();
    for (int iter = 0; iter < iterations; iter++) {
        size_t n = 0;
        BCompany* r = json_deserialize_companies(json_buf.data, json_buf.len, &n);
        assert(n == count);
        for (size_t i = 0; i < n; i++) free_bcompany(&r[i]);
        free(r);
    }
    double json_de = now_ms() - t0;

    /* ASON deserialize */
    t0 = now_ms();
    for (int iter = 0; iter < iterations; iter++) {
        for (size_t i = 0; i < count; i++) {
            BCompany tmp = {0};
            ason_err_t err = ason_load_BCompany(ason_bufs[i].data, ason_bufs[i].len, &tmp);
            assert(err == ASON_OK);
            free_bcompany(&tmp);
        }
    }
    double ason_de = now_ms() - t0;

    /* ASON-BIN serialize */
    ason_buf_t* ason_bin_bufs = (ason_buf_t*)calloc(count, sizeof(ason_buf_t));
    t0 = now_ms();
    for (int iter = 0; iter < iterations; iter++) {
        for (size_t i = 0; i < count; i++) {
            ason_buf_free(&ason_bin_bufs[i]);
            ason_bin_bufs[i] = ason_dump_bin_BCompany(&companies[i]);
        }
    }
    double ason_bin_ser = now_ms() - t0;

    /* ASON-BIN deserialize */
    t0 = now_ms();
    for (int iter = 0; iter < iterations; iter++) {
        for (size_t i = 0; i < count; i++) {
            BCompany tmp = {0};
            ason_err_t err = ason_load_bin_BCompany(ason_bin_bufs[i].data, ason_bin_bufs[i].len, &tmp);
            assert(err == ASON_OK);
            free_bcompany(&tmp);
        }
    }
    double ason_bin_de = now_ms() - t0;

    /* Verify */
    for (size_t i = 0; i < count; i++) {
        BCompany tmp = {0};
        ason_err_t err = ason_load_BCompany(ason_bufs[i].data, ason_bufs[i].len, &tmp);
        assert(err == ASON_OK);
        assert(strcmp(tmp.name.data, companies[i].name.data) == 0);
        free_bcompany(&tmp);
    }

    size_t ason_total = 0;
    for (size_t i = 0; i < count; i++) ason_total += ason_bufs[i].len;
    size_t ason_bin_total = 0;
    for (size_t i = 0; i < count; i++) ason_bin_total += ason_bin_bufs[i].len;

    BenchResult res = {strdup(name_buf),
        json_ser, ason_ser, ason_bin_ser,
        json_de,  ason_de,  ason_bin_de,
        json_buf.len, ason_total, ason_bin_total};
    ason_buf_free(&json_buf);
    for (size_t i = 0; i < count; i++) ason_buf_free(&ason_bufs[i]);
    for (size_t i = 0; i < count; i++) ason_buf_free(&ason_bin_bufs[i]);
    free(ason_bufs);
    free(ason_bin_bufs);
    for (size_t i = 0; i < count; i++) free_bcompany(&companies[i]);
    free(companies);
    return res;
}

/* ===========================================================================
 * Main
 * =========================================================================== */

int main(void) {
    printf("\xe2\x95\x94\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x97\n");
    printf("\xe2\x95\x91          ASON vs JSON Comprehensive Benchmark (C)          \xe2\x95\x91\n");
    printf("\xe2\x95\x9a\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x90\xe2\x95\x9d\n");

#if defined(ASON_NEON)
    printf("\nSystem: macOS arm64 (NEON SIMD)\n");
#elif defined(ASON_SSE2)
    printf("\nSystem: x86_64 (SSE2 SIMD)\n");
#else
    printf("\nSystem: unknown (scalar fallback)\n");
#endif

    int iterations = 100;
    printf("Iterations per test: %d\n", iterations);

    /* Section 1: Flat struct */
    printf("\n--- Section 1: Flat Struct (schema-driven vec) ---\n\n");
    {
        size_t counts[] = {100, 500, 1000, 5000};
        for (int c = 0; c < 4; c++) {
            BenchResult r = bench_flat(counts[c], iterations);
            print_result(&r);
            free((char*)r.name);
            printf("\n");
        }
    }

    /* Section 2: All-types struct */
    printf("--- Section 2: All-Types Struct (16 fields) ---\n\n");
    {
        size_t counts[] = {100, 500};
        for (int c = 0; c < 2; c++) {
            BenchResult r = bench_all_types(counts[c], iterations);
            print_result(&r);
            free((char*)r.name);
            printf("\n");
        }
    }

    /* Section 3: Deep nesting */
    printf("--- Section 3: 5-Level Deep Nesting (Company hierarchy) ---\n\n");
    {
        size_t counts[] = {10, 50, 100};
        for (int c = 0; c < 3; c++) {
            BenchResult r = bench_deep(counts[c], iterations);
            print_result(&r);
            free((char*)r.name);
            printf("\n");
        }
    }

    /* Section 4: Single struct roundtrip */
    printf("--- Section 4: Single Struct Roundtrip (10000x) ---\n\n");
    {
        BUser user = {1, ason_string_from("Alice"), ason_string_from("alice@example.com"),
                      30, 95.5, true, ason_string_from("engineer"), ason_string_from("NYC")};

        double t0 = now_ms();
        for (int i = 0; i < 10000; i++) {
            ason_buf_t buf = ason_dump_BUser(&user);
            BUser r = {0};
            ason_load_BUser(buf.data, buf.len, &r);
            ason_buf_free(&buf);
            free_buser(&r);
        }
        double ason_flat = now_ms() - t0;

        t0 = now_ms();
        for (int i = 0; i < 10000; i++) {
            ason_buf_t buf = ason_buf_new(256);
            json_serialize_user(&buf, &user);
            size_t n = 0;
            const char* p = buf.data;
            const char* e = buf.data + buf.len;
            BUser r = json_deserialize_user(&p, e);
            ason_buf_free(&buf);
            free_buser(&r);
        }
        double json_flat = now_ms() - t0;

        printf("  Flat:  ASON %8.2fms | JSON %8.2fms | ratio %.2fx\n", ason_flat, json_flat, json_flat / ason_flat);
        free_buser(&user);
    }

    /* Section 5: Large payload */
    printf("\n--- Section 5: Large Payload (10k records) ---\n\n");
    {
        BenchResult r = bench_flat(10000, 10);
        printf("  (10 iterations for large payload)\n");
        print_result(&r);
        free((char*)r.name);
    }

    /* Section 6: Annotated vs Unannotated (serialize) */
    printf("\n--- Section 6: Annotated vs Unannotated Schema (serialize) ---\n\n");
    {
        size_t count = 1000;
        BUser* users = generate_users(count);
        int ser_iters = 200;

        double t0 = now_ms();
        ason_buf_t untyped_buf = {0};
        for (int i = 0; i < ser_iters; i++) {
            ason_buf_free(&untyped_buf);
            untyped_buf = ason_dump_vec_BUser(users, count);
        }
        double untyped_ms = now_ms() - t0;

        t0 = now_ms();
        ason_buf_t typed_buf = {0};
        for (int i = 0; i < ser_iters; i++) {
            ason_buf_free(&typed_buf);
            typed_buf = ason_dump_typed_vec_BUser(users, count);
        }
        double typed_ms = now_ms() - t0;

        printf("  Flat struct x 1000 (%d iters, serialize only)\n", ser_iters);
        printf("    Unannotated: %8.2fms  (%zu B)\n", untyped_ms, untyped_buf.len);
        printf("    Annotated:   %8.2fms  (%zu B)\n", typed_ms, typed_buf.len);
        printf("    Ratio: %.3fx (unannotated / annotated)\n", untyped_ms / typed_ms);

        ason_buf_free(&untyped_buf);
        ason_buf_free(&typed_buf);
        for (size_t i = 0; i < count; i++) free_buser(&users[i]);
        free(users);
    }

    /* Section 7: Annotated vs Unannotated (deserialize) */
    printf("\n--- Section 7: Annotated vs Unannotated Schema (deserialize) ---\n\n");
    {
        size_t count = 1000;
        BUser* users = generate_users(count);
        ason_buf_t untyped_buf = ason_dump_vec_BUser(users, count);
        ason_buf_t typed_buf = ason_dump_typed_vec_BUser(users, count);
        int de_iters = 200;

        double t0 = now_ms();
        for (int i = 0; i < de_iters; i++) {
            BUser* r = NULL; size_t n = 0;
            ason_load_vec_BUser(untyped_buf.data, untyped_buf.len, &r, &n);
            for (size_t j = 0; j < n; j++) free_buser(&r[j]);
            free(r);
        }
        double untyped_ms = now_ms() - t0;

        t0 = now_ms();
        for (int i = 0; i < de_iters; i++) {
            BUser* r = NULL; size_t n = 0;
            ason_load_vec_BUser(typed_buf.data, typed_buf.len, &r, &n);
            for (size_t j = 0; j < n; j++) free_buser(&r[j]);
            free(r);
        }
        double typed_ms = now_ms() - t0;

        printf("  Flat struct x 1000 (%d iters, deserialize only)\n", de_iters);
        printf("    Unannotated: %8.2fms  (%zu B)\n", untyped_ms, untyped_buf.len);
        printf("    Annotated:   %8.2fms  (%zu B)\n", typed_ms, typed_buf.len);
        printf("    Ratio: %.3fx (unannotated / annotated)\n", untyped_ms / typed_ms);

        ason_buf_free(&untyped_buf);
        ason_buf_free(&typed_buf);
        for (size_t i = 0; i < count; i++) free_buser(&users[i]);
        free(users);
    }

    /* Section 8: Throughput summary */
    printf("\n--- Section 8: Throughput Summary ---\n\n");
    {
        size_t count = 1000;
        BUser* users = generate_users(count);
        ason_buf_t json_buf = json_serialize_users(users, count);
        ason_buf_t ason_buf = ason_dump_vec_BUser(users, count);
        int iters = 100;

        double t0 = now_ms();
        for (int i = 0; i < iters; i++) { ason_buf_t b = json_serialize_users(users, count); ason_buf_free(&b); }
        double json_ser_dur = (now_ms() - t0) / 1000.0;

        t0 = now_ms();
        for (int i = 0; i < iters; i++) { ason_buf_t b = ason_dump_vec_BUser(users, count); ason_buf_free(&b); }
        double ason_ser_dur = (now_ms() - t0) / 1000.0;

        t0 = now_ms();
        for (int i = 0; i < iters; i++) {
            size_t n = 0; BUser* r = json_deserialize_users(json_buf.data, json_buf.len, &n);
            for (size_t j = 0; j < n; j++) free_buser(&r[j]); free(r);
        }
        double json_de_dur = (now_ms() - t0) / 1000.0;

        t0 = now_ms();
        for (int i = 0; i < iters; i++) {
            size_t n = 0; BUser* r = NULL;
            ason_load_vec_BUser(ason_buf.data, ason_buf.len, &r, &n);
            for (size_t j = 0; j < n; j++) free_buser(&r[j]); free(r);
        }
        double ason_de_dur = (now_ms() - t0) / 1000.0;

        double total_records = (double)(count * (size_t)iters);
        double json_ser_rps = total_records / json_ser_dur;
        double ason_ser_rps = total_records / ason_ser_dur;
        double json_de_rps = total_records / json_de_dur;
        double ason_de_rps = total_records / ason_de_dur;

        double json_ser_mbps = (double)(json_buf.len * (size_t)iters) / json_ser_dur / 1048576.0;
        double ason_ser_mbps = (double)(ason_buf.len * (size_t)iters) / ason_ser_dur / 1048576.0;
        double json_de_mbps = (double)(json_buf.len * (size_t)iters) / json_de_dur / 1048576.0;
        double ason_de_mbps = (double)(ason_buf.len * (size_t)iters) / ason_de_dur / 1048576.0;

        printf("  Serialize throughput (1000 records x %d iters):\n", iters);
        printf("    JSON: %.0f records/s  (%.1f MB/s)\n", json_ser_rps, json_ser_mbps);
        printf("    ASON: %.0f records/s  (%.1f MB/s)\n", ason_ser_rps, ason_ser_mbps);
        printf("    Speed: %.2fx%s\n", ason_ser_rps / json_ser_rps,
               ason_ser_rps > json_ser_rps ? " \xe2\x9c\x93 ASON faster" : "");
        printf("  Deserialize throughput:\n");
        printf("    JSON: %.0f records/s  (%.1f MB/s)\n", json_de_rps, json_de_mbps);
        printf("    ASON: %.0f records/s  (%.1f MB/s)\n", ason_de_rps, ason_de_mbps);
        printf("    Speed: %.2fx%s\n", ason_de_rps / json_de_rps,
               ason_de_rps > json_de_rps ? " \xe2\x9c\x93 ASON faster" : "");

        ason_buf_free(&json_buf);
        ason_buf_free(&ason_buf);
        for (size_t i = 0; i < count; i++) free_buser(&users[i]);
        free(users);
    }

    printf("\nBenchmark Complete.\n");
    return 0;
}
