# Syntax Reference

Complete reference for the ASON text format.

## Document Structure

Every ASON document has the form:

```
schema ":" data
```

- **Schema** — wrapped in `{ }`, defines field names and optional type hints
- **Colon** — separates schema from data
- **Data** — one or more tuples wrapped in `( )`

## Schema

```ason
{field1, field2, field3}
{field1:type1, field2:type2}
```

Fields are comma-separated. Type annotations are optional. Whitespace around field names is ignored.

### Supported Type Annotations

| Annotation | Meaning |
|-----------|---------|
| `int` | Integer (any `i64`-range value) |
| `float` | IEEE 754 double |
| `str` | String (unquoted or quoted) |
| `bool` | Boolean (`true` / `false`) |
| `{...}` | Nested struct schema |
| `[type]` | Array of `type` |

## Data Tuples

```ason
(value1, value2, value3)
```

Values are comma-separated and must match the schema order. Whitespace around values is stripped. Multiple tuples are comma-separated:

```ason
{id:int, name:str}:
  (1, Alice),
  (2, Bob),
  (3, Carol)
```

## Scalar Values

### Integer

```ason
42
-100
0
```

Any value in the `i64` range. No quotes.

### Float

```ason
3.14
-0.5
1e10
```

IEEE 754 double-precision. Scientific notation supported.

### Boolean

```ason
true
false
```

Lowercase only.

### Null / None

An empty slot — nothing between two commas (or between a comma and the closing parenthesis):

```ason
{id, name, score}:
  (1, Alice, 9.5),
  (2, Bob,      )
```

`Bob`'s score is null/None.

### Unquoted String

Any sequence of characters that does not contain `,`, `(`, `)`, `[`, `]`, or `\`. Leading and trailing whitespace is auto-trimmed:

```ason
Alice Smith
hello world
```

### Quoted String

Wrap in double quotes to preserve whitespace or include reserved characters:

```ason
"  leading spaces  "
"value with, comma"
"line\nnewline"
```

Escape sequences inside quoted strings:

| Escape | Meaning |
|--------|---------|
| `\\` | Literal backslash |
| `\"` | Double quote |
| `\n` | Newline |
| `\t` | Tab |
| `\r` | Carriage return |
| `\,` | Literal comma (unquoted context) |

## Nested Structures

Nested objects are represented by nested schemas and tuples:

```ason
{id:int, address:{city:str, zip:str}}:
  (1, (Berlin, 10115)),
  (2, (Paris,  75001))
```

## Arrays

Array fields use `[...]` in both schema and data:

```ason
{id:int, tags:[str]}:
  (1, [rust, go]),
  (2, [python, c++])
```

Arrays can be empty:

```ason
(1, [])
```

## Enums

Enum variants are written in `Variant::Value` form:

```ason
Role::Admin
Status::Active
```

## Comments

ASON does not have a comment syntax. Use ASON inside a host format (TOML, YAML) that supports comments if needed.

## Single-Struct Shorthand

When serializing a single struct (not a list), schema and data appear together on one line:

```ason
{id:int, name:str, active:bool}:(1, Alice, true)
```

## Whitespace Rules

- Whitespace between tokens is ignored
- Unquoted string values have leading/trailing whitespace stripped
- Newlines between tuples are optional but recommended for readability
