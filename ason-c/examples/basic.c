#include <stdio.h>
#include <assert.h>
#include <string.h>
#include "ason.h"

/* ============================================================================
 * Struct definitions
 * ============================================================================ */

typedef struct {
    int64_t id;
    ason_string_t name;
    bool active;
} User;

ASON_FIELDS(User, 3,
    ASON_FIELD(User, id,     "id",     i64),
    ASON_FIELD(User, name,   "name",   str),
    ASON_FIELD(User, active, "active", bool))
ASON_FIELDS_BIN(User, 3)

typedef struct {
    int64_t id;
    ason_opt_str label;
} Item;

ASON_FIELDS(Item, 2,
    ASON_FIELD(Item, id,    "id",    i64),
    ASON_FIELD(Item, label, "label", opt_str))

typedef struct {
    ason_string_t name;
    ason_vec_str tags;
} Tagged;

ASON_FIELDS(Tagged, 2,
    ASON_FIELD(Tagged, name, "name", str),
    ASON_FIELD(Tagged, tags, "tags", vec_str))

/* ============================================================================
 * Main
 * ============================================================================ */

int main(void) {
    printf("=== ASON C Basic Examples ===\n\n");

    /* 1. Serialize a single struct */
    printf("1. Serialize single struct:\n");
    User user = {1, ason_string_from("Alice"), true};
    ason_buf_t buf = ason_encode_User(&user);
    printf("   %.*s\n\n", (int)buf.len, buf.data);
    ason_buf_free(&buf);

    /* 2. Serialize with type annotations */
    printf("2. Serialize with type annotations:\n");
    buf = ason_encode_typed_User(&user);
    printf("   %.*s\n\n", (int)buf.len, buf.data);
    assert(buf.len > 0);
    ason_buf_free(&buf);

    /* 3. Deserialize from ASON */
    printf("3. Deserialize single struct:\n");
    {
        const char* input = "{id:int,name:str,active:bool}:(1,Alice,true)";
        User u2 = {0};
        ason_err_t err = ason_decode_User(input, strlen(input), &u2);
        assert(err == ASON_OK);
        printf("   id=%lld name=%s active=%s\n\n",
               (long long)u2.id, u2.name.data, u2.active ? "true" : "false");
        ason_string_free(&u2.name);
    }

    /* 4. Serialize a vec of structs */
    printf("4. Serialize vec (schema-driven):\n");
    {
        User users[3] = {
            {1, ason_string_from("Alice"), true},
            {2, ason_string_from("Bob"), false},
            {3, ason_string_from("Carol Smith"), true},
        };
        buf = ason_encode_vec_User(users, 3);
        printf("   %.*s\n\n", (int)buf.len, buf.data);
        ason_buf_free(&buf);
        for (int i = 0; i < 3; i++) ason_string_free(&users[i].name);
    }

    /* 5. Deserialize vec */
    printf("5. Deserialize vec:\n");
    {
        const char* input = "[{id:int,name:str,active:bool}]:(1,Alice,true),(2,Bob,false),(3,\"Carol Smith\",true)";
        User* users = NULL;
        size_t count = 0;
        ason_err_t err = ason_decode_vec_User(input, strlen(input), &users, &count);
        assert(err == ASON_OK);
        for (size_t i = 0; i < count; i++) {
            printf("   id=%lld name=%s active=%s\n",
                   (long long)users[i].id, users[i].name.data,
                   users[i].active ? "true" : "false");
            ason_string_free(&users[i].name);
        }
        free(users);
        printf("\n");
    }

    /* 6. Multiline format */
    printf("6. Multiline format:\n");
    {
        const char* input = "[{id, name, active}]:\n  (1, Alice, true),\n  (2, Bob, false),\n  (3, \"Carol Smith\", true)";
        User* users = NULL;
        size_t count = 0;
        ason_err_t err = ason_decode_vec_User(input, strlen(input), &users, &count);
        assert(err == ASON_OK);
        for (size_t i = 0; i < count; i++) {
            printf("   id=%lld name=%s active=%s\n",
                   (long long)users[i].id, users[i].name.data,
                   users[i].active ? "true" : "false");
            ason_string_free(&users[i].name);
        }
        free(users);
        printf("\n");
    }

    /* 7. Roundtrip (ASON-text + ASON-bin) */
    printf("7. Roundtrip (ASON-text vs ASON-bin):\n");
    {
        User orig = {42, ason_string_from("Test User"), true};
        /* ASON text roundtrip */
        buf = ason_encode_User(&orig);
        User decoded = {0};
        ason_err_t err = ason_decode_User(buf.data, buf.len, &decoded);
        assert(err == ASON_OK);
        assert(decoded.id == 42);
        assert(strcmp(decoded.name.data, "Test User") == 0);
        assert(decoded.active == true);
        printf("   ASON text:    %.*s (%zu B)\n", (int)buf.len, buf.data, buf.len);
        ason_string_free(&decoded.name);

        /* ASON binary roundtrip */
        ason_buf_t bin = ason_encode_bin_User(&orig);
        User decoded_bin = {0};
        err = ason_decode_bin_User(bin.data, bin.len, &decoded_bin);
        assert(err == ASON_OK);
        assert(decoded_bin.id == 42);
        assert(strcmp(decoded_bin.name.data, "Test User") == 0);
        assert(decoded_bin.active == true);
        printf("   ASON binary:  %zu B\n", bin.len);
        printf("   BIN is %.0f%% smaller than text\n",
               (1.0 - (double)bin.len / (double)buf.len) * 100.0);
        printf("   \xe2\x9c\x93 both formats roundtrip OK\n\n");
        ason_buf_free(&buf);
        ason_buf_free(&bin);
        ason_string_free(&orig.name);
        ason_string_free(&decoded_bin.name);
    }

    /* 8. Optional fields */
    printf("8. Optional fields:\n");
    {
        const char* input1 = "{id,label}:(1,hello)";
        Item item1 = {0};
        ason_err_t err = ason_decode_Item(input1, strlen(input1), &item1);
        assert(err == ASON_OK);
        printf("   with value: id=%lld label=%s\n",
               (long long)item1.id, item1.label.has_value ? item1.label.value.data : "(null)");
        if (item1.label.has_value) ason_string_free(&item1.label.value);

        const char* input2 = "{id,label}:(2,)";
        Item item2 = {0};
        err = ason_decode_Item(input2, strlen(input2), &item2);
        assert(err == ASON_OK);
        assert(!item2.label.has_value);
        printf("   with null:  id=%lld label=%s\n\n",
               (long long)item2.id, item2.label.has_value ? item2.label.value.data : "(null)");
    }

    /* 9. Array fields */
    printf("9. Array fields:\n");
    {
        const char* input = "{name,tags}:(Alice,[rust,go,python])";
        Tagged t = {0};
        ason_err_t err = ason_decode_Tagged(input, strlen(input), &t);
        assert(err == ASON_OK);
        printf("   name=%s tags=[", t.name.data);
        for (size_t i = 0; i < t.tags.len; i++) {
            if (i > 0) printf(",");
            printf("%s", t.tags.data[i].data);
        }
        printf("]\n\n");
        ason_string_free(&t.name);
        for (size_t i = 0; i < t.tags.len; i++) ason_string_free(&t.tags.data[i]);
        ason_vec_str_free(&t.tags);
    }

    /* 10. Comments */
    printf("10. With comments:\n");
    {
        const char* input = "/* user list */ {id,name,active}:(1,Alice,true)";
        User u = {0};
        ason_err_t err = ason_decode_User(input, strlen(input), &u);
        assert(err == ASON_OK);
        printf("   id=%lld name=%s active=%s\n\n",
               (long long)u.id, u.name.data, u.active ? "true" : "false");
        ason_string_free(&u.name);
    }

    printf("=== All examples passed! ===\n");
    ason_string_free(&user.name);
    return 0;
}
