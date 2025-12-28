# ASON - Array-Schema Object Notation

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

A **token-efficient** data format that separates schema from data to minimize redundancy. Designed specifically for LLM interactions and batch data transmission.

## Why ASON?

When working with LLMs, every token counts. Traditional formats like JSON repeat field names for every record, wasting tokens and increasing costs. ASON solves this by declaring the schema once and using positional data.

```
JSON (62 tokens):
[{"name":"Alice","age":30},{"name":"Bob","age":25}]

ASON (~35 tokens):
{name,age}:(Alice,30),(Bob,25)
```

**~45% token reduction** on batch data!

## Format Comparison

### ASON vs JSON

| Feature | JSON | ASON |
|---------|------|------|
| Field names | Repeated per record | Declared once in schema |
| Batch efficiency | Poor (redundant keys) | Excellent (schema reuse) |
| Single-line friendly | Yes | Yes |
| Nested structures | Yes | Yes |
| Type inference | No (explicit) | Yes (auto-detect) |
| LLM token cost | High | Low (~30-50% less) |

**Example - 100 user records:**
```
JSON:  [{"id":1,"name":"Alice","email":"a@x.com"},{"id":2,"name":"Bob","email":"b@x.com"},...]
       → ~1500 tokens (field names repeated 100 times)

ASON:  {id,name,email}:(1,Alice,a@x.com),(2,Bob,b@x.com),...
       → ~800 tokens (field names declared once)
```

### ASON vs CSV

| Feature | CSV | ASON |
|---------|-----|------|
| Schema definition | Header row | Explicit schema syntax |
| Nested objects | ❌ Not supported | ✅ Supported |
| Arrays | ❌ Not supported | ✅ Supported |
| Type information | No | Yes (bool, int, float, string, null) |
| Comments | No | Yes (`/* ... */`) |
| Single value | Awkward | Natural |

**Example - Nested data:**
```
CSV cannot represent:
{name,addr{city,zip}}:(Alice,(NYC,10001))

→ {"name":"Alice","addr":{"city":"NYC","zip":10001}}
```

### ASON vs TOON

[TOON](https://github.com/BenoitZugmeyer/toon) (Token-Oriented Object Notation) is another LLM-optimized format. Key differences:

| Feature | TOON | ASON |
|---------|------|------|
| Schema location | Inline with data | Separate, prefix schema |
| Batch records | Repeated schema per record | Schema once, multiple records |
| Single-line support | Multi-line focused | **Native single-line** |
| Compression | Good | **Better for batch** |

**TOON:**
```toon
name: Alice
age: 30
---
name: Bob
age: 25
```

**ASON:**
```ason
{name,age}:(Alice,30),(Bob,25)
```

#### ASON's Single-Line Advantage

ASON is designed to be **single-line friendly** from the ground up:

1. **Stream-friendly**: Can be sent in one line without newlines
2. **Embedding-friendly**: Easy to embed in prompts, logs, or URLs
3. **Copy-paste friendly**: Single line is easier to copy and share
4. **Compact API responses**: No formatting overhead

```
# Multi-line format (like TOON) requires escaping or special handling in prompts:
"Parse this data:\nname: Alice\nage: 30"

# ASON works naturally in single line:
"Parse this data: {name,age}:(Alice,30)"
```

## Quick Syntax Reference

| Syntax | Example | Description |
|--------|---------|-------------|
| Object | `{name,age}:(Alice,30)` | Schema + data |
| Multiple records | `{id}:(1),(2),(3)` | Same schema, batch data |
| Nested object | `{user{name}}:((Alice))` | Nested structure |
| Array field | `{scores[]}:([90,85])` | Simple array |
| Object array | `{users[{id}]}:([(1),(2)])` | Array of objects |
| Plain array | `[1,2,3]` | No schema needed |
| Null value | `{a,b}:(1,)` | b = null |
| Empty string | `{a}:("")` | Explicit empty |
| Comments | `/* note */ {a}:(1)` | Block comments |

## Implementations

| Language | Path | Status |
|----------|------|--------|
| Rust | [`rust/`](rust/) | ✅ Complete |

## Use Cases

- **LLM API calls**: Reduce token usage by 30-50%
- **Batch data transfer**: Send multiple records efficiently
- **Configuration**: Compact config files
- **Logging**: Structured logs in single lines
- **Streaming**: Line-by-line data streams

## Specification

See [ASON Format Specification](docs/ASON_SPEC.md) for complete syntax and rules.

## License

MIT

