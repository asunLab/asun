#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <math.h>
#include <stdbool.h>
#include "ason.h"

static int tests_passed = 0;
static int tests_failed = 0;

#define TEST(name) do { printf("  " #name "... "); fflush(stdout); } while(0)
#define PASS() do { printf("OK\n"); tests_passed++; } while(0)
#define FAIL(msg) do { printf("FAIL: %s\n", msg); tests_failed++; return; } while(0)
#define ASSERT_OK(e) do { if ((e) != ASON_OK) FAIL("decode error"); } while(0)
#define ASSERT_EQ_I(a, b) do { if ((a) != (b)) { printf("FAIL: %s=%lld != %s=%lld\n", #a, (long long)(a), #b, (long long)(b)); tests_failed++; return; } } while(0)
#define ASSERT_EQ_U(a, b) do { if ((a) != (b)) FAIL(#a " != " #b); } while(0)
#define ASSERT_EQ_S(a, b) do { if (strcmp((a), (b)) != 0) { printf("FAIL: %s='%s' != %s='%s'\n", #a, (a), #b, (b)); tests_failed++; return; } } while(0)
#define ASSERT_TRUE(x) do { if (!(x)) FAIL(#x " is false"); } while(0)
#define ASSERT_FALSE(x) do { if (x) FAIL(#x " is true"); } while(0)
#define ASSERT_NEAR(a, b, eps) do { if (fabs((a)-(b)) > (eps)) FAIL(#a " !~ " #b); } while(0)

/* ===========================================================================
 * Source/Target struct pairs
 * =========================================================================== */

/* Dim 1: trailing fields (vec + single) */
typedef struct { int64_t id; ason_string_t name; int32_t age; bool active; double score; } SrcFull;
ASON_FIELDS(SrcFull, 5,
    ASON_FIELD(SrcFull, id, "id", i64),
    ASON_FIELD(SrcFull, name, "name", str),
    ASON_FIELD(SrcFull, age, "age", i32),
    ASON_FIELD(SrcFull, active, "active", bool),
    ASON_FIELD(SrcFull, score, "score", f64))
ASON_VEC_STRUCT_DEFINE(SrcFull)

typedef struct { int64_t id; ason_string_t name; } DstMini;
ASON_FIELDS(DstMini, 2,
    ASON_FIELD(DstMini, id, "id", i64),
    ASON_FIELD(DstMini, name, "name", str))
ASON_VEC_STRUCT_DEFINE(DstMini)

static void free_srcfull(SrcFull* s) { ason_string_free(&s->name); }
static void free_dstmini(DstMini* d) { if (d->name.data) ason_string_free(&d->name); }

/* Dim 3: skip trailing arrays+map */
typedef struct { int64_t id; ason_string_t name; ason_vec_str tags; ason_vec_i64 scores; } SrcRich;
ASON_FIELDS(SrcRich, 4,
    ASON_FIELD(SrcRich, id, "id", i64),
    ASON_FIELD(SrcRich, name, "name", str),
    ASON_FIELD(SrcRich, tags, "tags", vec_str),
    ASON_FIELD(SrcRich, scores, "scores", vec_i64))
ASON_VEC_STRUCT_DEFINE(SrcRich)

typedef struct { int64_t id; ason_string_t name; } DstThin;
ASON_FIELDS(DstThin, 2,
    ASON_FIELD(DstThin, id, "id", i64),
    ASON_FIELD(DstThin, name, "name", str))
ASON_VEC_STRUCT_DEFINE(DstThin)

static void free_srcrich(SrcRich* r) {
    ason_string_free(&r->name);
    for (size_t i = 0; i < r->tags.len; i++) ason_string_free(&r->tags.data[i]);
    ason_vec_str_free(&r->tags);
    ason_vec_i64_free(&r->scores);
}
static void free_dstthin(DstThin* d) { ason_string_free(&d->name); }

/* Dim 4: nested struct fewer fields */
typedef struct { int64_t x; int64_t y; double z; bool w; } SrcInner;
ASON_FIELDS(SrcInner, 4,
    ASON_FIELD(SrcInner, x, "x", i64),
    ASON_FIELD(SrcInner, y, "y", i64),
    ASON_FIELD(SrcInner, z, "z", f64),
    ASON_FIELD(SrcInner, w, "w", bool))

typedef struct { ason_string_t name; SrcInner inner; bool flag; } SrcOuter;
ASON_FIELDS(SrcOuter, 3,
    ASON_FIELD(SrcOuter, name, "name", str),
    ASON_FIELD_STRUCT(SrcOuter, inner, "inner", &SrcInner_ason_desc),
    ASON_FIELD(SrcOuter, flag, "flag", bool))

typedef struct { int64_t x; int64_t y; } DstInner;
ASON_FIELDS(DstInner, 2,
    ASON_FIELD(DstInner, x, "x", i64),
    ASON_FIELD(DstInner, y, "y", i64))

typedef struct { ason_string_t name; DstInner inner; } DstOuter;
ASON_FIELDS(DstOuter, 2,
    ASON_FIELD(DstOuter, name, "name", str),
    ASON_FIELD_STRUCT(DstOuter, inner, "inner", &DstInner_ason_desc))

static void free_srcouter(SrcOuter* o) { ason_string_free(&o->name); }
static void free_dstouter(DstOuter* o) { ason_string_free(&o->name); }

/* Dim 5: vec of nested structs */
typedef struct { ason_string_t title; bool done; int64_t priority; double weight; } SrcTask;
ASON_FIELDS(SrcTask, 4,
    ASON_FIELD(SrcTask, title, "title", str),
    ASON_FIELD(SrcTask, done, "done", bool),
    ASON_FIELD(SrcTask, priority, "priority", i64),
    ASON_FIELD(SrcTask, weight, "weight", f64))
ASON_VEC_STRUCT_DEFINE(SrcTask)

typedef struct { ason_string_t name; ason_vec_SrcTask tasks; } SrcProject;
ASON_FIELDS(SrcProject, 2,
    ASON_FIELD(SrcProject, name, "name", str),
    ASON_FIELD_VEC_STRUCT(SrcProject, tasks, "tasks", SrcTask))
ASON_VEC_STRUCT_DEFINE(SrcProject)

typedef struct { ason_string_t title; bool done; } DstTask;
ASON_FIELDS(DstTask, 2,
    ASON_FIELD(DstTask, title, "title", str),
    ASON_FIELD(DstTask, done, "done", bool))
ASON_VEC_STRUCT_DEFINE(DstTask)

typedef struct { ason_string_t name; ason_vec_DstTask tasks; } DstProject;
ASON_FIELDS(DstProject, 2,
    ASON_FIELD(DstProject, name, "name", str),
    ASON_FIELD_VEC_STRUCT(DstProject, tasks, "tasks", DstTask))
ASON_VEC_STRUCT_DEFINE(DstProject)

static void free_srctask(SrcTask* t) { ason_string_free(&t->title); }
static void free_dsttask(DstTask* t) { ason_string_free(&t->title); }
static void free_srcproject(SrcProject* p) {
    ason_string_free(&p->name);
    for (size_t i = 0; i < p->tasks.len; i++) free_srctask(&p->tasks.data[i]);
    ason_vec_SrcTask_free(&p->tasks);
}
static void free_dstproject(DstProject* p) {
    ason_string_free(&p->name);
    for (size_t i = 0; i < p->tasks.len; i++) free_dsttask(&p->tasks.data[i]);
    ason_vec_DstTask_free(&p->tasks);
}

/* Dim 6: deep 3-level */
typedef struct { int64_t a; ason_string_t b; bool c; } L3Full;
ASON_FIELDS(L3Full, 3,
    ASON_FIELD(L3Full, a, "a", i64),
    ASON_FIELD(L3Full, b, "b", str),
    ASON_FIELD(L3Full, c, "c", bool))

typedef struct { ason_string_t name; L3Full sub; int64_t code; } L2Full;
ASON_FIELDS(L2Full, 3,
    ASON_FIELD(L2Full, name, "name", str),
    ASON_FIELD_STRUCT(L2Full, sub, "sub", &L3Full_ason_desc),
    ASON_FIELD(L2Full, code, "code", i64))

typedef struct { int64_t id; L2Full child; ason_string_t extra; } L1Full;
ASON_FIELDS(L1Full, 3,
    ASON_FIELD(L1Full, id, "id", i64),
    ASON_FIELD_STRUCT(L1Full, child, "child", &L2Full_ason_desc),
    ASON_FIELD(L1Full, extra, "extra", str))

typedef struct { int64_t a; } L3Thin;
ASON_FIELDS(L3Thin, 1, ASON_FIELD(L3Thin, a, "a", i64))

typedef struct { ason_string_t name; L3Thin sub; } L2Thin;
ASON_FIELDS(L2Thin, 2,
    ASON_FIELD(L2Thin, name, "name", str),
    ASON_FIELD_STRUCT(L2Thin, sub, "sub", &L3Thin_ason_desc))

typedef struct { int64_t id; L2Thin child; } L1Thin;
ASON_FIELDS(L1Thin, 2,
    ASON_FIELD(L1Thin, id, "id", i64),
    ASON_FIELD_STRUCT(L1Thin, child, "child", &L2Thin_ason_desc))

static void free_l1full(L1Full* l) {
    ason_string_free(&l->child.name);
    ason_string_free(&l->child.sub.b);
    ason_string_free(&l->extra);
}
static void free_l1thin(L1Thin* l) {
    ason_string_free(&l->child.name);
}

/* Dim 7: field reorder */
typedef struct { int64_t a; ason_string_t b; bool c; } OrderABC;
ASON_FIELDS(OrderABC, 3,
    ASON_FIELD(OrderABC, a, "a", i64),
    ASON_FIELD(OrderABC, b, "b", str),
    ASON_FIELD(OrderABC, c, "c", bool))

typedef struct { bool c; int64_t a; ason_string_t b; } OrderCAB;
ASON_FIELDS(OrderCAB, 3,
    ASON_FIELD(OrderCAB, c, "c", bool),
    ASON_FIELD(OrderCAB, a, "a", i64),
    ASON_FIELD(OrderCAB, b, "b", str))

static void free_orderabc(OrderABC* o) { ason_string_free(&o->b); }
static void free_ordercab(OrderCAB* o) { ason_string_free(&o->b); }

/* Dim 8: reorder + drop (vec) */
typedef struct { int64_t id; ason_string_t name; double score; bool active; int64_t level; } BigRec;
ASON_FIELDS(BigRec, 5,
    ASON_FIELD(BigRec, id, "id", i64),
    ASON_FIELD(BigRec, name, "name", str),
    ASON_FIELD(BigRec, score, "score", f64),
    ASON_FIELD(BigRec, active, "active", bool),
    ASON_FIELD(BigRec, level, "level", i64))
ASON_VEC_STRUCT_DEFINE(BigRec)

typedef struct { double score; int64_t id; } SmallReorder;
ASON_FIELDS(SmallReorder, 2,
    ASON_FIELD(SmallReorder, score, "score", f64),
    ASON_FIELD(SmallReorder, id, "id", i64))
ASON_VEC_STRUCT_DEFINE(SmallReorder)

static void free_bigrec(BigRec* r) { ason_string_free(&r->name); }

/* Dim 9: target extra fields */
typedef struct { int64_t id; ason_string_t name; } SrcSmall;
ASON_FIELDS(SrcSmall, 2,
    ASON_FIELD(SrcSmall, id, "id", i64),
    ASON_FIELD(SrcSmall, name, "name", str))

typedef struct { int64_t id; ason_string_t name; bool missing; double extra; } DstBig;
ASON_FIELDS(DstBig, 4,
    ASON_FIELD(DstBig, id, "id", i64),
    ASON_FIELD(DstBig, name, "name", str),
    ASON_FIELD(DstBig, missing, "missing", bool),
    ASON_FIELD(DstBig, extra, "extra", f64))

static void free_srcsmall(SrcSmall* s) { ason_string_free(&s->name); }
static void free_dstbig(DstBig* d) { ason_string_free(&d->name); }

/* Dim 10: optional */
typedef struct { int64_t id; ason_opt_str label; ason_opt_f64 score; bool flag; } SrcOpt;
ASON_FIELDS(SrcOpt, 4,
    ASON_FIELD(SrcOpt, id, "id", i64),
    ASON_FIELD(SrcOpt, label, "label", opt_str),
    ASON_FIELD(SrcOpt, score, "score", opt_f64),
    ASON_FIELD(SrcOpt, flag, "flag", bool))

typedef struct { int64_t id; ason_opt_str label; } DstFewerOpt;
ASON_FIELDS(DstFewerOpt, 2,
    ASON_FIELD(DstFewerOpt, id, "id", i64),
    ASON_FIELD(DstFewerOpt, label, "label", opt_str))

static void free_srcopt(SrcOpt* s) { if (s->label.has_value) ason_string_free(&s->label.value); }
static void free_dstfeweropt(DstFewerOpt* d) { if (d->label.has_value) ason_string_free(&d->label.value); }

/* Dim 11: special strings */
typedef struct { int64_t id; ason_string_t name; ason_string_t bio; } SrcStr;
ASON_FIELDS(SrcStr, 3,
    ASON_FIELD(SrcStr, id, "id", i64),
    ASON_FIELD(SrcStr, name, "name", str),
    ASON_FIELD(SrcStr, bio, "bio", str))

typedef struct { int64_t id; } DstId;
ASON_FIELDS(DstId, 1, ASON_FIELD(DstId, id, "id", i64))

static void free_srcstr(SrcStr* s) { ason_string_free(&s->name); ason_string_free(&s->bio); }

/* Dim 12: trailing arrays */
typedef struct { int64_t id; ason_vec_i64 nums; ason_vec_str tags; } SrcArr;
ASON_FIELDS(SrcArr, 3,
    ASON_FIELD(SrcArr, id, "id", i64),
    ASON_FIELD(SrcArr, nums, "nums", vec_i64),
    ASON_FIELD(SrcArr, tags, "tags", vec_str))
ASON_VEC_STRUCT_DEFINE(SrcArr)

static void free_srcarr(SrcArr* s) {
    ason_vec_i64_free(&s->nums);
    for (size_t i = 0; i < s->tags.len; i++) ason_string_free(&s->tags.data[i]);
    ason_vec_str_free(&s->tags);
}

/* Dim 14: negative */
typedef struct { int64_t a; int64_t b; double c; ason_string_t d; } SrcNeg;
ASON_FIELDS(SrcNeg, 4,
    ASON_FIELD(SrcNeg, a, "a", i64),
    ASON_FIELD(SrcNeg, b, "b", i64),
    ASON_FIELD(SrcNeg, c, "c", f64),
    ASON_FIELD(SrcNeg, d, "d", str))

typedef struct { int64_t a; int64_t b; } DstNegThin;
ASON_FIELDS(DstNegThin, 2,
    ASON_FIELD(DstNegThin, a, "a", i64),
    ASON_FIELD(DstNegThin, b, "b", i64))

static void free_srcneg(SrcNeg* s) { ason_string_free(&s->d); }

/* Dim 16: map */
typedef struct { int64_t id; ason_string_t name; ason_map_si meta; } SrcMap;
ASON_FIELDS(SrcMap, 3,
    ASON_FIELD(SrcMap, id, "id", i64),
    ASON_FIELD(SrcMap, name, "name", str),
    ASON_FIELD(SrcMap, meta, "meta", map_si))

typedef struct { int64_t id; ason_string_t name; } DstNoMap;
ASON_FIELDS(DstNoMap, 2,
    ASON_FIELD(DstNoMap, id, "id", i64),
    ASON_FIELD(DstNoMap, name, "name", str))

static void free_srcmap(SrcMap* s) {
    ason_string_free(&s->name);
    for (size_t i = 0; i < s->meta.len; i++) ason_string_free(&s->meta.data[i].key);
    ason_map_si_free(&s->meta);
}
static void free_dstnomap(DstNoMap* d) { ason_string_free(&d->name); }

/* Dim 20: bools */
typedef struct { int64_t id; bool a; bool b; bool c; } SrcBools;
ASON_FIELDS(SrcBools, 4,
    ASON_FIELD(SrcBools, id, "id", i64),
    ASON_FIELD(SrcBools, a, "a", bool),
    ASON_FIELD(SrcBools, b, "b", bool),
    ASON_FIELD(SrcBools, c, "c", bool))
ASON_VEC_STRUCT_DEFINE(SrcBools)

/* Dim 21: five fields */
typedef struct { int64_t a; ason_string_t b; double c; bool d; int64_t e; } SrcFive;
ASON_FIELDS(SrcFive, 5,
    ASON_FIELD(SrcFive, a, "a", i64),
    ASON_FIELD(SrcFive, b, "b", str),
    ASON_FIELD(SrcFive, c, "c", f64),
    ASON_FIELD(SrcFive, d, "d", bool),
    ASON_FIELD(SrcFive, e, "e", i64))

typedef struct { double c; } DstMiddle;
ASON_FIELDS(DstMiddle, 1, ASON_FIELD(DstMiddle, c, "c", f64))

typedef struct { int64_t e; } DstLast;
ASON_FIELDS(DstLast, 1, ASON_FIELD(DstLast, e, "e", i64))

static void free_srcfive(SrcFive* s) { ason_string_free(&s->b); }

/* Dim 23: no overlap */
typedef struct { int64_t x; ason_string_t y; } SrcAlpha;
ASON_FIELDS(SrcAlpha, 2,
    ASON_FIELD(SrcAlpha, x, "x", i64),
    ASON_FIELD(SrcAlpha, y, "y", str))

typedef struct { int64_t p; ason_string_t q; } DstBeta;
ASON_FIELDS(DstBeta, 2,
    ASON_FIELD(DstBeta, p, "p", i64),
    ASON_FIELD(DstBeta, q, "q", str))

static void free_srcalpha(SrcAlpha* s) { ason_string_free(&s->y); }
static void free_dstbeta(DstBeta* d) { if (d->q.data) ason_string_free(&d->q); }

/* Dim 24: nested array of structs */
typedef struct { ason_string_t name; ason_vec_str skills; int64_t years; double rating; } SrcWorker;
ASON_FIELDS(SrcWorker, 4,
    ASON_FIELD(SrcWorker, name, "name", str),
    ASON_FIELD(SrcWorker, skills, "skills", vec_str),
    ASON_FIELD(SrcWorker, years, "years", i64),
    ASON_FIELD(SrcWorker, rating, "rating", f64))
ASON_VEC_STRUCT_DEFINE(SrcWorker)

typedef struct { ason_string_t lead; ason_vec_SrcWorker workers; double budget; } SrcTeam;
ASON_FIELDS(SrcTeam, 3,
    ASON_FIELD(SrcTeam, lead, "lead", str),
    ASON_FIELD_VEC_STRUCT(SrcTeam, workers, "workers", SrcWorker),
    ASON_FIELD(SrcTeam, budget, "budget", f64))

typedef struct { ason_string_t name; ason_vec_str skills; } DstWorker;
ASON_FIELDS(DstWorker, 2,
    ASON_FIELD(DstWorker, name, "name", str),
    ASON_FIELD(DstWorker, skills, "skills", vec_str))
ASON_VEC_STRUCT_DEFINE(DstWorker)

typedef struct { ason_string_t lead; ason_vec_DstWorker workers; } DstTeam;
ASON_FIELDS(DstTeam, 2,
    ASON_FIELD(DstTeam, lead, "lead", str),
    ASON_FIELD_VEC_STRUCT(DstTeam, workers, "workers", DstWorker))

static void free_srcworker(SrcWorker* w) {
    ason_string_free(&w->name);
    for (size_t i = 0; i < w->skills.len; i++) ason_string_free(&w->skills.data[i]);
    ason_vec_str_free(&w->skills);
}
static void free_dstworker(DstWorker* w) {
    ason_string_free(&w->name);
    for (size_t i = 0; i < w->skills.len; i++) ason_string_free(&w->skills.data[i]);
    ason_vec_str_free(&w->skills);
}
static void free_srcteam(SrcTeam* t) {
    ason_string_free(&t->lead);
    for (size_t i = 0; i < t->workers.len; i++) free_srcworker(&t->workers.data[i]);
    ason_vec_SrcWorker_free(&t->workers);
}
static void free_dstteam(DstTeam* t) {
    ason_string_free(&t->lead);
    for (size_t i = 0; i < t->workers.len; i++) free_dstworker(&t->workers.data[i]);
    ason_vec_DstWorker_free(&t->workers);
}

/* Dim 26: wide (10→1) */
typedef struct { int64_t f1; ason_string_t f2; bool f3; int64_t f4; ason_string_t f5; bool f6; int64_t f7; ason_string_t f8; bool f9; int64_t f10; } SrcWide;
ASON_FIELDS(SrcWide, 10,
    ASON_FIELD(SrcWide, f1, "f1", i64), ASON_FIELD(SrcWide, f2, "f2", str),
    ASON_FIELD(SrcWide, f3, "f3", bool), ASON_FIELD(SrcWide, f4, "f4", i64),
    ASON_FIELD(SrcWide, f5, "f5", str), ASON_FIELD(SrcWide, f6, "f6", bool),
    ASON_FIELD(SrcWide, f7, "f7", i64), ASON_FIELD(SrcWide, f8, "f8", str),
    ASON_FIELD(SrcWide, f9, "f9", bool), ASON_FIELD(SrcWide, f10, "f10", i64))

typedef struct { int64_t f1; } DstNarrow;
ASON_FIELDS(DstNarrow, 1, ASON_FIELD(DstNarrow, f1, "f1", i64))

static void free_srcwide(SrcWide* w) {
    ason_string_free(&w->f2); ason_string_free(&w->f5); ason_string_free(&w->f8);
}

/* Dim 28: ason-like syntax in string */
typedef struct { int64_t id; ason_string_t data; ason_string_t code; } SrcAsonLike;
ASON_FIELDS(SrcAsonLike, 3,
    ASON_FIELD(SrcAsonLike, id, "id", i64),
    ASON_FIELD(SrcAsonLike, data, "data", str),
    ASON_FIELD(SrcAsonLike, code, "code", str))

static void free_srcasonlike(SrcAsonLike* s) { ason_string_free(&s->data); ason_string_free(&s->code); }

/* Dim 29: unicode */
typedef struct { int64_t id; ason_string_t name; ason_string_t bio; } SrcUnicode;
ASON_FIELDS(SrcUnicode, 3,
    ASON_FIELD(SrcUnicode, id, "id", i64),
    ASON_FIELD(SrcUnicode, name, "name", str),
    ASON_FIELD(SrcUnicode, bio, "bio", str))

static void free_srcunicode(SrcUnicode* s) { ason_string_free(&s->name); ason_string_free(&s->bio); }

/* Dim 30: roundtrip A→B→A */
typedef struct { int64_t id; ason_string_t name; bool active; } VerA;
ASON_FIELDS(VerA, 3,
    ASON_FIELD(VerA, id, "id", i64),
    ASON_FIELD(VerA, name, "name", str),
    ASON_FIELD(VerA, active, "active", bool))

typedef struct { int64_t id; ason_string_t name; } VerB;
ASON_FIELDS(VerB, 2,
    ASON_FIELD(VerB, id, "id", i64),
    ASON_FIELD(VerB, name, "name", str))

static void free_vera(VerA* v) { ason_string_free(&v->name); }
static void free_verb(VerB* v) { ason_string_free(&v->name); }

/* Dim 31: array in middle */
typedef struct { int64_t id; ason_vec_str items; int64_t score; } SrcArrMid;
ASON_FIELDS(SrcArrMid, 3,
    ASON_FIELD(SrcArrMid, id, "id", i64),
    ASON_FIELD(SrcArrMid, items, "items", vec_str),
    ASON_FIELD(SrcArrMid, score, "score", i64))
ASON_VEC_STRUCT_DEFINE(SrcArrMid)

typedef struct { int64_t id; ason_vec_str items; } DstArrMidThin;
ASON_FIELDS(DstArrMidThin, 2,
    ASON_FIELD(DstArrMidThin, id, "id", i64),
    ASON_FIELD(DstArrMidThin, items, "items", vec_str))
ASON_VEC_STRUCT_DEFINE(DstArrMidThin)

static void free_srcarrmid(SrcArrMid* s) {
    for (size_t i = 0; i < s->items.len; i++) ason_string_free(&s->items.data[i]);
    ason_vec_str_free(&s->items);
}
static void free_dstarrmidthin(DstArrMidThin* d) {
    for (size_t i = 0; i < d->items.len; i++) ason_string_free(&d->items.data[i]);
    ason_vec_str_free(&d->items);
}

/* Dim 32: skip nested struct as tuple */
typedef struct { int64_t a; ason_string_t b; } InnerSkip;
ASON_FIELDS(InnerSkip, 2,
    ASON_FIELD(InnerSkip, a, "a", i64),
    ASON_FIELD(InnerSkip, b, "b", str))

typedef struct { int64_t id; InnerSkip inner; ason_string_t tail; } SrcNested;
ASON_FIELDS(SrcNested, 3,
    ASON_FIELD(SrcNested, id, "id", i64),
    ASON_FIELD_STRUCT(SrcNested, inner, "inner", &InnerSkip_ason_desc),
    ASON_FIELD(SrcNested, tail, "tail", str))

typedef struct { int64_t id; } DstFlat;
ASON_FIELDS(DstFlat, 1, ASON_FIELD(DstFlat, id, "id", i64))

static void free_srcnested(SrcNested* s) { ason_string_free(&s->inner.b); ason_string_free(&s->tail); }

/* Dim 19: nested vec + trailing outer */
typedef struct { int64_t id; ason_string_t name; int32_t age; bool gender; } DetailFull;
ASON_FIELDS(DetailFull, 4,
    ASON_FIELD(DetailFull, id, "ID", i64),
    ASON_FIELD(DetailFull, name, "Name", str),
    ASON_FIELD(DetailFull, age, "Age", i32),
    ASON_FIELD(DetailFull, gender, "Gender", bool))
ASON_VEC_STRUCT_DEFINE(DetailFull)

typedef struct { ason_vec_DetailFull details; int64_t code; ason_string_t label; } UserFull2;
ASON_FIELDS(UserFull2, 3,
    ASON_FIELD_VEC_STRUCT(UserFull2, details, "details", DetailFull),
    ASON_FIELD(UserFull2, code, "code", i64),
    ASON_FIELD(UserFull2, label, "label", str))
ASON_VEC_STRUCT_DEFINE(UserFull2)

typedef struct { int64_t id; ason_string_t name; } PersonThin;
ASON_FIELDS(PersonThin, 2,
    ASON_FIELD(PersonThin, id, "ID", i64),
    ASON_FIELD(PersonThin, name, "Name", str))
ASON_VEC_STRUCT_DEFINE(PersonThin)

typedef struct { ason_vec_PersonThin details; } HumanThin;
ASON_FIELDS(HumanThin, 1,
    ASON_FIELD_VEC_STRUCT(HumanThin, details, "details", PersonThin))
ASON_VEC_STRUCT_DEFINE(HumanThin)

static void free_detailfull(DetailFull* d) { ason_string_free(&d->name); }
static void free_personthin(PersonThin* d) { ason_string_free(&d->name); }
static void free_userfull2(UserFull2* u) {
    for (size_t i = 0; i < u->details.len; i++) free_detailfull(&u->details.data[i]);
    ason_vec_DetailFull_free(&u->details);
    ason_string_free(&u->label);
}
static void free_humanthin(HumanThin* h) {
    for (size_t i = 0; i < h->details.len; i++) free_personthin(&h->details.data[i]);
    ason_vec_PersonThin_free(&h->details);
}

/* Dim 25: typed mixed */
typedef struct { int64_t a; ason_string_t b; double c; bool d; } SrcTyped;
ASON_FIELDS(SrcTyped, 4,
    ASON_FIELD(SrcTyped, a, "a", i64),
    ASON_FIELD(SrcTyped, b, "b", str),
    ASON_FIELD(SrcTyped, c, "c", f64),
    ASON_FIELD(SrcTyped, d, "d", bool))

typedef struct { ason_string_t b; bool d; } DstMixed;
ASON_FIELDS(DstMixed, 2,
    ASON_FIELD(DstMixed, b, "b", str),
    ASON_FIELD(DstMixed, d, "d", bool))

static void free_srctyped(SrcTyped* s) { ason_string_free(&s->b); }
static void free_dstmixed(DstMixed* d) { ason_string_free(&d->b); }

/* ===========================================================================
 * Tests
 * =========================================================================== */

void test_cross_trailing_vec(void) {
    TEST(trailing_fields_vec);
    SrcFull src[] = {
        {1, ason_string_from("Alice"), 30, true, 95.5},
        {2, ason_string_from("Bob"),   25, false, 87.0}
    };
    ason_buf_t buf = ason_encode_vec_SrcFull(src, 2);
    DstMini* dst = NULL; size_t n = 0;
    ASSERT_OK(ason_decode_vec_DstMini(buf.data, buf.len, &dst, &n));
    ASSERT_EQ_U(n, 2u);
    ASSERT_EQ_I(dst[0].id, 1); ASSERT_EQ_S(dst[0].name.data, "Alice");
    ASSERT_EQ_I(dst[1].id, 2); ASSERT_EQ_S(dst[1].name.data, "Bob");
    ason_buf_free(&buf);
    for (size_t i = 0; i < 2; i++) free_srcfull(&src[i]);
    for (size_t i = 0; i < n; i++) free_dstmini(&dst[i]);
    free(dst);
    PASS();
}

void test_cross_trailing_single(void) {
    TEST(trailing_fields_single);
    SrcFull src = {99, ason_string_from("Zara"), 40, true, 100.0};
    ason_buf_t buf = ason_encode_SrcFull(&src);
    DstMini dst = {0};
    ASSERT_OK(ason_decode_DstMini(buf.data, buf.len, &dst));
    ASSERT_EQ_I(dst.id, 99); ASSERT_EQ_S(dst.name.data, "Zara");
    ason_buf_free(&buf); free_srcfull(&src); free_dstmini(&dst);
    PASS();
}

void test_cross_skip_array_map(void) {
    TEST(skip_trailing_array_map);
    SrcRich src = {1, ason_string_from("Alice"), {0}, {0}};
    /* manually build tags */
    ason_vec_str_push(&src.tags, ason_string_from("go"));
    ason_vec_str_push(&src.tags, ason_string_from("rust"));
    ason_vec_i64_push(&src.scores, 90);
    ason_vec_i64_push(&src.scores, 85);
    ason_buf_t buf = ason_encode_SrcRich(&src);
    DstThin dst = {0};
    ASSERT_OK(ason_decode_DstThin(buf.data, buf.len, &dst));
    ASSERT_EQ_I(dst.id, 1); ASSERT_EQ_S(dst.name.data, "Alice");
    ason_buf_free(&buf); free_srcrich(&src); free_dstthin(&dst);
    PASS();
}

void test_cross_nested_fewer(void) {
    TEST(nested_fewer_fields);
    SrcOuter src = {ason_string_from("test"), {10, 20, 3.14, true}, true};
    ason_buf_t buf = ason_encode_SrcOuter(&src);
    DstOuter dst = {0};
    ASSERT_OK(ason_decode_DstOuter(buf.data, buf.len, &dst));
    ASSERT_EQ_S(dst.name.data, "test");
    ASSERT_EQ_I(dst.inner.x, 10); ASSERT_EQ_I(dst.inner.y, 20);
    ason_buf_free(&buf); free_srcouter(&src); free_dstouter(&dst);
    PASS();
}

void test_cross_vec_nested_skip(void) {
    TEST(vec_nested_skip);
    SrcProject src[] = {
        {ason_string_from("Alpha"), {0}},
        {ason_string_from("Beta"), {0}},
    };
    SrcTask t1 = {ason_string_from("Design"), true, 1, 0.5};
    SrcTask t2 = {ason_string_from("Code"), false, 2, 0.8};
    SrcTask t3 = {ason_string_from("Test"), false, 3, 1.0};
    ason_vec_SrcTask_push(&src[0].tasks, t1);
    ason_vec_SrcTask_push(&src[0].tasks, t2);
    ason_vec_SrcTask_push(&src[1].tasks, t3);
    ason_buf_t buf = ason_encode_vec_SrcProject(src, 2);
    DstProject* dst = NULL; size_t n = 0;
    ASSERT_OK(ason_decode_vec_DstProject(buf.data, buf.len, &dst, &n));
    ASSERT_EQ_U(n, 2u);
    ASSERT_EQ_S(dst[0].name.data, "Alpha");
    ASSERT_EQ_U(dst[0].tasks.len, 2u);
    ASSERT_EQ_S(dst[0].tasks.data[0].title.data, "Design");
    ASSERT_TRUE(dst[0].tasks.data[0].done);
    ASSERT_EQ_S(dst[0].tasks.data[1].title.data, "Code");
    ASSERT_FALSE(dst[0].tasks.data[1].done);
    ASSERT_EQ_S(dst[1].name.data, "Beta");
    ASSERT_EQ_U(dst[1].tasks.len, 1u);
    ason_buf_free(&buf);
    for (size_t i = 0; i < 2; i++) free_srcproject(&src[i]);
    for (size_t i = 0; i < n; i++) free_dstproject(&dst[i]);
    free(dst);
    PASS();
}

void test_cross_deep_3(void) {
    TEST(deep_3_levels);
    L1Full src = {1, {ason_string_from("mid"), {42, ason_string_from("hello"), true}, 7}, ason_string_from("dropped")};
    ason_buf_t buf = ason_encode_L1Full(&src);
    L1Thin dst = {0};
    ASSERT_OK(ason_decode_L1Thin(buf.data, buf.len, &dst));
    ASSERT_EQ_I(dst.id, 1);
    ASSERT_EQ_S(dst.child.name.data, "mid");
    ASSERT_EQ_I(dst.child.sub.a, 42);
    ason_buf_free(&buf); free_l1full(&src); free_l1thin(&dst);
    PASS();
}

void test_cross_reorder(void) {
    TEST(field_reorder);
    OrderABC src = {1, ason_string_from("hi"), true};
    ason_buf_t buf = ason_encode_OrderABC(&src);
    OrderCAB dst = {0};
    ASSERT_OK(ason_decode_OrderCAB(buf.data, buf.len, &dst));
    ASSERT_EQ_I(dst.a, 1); ASSERT_EQ_S(dst.b.data, "hi"); ASSERT_TRUE(dst.c);
    ason_buf_free(&buf); free_orderabc(&src); free_ordercab(&dst);
    PASS();
}

void test_cross_reorder_drop(void) {
    TEST(reorder_drop);
    BigRec src[] = {
        {1, ason_string_from("A"), 9.5, true, 3},
        {2, ason_string_from("B"), 8.0, false, 1}
    };
    ason_buf_t buf = ason_encode_vec_BigRec(src, 2);
    SmallReorder* dst = NULL; size_t n = 0;
    ASSERT_OK(ason_decode_vec_SmallReorder(buf.data, buf.len, &dst, &n));
    ASSERT_EQ_U(n, 2u);
    ASSERT_NEAR(dst[0].score, 9.5, 1e-10); ASSERT_EQ_I(dst[0].id, 1);
    ASSERT_NEAR(dst[1].score, 8.0, 1e-10); ASSERT_EQ_I(dst[1].id, 2);
    ason_buf_free(&buf);
    for (size_t i = 0; i < 2; i++) free_bigrec(&src[i]);
    free(dst);
    PASS();
}

void test_cross_target_extra(void) {
    TEST(target_extra_fields);
    SrcSmall src = {42, ason_string_from("Alice")};
    ason_buf_t buf = ason_encode_SrcSmall(&src);
    DstBig dst = {0};
    ASSERT_OK(ason_decode_DstBig(buf.data, buf.len, &dst));
    ASSERT_EQ_I(dst.id, 42); ASSERT_EQ_S(dst.name.data, "Alice");
    ASSERT_FALSE(dst.missing); ASSERT_NEAR(dst.extra, 0.0, 1e-10);
    ason_buf_free(&buf); free_srcsmall(&src); free_dstbig(&dst);
    PASS();
}

void test_cross_optional_skip(void) {
    TEST(optional_skip);
    SrcOpt src = {1, {true, ason_string_from("hello")}, {true, 95.5}, true};
    ason_buf_t buf = ason_encode_SrcOpt(&src);
    DstFewerOpt dst = {0};
    ASSERT_OK(ason_decode_DstFewerOpt(buf.data, buf.len, &dst));
    ASSERT_EQ_I(dst.id, 1);
    ASSERT_TRUE(dst.label.has_value);
    ASSERT_EQ_S(dst.label.value.data, "hello");
    ason_buf_free(&buf); free_srcopt(&src); free_dstfeweropt(&dst);
    PASS();
}

void test_cross_optional_nil(void) {
    TEST(optional_nil_skip);
    SrcOpt src = {2, {false, {0}}, {false, 0}, false};
    ason_buf_t buf = ason_encode_SrcOpt(&src);
    DstFewerOpt dst = {0};
    ASSERT_OK(ason_decode_DstFewerOpt(buf.data, buf.len, &dst));
    ASSERT_EQ_I(dst.id, 2);
    ASSERT_FALSE(dst.label.has_value);
    ason_buf_free(&buf); free_srcopt(&src); free_dstfeweropt(&dst);
    PASS();
}

void test_cross_special_string(void) {
    TEST(skip_special_string);
    SrcStr src = {1, ason_string_from("comma,here"), ason_string_from("paren(test)")};
    ason_buf_t buf = ason_encode_SrcStr(&src);
    DstId dst = {0};
    ASSERT_OK(ason_decode_DstId(buf.data, buf.len, &dst));
    ASSERT_EQ_I(dst.id, 1);
    ason_buf_free(&buf); free_srcstr(&src);
    PASS();
}

void test_cross_trailing_arrays(void) {
    TEST(skip_trailing_arrays);
    SrcArr src[] = {
        {1, {0}, {0}},
        {2, {0}, {0}},
    };
    ason_vec_i64_push(&src[0].nums, 1); ason_vec_i64_push(&src[0].nums, 2);
    ason_vec_str_push(&src[0].tags, ason_string_from("a"));
    ason_vec_i64_push(&src[1].nums, 4);
    ason_vec_str_push(&src[1].tags, ason_string_from("c"));
    ason_buf_t buf = ason_encode_vec_SrcArr(src, 2);
    DstId* dst = NULL; size_t n = 0;
    ASSERT_OK(ason_decode_vec_DstId(buf.data, buf.len, &dst, &n));
    ASSERT_EQ_U(n, 2u);
    ASSERT_EQ_I(dst[0].id, 1); ASSERT_EQ_I(dst[1].id, 2);
    ason_buf_free(&buf);
    for (size_t i = 0; i < 2; i++) free_srcarr(&src[i]);
    free(dst);
    PASS();
}

void test_cross_float_roundtrip(void) {
    TEST(float_roundtrip);
    typedef struct { int64_t id; double value; } FRT;
    /* encode manually */
    const char* input = "{id,value}:(1,3.14159)";
    FRT dst = {0};
    /* We can't use generated functions for local struct, so parse directly */
    /* Instead use SrcStr trick — encode a struct with float and decode same */
    SrcNeg src_f = {1, 0, 3.14159, ason_string_from("x")};
    ason_buf_t buf = ason_encode_SrcNeg(&src_f);
    SrcNeg dst_f = {0};
    ASSERT_OK(ason_decode_SrcNeg(buf.data, buf.len, &dst_f));
    ASSERT_NEAR(dst_f.c, 3.14159, 1e-10);
    ason_buf_free(&buf); free_srcneg(&src_f); free_srcneg(&dst_f);
    (void)input; (void)dst;
    PASS();
}

void test_cross_negative(void) {
    TEST(negative_skip);
    SrcNeg src = {-1, -999999, -3.14, ason_string_from("neg")};
    ason_buf_t buf = ason_encode_SrcNeg(&src);
    DstNegThin dst = {0};
    ASSERT_OK(ason_decode_DstNegThin(buf.data, buf.len, &dst));
    ASSERT_EQ_I(dst.a, -1); ASSERT_EQ_I(dst.b, -999999);
    ason_buf_free(&buf); free_srcneg(&src);
    PASS();
}

void test_cross_empty_string(void) {
    TEST(empty_string);
    SrcStr src = {1, ason_string_from(""), ason_string_from("")};
    ason_buf_t buf = ason_encode_SrcStr(&src);
    DstId dst = {0};
    ASSERT_OK(ason_decode_DstId(buf.data, buf.len, &dst));
    ASSERT_EQ_I(dst.id, 1);
    ason_buf_free(&buf); free_srcstr(&src);
    PASS();
}

void test_cross_skip_map(void) {
    TEST(skip_map);
    SrcMap src = {1, ason_string_from("Alice"), {0}};
    ason_map_si_entry_t e1 = {ason_string_from("age"), 30};
    ason_map_si_entry_t e2 = {ason_string_from("score"), 95};
    ason_map_si_push(&src.meta, e1);
    ason_map_si_push(&src.meta, e2);
    ason_buf_t buf = ason_encode_SrcMap(&src);
    DstNoMap dst = {0};
    ASSERT_OK(ason_decode_DstNoMap(buf.data, buf.len, &dst));
    ASSERT_EQ_I(dst.id, 1); ASSERT_EQ_S(dst.name.data, "Alice");
    ason_buf_free(&buf); free_srcmap(&src); free_dstnomap(&dst);
    PASS();
}

void test_cross_typed_vec(void) {
    TEST(typed_vec);
    SrcFull src[] = {{1, ason_string_from("Alice"), 30, true, 95.5}};
    ason_buf_t buf = ason_encode_typed_vec_SrcFull(src, 1);
    DstMini* dst = NULL; size_t n = 0;
    ASSERT_OK(ason_decode_vec_DstMini(buf.data, buf.len, &dst, &n));
    ASSERT_EQ_U(n, 1u);
    ASSERT_EQ_I(dst[0].id, 1); ASSERT_EQ_S(dst[0].name.data, "Alice");
    ason_buf_free(&buf); free_srcfull(&src[0]);
    for (size_t i = 0; i < n; i++) free_dstmini(&dst[i]);
    free(dst);
    PASS();
}

void test_cross_typed_single(void) {
    TEST(typed_single);
    SrcFull src = {42, ason_string_from("Bob"), 25, false, 88.0};
    ason_buf_t buf = ason_encode_typed_SrcFull(&src);
    DstMini dst = {0};
    ASSERT_OK(ason_decode_DstMini(buf.data, buf.len, &dst));
    ASSERT_EQ_I(dst.id, 42); ASSERT_EQ_S(dst.name.data, "Bob");
    ason_buf_free(&buf); free_srcfull(&src); free_dstmini(&dst);
    PASS();
}

void test_cross_nested_vec_trailing(void) {
    TEST(nested_vec_trailing_outer);
    UserFull2 src[] = {{.details = {0}, .code = 42, .label = ason_string_from("test")}};
    DetailFull d1 = {1, ason_string_from("Alice"), 30, true};
    DetailFull d2 = {2, ason_string_from("Bob"), 25, false};
    ason_vec_DetailFull_push(&src[0].details, d1);
    ason_vec_DetailFull_push(&src[0].details, d2);
    ason_buf_t buf = ason_encode_vec_UserFull2(src, 1);
    HumanThin* dst = NULL; size_t n = 0;
    ASSERT_OK(ason_decode_vec_HumanThin(buf.data, buf.len, &dst, &n));
    ASSERT_EQ_U(n, 1u);
    ASSERT_EQ_U(dst[0].details.len, 2u);
    ASSERT_EQ_I(dst[0].details.data[0].id, 1);
    ASSERT_EQ_S(dst[0].details.data[0].name.data, "Alice");
    ASSERT_EQ_I(dst[0].details.data[1].id, 2);
    ASSERT_EQ_S(dst[0].details.data[1].name.data, "Bob");
    ason_buf_free(&buf); free_userfull2(&src[0]);
    for (size_t i = 0; i < n; i++) free_humanthin(&dst[i]);
    free(dst);
    PASS();
}

void test_cross_skip_bools(void) {
    TEST(skip_bools);
    SrcBools src[] = {{1, true, false, true}, {2, false, true, false}};
    ason_buf_t buf = ason_encode_vec_SrcBools(src, 2);
    DstId* dst = NULL; size_t n = 0;
    ASSERT_OK(ason_decode_vec_DstId(buf.data, buf.len, &dst, &n));
    ASSERT_EQ_U(n, 2u);
    ASSERT_EQ_I(dst[0].id, 1); ASSERT_EQ_I(dst[1].id, 2);
    ason_buf_free(&buf); free(dst);
    PASS();
}

void test_cross_pick_middle(void) {
    TEST(pick_middle);
    SrcFive src = {1, ason_string_from("hi"), 3.14, true, 99};
    ason_buf_t buf = ason_encode_SrcFive(&src);
    DstMiddle dst = {0};
    ASSERT_OK(ason_decode_DstMiddle(buf.data, buf.len, &dst));
    ASSERT_NEAR(dst.c, 3.14, 1e-10);
    ason_buf_free(&buf); free_srcfive(&src);
    PASS();
}

void test_cross_pick_last(void) {
    TEST(pick_last);
    SrcFive src = {1, ason_string_from("hi"), 3.14, true, 42};
    ason_buf_t buf = ason_encode_SrcFive(&src);
    DstLast dst = {0};
    ASSERT_OK(ason_decode_DstLast(buf.data, buf.len, &dst));
    ASSERT_EQ_I(dst.e, 42);
    ason_buf_free(&buf); free_srcfive(&src);
    PASS();
}

void test_cross_no_overlap(void) {
    TEST(no_overlap);
    SrcAlpha src = {1, ason_string_from("hello")};
    ason_buf_t buf = ason_encode_SrcAlpha(&src);
    DstBeta dst = {0};
    ASSERT_OK(ason_decode_DstBeta(buf.data, buf.len, &dst));
    ASSERT_EQ_I(dst.p, 0);
    ASSERT_TRUE(dst.q.data == NULL || strcmp(dst.q.data, "") == 0);
    ason_buf_free(&buf); free_srcalpha(&src); free_dstbeta(&dst);
    PASS();
}

void test_cross_nested_array_structs(void) {
    TEST(nested_array_structs);
    SrcTeam src = {ason_string_from("Alice"), {0}, 100000.0};
    SrcWorker w1 = {ason_string_from("Bob"), {0}, 5, 4.5};
    ason_vec_str_push(&w1.skills, ason_string_from("go"));
    ason_vec_str_push(&w1.skills, ason_string_from("rust"));
    SrcWorker w2 = {ason_string_from("Carol"), {0}, 3, 3.8};
    ason_vec_str_push(&w2.skills, ason_string_from("python"));
    ason_vec_SrcWorker_push(&src.workers, w1);
    ason_vec_SrcWorker_push(&src.workers, w2);
    ason_buf_t buf = ason_encode_SrcTeam(&src);
    DstTeam dst = {0};
    ASSERT_OK(ason_decode_DstTeam(buf.data, buf.len, &dst));
    ASSERT_EQ_S(dst.lead.data, "Alice");
    ASSERT_EQ_U(dst.workers.len, 2u);
    ASSERT_EQ_S(dst.workers.data[0].name.data, "Bob");
    ASSERT_EQ_U(dst.workers.data[0].skills.len, 2u);
    ASSERT_EQ_S(dst.workers.data[0].skills.data[0].data, "go");
    ASSERT_EQ_S(dst.workers.data[1].name.data, "Carol");
    ASSERT_EQ_U(dst.workers.data[1].skills.len, 1u);
    ason_buf_free(&buf); free_srcteam(&src); free_dstteam(&dst);
    PASS();
}

void test_cross_typed_mixed(void) {
    TEST(typed_mixed);
    SrcTyped src = {1, ason_string_from("test"), 2.5, true};
    ason_buf_t buf = ason_encode_typed_SrcTyped(&src);
    DstMixed dst = {0};
    ASSERT_OK(ason_decode_DstMixed(buf.data, buf.len, &dst));
    ASSERT_EQ_S(dst.b.data, "test"); ASSERT_TRUE(dst.d);
    ason_buf_free(&buf); free_srctyped(&src); free_dstmixed(&dst);
    PASS();
}

void test_cross_many_trailing(void) {
    TEST(many_trailing);
    SrcWide src = {42, ason_string_from("a"), true, 4, ason_string_from("b"), false, 7, ason_string_from("c"), true, 10};
    ason_buf_t buf = ason_encode_SrcWide(&src);
    DstNarrow dst = {0};
    ASSERT_OK(ason_decode_DstNarrow(buf.data, buf.len, &dst));
    ASSERT_EQ_I(dst.f1, 42);
    ason_buf_free(&buf); free_srcwide(&src);
    PASS();
}

void test_cross_single_row(void) {
    TEST(vec_single_row);
    SrcFull src[] = {{1, ason_string_from("Alice"), 30, true, 95.5}};
    ason_buf_t buf = ason_encode_vec_SrcFull(src, 1);
    DstMini* dst = NULL; size_t n = 0;
    ASSERT_OK(ason_decode_vec_DstMini(buf.data, buf.len, &dst, &n));
    ASSERT_EQ_U(n, 1u);
    ASSERT_EQ_I(dst[0].id, 1); ASSERT_EQ_S(dst[0].name.data, "Alice");
    ason_buf_free(&buf); free_srcfull(&src[0]);
    for (size_t i = 0; i < n; i++) free_dstmini(&dst[i]);
    free(dst);
    PASS();
}

void test_cross_ason_syntax_string(void) {
    TEST(ason_syntax_string);
    SrcAsonLike src = {1, ason_string_from("{a,b}:(1,2)"), ason_string_from("[(x,y)]")};
    ason_buf_t buf = ason_encode_SrcAsonLike(&src);
    DstId dst = {0};
    ASSERT_OK(ason_decode_DstId(buf.data, buf.len, &dst));
    ASSERT_EQ_I(dst.id, 1);
    ason_buf_free(&buf); free_srcasonlike(&src);
    PASS();
}

void test_cross_unicode(void) {
    TEST(unicode_trailing);
    SrcUnicode src = {1, ason_string_from("日本語"), ason_string_from("中文描述")};
    ason_buf_t buf = ason_encode_SrcUnicode(&src);
    DstId dst = {0};
    ASSERT_OK(ason_decode_DstId(buf.data, buf.len, &dst));
    ASSERT_EQ_I(dst.id, 1);
    ason_buf_free(&buf); free_srcunicode(&src);
    PASS();
}

void test_cross_roundtrip_abba(void) {
    TEST(roundtrip_abba);
    VerA a = {1, ason_string_from("test"), true};
    ason_buf_t ba = ason_encode_VerA(&a);
    VerB b = {0};
    ASSERT_OK(ason_decode_VerB(ba.data, ba.len, &b));
    ASSERT_EQ_I(b.id, 1); ASSERT_EQ_S(b.name.data, "test");
    ason_buf_t bb = ason_encode_VerB(&b);
    VerA a2 = {0};
    ASSERT_OK(ason_decode_VerA(bb.data, bb.len, &a2));
    ASSERT_EQ_I(a2.id, 1); ASSERT_EQ_S(a2.name.data, "test"); ASSERT_FALSE(a2.active);
    ason_buf_free(&ba); ason_buf_free(&bb);
    free_vera(&a); free_verb(&b); free_vera(&a2);
    PASS();
}

void test_cross_empty_array_middle(void) {
    TEST(empty_array_middle);
    SrcArrMid src[] = {
        {1, {0}, 10},
        {2, {0}, 20},
    };
    ason_vec_str_push(&src[1].items, ason_string_from("a"));
    ason_vec_str_push(&src[1].items, ason_string_from("b"));
    ason_buf_t buf = ason_encode_vec_SrcArrMid(src, 2);
    DstArrMidThin* dst = NULL; size_t n = 0;
    ASSERT_OK(ason_decode_vec_DstArrMidThin(buf.data, buf.len, &dst, &n));
    ASSERT_EQ_U(n, 2u);
    ASSERT_EQ_I(dst[0].id, 1); ASSERT_EQ_U(dst[0].items.len, 0u);
    ASSERT_EQ_I(dst[1].id, 2); ASSERT_EQ_U(dst[1].items.len, 2u);
    ASSERT_EQ_S(dst[1].items.data[0].data, "a");
    ason_buf_free(&buf);
    for (size_t i = 0; i < 2; i++) free_srcarrmid(&src[i]);
    for (size_t i = 0; i < n; i++) free_dstarrmidthin(&dst[i]);
    free(dst);
    PASS();
}

void test_cross_skip_nested_tuple(void) {
    TEST(skip_nested_tuple);
    SrcNested src = {1, {10, ason_string_from("nested")}, ason_string_from("end")};
    ason_buf_t buf = ason_encode_SrcNested(&src);
    DstFlat dst = {0};
    ASSERT_OK(ason_decode_DstFlat(buf.data, buf.len, &dst));
    ASSERT_EQ_I(dst.id, 1);
    ason_buf_free(&buf); free_srcnested(&src);
    PASS();
}

void test_cross_many_rows(void) {
    TEST(many_rows);
    SrcFull* src = (SrcFull*)calloc(100, sizeof(SrcFull));
    for (int i = 0; i < 100; i++) {
        src[i].id = i;
        src[i].name = ason_string_from("user");
        src[i].age = i;
        src[i].active = i % 2 == 0;
        src[i].score = i * 0.1;
    }
    ason_buf_t buf = ason_encode_vec_SrcFull(src, 100);
    DstMini* dst = NULL; size_t n = 0;
    ASSERT_OK(ason_decode_vec_DstMini(buf.data, buf.len, &dst, &n));
    ASSERT_EQ_U(n, 100u);
    for (int i = 0; i < 100; i++) {
        ASSERT_EQ_I(dst[i].id, i);
        ASSERT_EQ_S(dst[i].name.data, "user");
    }
    ason_buf_free(&buf);
    for (int i = 0; i < 100; i++) free_srcfull(&src[i]);
    free(src);
    for (size_t i = 0; i < n; i++) free_dstmini(&dst[i]);
    free(dst);
    PASS();
}

void test_cross_typed_subset_reorder(void) {
    TEST(typed_subset_reorder);
    BigRec src[] = {
        {1, ason_string_from("A"), 9.5, true, 3},
        {2, ason_string_from("B"), 8.0, false, 1}
    };
    ason_buf_t buf = ason_encode_typed_vec_BigRec(src, 2);
    SmallReorder* dst = NULL; size_t n = 0;
    ASSERT_OK(ason_decode_vec_SmallReorder(buf.data, buf.len, &dst, &n));
    ASSERT_EQ_U(n, 2u);
    ASSERT_NEAR(dst[0].score, 9.5, 1e-10); ASSERT_EQ_I(dst[0].id, 1);
    ASSERT_NEAR(dst[1].score, 8.0, 1e-10); ASSERT_EQ_I(dst[1].id, 2);
    ason_buf_free(&buf);
    for (size_t i = 0; i < 2; i++) free_bigrec(&src[i]);
    free(dst);
    PASS();
}

void test_cross_zero_value(void) {
    TEST(zero_value);
    SrcFull src = {0, ason_string_from(""), 0, false, 0.0};
    ason_buf_t buf = ason_encode_SrcFull(&src);
    DstMini dst = {0};
    ASSERT_OK(ason_decode_DstMini(buf.data, buf.len, &dst));
    ASSERT_EQ_I(dst.id, 0);
    ASSERT_TRUE(dst.name.data == NULL || strcmp(dst.name.data, "") == 0);
    ason_buf_free(&buf); free_srcfull(&src); free_dstmini(&dst);
    PASS();
}

int main(void) {
    printf("=== ASON C Cross-Compat Test Suite ===\n\n");

    test_cross_trailing_vec();
    test_cross_trailing_single();
    test_cross_skip_array_map();
    test_cross_nested_fewer();
    test_cross_vec_nested_skip();
    test_cross_deep_3();
    test_cross_reorder();
    test_cross_reorder_drop();
    test_cross_target_extra();
    test_cross_optional_skip();
    test_cross_optional_nil();
    test_cross_special_string();
    test_cross_trailing_arrays();
    test_cross_float_roundtrip();
    test_cross_negative();
    test_cross_empty_string();
    test_cross_skip_map();
    test_cross_typed_vec();
    test_cross_typed_single();
    test_cross_nested_vec_trailing();
    test_cross_skip_bools();
    test_cross_pick_middle();
    test_cross_pick_last();
    test_cross_no_overlap();
    test_cross_nested_array_structs();
    test_cross_typed_mixed();
    test_cross_many_trailing();
    test_cross_single_row();
    test_cross_ason_syntax_string();
    test_cross_unicode();
    test_cross_roundtrip_abba();
    test_cross_empty_array_middle();
    test_cross_skip_nested_tuple();
    test_cross_many_rows();
    test_cross_typed_subset_reorder();
    test_cross_zero_value();

    printf("\n=== Results: %d passed, %d failed ===\n", tests_passed, tests_failed);
    return tests_failed > 0 ? 1 : 0;
}
