# ASON — Array-Schema Object Notation

> _"The efficiency of arrays, the structure of objects."_

ASON is a high-performance serialization format designed for large-scale data exchange and LLM
(Large Language Model) interaction. By separating **schema from data**, ASON eliminates the
redundant key repetition of JSON, introduces a Markdown-like row-oriented syntax, and achieves
dramatic token reduction while preserving human readability.

---

## Why ASON?

### Token Efficiency

```json
// JSON — 100 tokens
{
  "users": [
    { "id": 1, "name": "Alice", "active": true },
    { "id": 2, "name": "Bob", "active": false }
  ]
}
```

```ason
// ASON — ~35 tokens (65% savings)
{id:int, name:str, active:bool}:
  (1, Alice, true),
  (2, Bob,   false)
```

### Performance vs JSON

| Metric                | ASON Text           | ASON-BIN           |
| --------------------- | ------------------- | ------------------ |
| Serialization speed   | **2–4× faster**     | **7–10× faster**   |
| Deserialization speed | **1.2–2.5× faster** | **2–2.5× faster**  |
| Payload size          | **53–61% smaller**  | **38–49% smaller** |

---

## Installation

Add to `Cargo.toml`:

```toml
[dependencies]
ason = { path = "rust" }   # local
serde = { version = "1", features = ["derive"] }
```

---

## Text Format API

### Data types

| Type            | Example         | Notes                         |
| --------------- | --------------- | ----------------------------- |
| Integer         | `42`, `-100`    | Any `i64`-range value         |
| Float           | `3.14`, `-0.5`  | IEEE 754 double               |
| Bool            | `true`, `false` | Lowercase only                |
| Null            | _(empty)_       | Empty slot between commas     |
| Unquoted string | `Alice Smith`   | Auto-trimmed, escape `,()[]\` |
| Quoted string   | `" spaces "`    | Preserves whitespace          |
| Array           | `[a, b, c]`     | Nested lists                  |
| Nested struct   | `{...}:(...)`   | Recursive schema              |

### `StructSchema` trait

ASON uses a compile-time schema trait so the serializer knows field names and types:

```rust
use ason::{Result, StructSchema, from_str, from_str_vec, to_string, to_string_vec};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, PartialEq)]
struct User {
    id: i64,
    name: String,
    active: bool,
}

impl StructSchema for User {
    fn field_names() -> &'static [&'static str] { &["id", "name", "active"] }
    fn field_types() -> &'static [&'static str] { &["int", "str", "bool"] }
    fn serialize_fields(&self, ser: &mut ason::serialize::Serializer) -> Result<()> {
        use serde::Serialize;
        self.id.serialize(&mut *ser)?;
        self.name.serialize(&mut *ser)?;
        self.active.serialize(&mut *ser)?;
        Ok(())
    }
}
```

### Serialization

```rust
// Single struct  →  "{id,name,active}:(1,Alice,true)"
let s = to_string(&user)?;

// Vec<T>  →  "{id,name,active}:\n  (1,Alice,true),\n  (2,Bob,false)"
let s = to_string_vec(&users)?;

// With type annotations
let s = to_string_typed(&user)?;       // "{id:int,name:str,active:bool}:..."
let s = to_string_vec_typed(&users)?;
```

### Deserialization

```rust
// Single struct (accepts both annotated and plain schemas)
let user: User = from_str("{id,name,active}:(1,Alice,true)")?;

// Vec<T>
let users: Vec<User> = from_str_vec(&s)?;
```

---

## Binary Format (ASON-BIN)

ASON-BIN is a compact binary encoding of any `serde`-compatible type. It provides the largest
performance gains, making it ideal for internal service communication, caches, and storage.

### API

```rust
use ason::{to_bin, to_bin_vec, from_bin, from_bin_vec};
```

| Function       | Signature                             | Description                              |
| -------------- | ------------------------------------- | ---------------------------------------- |
| `to_bin`       | `(value: &T) -> Result<Vec<u8>>`      | Serialize any `T: Serialize` to bytes    |
| `from_bin`     | `(data: &'de [u8]) -> Result<T>`      | Deserialize with zero-copy string slices |
| `to_bin_vec`   | `(values: &[T]) -> Result<Vec<u8>>`   | Serialize a slice to bytes               |
| `from_bin_vec` | `(data: &'de [u8]) -> Result<Vec<T>>` | Deserialize a sequence from bytes        |

### Usage

```rust
use ason::{from_bin, from_bin_vec, to_bin, to_bin_vec};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, PartialEq)]
struct User {
    id: i64,
    name: String,
    active: bool,
}

fn main() -> ason::Result<()> {
    // Single value
    let user = User { id: 1, name: "Alice".into(), active: true };
    let bytes = to_bin(&user)?;
    let restored: User = from_bin(&bytes)?;
    assert_eq!(user, restored);

    // Vec
    let users = vec![
        User { id: 1, name: "Alice".into(), active: true },
        User { id: 2, name: "Bob".into(),   active: false },
    ];
    let bytes = to_bin_vec(&users)?;
    let restored: Vec<User> = from_bin_vec(&bytes)?;
    assert_eq!(users, restored);

    Ok(())
}
```

### Zero-Copy Deserialization

When your struct borrows string data with a `'de` lifetime, `from_bin` returns slices directly
into the input buffer — no heap allocation for string fields:

```rust
#[derive(Debug, Deserialize)]
struct BorrowedUser<'a> {
    id: i64,
    name: &'a str,   // zero-copy — points directly into the input bytes
    active: bool,
}

let bytes = to_bin(&User { id: 1, name: "Alice".into(), active: true })?;
let u: BorrowedUser = from_bin(&bytes)?;
// u.name is a &str slice into `bytes`, no String allocation
```

### Wire Format

All integers are **little-endian**. Strings are length-prefixed (u32 LE + UTF-8 bytes).

| Type             | Encoding                               |
| ---------------- | -------------------------------------- |
| `bool`           | 1 byte (0/1)                           |
| `i8` / `u8`      | 1 byte                                 |
| `i16` / `u16`    | 2 bytes LE                             |
| `i32` / `u32`    | 4 bytes LE                             |
| `i64` / `u64`    | 8 bytes LE                             |
| `f32`            | 4 bytes IEEE 754 LE                    |
| `f64`            | 8 bytes IEEE 754 LE                    |
| `char`           | 4 bytes (u32 codepoint LE)             |
| `str` / `String` | `[u32 len][UTF-8 bytes]`               |
| `Option<T>`      | `[u8 tag: 0=None / 1=Some][payload]`   |
| `Vec<T>` / seq   | `[u32 count][elements...]`             |
| `HashMap` / map  | `[u32 count][key value pairs...]`      |
| `struct`         | fields in declaration order, no prefix |
| `tuple`          | elements in order, no prefix           |
| `enum`           | `[u32 variant_index][payload]`         |
| `unit`           | 0 bytes                                |

### Performance (Apple M-series, release build)

**Flat struct (8 fields)**

| Test                | JSON      | ASON Text | ASON-BIN     | BIN vs JSON |
| ------------------- | --------- | --------- | ------------ | ----------- |
| Serialize × 1000    | 2.19 ms   | 1.07 ms   | **0.28 ms**  | **7.7×**    |
| Deserialize × 1000  | 6.05 ms   | 5.10 ms   | **2.96 ms**  | **2.0×**    |
| Payload size × 1000 | 121,675 B | 56,716 B  | **74,454 B** | 39% smaller |

**Deep struct (5-level nested, × 100)**

| Test               | JSON      | ASON Text | ASON-BIN      | BIN vs JSON |
| ------------------ | --------- | --------- | ------------- | ----------- |
| Serialize × 100    | 4.19 ms   | 2.67 ms   | **0.50 ms**   | **8.4×**    |
| Deserialize × 100  | 9.37 ms   | 8.55 ms   | **3.72 ms**   | **2.5×**    |
| Payload size × 100 | 438,112 B | 174,611 B | **225,434 B** | 49% smaller |

**Single roundtrip (User, × 100,000)**

| Format    | Time/op | vs JSON         |
| --------- | ------- | --------------- |
| ASON-BIN  | ~182 ns | **2.1× faster** |
| JSON      | ~375 ns | —               |
| ASON Text | ~552 ns | 0.7×            |

### SIMD Acceleration

`to_bin` uses platform SIMD intrinsics (NEON on ARM, SSE2 on x86-64) to bulk-copy string
payloads ≥ 32 bytes in 16-byte chunks via `simd_bulk_extend`, avoiding scalar byte loops.

---

## Running Examples

```bash
# Basic usage
cargo run --release --example basic

# Binary format (zero-copy demo, wire layout, benchmarks)
cargo run --release --example binary

# Comprehensive benchmark suite (all 9 sections)
cargo run --release --example bench

# Complex nested structures
cargo run --release --example complex
```

---

## ASON Text Format Syntax

```ason
// Single struct
{id, name, active}:(1, Alice, true)

// Vec of structs  (schema declared once)
{id:int, name:str, active:bool}:
  (1, Alice, true),
  (2, Bob,   false),
  (3, Carol, true)

// Nested struct
{id:int, address:{city:str, zip:str}}:
  (1, (Berlin, 10115)),
  (2, (Paris,  75001))

// Nested arrays
{id:int, tags:[str]}:
  (1, [rust, go]),
  (2, [python])

// Enum value
Role::Admin

// Option  (empty slot = None)
{id, name, score}:
  (1, Alice,  9.5),
  (2, Bob,       )    ← score is None
```

---

## License

MIT
