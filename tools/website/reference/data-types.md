# Data Types

ASON supports a rich set of scalar and composite types. This page documents their text-format representation and cross-language equivalents.

## Scalar Types

| ASON type | Example | Notes |
|-----------|---------|-------|
| `int` | `42`, `-100`, `0` | Any `i64`-range integer |
| `float` | `3.14`, `-0.5`, `1e10` | IEEE 754 double |
| `bool` | `true`, `false` | Lowercase only |
| `str` (unquoted) | `Alice Smith` | Auto-trimmed, `\,` to escape comma |
| `str` (quoted) | `" spaces "` | Preserves whitespace, supports `\"` `\n` `\t` |
| null / None | _(empty slot)_ | Empty between commas |

## Composite Types

### Nested Struct

```ason
// Schema
{id:int, address:{city:str, zip:str}}

// Data
(1, (Berlin, 10115))
```

### Array / List

```ason
// Schema
{id:int, tags:[str]}

// Data
(1, [rust, go, wasm])
```

Arrays can be nested:

```ason
{matrix:[[int]]}:
  ([[1, 2], [3, 4]]),
  ([[5, 6], [7, 8]])
```

### Enum

```ason
Role::Admin
Status::Active
Color::Rgb
```

Enum variants are namespaced with `::`.

### Option / Nullable

An empty slot (nothing between commas) represents `None` / `null`:

```ason
{id:int, score:float}:
  (1, 9.5),
  (2,    )
```

## Cross-Language Type Mapping

| ASON type | Rust | Go | Python | Java | C |
|-----------|------|----|--------|------|---|
| `int` | `i64` | `int64` | `int` | `long` | `int64_t` |
| `float` | `f64` | `float64` | `float` | `double` | `double` |
| `bool` | `bool` | `bool` | `bool` | `boolean` | `bool` |
| `str` | `String` / `&str` | `string` | `str` | `String` | `char*` |
| null | `Option<T>` | pointer / `*T` | `Optional` | `Optional<T>` | nullable ptr |
| `[T]` | `Vec<T>` | `[]T` | `list` | `List<T>` | `T[]` |
| nested struct | struct | struct | dataclass | class | struct |

## Integer Sizes (ASON-BIN)

In the binary format, integers use their native sizes:

| Type | Bytes |
|------|-------|
| `i8` / `u8` | 1 |
| `i16` / `u16` | 2 |
| `i32` / `u32` | 4 |
| `i64` / `u64` | 8 |

All integers are **little-endian** in ASON-BIN.
