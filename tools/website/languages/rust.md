# Rust Guide

The Rust implementation is the reference implementation of ASON. It uses `serde` for serialization and supports all ASON features including SIMD-accelerated binary encoding.

## Installation

```toml
# Cargo.toml
[dependencies]
ason = "0.1"
serde = { version = "1", features = ["derive"] }
```

## StructSchema Trait

ASON requires your struct to implement `StructSchema` so the serializer knows field names and types at runtime:

```rust
use ason::{StructSchema, Result};
use serde::{Serialize, Deserialize};

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

## Text Format API

```rust
use ason::{to_string, to_string_vec, to_string_typed, to_string_vec_typed,
           from_str, from_str_vec};

// Serialize single struct
let s = to_string(&user)?;
// → "{id,name,active}:(1,Alice,true)"

// Serialize Vec<T>
let s = to_string_vec(&users)?;
// → "{id,name,active}:\n  (1,Alice,true),\n  (2,Bob,false)"

// With type annotations
let s = to_string_typed(&user)?;
// → "{id:int,name:str,active:bool}:(1,Alice,true)"

// Deserialize single struct
let user: User = from_str("{id,name,active}:(1,Alice,true)")?;

// Deserialize Vec<T>
let users: Vec<User> = from_str_vec(&s)?;
```

## Binary Format API

```rust
use ason::{to_bin, to_bin_vec, from_bin, from_bin_vec};

// Single value
let bytes: Vec<u8> = to_bin(&user)?;
let user: User      = from_bin(&bytes)?;

// Slice
let bytes: Vec<u8>  = to_bin_vec(&users)?;
let users: Vec<User> = from_bin_vec(&bytes)?;
```

### Zero-Copy Binary Deserialization

```rust
#[derive(Deserialize)]
struct BorrowedUser<'a> {
    id: i64,
    name: &'a str,   // zero-copy slice into input bytes
    active: bool,
}

let bytes = to_bin(&user)?;
let u: BorrowedUser = from_bin(&bytes)?;
```

## Pretty Printing

```rust
use ason::to_string_pretty;

let pretty = to_string_pretty(&user)?;
```

## Running Examples

```bash
cargo run --release --example basic
cargo run --release --example bench
cargo run --release --example complex
cargo run --release --example cross_compat
```

## Running Tests

```bash
cargo test
```
