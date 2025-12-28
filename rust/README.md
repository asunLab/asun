# ASON - Array-Schema Object Notation

[![Rust](https://img.shields.io/badge/rust-1.85+-orange.svg)](https://www.rust-lang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A **token-efficient** data format that reduces redundancy by separating schema from data. Optimized for LLM interactions.

## Features

- 🚀 **Token Efficient** - Field names declared once, reducing 30-50% tokens
- 📦 **Schema Separation** - Structure and data separated, batch data more compact
- 🔧 **Zero-Copy Parsing** - Optional zero-copy mode for high-performance scenarios
- 🦀 **Serde Support** - Seamless integration with Rust ecosystem

## Quick Start

```toml
[dependencies]
ason = "0.1"
```

## Basic Usage

### Parsing ASON

```rust
use ason::parse;

// Single object
let value = parse("{name,age}:(Alice,30)").unwrap();
assert_eq!(value.get("name").unwrap().as_str(), Some("Alice"));
assert_eq!(value.get("age").unwrap().as_i64(), Some(30));

// Multiple records
let users = parse("{name,age}:(Alice,30),(Bob,25)").unwrap();
let arr = users.as_array().unwrap();
assert_eq!(arr.len(), 2);
```

### Serialization

```rust
use ason::{Value, to_string, to_string_pretty};
use indexmap::IndexMap;

let mut obj = IndexMap::new();
obj.insert("name".to_string(), Value::String("Alice".to_string()));
obj.insert("age".to_string(), Value::Integer(30));
let value = Value::Object(obj);

assert_eq!(to_string(&value), "(Alice,30)");
```

### Serde Integration

```rust
use ason::{parse, from_value, to_value};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
struct User {
    name: String,
    age: i64,
}

// ASON → Rust struct
let value = parse("{name,age}:(Alice,30)").unwrap();
let user: User = from_value(&value).unwrap();

// Rust struct → ASON Value
let value = to_value(&user).unwrap();
```

### Zero-Copy Parsing (High Performance)

```rust
use ason::zero_copy;

let input = "{name,age}:(Alice,30)";
let value = zero_copy::parse(input).unwrap();

// Strings borrow directly from input, no memory allocation
assert_eq!(value.get("name").unwrap().as_str(), Some("Alice"));

// Convert to owned when independent lifetime needed
let owned = value.into_owned();
```

## ASON Syntax Overview

| Syntax | Example | Description |
|--------|---------|-------------|
| Object | `{name,age}:(Alice,30)` | Schema + data |
| Multiple records | `{id}:(1),(2),(3)` | Same structure batch data |
| Nested | `{user{name}}:((Alice))` | Nested object |
| Array field | `{scores[]}:([90,85])` | Simple array |
| Object array | `{users[{id}]}:([(1),(2)])` | Array of objects |
| Plain array | `[1,2,3]` | No schema |
| Null | `{a,b}:(1,)` | b = null |

### Comparison with JSON

```
JSON (62 tokens):
[{"name":"Alice","age":30},{"name":"Bob","age":25}]

ASON (~35 tokens):
{name,age}:(Alice,30),(Bob,25)
```

## API Overview

### Parsing

| Function | Description |
|----------|-------------|
| `parse(input)` | Parse ASON string |
| `zero_copy::parse(input)` | Zero-copy parsing |

### Serialization

| Function | Description |
|----------|-------------|
| `to_string(&value)` | Compact output |
| `to_string_pretty(&value)` | Formatted output |

### Serde

| Function | Description |
|----------|-------------|
| `to_value(&T)` | Rust value → ASON Value |
| `from_value(Value)` | ASON Value → Rust value |

## Run Examples

```bash
cargo run --example basic
```

## License

MIT

