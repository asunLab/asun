#include "ason.h"
#include <assert.h>
#include <stdio.h>
#include <string.h>

// ----------------------------------------------------------------------------
// Generation structures
// ----------------------------------------------------------------------------
typedef struct {
  int64_t id;
  ason_string_t name;
  int32_t age;
  bool gender;
} Detail;

ASON_FIELDS(Detail, 4, ASON_FIELD(Detail, id, "ID", i64),
            ASON_FIELD(Detail, name, "Name", str),
            ASON_FIELD(Detail, age, "Age", i32),
            ASON_FIELD(Detail, gender, "Gender", bool))

ASON_VEC_STRUCT_DEFINE(Detail)

typedef struct {
  ason_vec_Detail details;
} User;

ASON_FIELDS(User, 1, ASON_FIELD_VEC_STRUCT(User, details, "details", Detail))

// ----------------------------------------------------------------------------
// Consumption structures
// ----------------------------------------------------------------------------
typedef struct {
  int64_t id;
  ason_string_t name;
  int32_t age;
} Person;

ASON_FIELDS(Person, 3, ASON_FIELD(Person, id, "ID", i64),
            ASON_FIELD(Person, name, "Name", str),
            ASON_FIELD(Person, age, "Age", i32))

ASON_VEC_STRUCT_DEFINE(Person)

typedef struct {
  ason_vec_Person details;
} Human;

ASON_FIELDS(Human, 1, ASON_FIELD_VEC_STRUCT(Human, details, "details", Person))

// ----------------------------------------------------------------------------
// Main
// ----------------------------------------------------------------------------
int main(void) {
  // 1. Setup User data
  User u = {0};
  u.details = ason_vec_Detail_new();

  Detail d1 = {1, ason_string_from("Alice"), 30, true};
  Detail d2 = {2, ason_string_from("Bob"), 25, false};

  ason_vec_Detail_push(&u.details, d1);
  ason_vec_Detail_push(&u.details, d2);

  User users[1] = {u};

  // 2. Encode
  ason_buf_t buf = ason_encode_vec_User(users, 1);
  printf("Encoded ASON:\n%.*s\n", (int)buf.len, buf.data);

  // 3. Decode into Human list
  Human *humans = NULL;
  size_t count = 0;
  ason_err_t err = ason_decode_vec_Human(buf.data, buf.len, &humans, &count);
  assert(err == ASON_OK);

  printf("\nDecoded into Human list:\n");
  for (size_t i = 0; i < count; i++) {
    printf("Human{details=[");
    for (size_t j = 0; j < humans[i].details.len; j++) {
      if (j > 0)
        printf(", ");
      printf("Person{ID=%lld, Name=\"%s\", Age=%d}",
             (long long)humans[i].details.data[j].id,
             humans[i].details.data[j].name.data,
             humans[i].details.data[j].age);
    }
    printf("]}\n");
  }

  // Cleanup
  ason_buf_free(&buf);
  for (size_t i = 0; i < count; i++) {
    for (size_t j = 0; j < humans[i].details.len; j++) {
      ason_string_free(&humans[i].details.data[j].name);
    }
    ason_vec_Person_free(&humans[i].details);
  }
  free(humans);

  ason_string_free(&d1.name);
  ason_string_free(&d2.name);
  ason_vec_Detail_free(&u.details);

  return 0;
}
