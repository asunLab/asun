#include "ason.h"
#include <assert.h>
#include <stdio.h>
#include <string.h>

/* ============================================================================
 * Struct definitions
 * ============================================================================
 */

/* Department / Employee */
typedef struct {
  ason_string_t title;
} Department;
ASON_FIELDS(Department, 1, ASON_FIELD(Department, title, "title", str))

typedef struct {
  int64_t id;
  ason_string_t name;
  Department dept;
  ason_vec_str skills;
  bool active;
} Employee;

ASON_VEC_STRUCT_DEFINE(Department)

ASON_FIELDS(Employee, 5, ASON_FIELD(Employee, id, "id", i64),
            ASON_FIELD(Employee, name, "name", str),
            ASON_FIELD_STRUCT(Employee, dept, "dept", &Department_ason_desc),
            ASON_FIELD(Employee, skills, "skills", vec_str),
            ASON_FIELD(Employee, active, "active", bool))

/* WithMap */
typedef struct {
  ason_string_t name;
  ason_map_si attrs;
} WithMap;

ASON_FIELDS(WithMap, 2, ASON_FIELD(WithMap, name, "name", str),
            ASON_FIELD(WithMap, attrs, "attrs", map_si))

/* Address / Nested */
typedef struct {
  ason_string_t city;
  int64_t zip;
} Address;
ASON_FIELDS(Address, 2, ASON_FIELD(Address, city, "city", str),
            ASON_FIELD(Address, zip, "zip", i64))

typedef struct {
  ason_string_t name;
  Address addr;
} Nested;
ASON_FIELDS(Nested, 2, ASON_FIELD(Nested, name, "name", str),
            ASON_FIELD_STRUCT(Nested, addr, "addr", &Address_ason_desc))

/* AllTypes */
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
  ason_vec_vec_i64 nested_vec;
} AllTypes;

ASON_FIELDS(AllTypes, 17, ASON_FIELD(AllTypes, b, "b", bool),
            ASON_FIELD(AllTypes, i8v, "i8v", i8),
            ASON_FIELD(AllTypes, i16v, "i16v", i16),
            ASON_FIELD(AllTypes, i32v, "i32v", i32),
            ASON_FIELD(AllTypes, i64v, "i64v", i64),
            ASON_FIELD(AllTypes, u8v, "u8v", u8),
            ASON_FIELD(AllTypes, u16v, "u16v", u16),
            ASON_FIELD(AllTypes, u32v, "u32v", u32),
            ASON_FIELD(AllTypes, u64v, "u64v", u64),
            ASON_FIELD(AllTypes, f32v, "f32v", f32),
            ASON_FIELD(AllTypes, f64v, "f64v", f64),
            ASON_FIELD(AllTypes, s, "s", str),
            ASON_FIELD(AllTypes, opt_some, "opt_some", opt_i64),
            ASON_FIELD(AllTypes, opt_none, "opt_none", opt_i64),
            ASON_FIELD(AllTypes, vec_int, "vec_int", vec_i64),
            ASON_FIELD(AllTypes, vec_str, "vec_str", vec_str),
            ASON_FIELD(AllTypes, nested_vec, "nested_vec", vec_vec_i64))

/* 5-level: Country > Region > City > District > Street > Building */
typedef struct {
  ason_string_t name;
  int64_t floors;
  bool residential;
  double height_m;
} Building;

ASON_FIELDS(Building, 4, ASON_FIELD(Building, name, "name", str),
            ASON_FIELD(Building, floors, "floors", i64),
            ASON_FIELD(Building, residential, "residential", bool),
            ASON_FIELD(Building, height_m, "height_m", f64))
ASON_VEC_STRUCT_DEFINE(Building)

typedef struct {
  ason_string_t name;
  double length_km;
  ason_vec_Building buildings;
} Street;

ASON_FIELDS(Street, 3, ASON_FIELD(Street, name, "name", str),
            ASON_FIELD(Street, length_km, "length_km", f64),
            ASON_FIELD_VEC_STRUCT(Street, buildings, "buildings", Building))
ASON_VEC_STRUCT_DEFINE(Street)

typedef struct {
  ason_string_t name;
  int64_t population;
  ason_vec_Street streets;
} District;

ASON_FIELDS(District, 3, ASON_FIELD(District, name, "name", str),
            ASON_FIELD(District, population, "population", i64),
            ASON_FIELD_VEC_STRUCT(District, streets, "streets", Street))
ASON_VEC_STRUCT_DEFINE(District)

typedef struct {
  ason_string_t name;
  int64_t population;
  double area_km2;
  ason_vec_District districts;
} City;

ASON_FIELDS(City, 4, ASON_FIELD(City, name, "name", str),
            ASON_FIELD(City, population, "population", i64),
            ASON_FIELD(City, area_km2, "area_km2", f64),
            ASON_FIELD_VEC_STRUCT(City, districts, "districts", District))
ASON_VEC_STRUCT_DEFINE(City)

typedef struct {
  ason_string_t name;
  ason_vec_City cities;
} Region;

ASON_FIELDS(Region, 2, ASON_FIELD(Region, name, "name", str),
            ASON_FIELD_VEC_STRUCT(Region, cities, "cities", City))
ASON_VEC_STRUCT_DEFINE(Region)

typedef struct {
  ason_string_t name;
  ason_string_t code;
  int64_t population;
  double gdp_trillion;
  ason_vec_Region regions;
} Country;

ASON_FIELDS(Country, 5, ASON_FIELD(Country, name, "name", str),
            ASON_FIELD(Country, code, "code", str),
            ASON_FIELD(Country, population, "population", i64),
            ASON_FIELD(Country, gdp_trillion, "gdp_trillion", f64),
            ASON_FIELD_VEC_STRUCT(Country, regions, "regions", Region))

/* Binary support for 5-level hierarchy */
ASON_FIELDS_BIN(Building, 4)
ASON_FIELDS_BIN(Street, 3)
ASON_FIELDS_BIN(District, 3)
ASON_FIELDS_BIN(City, 4)
ASON_FIELDS_BIN(Region, 2)
ASON_FIELDS_BIN(Country, 5)

/* 7-level: Universe > Galaxy > SolarSystem > Planet > Continent > Nation >
 * State */
typedef struct {
  ason_string_t name;
  ason_string_t capital;
  int64_t population;
} State;
ASON_FIELDS(State, 3, ASON_FIELD(State, name, "name", str),
            ASON_FIELD(State, capital, "capital", str),
            ASON_FIELD(State, population, "population", i64))
ASON_VEC_STRUCT_DEFINE(State)

typedef struct {
  ason_string_t name;
  ason_vec_State states;
} Nation;
ASON_FIELDS(Nation, 2, ASON_FIELD(Nation, name, "name", str),
            ASON_FIELD_VEC_STRUCT(Nation, states, "states", State))
ASON_VEC_STRUCT_DEFINE(Nation)

typedef struct {
  ason_string_t name;
  ason_vec_Nation nations;
} Continent;
ASON_FIELDS(Continent, 2, ASON_FIELD(Continent, name, "name", str),
            ASON_FIELD_VEC_STRUCT(Continent, nations, "nations", Nation))
ASON_VEC_STRUCT_DEFINE(Continent)

typedef struct {
  ason_string_t name;
  double radius_km;
  bool has_life;
  ason_vec_Continent continents;
} Planet;
ASON_FIELDS(Planet, 4, ASON_FIELD(Planet, name, "name", str),
            ASON_FIELD(Planet, radius_km, "radius_km", f64),
            ASON_FIELD(Planet, has_life, "has_life", bool),
            ASON_FIELD_VEC_STRUCT(Planet, continents, "continents", Continent))
ASON_VEC_STRUCT_DEFINE(Planet)

typedef struct {
  ason_string_t name;
  ason_string_t star_type;
  ason_vec_Planet planets;
} SolarSystem;
ASON_FIELDS(SolarSystem, 3, ASON_FIELD(SolarSystem, name, "name", str),
            ASON_FIELD(SolarSystem, star_type, "star_type", str),
            ASON_FIELD_VEC_STRUCT(SolarSystem, planets, "planets", Planet))
ASON_VEC_STRUCT_DEFINE(SolarSystem)

typedef struct {
  ason_string_t name;
  double star_count_billions;
  ason_vec_SolarSystem systems;
} Galaxy;
ASON_FIELDS(Galaxy, 3, ASON_FIELD(Galaxy, name, "name", str),
            ASON_FIELD(Galaxy, star_count_billions, "star_count_billions", f64),
            ASON_FIELD_VEC_STRUCT(Galaxy, systems, "systems", SolarSystem))
ASON_VEC_STRUCT_DEFINE(Galaxy)

typedef struct {
  ason_string_t name;
  double age_billion_years;
  ason_vec_Galaxy galaxies;
} Universe;
ASON_FIELDS(Universe, 3, ASON_FIELD(Universe, name, "name", str),
            ASON_FIELD(Universe, age_billion_years, "age_billion_years", f64),
            ASON_FIELD_VEC_STRUCT(Universe, galaxies, "galaxies", Galaxy))

/* Binary support for 7-level hierarchy */
ASON_FIELDS_BIN(State, 3)
ASON_FIELDS_BIN(Nation, 2)
ASON_FIELDS_BIN(Continent, 2)
ASON_FIELDS_BIN(Planet, 4)
ASON_FIELDS_BIN(SolarSystem, 3)
ASON_FIELDS_BIN(Galaxy, 3)
ASON_FIELDS_BIN(Universe, 3)

/* ServiceConfig */
typedef struct {
  ason_string_t host;
  int64_t port;
  int64_t max_connections;
  bool ssl;
  double timeout_ms;
} DbConfig;
ASON_FIELDS(DbConfig, 5, ASON_FIELD(DbConfig, host, "host", str),
            ASON_FIELD(DbConfig, port, "port", i64),
            ASON_FIELD(DbConfig, max_connections, "max_connections", i64),
            ASON_FIELD(DbConfig, ssl, "ssl", bool),
            ASON_FIELD(DbConfig, timeout_ms, "timeout_ms", f64))

typedef struct {
  bool enabled;
  int64_t ttl_seconds;
  int64_t max_size_mb;
} CacheConfig;
ASON_FIELDS(CacheConfig, 3, ASON_FIELD(CacheConfig, enabled, "enabled", bool),
            ASON_FIELD(CacheConfig, ttl_seconds, "ttl_seconds", i64),
            ASON_FIELD(CacheConfig, max_size_mb, "max_size_mb", i64))

typedef struct {
  ason_string_t level;
  ason_opt_str file;
  bool rotate;
} LogConfig;
ASON_FIELDS(LogConfig, 3, ASON_FIELD(LogConfig, level, "level", str),
            ASON_FIELD(LogConfig, file, "file", opt_str),
            ASON_FIELD(LogConfig, rotate, "rotate", bool))

typedef struct {
  ason_string_t name;
  ason_string_t version;
  DbConfig db;
  CacheConfig cache;
  LogConfig log;
  ason_vec_str features;
  ason_map_ss env;
} ServiceConfig;

ASON_FIELDS(ServiceConfig, 7, ASON_FIELD(ServiceConfig, name, "name", str),
            ASON_FIELD(ServiceConfig, version, "version", str),
            ASON_FIELD_STRUCT(ServiceConfig, db, "db", &DbConfig_ason_desc),
            ASON_FIELD_STRUCT(ServiceConfig, cache, "cache",
                              &CacheConfig_ason_desc),
            ASON_FIELD_STRUCT(ServiceConfig, log, "log", &LogConfig_ason_desc),
            ASON_FIELD(ServiceConfig, features, "features", vec_str),
            ASON_FIELD(ServiceConfig, env, "env", map_ss))

/* Note, Measurement, Nums, Special, Matrix3D, WithVec — for edge cases */
typedef struct {
  ason_string_t text;
} Note;
ASON_FIELDS(Note, 1, ASON_FIELD(Note, text, "text", str))

typedef struct {
  int64_t id;
  double value;
  ason_string_t label;
} Measurement;
ASON_FIELDS(Measurement, 3, ASON_FIELD(Measurement, id, "id", i64),
            ASON_FIELD(Measurement, value, "value", f64),
            ASON_FIELD(Measurement, label, "label", str))

typedef struct {
  int64_t a;
  double b;
  int64_t c;
} Nums;
ASON_FIELDS(Nums, 3, ASON_FIELD(Nums, a, "a", i64),
            ASON_FIELD(Nums, b, "b", f64), ASON_FIELD(Nums, c, "c", i64))

typedef struct {
  ason_string_t val;
} Special;
ASON_FIELDS(Special, 1, ASON_FIELD(Special, val, "val", str))

typedef struct {
  ason_vec_vec_i64 data;
} Matrix3D;
ASON_FIELDS(Matrix3D, 1, ASON_FIELD(Matrix3D, data, "data", vec_vec_i64))

typedef struct {
  ason_vec_i64 items;
} WithVec;
ASON_FIELDS(WithVec, 1, ASON_FIELD(WithVec, items, "items", vec_i64))

/* ============================================================================
 * Helper: free deeply nested structs
 * ============================================================================
 */

static void free_building(Building *b) { ason_string_free(&b->name); }
static void free_street(Street *s) {
  ason_string_free(&s->name);
  for (size_t i = 0; i < s->buildings.len; i++)
    free_building(&s->buildings.data[i]);
  ason_vec_Building_free(&s->buildings);
}
static void free_district(District *d) {
  ason_string_free(&d->name);
  for (size_t i = 0; i < d->streets.len; i++)
    free_street(&d->streets.data[i]);
  ason_vec_Street_free(&d->streets);
}
static void free_city(City *c) {
  ason_string_free(&c->name);
  for (size_t i = 0; i < c->districts.len; i++)
    free_district(&c->districts.data[i]);
  ason_vec_District_free(&c->districts);
}
static void free_region(Region *r) {
  ason_string_free(&r->name);
  for (size_t i = 0; i < r->cities.len; i++)
    free_city(&r->cities.data[i]);
  ason_vec_City_free(&r->cities);
}
static void free_country(Country *c) {
  ason_string_free(&c->name);
  ason_string_free(&c->code);
  for (size_t i = 0; i < c->regions.len; i++)
    free_region(&c->regions.data[i]);
  ason_vec_Region_free(&c->regions);
}

/* ============================================================================
 * Main
 * ============================================================================
 */

int main(void) {
  printf("=== ASON C Complex Examples ===\n\n");
  int passed = 0;

  /* 1. Nested struct */
  printf("1. Nested struct:\n");
  {
    const char *input =
        "{id,name,dept:{title},skills,active}:(1,Alice,(Manager),[rust],true)";
    Employee emp = {0};
    ason_err_t err = ason_decode_Employee(input, strlen(input), &emp);
    assert(err == ASON_OK);
    printf("   id=%lld name=%s dept=%s skills=[", (long long)emp.id,
           emp.name.data, emp.dept.title.data);
    for (size_t i = 0; i < emp.skills.len; i++) {
      if (i)
        printf(",");
      printf("%s", emp.skills.data[i].data);
    }
    printf("] active=%s\n\n", emp.active ? "true" : "false");
    ason_string_free(&emp.name);
    ason_string_free(&emp.dept.title);
    for (size_t i = 0; i < emp.skills.len; i++)
      ason_string_free(&emp.skills.data[i]);
    ason_vec_str_free(&emp.skills);
    passed++;
  }

  /* 2. Vec with nested structs */
  printf("2. Vec with nested structs:\n");
  {
    const char *input =
        "[{id:int,name:str,dept:{title:str},skills:[str],active:bool}]:"
        "(1,Alice,(Manager),[Rust,Go],true),"
        "(2,Bob,(Engineer),[Python],false),"
        "(3,\"Carol Smith\",(Director),[Leadership,Strategy],true)";
    Employee *emps = NULL;
    size_t cnt = 0;
    ason_err_t err = ason_decode_vec_Employee(input, strlen(input), &emps, &cnt);
    assert(err == ASON_OK);
    for (size_t i = 0; i < cnt; i++) {
      printf("   id=%lld name=%s dept=%s\n", (long long)emps[i].id,
             emps[i].name.data, emps[i].dept.title.data);
      ason_string_free(&emps[i].name);
      ason_string_free(&emps[i].dept.title);
      for (size_t j = 0; j < emps[i].skills.len; j++)
        ason_string_free(&emps[i].skills.data[j]);
      ason_vec_str_free(&emps[i].skills);
    }
    free(emps);
    printf("\n");
    passed++;
  }

  /* 3. Map/Dict field */
  printf("3. Map/Dict field:\n");
  {
    const char *input = "{name,attrs}:(Alice,[(age,30),(score,95)])";
    WithMap item = {0};
    ason_err_t err = ason_decode_WithMap(input, strlen(input), &item);
    assert(err == ASON_OK);
    printf("   name=%s attrs={", item.name.data);
    for (size_t i = 0; i < item.attrs.len; i++) {
      if (i)
        printf(",");
      printf("%s:%lld", item.attrs.data[i].key.data,
             (long long)item.attrs.data[i].val);
    }
    printf("}\n\n");
    ason_string_free(&item.name);
    for (size_t i = 0; i < item.attrs.len; i++)
      ason_string_free(&item.attrs.data[i].key);
    ason_map_si_free(&item.attrs);
    passed++;
  }

  /* 4. Nested struct roundtrip */
  printf("4. Nested struct roundtrip:\n");
  {
    Nested n = {ason_string_from("Alice"), {ason_string_from("NYC"), 10001}};
    ason_buf_t buf = ason_encode_Nested(&n);
    printf("   serialized:   %.*s\n", (int)buf.len, buf.data);
    Nested n2 = {0};
    ason_err_t err = ason_decode_Nested(buf.data, buf.len, &n2);
    assert(err == ASON_OK);
    assert(strcmp(n2.name.data, "Alice") == 0);
    assert(strcmp(n2.addr.city.data, "NYC") == 0);
    assert(n2.addr.zip == 10001);
    printf("   roundtrip OK\n\n");
    ason_buf_free(&buf);
    ason_string_free(&n.name);
    ason_string_free(&n.addr.city);
    ason_string_free(&n2.name);
    ason_string_free(&n2.addr.city);
    passed++;
  }

  /* 5. Escaped strings */
  printf("5. Escaped strings:\n");
  {
    Note note = {ason_string_from("say \"hi\", then (wave)\tnewline\nend")};
    ason_buf_t buf = ason_encode_Note(&note);
    printf("   serialized:   %.*s\n", (int)buf.len, buf.data);
    Note note2 = {0};
    ason_err_t err = ason_decode_Note(buf.data, buf.len, &note2);
    assert(err == ASON_OK);
    assert(strcmp(note.text.data, note2.text.data) == 0);
    printf("   escape roundtrip OK\n\n");
    ason_buf_free(&buf);
    ason_string_free(&note.text);
    ason_string_free(&note2.text);
    passed++;
  }

  /* 6. Float fields */
  printf("6. Float fields:\n");
  {
    Measurement m = {2, 95.0, ason_string_from("score")};
    ason_buf_t buf = ason_encode_Measurement(&m);
    printf("   serialized: %.*s\n", (int)buf.len, buf.data);
    Measurement m2 = {0};
    ason_err_t err = ason_decode_Measurement(buf.data, buf.len, &m2);
    assert(err == ASON_OK);
    assert(m2.id == 2);
    assert(fabs(m2.value - 95.0) < 0.001);
    printf("   float roundtrip OK\n\n");
    ason_buf_free(&buf);
    ason_string_free(&m.label);
    ason_string_free(&m2.label);
    passed++;
  }

  /* 7. Negative numbers */
  printf("7. Negative numbers:\n");
  {
    Nums n = {-42, -3.15, -9223372036854775807LL};
    ason_buf_t buf = ason_encode_Nums(&n);
    printf("   serialized:   %.*s\n", (int)buf.len, buf.data);
    Nums n2 = {0};
    ason_err_t err = ason_decode_Nums(buf.data, buf.len, &n2);
    assert(err == ASON_OK);
    assert(n2.a == -42);
    assert(n2.c == -9223372036854775807LL);
    printf("   negative roundtrip OK\n\n");
    ason_buf_free(&buf);
    passed++;
  }

  /* 8. All types struct */
  printf("8. All types struct:\n");
  {
    AllTypes all = {0};
    all.b = true;
    all.i8v = -128;
    all.i16v = -32768;
    all.i32v = -2147483647 - 1;
    all.i64v = -9223372036854775807LL;
    all.u8v = 255;
    all.u16v = 65535;
    all.u32v = 4294967295U;
    all.u64v = 18446744073709551615ULL;
    all.f32v = 3.15f;
    all.f64v = 2.718281828459045;
    all.s = ason_string_from("hello, world (test) [arr]");
    all.opt_some = (ason_opt_i64){true, 42};
    all.opt_none = (ason_opt_i64){false, 0};
    all.vec_int = ason_vec_i64_new();
    int64_t vi[] = {1, 2, 3, -4, 0};
    for (int i = 0; i < 5; i++)
      ason_vec_i64_push(&all.vec_int, vi[i]);
    all.vec_str = ason_vec_str_new();
    ason_vec_str_push(&all.vec_str, ason_string_from("alpha"));
    ason_vec_str_push(&all.vec_str, ason_string_from("beta gamma"));
    ason_vec_str_push(&all.vec_str, ason_string_from("delta"));
    all.nested_vec = ason_vec_vec_i64_new();
    ason_vec_i64 nv1 = ason_vec_i64_new();
    ason_vec_i64_push(&nv1, 1);
    ason_vec_i64_push(&nv1, 2);
    ason_vec_i64 nv2 = ason_vec_i64_new();
    ason_vec_i64_push(&nv2, 3);
    ason_vec_i64_push(&nv2, 4);
    ason_vec_i64_push(&nv2, 5);
    ason_vec_vec_i64_push(&all.nested_vec, nv1);
    ason_vec_vec_i64_push(&all.nested_vec, nv2);

    ason_buf_t buf = ason_encode_AllTypes(&all);
    printf("   serialized (%zu bytes):\n   %.*s\n", buf.len, (int)buf.len,
           buf.data);

    AllTypes all2 = {0};
    ason_err_t err = ason_decode_AllTypes(buf.data, buf.len, &all2);
    assert(err == ASON_OK);
    assert(all2.b == true);
    assert(all2.i8v == -128);
    assert(all2.u64v == 18446744073709551615ULL);
    assert(all2.vec_int.len == 5);
    assert(all2.nested_vec.len == 2);
    printf("   all-types roundtrip OK\n\n");

    ason_buf_free(&buf);
    ason_string_free(&all.s);
    for (size_t i = 0; i < all.vec_str.len; i++)
      ason_string_free(&all.vec_str.data[i]);
    ason_vec_str_free(&all.vec_str);
    ason_vec_i64_free(&all.vec_int);
    for (size_t i = 0; i < all.nested_vec.len; i++)
      ason_vec_i64_free(&all.nested_vec.data[i]);
    ason_vec_vec_i64_free(&all.nested_vec);
    ason_string_free(&all2.s);
    for (size_t i = 0; i < all2.vec_str.len; i++)
      ason_string_free(&all2.vec_str.data[i]);
    ason_vec_str_free(&all2.vec_str);
    ason_vec_i64_free(&all2.vec_int);
    for (size_t i = 0; i < all2.nested_vec.len; i++)
      ason_vec_i64_free(&all2.nested_vec.data[i]);
    ason_vec_vec_i64_free(&all2.nested_vec);
    passed++;
  }

  /* 9. 5-level deep nesting */
  printf("9. Five-level nesting:\n");
  {
    Country country = {0};
    country.name = ason_string_from("Rustland");
    country.code = ason_string_from("RL");
    country.population = 50000000;
    country.gdp_trillion = 1.5;
    country.regions = ason_vec_Region_new();

    Region r1 = {ason_string_from("Northern"), ason_vec_City_new()};
    City c1 = {ason_string_from("Ferriton"), 2000000, 350.5,
               ason_vec_District_new()};
    District d1 = {ason_string_from("Downtown"), 500000, ason_vec_Street_new()};
    Street s1 = {ason_string_from("Main St"), 2.5, ason_vec_Building_new()};
    Building b1 = {ason_string_from("Tower A"), 50, false, 200.0};
    Building b2 = {ason_string_from("Apt Block 1"), 12, true, 40.5};
    ason_vec_Building_push(&s1.buildings, b1);
    ason_vec_Building_push(&s1.buildings, b2);
    Street s2 = {ason_string_from("Oak Ave"), 1.2, ason_vec_Building_new()};
    Building b3 = {ason_string_from("Library"), 3, false, 15.0};
    ason_vec_Building_push(&s2.buildings, b3);
    ason_vec_Street_push(&d1.streets, s1);
    ason_vec_Street_push(&d1.streets, s2);
    District d2 = {ason_string_from("Harbor"), 150000, ason_vec_Street_new()};
    Street s3 = {ason_string_from("Dock Rd"), 0.8, ason_vec_Building_new()};
    Building b4 = {ason_string_from("Warehouse 7"), 1, false, 8.0};
    ason_vec_Building_push(&s3.buildings, b4);
    ason_vec_Street_push(&d2.streets, s3);
    ason_vec_District_push(&c1.districts, d1);
    ason_vec_District_push(&c1.districts, d2);
    ason_vec_City_push(&r1.cities, c1);
    ason_vec_Region_push(&country.regions, r1);

    Region r2 = {ason_string_from("Southern"), ason_vec_City_new()};
    City c2 = {ason_string_from("Crabville"), 800000, 120.0,
               ason_vec_District_new()};
    District d3 = {ason_string_from("Old Town"), 200000, ason_vec_Street_new()};
    Street s4 = {ason_string_from("Heritage Ln"), 0.5, ason_vec_Building_new()};
    Building b5 = {ason_string_from("Museum"), 2, false, 12.0};
    Building b6 = {ason_string_from("Town Hall"), 4, false, 20.0};
    ason_vec_Building_push(&s4.buildings, b5);
    ason_vec_Building_push(&s4.buildings, b6);
    ason_vec_Street_push(&d3.streets, s4);
    ason_vec_District_push(&c2.districts, d3);
    ason_vec_City_push(&r2.cities, c2);
    ason_vec_Region_push(&country.regions, r2);

    ason_buf_t buf = ason_encode_Country(&country);
    printf("   serialized (%zu bytes)\n", buf.len);

    Country country2 = {0};
    ason_err_t err = ason_decode_Country(buf.data, buf.len, &country2);
    assert(err == ASON_OK);
    assert(strcmp(country2.name.data, "Rustland") == 0);
    assert(country2.regions.len == 2);
    printf("   ✓ 5-level ASON-text roundtrip OK\n");

    /* ASON binary roundtrip */
    ason_buf_t bin = ason_encode_bin_Country(&country);
    Country country3 = {0};
    ason_err_t err2 = ason_decode_bin_Country(bin.data, bin.len, &country3);
    assert(err2 == ASON_OK);
    assert(strcmp(country3.name.data, "Rustland") == 0);
    printf("   ✓ 5-level ASON-bin roundtrip OK\n");
    printf("   ASON text: %zu B | ASON bin: %zu B\n", buf.len, bin.len);
    printf("   BIN is %.0f%% smaller than text\n\n",
           (1.0 - (double)bin.len / (double)buf.len) * 100.0);

    ason_buf_free(&buf);
    ason_buf_free(&bin);
    free_country(&country);
    free_country(&country2);
    free_country(&country3);
    passed++;
  }

  /* 10. 7-level deep */
  printf("10. Seven-level nesting:\n");
  {
    Universe u = {0};
    u.name = ason_string_from("Observable");
    u.age_billion_years = 13.8;
    u.galaxies = ason_vec_Galaxy_new();

    Galaxy g = {ason_string_from("Milky Way"), 250.0,
                ason_vec_SolarSystem_new()};
    SolarSystem ss = {ason_string_from("Sol"), ason_string_from("G2V"),
                      ason_vec_Planet_new()};

    Planet earth = {ason_string_from("Earth"), 6371.0, true,
                    ason_vec_Continent_new()};
    Continent asia = {ason_string_from("Asia"), ason_vec_Nation_new()};
    Nation japan = {ason_string_from("Japan"), ason_vec_State_new()};
    State tokyo = {ason_string_from("Tokyo"), ason_string_from("Shinjuku"),
                   14000000};
    State osaka = {ason_string_from("Osaka"), ason_string_from("Osaka City"),
                   8800000};
    ason_vec_State_push(&japan.states, tokyo);
    ason_vec_State_push(&japan.states, osaka);
    ason_vec_Nation_push(&asia.nations, japan);
    Nation china = {ason_string_from("China"), ason_vec_State_new()};
    State beijing = {ason_string_from("Beijing"), ason_string_from("Beijing"),
                     21500000};
    ason_vec_State_push(&china.states, beijing);
    ason_vec_Nation_push(&asia.nations, china);
    ason_vec_Continent_push(&earth.continents, asia);

    Continent europe = {ason_string_from("Europe"), ason_vec_Nation_new()};
    Nation germany = {ason_string_from("Germany"), ason_vec_State_new()};
    State bavaria = {ason_string_from("Bavaria"), ason_string_from("Munich"),
                     13000000};
    State berlin = {ason_string_from("Berlin"), ason_string_from("Berlin"),
                    3600000};
    ason_vec_State_push(&germany.states, bavaria);
    ason_vec_State_push(&germany.states, berlin);
    ason_vec_Nation_push(&europe.nations, germany);
    ason_vec_Continent_push(&earth.continents, europe);

    Planet mars = {ason_string_from("Mars"), 3389.5, false,
                   ason_vec_Continent_new()};

    ason_vec_Planet_push(&ss.planets, earth);
    ason_vec_Planet_push(&ss.planets, mars);
    ason_vec_SolarSystem_push(&g.systems, ss);
    ason_vec_Galaxy_push(&u.galaxies, g);

    ason_buf_t buf = ason_encode_Universe(&u);
    printf("   serialized (%zu bytes)\n", buf.len);

    Universe u2 = {0};
    ason_err_t err = ason_decode_Universe(buf.data, buf.len, &u2);
    assert(err == ASON_OK);
    assert(strcmp(u2.name.data, "Observable") == 0);
    printf("   ✓ 7-level ASON-text roundtrip OK\n");

    /* ASON binary roundtrip */
    ason_buf_t bin = ason_encode_bin_Universe(&u);
    Universe u3 = {0};
    ason_err_t err2 = ason_decode_bin_Universe(bin.data, bin.len, &u3);
    assert(err2 == ASON_OK);
    assert(strcmp(u3.name.data, "Observable") == 0);
    printf("   ✓ 7-level ASON-bin roundtrip OK\n");
    printf("   ASON text: %zu B | ASON bin: %zu B\n", buf.len, bin.len);
    printf("   BIN is %.0f%% smaller than text\n\n",
           (1.0 - (double)bin.len / (double)buf.len) * 100.0);

    ason_buf_free(&buf);
    ason_buf_free(&bin);
    /* Note: deep free omitted for brevity — would need recursive free */
    passed++;
  }

  /* 11. Service config */
  printf("11. Complex config struct:\n");
  {
    ServiceConfig cfg = {0};
    cfg.name = ason_string_from("my-service");
    cfg.version = ason_string_from("2.1.0");
    cfg.db =
        (DbConfig){ason_string_from("db.example.com"), 5432, 100, true, 3000.5};
    cfg.cache = (CacheConfig){true, 3600, 512};
    cfg.log = (LogConfig){ason_string_from("info"),
                          {true, ason_string_from("/var/log/app.log")},
                          true};
    cfg.features = ason_vec_str_new();
    ason_vec_str_push(&cfg.features, ason_string_from("auth"));
    ason_vec_str_push(&cfg.features, ason_string_from("rate-limit"));
    ason_vec_str_push(&cfg.features, ason_string_from("websocket"));
    cfg.env = ason_map_ss_new();
    ason_map_ss_entry_t e1 = {ason_string_from("RUST_LOG"),
                              ason_string_from("debug")};
    ason_map_ss_entry_t e2 = {
        ason_string_from("DATABASE_URL"),
        ason_string_from("postgres://localhost:5432/mydb")};
    ason_map_ss_entry_t e3 = {ason_string_from("SECRET_KEY"),
                              ason_string_from("abc123!@#")};
    ason_map_ss_push(&cfg.env, e1);
    ason_map_ss_push(&cfg.env, e2);
    ason_map_ss_push(&cfg.env, e3);

    ason_buf_t buf = ason_encode_ServiceConfig(&cfg);
    printf("   serialized (%zu bytes):\n   %.*s\n", buf.len,
           (int)(buf.len > 200 ? 200 : buf.len), buf.data);

    ServiceConfig cfg2 = {0};
    ason_err_t err = ason_decode_ServiceConfig(buf.data, buf.len, &cfg2);
    assert(err == ASON_OK);
    assert(strcmp(cfg2.name.data, "my-service") == 0);
    assert(cfg2.db.port == 5432);
    printf("   config roundtrip OK\n\n");

    ason_buf_free(&buf);
    ason_string_free(&cfg.name);
    ason_string_free(&cfg.version);
    ason_string_free(&cfg.db.host);
    ason_string_free(&cfg.log.level);
    if (cfg.log.file.has_value)
      ason_string_free(&cfg.log.file.value);
    for (size_t i = 0; i < cfg.features.len; i++)
      ason_string_free(&cfg.features.data[i]);
    ason_vec_str_free(&cfg.features);
    for (size_t i = 0; i < cfg.env.len; i++) {
      ason_string_free(&cfg.env.data[i].key);
      ason_string_free(&cfg.env.data[i].val);
    }
    ason_map_ss_free(&cfg.env);
    ason_string_free(&cfg2.name);
    ason_string_free(&cfg2.version);
    ason_string_free(&cfg2.db.host);
    ason_string_free(&cfg2.log.level);
    if (cfg2.log.file.has_value)
      ason_string_free(&cfg2.log.file.value);
    for (size_t i = 0; i < cfg2.features.len; i++)
      ason_string_free(&cfg2.features.data[i]);
    ason_vec_str_free(&cfg2.features);
    for (size_t i = 0; i < cfg2.env.len; i++) {
      ason_string_free(&cfg2.env.data[i].key);
      ason_string_free(&cfg2.env.data[i].val);
    }
    ason_map_ss_free(&cfg2.env);
    passed++;
  }

  /* 12. Edge cases */
  printf("12. Edge cases:\n");
  {
    /* Empty vec */
    WithVec wv = {{0}};
    wv.items = ason_vec_i64_new();
    ason_buf_t buf = ason_encode_WithVec(&wv);
    printf("   empty vec: %.*s\n", (int)buf.len, buf.data);
    WithVec wv2 = {0};
    ason_err_t err = ason_decode_WithVec(buf.data, buf.len, &wv2);
    assert(err == ASON_OK);
    assert(wv2.items.len == 0);
    ason_buf_free(&buf);
    ason_vec_i64_free(&wv.items);
    ason_vec_i64_free(&wv2.items);

    /* Special chars */
    Special sp = {
        ason_string_from("tabs\there, newlines\nhere, quotes\"and\\backslash")};
    buf = ason_encode_Special(&sp);
    printf("   special chars: %.*s\n", (int)buf.len, buf.data);
    Special sp2 = {0};
    err = ason_decode_Special(buf.data, buf.len, &sp2);
    assert(err == ASON_OK);
    assert(strcmp(sp.val.data, sp2.val.data) == 0);
    ason_buf_free(&buf);
    ason_string_free(&sp.val);
    ason_string_free(&sp2.val);

    /* Bool-like string */
    Special sp3 = {ason_string_from("true")};
    buf = ason_encode_Special(&sp3);
    printf("   bool-like string: %.*s\n", (int)buf.len, buf.data);
    Special sp4 = {0};
    err = ason_decode_Special(buf.data, buf.len, &sp4);
    assert(err == ASON_OK);
    assert(strcmp(sp4.val.data, "true") == 0);
    ason_buf_free(&buf);
    ason_string_free(&sp3.val);
    ason_string_free(&sp4.val);

    /* Number-like string */
    Special sp5 = {ason_string_from("12345")};
    buf = ason_encode_Special(&sp5);
    printf("   number-like string: %.*s\n", (int)buf.len, buf.data);
    Special sp6 = {0};
    err = ason_decode_Special(buf.data, buf.len, &sp6);
    assert(err == ASON_OK);
    assert(strcmp(sp6.val.data, "12345") == 0);
    ason_buf_free(&buf);
    ason_string_free(&sp5.val);
    ason_string_free(&sp6.val);

    printf("   all edge cases OK\n\n");
    passed++;
  }

  /* 13. Triple-nested arrays */
  printf("13. Triple-nested arrays:\n");
  {
    Matrix3D m3 = {{0}};
    m3.data = ason_vec_vec_i64_new();
    ason_vec_i64 r1 = ason_vec_i64_new();
    ason_vec_i64_push(&r1, 1);
    ason_vec_i64_push(&r1, 2);
    ason_vec_i64 r2 = ason_vec_i64_new();
    ason_vec_i64_push(&r2, 3);
    ason_vec_i64_push(&r2, 4);
    ason_vec_i64 r3 = ason_vec_i64_new();
    ason_vec_i64_push(&r3, 5);
    ason_vec_i64_push(&r3, 6);
    ason_vec_i64_push(&r3, 7);
    ason_vec_vec_i64_push(&m3.data, r1);
    ason_vec_vec_i64_push(&m3.data, r2);
    ason_vec_vec_i64_push(&m3.data, r3);

    ason_buf_t buf = ason_encode_Matrix3D(&m3);
    printf("   %.*s\n", (int)buf.len, buf.data);
    Matrix3D m3b = {0};
    ason_err_t err = ason_decode_Matrix3D(buf.data, buf.len, &m3b);
    assert(err == ASON_OK);
    assert(m3b.data.len == 3);
    assert(m3b.data.data[2].len == 3);
    printf("   triple-nested array roundtrip OK\n\n");

    ason_buf_free(&buf);
    for (size_t i = 0; i < m3.data.len; i++)
      ason_vec_i64_free(&m3.data.data[i]);
    ason_vec_vec_i64_free(&m3.data);
    for (size_t i = 0; i < m3b.data.len; i++)
      ason_vec_i64_free(&m3b.data.data[i]);
    ason_vec_vec_i64_free(&m3b.data);
    passed++;
  }

  /* 14. Comments */
  printf("14. Comments:\n");
  {
    const char *input =
        "/* Top-level */ {id,name,dept:{title},skills,active}:/* inline */ "
        "(1,Alice,(HR),[rust],true)";
    Employee emp = {0};
    ason_err_t err = ason_decode_Employee(input, strlen(input), &emp);
    assert(err == ASON_OK);
    printf("   with inline comment: id=%lld name=%s dept=%s\n",
           (long long)emp.id, emp.name.data, emp.dept.title.data);
    printf("   comment parsing OK\n\n");
    ason_string_free(&emp.name);
    ason_string_free(&emp.dept.title);
    for (size_t i = 0; i < emp.skills.len; i++)
      ason_string_free(&emp.skills.data[i]);
    ason_vec_str_free(&emp.skills);
    passed++;
  }

  printf("\n=== All %d complex examples passed! ===\n", passed);
  return 0;
}
