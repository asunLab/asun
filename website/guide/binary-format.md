# Binary Format (ASON-BIN)

ASON-BIN is a compact binary encoding for any `serde`-compatible type. It is the fastest ASON format — ideal for internal service communication, caches, and storage.

## When to Use ASON-BIN

| Use case | Recommended format |
|----------|--------------------|
| LLM prompts / responses | ASON Text |
| Human-readable config / logs | ASON Text |
| Internal microservice RPC | **ASON-BIN** |
| Redis / disk cache | **ASON-BIN** |
| High-frequency data pipelines | **ASON-BIN** |

## Performance

**Flat struct (8 fields, × 1 000 iterations)**

| Test | JSON | ASON Text | ASON-BIN | BIN vs JSON |
|------|------|-----------|----------|-------------|
| Serialize | 2.19 ms | 1.07 ms | **0.28 ms** | **7.7×** |
| Deserialize | 6.05 ms | 5.10 ms | **2.96 ms** | **2.0×** |
| Payload size | 121 675 B | 56 716 B | **74 454 B** | **39% smaller** |

**Deep struct (5-level nested, × 100 iterations)**

| Test | JSON | ASON Text | ASON-BIN | BIN vs JSON |
|------|------|-----------|----------|-------------|
| Serialize | 4.19 ms | 2.67 ms | **0.50 ms** | **8.4×** |
| Deserialize | 9.37 ms | 8.55 ms | **3.72 ms** | **2.5×** |
| Payload size | 438 112 B | 174 611 B | **225 434 B** | **49% smaller** |

## Wire Format

All integers are **little-endian**. Strings are length-prefixed.

| Type | Encoding |
|------|----------|
| `bool` | 1 byte (0 / 1) |
| `i8` / `u8` | 1 byte |
| `i16` / `u16` | 2 bytes LE |
| `i32` / `u32` | 4 bytes LE |
| `i64` / `u64` | 8 bytes LE |
| `f32` | 4 bytes IEEE 754 LE |
| `f64` | 8 bytes IEEE 754 LE |
| `str` / `String` | `[u32 len][UTF-8 bytes]` |
| `Option<T>` | `[u8: 0=None / 1=Some][payload]` |
| `Vec<T>` / seq | `[u32 count][elements…]` |
| `struct` | fields in declaration order, no prefix |
| `enum` | `[u32 variant_index][payload]` |
| `unit` | 0 bytes |

## Zero-Copy Deserialization

When struct fields borrow from the input buffer using a `'de` lifetime, `from_bin` returns string slices directly into the input bytes — no heap allocation:

```rust
#[derive(Deserialize)]
struct BorrowedUser<'a> {
    id: i64,
    name: &'a str,  // points directly into the input buffer
    active: bool,
}

let bytes = to_bin(&user)?;
let u: BorrowedUser = from_bin(&bytes)?;
// u.name is a &str slice — zero String allocation
```

## SIMD Acceleration

`to_bin` uses platform SIMD intrinsics to bulk-copy string payloads ≥ 32 bytes:

- **ARM**: NEON 128-bit loads/stores
- **x86-64**: SSE2 128-bit loads/stores

Strings are copied in 16-byte chunks via `simd_bulk_extend`, avoiding scalar byte loops.

## Rust API

```rust
use ason::{to_bin, to_bin_vec, from_bin, from_bin_vec};

// Single value
let bytes: Vec<u8> = to_bin(&value)?;
let value: T       = from_bin(&bytes)?;

// Slice
let bytes: Vec<u8> = to_bin_vec(&values)?;
let values: Vec<T> = from_bin_vec(&bytes)?;
```

## Interoperability

ASON-BIN is a **private wire format** — it is not intended for cross-language interchange. Use ASON Text when different language runtimes need to communicate, or when human inspection is required.

Within a single language (e.g., a Rust service writing to Redis and reading back), ASON-BIN is always the best choice.
