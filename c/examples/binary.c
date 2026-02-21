/*
 * ASON Binary Format Example (C)
 *
 * Demonstrates ason_dump_bin_<T> / ason_load_bin_<T> for fast binary
 * serialization. Wire format: little-endian fixed-width scalars,
 * 4-byte length-prefixed strings and arrays — no field names, purely
 * positional.
 *
 * Usage: add ASON_FIELDS_BIN(MyStruct, N) right after ASON_FIELDS(...)
 * to inject:
 *   ason_buf_t ason_dump_bin_MyStruct(const MyStruct*)
 *   ason_err_t ason_load_bin_MyStruct(const char* data, size_t len, MyStruct*)
 *   ason_buf_t ason_dump_bin_vec_MyStruct(const MyStruct*, size_t)
 *   ason_err_t ason_load_bin_vec_MyStruct(const char*, size_t, MyStruct**, size_t*)
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>
#include <assert.h>
#include "ason.h"

/* ------------------------------------------------------------------ */
/* Example structs                                                       */
/* ------------------------------------------------------------------ */

typedef struct {
    ason_string_t name;
    int64_t       age;
    double        score;
    bool          active;
} User;

ASON_FIELDS(User, 4,
    ASON_FIELD(User, name,   "name",   str),
    ASON_FIELD(User, age,    "age",    i64),
    ASON_FIELD(User, score,  "score",  f64),
    ASON_FIELD(User, active, "active", bool))

ASON_FIELDS_BIN(User, 4)   /* injects ason_dump_bin_User / ason_load_bin_User */

/* ------------------------------------------------------------------ */

typedef struct {
    ason_string_t title;
    int32_t       year;
    double        rating;
    ason_vec_str  tags;
} Movie;

ASON_FIELDS(Movie, 4,
    ASON_FIELD(Movie, title,  "title",  str),
    ASON_FIELD(Movie, year,   "year",   i32),
    ASON_FIELD(Movie, rating, "rating", f64),
    ASON_FIELD(Movie, tags,   "tags",   vec_str))

ASON_FIELDS_BIN(Movie, 4)

/* ------------------------------------------------------------------ */
/* Helpers                                                              */
/* ------------------------------------------------------------------ */

static void free_user(User* u) {
    ason_string_free(&u->name);
}

static void free_movie(Movie* m) {
    ason_string_free(&m->title);
    for (size_t i = 0; i < m->tags.len; i++) ason_string_free(&m->tags.data[i]);
    ason_vec_str_free(&m->tags);
}

static void print_user(const User* u) {
    printf("  User { name=\"%.*s\", age=%lld, score=%.2f, active=%s }\n",
           (int)u->name.len, u->name.data,
           (long long)u->age, u->score,
           u->active ? "true" : "false");
}

static void print_movie(const Movie* m) {
    printf("  Movie { title=\"%.*s\", year=%d, rating=%.1f, tags=[",
           (int)m->title.len, m->title.data, m->year, m->rating);
    for (size_t i = 0; i < m->tags.len; i++) {
        if (i) printf(", ");
        printf("\"%.*s\"", (int)m->tags.data[i].len, m->tags.data[i].data);
    }
    printf("] }\n");
}

/* ------------------------------------------------------------------ */
/* main                                                                 */
/* ------------------------------------------------------------------ */

int main(void) {
    printf("=== ASON Binary Format Example (C) ===\n\n");

    /* ---- 1. Single struct roundtrip ---- */
    printf("1. Single struct roundtrip\n");
    {
        User u = {
            .name   = ason_string_from_len("Alice", strlen("Alice")),
            .age    = 30,
            .score  = 95.5,
            .active = true,
        };

        /* serialize to binary */
        ason_buf_t buf = ason_dump_bin_User(&u);
        printf("   Serialized %zu bytes\n", buf.len);

        /* deserialize */
        User u2 = {0};
        ason_err_t err = ason_load_bin_User(buf.data, buf.len, &u2);
        assert(err == ASON_OK);
        print_user(&u2);

        assert(u2.age    == u.age);
        assert(u2.score  == u.score);
        assert(u2.active == u.active);
        assert(u2.name.len == u.name.len);
        assert(memcmp(u2.name.data, u.name.data, u.name.len) == 0);

        ason_buf_free(&buf);
        free_user(&u);
        free_user(&u2);
    }
    printf("   OK\n\n");

    /* ---- 2. Struct with vector fields ---- */
    printf("2. Struct with vector fields (Movie)\n");
    {
        /* build tags vec */
        ason_vec_str tags = {0};
        ason_vec_str_push(&tags, ason_string_from_len("sci-fi", strlen("sci-fi")));
        ason_vec_str_push(&tags, ason_string_from_len("thriller", strlen("thriller")));
        ason_vec_str_push(&tags, ason_string_from_len("masterpiece", strlen("masterpiece")));

        Movie m = {
            .title  = ason_string_from_len("Interstellar", strlen("Interstellar")),
            .year   = 2014,
            .rating = 9.0,
            .tags   = tags,
        };

        ason_buf_t buf = ason_dump_bin_Movie(&m);
        printf("   Serialized %zu bytes\n", buf.len);

        Movie m2 = {0};
        ason_err_t err = ason_load_bin_Movie(buf.data, buf.len, &m2);
        assert(err == ASON_OK);
        print_movie(&m2);

        assert(m2.year   == m.year);
        assert(m2.rating == m.rating);
        assert(m2.tags.len == m.tags.len);

        ason_buf_free(&buf);
        free_movie(&m);
        free_movie(&m2);
    }
    printf("   OK\n\n");

    /* ---- 3. Array of structs (batch) ---- */
    printf("3. Array of structs (vec roundtrip)\n");
    {
        const int N = 5;
        User users[5];
        char name_buf[8];
        for (int i = 0; i < N; i++) {
            snprintf(name_buf, 8, "User%d", i);
            users[i].name   = ason_string_from_len(name_buf, strlen(name_buf));
            users[i].age    = 20 + i;
            users[i].score  = 80.0 + i * 2.5;
            users[i].active = (i % 2 == 0);
        }

        ason_buf_t buf = ason_dump_bin_vec_User(users, N);
        printf("   Serialized %d users → %zu bytes (%.1f B/user)\n",
               N, buf.len, (double)buf.len / N);

        User* out = NULL;
        size_t out_n = 0;
        ason_err_t err = ason_load_bin_vec_User(buf.data, buf.len, &out, &out_n);
        assert(err == ASON_OK);
        assert(out_n == (size_t)N);

        for (size_t i = 0; i < out_n; i++) {
            print_user(&out[i]);
            free_user(&out[i]);
        }
        free(out);
        ason_buf_free(&buf);
        for (int i = 0; i < N; i++) free_user(&users[i]);
    }
    printf("   OK\n\n");

    /* ---- 4. Compare size with ASON text ---- */
    printf("4. Size comparison: ASON-text vs ASON-BIN\n");
    {
        ason_vec_str tags = {0};
        ason_vec_str_push(&tags, ason_string_from_len("action", strlen("action")));
        ason_vec_str_push(&tags, ason_string_from_len("adventure", strlen("adventure")));

        Movie m = {
            .title  = ason_string_from_len("The Dark Knight", strlen("The Dark Knight")),
            .year   = 2008,
            .rating = 9.0,
            .tags   = tags,
        };

        ason_buf_t text_buf = ason_dump_Movie(&m);
        ason_buf_t bin_buf  = ason_dump_bin_Movie(&m);

        printf("   ASON text : %zu bytes: %.*s\n",
               text_buf.len, (int)text_buf.len, text_buf.data);
        printf("   ASON bin  : %zu bytes (%.0f%% smaller)\n",
               bin_buf.len,
               (1.0 - (double)bin_buf.len / text_buf.len) * 100.0);

        ason_buf_free(&text_buf);
        ason_buf_free(&bin_buf);
        free_movie(&m);
    }
    printf("\nAll examples passed.\n");
    return 0;
}
