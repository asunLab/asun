# ason-go

[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.24+-00ADD8.svg)](https://go.dev)

A high-performance Marshal/Unmarshal library for [ASON](https://github.com/athxx/ason) (Array-Schema Object Notation) in Go — a token-efficient, schema-driven data format designed for LLM interactions and large-scale data transmission.

**Zerocopy design, zero dependencies (stdlib only), extreme performance.**

[中文文档](README_CN.md)

## What is ASON?

ASON separates **schema** from **data**, eliminating repetitive keys found in JSON. The schema is declared once, and data rows carry only values:

```text
JSON (100 tokens):
{"users":[{"id":1,"name":"Alice","active":true},{"id":2,"name":"Bob","active":false}]}

ASON (~35 tokens, 65% saving):
[{id:int, name:str, active:bool}]:(1,Alice,true),(2,Bob,false)
```

| Aspect              | JSON         | ASON                  |
| ------------------- | ------------ | --------------------- |
| Token efficiency    | 100%         | 30–70% ✓              |
| Key repetition      | Every object | Declared once ✓       |
| Human readable      | Yes          | Yes ✓                 |
| Nested structs      | ✓            | ✓                     |
| Type annotations    | No           | Optional ✓            |
| Serialization speed | 1x           | **1.4–1.9x faster** ✓ |
| Deserialization     | 1x           | **2.9–4.4x faster** ✓ |
| Data size           | 100%         | **40–50%** ✓          |

## Quick Start

```bash
go get github.com/example/ason
```

### Marshal & Unmarshal a Struct

```go
package main

import (
    "fmt"
    ason "github.com/example/ason"
)

type User struct {
    ID     int64  `ason:"id"`
    Name   string `ason:"name"`
    Active bool   `ason:"active"`
}

func main() {
    user := User{ID: 1, Name: "Alice", Active: true}

    // Marshal
    b, _ := ason.Marshal(&user)
    fmt.Println(string(b))
    // Output: {id,name,active}:(1,Alice,true)

    // Unmarshal
    var user2 User
    ason.Unmarshal(b, &user2)
    fmt.Println(user == user2) // true
}
```

### Marshal with Type Annotations

Use `MarshalTyped` to output a type-annotated schema — useful for documentation, LLM prompts, and cross-language exchange:

```go
b, _ := ason.MarshalTyped(&user)
// Output: {id:int,name:str,active:bool}:(1,Alice,true)

// Unmarshal accepts both annotated and unannotated schemas
var user2 User
ason.Unmarshal(b, &user2)
```

### Marshal & Unmarshal a Slice (Schema-Driven)

For `[]T`, ASON writes the schema **once** and emits each element as a compact tuple — the key advantage over JSON:

```go
users := []User{
    {ID: 1, Name: "Alice", Active: true},
    {ID: 2, Name: "Bob", Active: false},
}

// Unannotated schema
b, _ := ason.MarshalSlice(users)
// Output: [{id,name,active}]:(1,Alice,true),(2,Bob,false)

// Type-annotated schema
b2, _ := ason.MarshalSliceTyped(users, []string{"int", "str", "bool"})
// Output: [{id:int,name:str,active:bool}]:(1,Alice,true),(2,Bob,false)

// Unmarshal — accepts both forms
var parsed []User
ason.UnmarshalSlice(b, &parsed)
```

## Binary Format (ASON-BIN)

ASON-BIN is a compact binary encoding of any Go value. It provides the largest performance gains, making it ideal for internal service communication, caches, and storage.

### API

```go
import ason "github.com/example/ason"

// Marshal any value to bytes
b, err := ason.MarshalBinary(&user)

// Unmarshal from bytes (zero-copy for strings and []byte)
var u2 User
err = ason.UnmarshalBinary(b, &u2)
```

### Zero-Copy Deserialization

When unmarshaling ASON-BIN, `string` and `[]byte` fields are created using zero-copy techniques (`unsafe.String` and slice reslicing). They point directly into the input byte slice, avoiding heap allocations and memory copying.

### Performance (Apple M-series)

**Flat struct (8 fields)**

| Test | JSON | ASON Text | ASON-BIN | BIN vs JSON |
|---|---|---|---|---|
| Serialize × 1000 | 0.17 ms | 0.08 ms | **0.08 ms** | **2.3× faster** |
| Deserialize × 1000 | 0.83 ms | 0.21 ms | **0.07 ms** | **12.6× faster** |
| Payload size × 1000 | 122,071 B | 57,112 B | **74,784 B** | 39% smaller |

**Deep struct (5-level nested, × 100)**

| Test | JSON | ASON Text | ASON-BIN | BIN vs JSON |
|---|---|---|---|---|
| Serialize × 100 | 0.35 ms | 0.20 ms | **0.05 ms** | **7.0× faster** |
| Deserialize × 100 | 1.20 ms | 0.45 ms | **0.08 ms** | **15.0× faster** |
| Payload size × 100 | 43,811 B | 17,461 B | **22,543 B** | 49% smaller |

## Supported Types

| Type                            | ASON Representation | Example                      |
| ------------------------------- | ------------------- | ---------------------------- |
| Integers (int8–int64, uint8–64) | Plain number        | `42`, `-100`                 |
| Floats (float32, float64)       | Decimal number      | `3.14`, `-0.5`               |
| Bool                            | Literal             | `true`, `false`              |
| String                          | Unquoted or quoted  | `Alice`, `"Carol Smith"`     |
| Pointer (\*T)                   | Value or empty      | `hello` or _(blank)_ for nil |
| Slice ([]T)                     | `[v1,v2,v3]`        | `[rust,go,python]`           |
| Map (map[K]V)                   | `[(k1,v1),(k2,v2)]` | `[(age,30),(score,95)]`      |
| Nested struct                   | `(field1,field2)`   | `(Engineering,500000)`       |

### Struct Tags

The library reads `ason` tags (priority), falling back to `json` tags. Use `"-"` to skip a field:

```go
type Employee struct {
    ID     int64      `ason:"id"`
    Name   string     `ason:"name"`
    Dept   Department `ason:"dept"`
    Secret string     `ason:"-"`  // skipped
}
```

### Nested Structs

```go
type Dept struct {
    Title string `ason:"title"`
}

type Employee struct {
    Name string `ason:"name"`
    Dept Dept   `ason:"dept"`
}

// Schema reflects nesting:
// {name,dept}:(Alice,(Engineering))
```

### Optional Fields (Pointers)

```go
type Item struct {
    ID    int64   `ason:"id"`
    Label *string `ason:"label"`
}

// With value:  {id,label}:(1,hello)
// With nil:    {id,label}:(1,)
```

### Slices & Maps

```go
type Tagged struct {
    Name string   `ason:"name"`
    Tags []string `ason:"tags"`
}
// {name,tags}:(Alice,[rust,go,python])

type Profile struct {
    Name  string           `ason:"name"`
    Attrs map[string]int64 `ason:"attrs"`
}
// {name,attrs}:(Alice,[(age,30),(score,95)])
```

### Type Annotations (Optional)

ASON schema supports **optional** type annotations. Both forms are fully equivalent — the unmarshaler handles them identically:

```text
// Without annotations (default output of Marshal / MarshalSlice)
{id,name,salary,active}:(1,Alice,5000.50,true)

// With annotations (output of MarshalTyped / MarshalSliceTyped)
{id:int,name:str,salary:float,active:bool}:(1,Alice,5000.50,true)
```

Annotations are **purely decorative metadata** — they do not affect parsing or unmarshaling behavior. The unmarshaler simply skips the `:type` portion when present.

**When to use annotations:**

- LLM prompts — helps models understand and generate correct data
- API documentation — self-describing schema without external docs
- Cross-language exchange — eliminates type ambiguity (is `42` an int or float?)
- Debugging — see data types at a glance

**Performance impact:** negligible. The overhead is constant and does not grow with data volume — annotations only affect the schema header, not the data body.

### Comments

```text
/* user list */
[{id:int, name:str, active:bool}]:(1,Alice,true),(2,Bob,false)
```

### Multiline Format

```text
[{id:int, name:str, active:bool}]:
  (1, Alice, true),
  (2, Bob, false),
  (3, "Carol Smith", true)
```

## API Reference

| Function                                        | Description                                                 |
| ----------------------------------------------- | ----------------------------------------------------------- |
| `Marshal(v any) ([]byte, error)`                | Marshal a struct → unannotated schema `{id,name}:`          |
| `MarshalTyped(v any) ([]byte, error)`           | Marshal a struct → annotated schema `{id:int,name:str}:`    |
| `Unmarshal(data []byte, v any) error`           | Unmarshal a struct (accepts both annotated and unannotated) |
| `MarshalSlice(v any) ([]byte, error)`           | Marshal a slice → unannotated schema (schema-driven)        |
| `MarshalSliceTyped(v, types) ([]byte, error)`   | Marshal a slice → annotated schema                          |
| `MarshalSliceFields(v, fields) ([]byte, error)` | Marshal a slice with custom field names                     |
| `UnmarshalSlice(data []byte, v any) error`      | Unmarshal a slice (accepts both annotated and unannotated)  |

## Performance

Benchmarked on Apple Silicon (M-series), comparing against `encoding/json`:

### Serialization (ASON is 1.4–1.9x faster)

| Scenario            | JSON     | ASON     | Speedup   |
| ------------------- | -------- | -------- | --------- |
| Flat struct × 100   | 2.74 ms  | 1.41 ms  | **1.95x** |
| Flat struct × 500   | 9.72 ms  | 5.27 ms  | **1.85x** |
| Flat struct × 1000  | 15.52 ms | 8.87 ms  | **1.75x** |
| Flat struct × 5000  | 79.15 ms | 55.15 ms | **1.44x** |
| 5-level deep × 10   | 6.80 ms  | 4.19 ms  | **1.62x** |
| 5-level deep × 50   | 32.90 ms | 21.15 ms | **1.56x** |
| 5-level deep × 100  | 65.79 ms | 41.88 ms | **1.57x** |
| Large payload (10k) | 17.07 ms | 10.57 ms | **1.62x** |

### Deserialization (ASON is 2.9–4.4x faster)

| Scenario            | JSON      | ASON      | Speedup   |
| ------------------- | --------- | --------- | --------- |
| Flat struct × 100   | 13.29 ms  | 3.04 ms   | **4.37x** |
| Flat struct × 500   | 44.46 ms  | 11.76 ms  | **3.78x** |
| Flat struct × 1000  | 84.93 ms  | 22.88 ms  | **3.71x** |
| Flat struct × 5000  | 440.20 ms | 120.71 ms | **3.65x** |
| 5-level deep × 10   | 35.02 ms  | 11.47 ms  | **3.05x** |
| 5-level deep × 50   | 175.52 ms | 58.04 ms  | **3.02x** |
| 5-level deep × 100  | 350.60 ms | 118.59 ms | **2.96x** |
| Large payload (10k) | 91.14 ms  | 27.53 ms  | **3.31x** |

### Single Struct Roundtrip (10,000 iterations)

| Scenario | JSON     | ASON     | Speedup   |
| -------- | -------- | -------- | --------- |
| Flat     | 11.37 ms | 4.95 ms  | **2.30x** |
| Deep     | 48.20 ms | 20.59 ms | **2.34x** |

### Throughput (1000 records × 100 iterations)

| Operation   | JSON            | ASON             | Speedup          |
| ----------- | --------------- | ---------------- | ---------------- |
| Serialize   | 6,158,789 rec/s | 11,426,939 rec/s | **1.86x faster** |
| Deserialize | 1,133,686 rec/s | 4,291,293 rec/s  | **3.79x faster** |

### Size Savings

| Scenario            | JSON   | ASON   | Saving  |
| ------------------- | ------ | ------ | ------- |
| Flat struct × 100   | 12 KB  | 5.6 KB | **54%** |
| Flat struct × 1000  | 122 KB | 57 KB  | **53%** |
| 5-level deep × 10   | 43 KB  | 17 KB  | **60%** |
| 5-level deep × 100  | 432 KB | 175 KB | **60%** |
| Large payload (10k) | 1.2 MB | 0.6 MB | **53%** |

### Why is ASON Faster?

1. **Zero key-hashing** — Schema is parsed once; data fields are mapped by position index `O(1)`, no per-row key string hashing.
2. **Schema-driven parsing** — The unmarshaler knows the expected type of each field from the struct, enabling direct parsing (`parseInt64()`) instead of runtime type inference. CPU branch prediction hits ~100%.
3. **Minimal memory allocation** — `sync.Pool` buffer reuse, `[256]bool` lookup tables for character classification, `unsafe.String` for zerocopy byte-to-string conversion.
4. **Struct info caching** — Reflection results cached in `sync.Map`, eliminating repeated reflection overhead.
5. **No intermediate representation** — Data is written directly to byte buffers without any DOM/AST layer.

## Implementation Highlights

- **Zero dependencies** — only Go stdlib (`reflect`, `sync`, `unsafe`, `math`, `strconv`)
- **Zerocopy** — `unsafe.String` for string conversions, avoids unnecessary allocations
- **Fast number formatting** — custom `appendI64`/`appendU64`/`appendFloat64` with fast paths for common cases (1-decimal, 2-decimal, integer-valued floats)
- **Lookup tables** — `[256]bool` for quoting decisions, `[256]byte` for escape characters
- **Buffer pooling** — `sync.Pool` for buffer reuse across marshal calls
- **Cached struct info** — `sync.Map` caches struct field metadata per type

Run the benchmark yourself:

```bash
go run examples/bench/main.go
```

## Examples

```bash
# Basic usage (11 examples)
go run examples/basic/main.go

# Comprehensive (17 examples — nested structs, 7-level nesting, large structures, edge cases)
go run examples/complex/main.go

# Performance benchmark (ASON vs JSON, throughput, memory, 8 sections)
go run examples/bench/main.go
```

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
