# ason-c

[![C11](https://img.shields.io/badge/C-11-blue.svg)](https://en.cppreference.com/w/c/11)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

A high-performance C11 library for [ASON](https://github.com/athxx/ason) (Array-Schema Object Notation) — a token-efficient, schema-driven data format designed for LLM interactions and large-scale data transmission. SIMD-accelerated (NEON / SSE2), zero-copy parsing, zero external dependencies.

[中文文档](README_CN.md)

## What is ASON?

ASON separates **schema** from **data**, eliminating repetitive keys found in JSON. The schema is declared once, and data rows carry only values:

```text
JSON (100 tokens):
{"users":[{"id":1,"name":"Alice","active":true},{"id":2,"name":"Bob","active":false}]}

ASON (~35 tokens, 65% saving):
{id:int, name:str, active:bool}:(1,Alice,true),(2,Bob,false)
```

| Aspect                | JSON         | ASON               |
| --------------------- | ------------ | ------------------ |
| Token efficiency      | 100%         | 30–70% ✓           |
| Key repetition        | Every object | Declared once ✓    |
| Human readable        | Yes          | Yes ✓              |
| Nested structs        | ✓            | ✓                  |
| Type annotations      | No           | Optional ✓         |
| Serialization speed   | 1x           | **~1.7x faster** ✓ |
| Deserialization speed | 1x           | **~2.9x faster** ✓ |
| Data size             | 100%         | **47–60%** ✓       |

## Features

- **Pure C11** — no C++ required, compiles with any C11-compliant compiler
- **Zero dependencies** — standard library only, no third-party libs
- **SIMD-accelerated** — NEON (ARM64/Apple Silicon) and SSE2 (x86_64) with scalar fallback
- **Zero-copy parsing** — pointer-into-input parsing for unescaped strings, minimal allocations
- **Macro-based reflection** — `ASON_FIELDS` macro for struct registration with compile-time field descriptors
- **Full type support** — bool, all integers (i8–i64, u8–u64), float, double, string, optional, vector, map, nested structs, vector-of-structs
- **Schema-driven** — schema declared once, unlimited data rows
- **Type annotations** — optional type hints in schema for documentation and LLM prompts

## Quick Start

### 1. Add to Your Project

```bash
# Copy include/ason.h and src/ason.c to your project
cp include/ason.h /your/project/include/
cp src/ason.c /your/project/src/
```

### 2. Define Your Struct

```c
#include "ason.h"

typedef struct {
    int64_t id;
    ason_string_t name;
    bool active;
} User;

ASON_FIELDS(User, 3,
    ASON_FIELD(User, id,     "id",     i64),
    ASON_FIELD(User, name,   "name",   str),
    ASON_FIELD(User, active, "active", bool))
```

### 3. Serialize & Deserialize

```c
// Serialize
User user = {1, ason_string_from("Alice"), true};
ason_buf_t buf = ason_dump_User(&user);
// → "{id,name,active}:(1,Alice,true)"

// Serialize with type annotations
ason_buf_t buf2 = ason_dump_typed_User(&user);
// → "{id:int,name:str,active:bool}:(1,Alice,true)"

// Deserialize
User user2 = {0};
ason_err_t err = ason_load_User(buf.data, buf.len, &user2);
assert(err == ASON_OK);
assert(user2.id == 1);

// Free resources
ason_buf_free(&buf);
ason_buf_free(&buf2);
ason_string_free(&user.name);
ason_string_free(&user2.name);
```

### 4. Vec Serialization (Schema-Driven)

For arrays of structs, ASON writes the schema **once** and emits each element as a compact tuple — the key advantage over JSON:

```c
User users[] = {
    {1, ason_string_from("Alice"), true},
    {2, ason_string_from("Bob"),   false},
    {3, ason_string_from("Carol Smith"), true},
};

// Schema written once, data as tuples
ason_buf_t buf = ason_dump_vec_User(users, 3);
// → "{id,name,active}:(1,Alice,true),(2,Bob,false),(3,Carol Smith,true)"

// Deserialize
User* loaded = NULL;
size_t count = 0;
ason_err_t err = ason_load_vec_User(buf.data, buf.len, &loaded, &count);
assert(err == ASON_OK && count == 3);

// Free
ason_buf_free(&buf);
for (size_t i = 0; i < count; i++) ason_string_free(&loaded[i].name);
free(loaded);
```

## Supported Types

| Type                      | C Type             | ASON Field Type         | Example                       |
| ------------------------- | ------------------ | ----------------------- | ----------------------------- |
| Integers (i8–i64, u8–u64) | `int8_t`–`int64_t` | `i8`–`i64`, `u8`–`u64`  | `42`, `-100`                  |
| Floats                    | `float`, `double`  | `f32`, `f64`            | `3.14`, `-0.5`                |
| Bool                      | `bool`             | `bool`                  | `true`, `false`               |
| String                    | `ason_string_t`    | `str`                   | `Alice`, `"Carol Smith"`      |
| Optional int              | `ason_opt_i64`     | `opt_i64`               | `42` or _(blank)_ for None    |
| Optional string           | `ason_opt_str`     | `opt_str`               | `hello` or _(blank)_ for None |
| Optional float            | `ason_opt_f64`     | `opt_f64`               | `3.14` or _(blank)_ for None  |
| Vector of int             | `ason_vec_i64`     | `vec_i64`               | `[1,2,3]`                     |
| Vector of string          | `ason_vec_str`     | `vec_str`               | `[rust,go,python]`            |
| Vector of float           | `ason_vec_f64`     | `vec_f64`               | `[1.5,2.7]`                   |
| Vector of bool            | `ason_vec_bool`    | `vec_bool`              | `[true,false]`                |
| Nested vector             | `ason_vec_vec_i64` | `vec_vec_i64`           | `[[1,2],[3,4]]`               |
| Map (str→int)             | `ason_map_si`      | `map_si`                | `[(age,30),(score,95)]`       |
| Map (str→str)             | `ason_map_ss`      | `map_ss`                | `[(k1,v1),(k2,v2)]`           |
| Nested struct             | Any struct         | `ASON_FIELD_STRUCT`     | `(field1,field2)`             |
| Vector of structs         | `ason_vec_T`       | `ASON_FIELD_VEC_STRUCT` | `[(v1,v2),(v3,v4)]`           |
| Char                      | `char`             | `char`                  | `A`                           |

### Nested Structs

```c
typedef struct { ason_string_t title; } Dept;
ASON_FIELDS(Dept, 1, ASON_FIELD(Dept, title, "title", str))

typedef struct {
    ason_string_t name;
    Dept dept;
} Employee;
ASON_FIELDS(Employee, 2,
    ASON_FIELD(Employee, name, "name", str),
    ASON_FIELD_STRUCT(Employee, dept, "dept", &Dept_ason_desc))
// → {name,dept}:(Alice,(Engineering))
```

### Optional Fields

```c
typedef struct {
    int64_t id;
    ason_opt_str label;
} Item;
ASON_FIELDS(Item, 2,
    ASON_FIELD(Item, id,    "id",    i64),
    ASON_FIELD(Item, label, "label", opt_str))
// With value:   {id,label}:(1,hello)
// With None:    {id,label}:(1,)
```

### Arrays & Maps

```c
typedef struct {
    ason_string_t name;
    ason_vec_str tags;
} Tagged;
ASON_FIELDS(Tagged, 2,
    ASON_FIELD(Tagged, name, "name", str),
    ASON_FIELD(Tagged, tags, "tags", vec_str))
// → {name,tags}:(Alice,[rust,go,python])

typedef struct {
    ason_string_t name;
    ason_map_si attrs;
} Profile;
ASON_FIELDS(Profile, 2,
    ASON_FIELD(Profile, name,  "name",  str),
    ASON_FIELD(Profile, attrs, "attrs", map_si))
// → {name,attrs}:(Alice,[(age,30),(score,95)])
```

### Vector of Structs

```c
typedef struct { ason_string_t name; int64_t val; } Item;
ASON_FIELDS(Item, 2,
    ASON_FIELD(Item, name, "name", str),
    ASON_FIELD(Item, val,  "val",  i64))
ASON_VEC_STRUCT_DEFINE(Item)

typedef struct {
    ason_string_t label;
    ason_vec_Item items;
} Group;
ASON_FIELDS(Group, 2,
    ASON_FIELD(Group, label, "label", str),
    ASON_FIELD_VEC_STRUCT(Group, items, "items", Item))
// → {label,items}:(group1,[(a1,1),(a2,2)])
```

### Type Annotations (Optional)

Both annotated and unannotated schemas are fully equivalent — the deserializer handles them identically:

```text
// Without annotations (default output of dump / dump_vec)
{id,name,salary,active}:(1,Alice,5000.50,true)

// With annotations (output of dump_typed / dump_typed_vec)
{id:int,name:str,salary:float,active:bool}:(1,Alice,5000.50,true)
```

### Comments

```text
/* user list */
{id:int, name:str, active:bool}:(1,Alice,true),(2,Bob,false)
```

### Multiline Format

```text
{id:int, name:str, active:bool}:
  (1, Alice, true),
  (2, Bob, false),
  (3, "Carol Smith", true)
```

## API Reference

| Function                              | Description                                                   |
| ------------------------------------- | ------------------------------------------------------------- |
| `ason_dump_T(obj)`                    | Serialize a struct → unannotated schema `{id,name}:`          |
| `ason_dump_typed_T(obj)`              | Serialize a struct → annotated schema `{id:int,name:str}:`    |
| `ason_dump_vec_T(arr, count)`         | Serialize an array → unannotated schema                       |
| `ason_dump_typed_vec_T(arr, count)`   | Serialize an array → annotated schema                         |
| `ason_load_T(str, len, out)`          | Deserialize a struct (accepts both annotated and unannotated) |
| `ason_load_vec_T(str, len, out, cnt)` | Deserialize an array (accepts both annotated and unannotated) |

### Reflection Macros

```c
// Register struct fields
ASON_FIELDS(StructType, field_count,
    ASON_FIELD(StructType, member, "schema_name", field_type),
    ASON_FIELD_STRUCT(StructType, member, "schema_name", &SubType_ason_desc),
    ASON_FIELD_VEC_STRUCT(StructType, member, "schema_name", ElemType))

// Define vector-of-struct type (required before ASON_FIELD_VEC_STRUCT)
ASON_VEC_STRUCT_DEFINE(ElemType)
```

### Memory Management

C requires explicit memory management. Use the provided free helpers:

```c
ason_buf_free(&buf);          // Free serialization buffer
ason_string_free(&str);       // Free owned string
ason_vec_i64_free(&vec);      // Free vector
// For loaded arrays: free each element's strings, then free(arr)
```

## Performance

Benchmarked on Apple Silicon M-series (arm64, NEON SIMD), `-O3` Release mode, comparing ASON against a minimal inline JSON serializer/deserializer:

### Serialization (ASON is 1.3–1.8x faster)

| Scenario            | JSON     | ASON     | Speedup   |
| ------------------- | -------- | -------- | --------- |
| Flat struct × 100   | 2.22 ms  | 1.21 ms  | **1.84x** |
| Flat struct × 500   | 8.49 ms  | 4.59 ms  | **1.85x** |
| Flat struct × 1000  | 12.12 ms | 7.26 ms  | **1.67x** |
| Flat struct × 5000  | 59.92 ms | 35.99 ms | **1.67x** |
| 5-level deep × 10   | 3.44 ms  | 2.55 ms  | **1.35x** |
| 5-level deep × 50   | 17.95 ms | 13.63 ms | **1.32x** |
| 5-level deep × 100  | 35.26 ms | 26.45 ms | **1.33x** |
| Large payload (10k) | 12.42 ms | 7.86 ms  | **1.58x** |

### Deserialization (ASON is 2.9–3.3x faster)

| Scenario            | JSON      | ASON     | Speedup   |
| ------------------- | --------- | -------- | --------- |
| Flat struct × 100   | 8.98 ms   | 2.78 ms  | **3.23x** |
| Flat struct × 500   | 31.91 ms  | 9.72 ms  | **3.28x** |
| Flat struct × 1000  | 56.82 ms  | 19.04 ms | **2.98x** |
| Flat struct × 5000  | 286.56 ms | 97.89 ms | **2.93x** |
| Large payload (10k) | 58.08 ms  | 20.16 ms | **2.88x** |

### Throughput

| Metric                | JSON           | ASON            | Speedup   |
| --------------------- | -------------- | --------------- | --------- |
| Serialize records/s   | 8.2M records/s | 13.3M records/s | **1.63x** |
| Serialize MB/s        | 947 MB/s       | 718 MB/s        | —         |
| Deserialize records/s | 1.7M records/s | 5.0M records/s  | **2.92x** |

### Size Savings

| Scenario            | JSON     | ASON     | Saving  |
| ------------------- | -------- | -------- | ------- |
| Flat struct × 100   | 12,071 B | 5,612 B  | **54%** |
| Flat struct × 1000  | 122 KB   | 57 KB    | **53%** |
| 5-level deep × 10   | 43,244 B | 16,893 B | **61%** |
| 5-level deep × 100  | 438 KB   | 175 KB   | **60%** |
| Large payload (10k) | 1.2 MB   | 0.6 MB   | **53%** |

### Why is ASON Faster?

1. **Zero key-hashing** — Schema is parsed once; data fields are mapped by position index O(1), no per-row key string hashing.
2. **Schema-driven parsing** — The deserializer knows the expected type of each field, enabling direct parsing instead of runtime type inference. CPU branch prediction hits ~100%.
3. **Minimal memory allocation** — All data rows share one schema reference. No repeated key string allocation.
4. **SIMD string scanning** — NEON/SSE2 vectorized character scanning for quoting detection, escape sequences, and quoted string parsing. Processes 16 bytes per cycle.
5. **Fast number formatting** — Two-digit lookup table for integer serialization; fast-path float formatting for common cases (integers, 1–2 decimal places).
6. **Zero-copy parsing** — Unescaped strings return pointers directly into the input buffer, avoiding string copies.

## Test Suite

36 comprehensive tests covering all functionality:

### Serialization & Deserialization (4 tests)

- Simple struct roundtrip
- Typed schema roundtrip (with type annotations)
- Vec roundtrip (schema-driven)
- Vec typed roundtrip

### Optional Fields (4 tests)

- Present value parsing
- Absent value (None)
- Optional field serialization
- Empty optional between commas `(1,,3)`

### Collections (5 tests)

- Vector field parsing and roundtrip
- Empty vector handling
- Nested vectors `[[1,2],[3,4]]`
- Map field parsing `[(key,val)]`
- Map roundtrip

### Nested Structs (3 tests)

- Nested struct parsing
- Nested struct roundtrip
- 3-level deep nesting with vector-of-structs (DeepC > DeepB > DeepA)

### String Handling (7 tests)

- Quoted string parsing
- Full escape sequence support (`\"`, `\\`, `\n`, `\t`, `\,`, `\(`, `\)`, `\[`, `\]`)
- Automatic quoting detection (SIMD-accelerated)
- Unquoted string trimming
- Strings with internal spaces
- Leading/trailing space quoting
- Backslash escape roundtrip

### Number Handling (4 tests)

- Float precision (double, float)
- Integer-valued float formatting (e.g., `42.0`)
- Negative numbers (integers and floats)
- Large unsigned integers (uint64 max)

### Parsing Features (6 tests)

- Boolean values (`true`/`false`)
- Block comment parsing `/* ... */`
- Whitespace handling
- Multiline format parsing
- Typed schema parsing (with type annotations)
- Schema field mismatch (extra/missing fields)

### Edge Cases (3 tests)

- Error handling (invalid input returns error code)
- Empty vector serialization `[]`
- Leading/trailing space quoting roundtrip

### Complex Examples (14 tests)

- Nested struct with arrays
- Vec with nested structs
- Map/dict fields
- Escaped strings roundtrip
- All-types struct (17 fields: bool, i8–i64, u8–u64, float, double, string, optional, vectors, nested vectors)
- 5-level deep nesting (Country > Region > City > District > Street > Building)
- 7-level deep nesting (Universe > Galaxy > SolarSystem > Planet > Continent > Nation > State)
- Complex config struct (nested + map + optional)
- Edge cases (empty vec, special chars, bool-like strings, number-like strings)
- Triple-nested arrays `[[1,2],[3,4],[5,6,7]]`
- Comment parsing in real data

## Build & Run

```bash
# Build all targets (Release mode for best performance)
mkdir -p build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
make -j$(nproc)

# Run tests
./ason_test

# Run examples
./basic              # 10 basic examples
./complex_example    # 14 complex examples

# Run benchmark
./bench              # ASON vs JSON comprehensive benchmark
```

## ASON Binary Format

ASON-BIN is a compact binary serialization format that shares the same schema as ASON text. It uses little-endian fixed-width encoding with no field names — purely positional — for maximum throughput.

### Wire Format

| Type       | Encoding                              |
| ---------- | ------------------------------------- |
| `bool`     | 1 byte (0 or 1)                       |
| `i8/u8`    | 1 byte                                |
| `i16/u16`  | 2 bytes LE                            |
| `i32/u32`  | 4 bytes LE                            |
| `i64/u64`  | 8 bytes LE                            |
| `f32`      | 4 bytes LE                            |
| `f64`      | 8 bytes LE                            |
| `str`      | u32 length (LE) + UTF-8 bytes         |
| `vec<T>`   | u32 count (LE) + N × element encoding |
| struct     | fields encoded sequentially           |

### API

Add `ASON_FIELDS_BIN(StructType, nfields)` right after `ASON_FIELDS(...)` to inject four functions:

```c
#include "ason.h"

typedef struct { ason_string_t name; int64_t age; double score; bool active; } User;

ASON_FIELDS(User, 4,
    ASON_FIELD(User, name,   "name",   str),
    ASON_FIELD(User, age,    "age",    i64),
    ASON_FIELD(User, score,  "score",  f64),
    ASON_FIELD(User, active, "active", bool))

ASON_FIELDS_BIN(User, 4)   /* injects: */
/*
 *  ason_buf_t  ason_dump_bin_User(const User*)
 *  ason_buf_t  ason_dump_bin_vec_User(const User*, size_t count)
 *  ason_err_t  ason_load_bin_User(const char* data, size_t len, User* out)
 *  ason_err_t  ason_load_bin_vec_User(const char* data, size_t len,
 *                                     User** out_arr, size_t* out_count)
 */
```

#### Serialize

```c
User u = { .name = ason_string_from_len("Alice", 5), .age = 30,
           .score = 95.5, .active = true };

ason_buf_t buf = ason_dump_bin_User(&u);
/* use buf.data / buf.len for network/file I/O */
ason_buf_free(&buf);
ason_string_free(&u.name);
```

#### Deserialize

```c
ason_err_t err = ason_load_bin_User(buf.data, buf.len, &u2);
assert(err == ASON_OK);
/* free heap strings when done */
ason_string_free(&u2.name);
```

#### Batch (array of structs)

```c
/* serialize */
ason_buf_t batch = ason_dump_bin_vec_User(users, count);

/* deserialize */
User* out = NULL;  size_t n = 0;
ason_load_bin_vec_User(batch.data, batch.len, &out, &n);
for (size_t i = 0; i < n; i++) ason_string_free(&out[i].name);
free(out);
ason_buf_free(&batch);
```

### Performance (macOS arm64 NEON, 100 iterations)

| Benchmark             | JSON Ser   | BIN Ser    | Ser Ratio | JSON De    | BIN De     | De Ratio | Size Saving |
| --------------------- | ---------- | ---------- | --------- | ---------- | ---------- | -------- | ----------- |
| Flat struct ×100      | 2.07 ms    | 0.30 ms    | **6.97x** | 8.86 ms    | 1.79 ms    | **4.96x**| 38%         |
| Flat struct ×1000     | 12.51 ms   | 2.31 ms    | **5.40x** | 58.84 ms   | 14.20 ms   | **4.14x**| 39%         |
| Flat struct ×5000     | 64.04 ms   | 11.35 ms   | **5.64x** | 295.65 ms  | 67.99 ms   | **4.35x**| 39%         |
| All-types ×100        | 1.41 ms    | 0.56 ms    | **2.51x** | 2.12 ms    | 1.72 ms    | **1.23x**| 31%         |
| 5-level deep ×100     | 36.11 ms   | 26.11 ms   | **1.38x** | 114.91 ms  | 58.34 ms   | **1.97x**| 61%         |
| Large payload ×10k    | 12.74 ms   | 2.22 ms    | **5.73x** | 58.11 ms   | 13.67 ms   | **4.25x**| 39%         |

> BIN is fastest for flat/homogeneous payloads. Deep nested hierarchies (5-level) still see **~2× deserialize speedup** with **61% size reduction**.

## ASON Format Specification

See the full [ASON Spec](https://github.com/athxx/ason/blob/main/docs/ASON_SPEC_CN.md) for syntax rules, BNF grammar, escape rules, type system, and LLM integration best practices.

### Syntax Quick Reference

| Element       | Schema                      | Data                |
| ------------- | --------------------------- | ------------------- |
| Object        | `{field1:type,field2:type}` | `(val1,val2)`       |
| Array         | `field:[type]`              | `[v1,v2,v3]`        |
| Object array  | `field:[{f1:type,f2:type}]` | `[(v1,v2),(v3,v4)]` |
| Map           | `field:map[K,V]`            | `[(k1,v1),(k2,v2)]` |
| Nested object | `field:{f1:type,f2:type}`   | `(v1,(v3,v4))`      |
| Null          | —                           | _(blank)_           |
| Empty string  | —                           | `""`                |
| Comment       | —                           | `/* ... */`         |

## License

MIT
