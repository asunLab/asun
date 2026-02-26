# ason-python

[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Python](https://img.shields.io/badge/python-3.9+-3776AB.svg)](https://python.org)

A high-performance dump/load library for [ASON](https://github.com/athxx/ason) (Array-Schema Object Notation) in Python — a token-efficient, schema-driven data format designed for LLM interactions and large-scale data transmission.

**Pure Python, zero dependencies (stdlib only), dataclass-driven.**

[中文文档](README_CN.md)

## What is ASON?

ASON separates **schema** from **data**, eliminating repetitive keys found in JSON. The schema is declared once, and data rows carry only values:

```text
JSON (100 tokens):
{"users":[{"id":1,"name":"Alice","active":true},{"id":2,"name":"Bob","active":false}]}

ASON (~35 tokens, 65% saving):
[{id:int, name:str, active:bool}]:(1,Alice,true),(2,Bob,false)
```

| Aspect           | JSON         | ASON            |
| ---------------- | ------------ | --------------- |
| Token efficiency | 100%         | 30–70% ✓        |
| Key repetition   | Every object | Declared once ✓ |
| Human readable   | Yes          | Yes ✓           |
| Nested structs   | ✓            | ✓               |
| Type annotations | No           | Optional ✓      |
| Data size        | 100%         | **40–50%** ✓    |

## Quick Start

Copy `ason.py` into your project — no installation required.

### Dump & Load a Struct

```python
from dataclasses import dataclass
import ason

@dataclass
class User:
    id: int
    name: str
    active: bool

user = User(id=1, name="Alice", active=True)

# Dump
s = ason.dump(user)
print(s)
# Output: {id,name,active}:(1,Alice,true)

# Load
user2 = ason.load(s, User)
print(user == user2)  # True
```

### Dump with Type Annotations

Use `dump_typed` to output a type-annotated schema — useful for documentation, LLM prompts, and cross-language exchange:

```python
s = ason.dump_typed(user)
# Output: {id:int,name:str,active:bool}:(1,Alice,true)

# load accepts both annotated and unannotated schemas
user2 = ason.load(s, User)
```

### Dump & Load a List (Schema-Driven)

For `list[T]`, ASON writes the schema **once** and emits each element as a compact tuple — the key advantage over JSON:

```python
users = [
    User(id=1, name="Alice", active=True),
    User(id=2, name="Bob", active=False),
]

# Unannotated schema
s = ason.dump_slice(users)
# Output: [{id,name,active}]:(1,Alice,true),(2,Bob,false)

# Type-annotated schema
s2 = ason.dump_slice_typed(users)
# Output: [{id:int,name:str,active:bool}]:(1,Alice,true),(2,Bob,false)

# Load — accepts both forms
parsed = ason.load_slice(s, User)
```

## Supported Types

| Type                  | ASON Representation | Example                      |
| --------------------- | ------------------- | ---------------------------- |
| `int`                 | Plain number        | `42`, `-100`                 |
| `float`               | Decimal number      | `3.14`, `-0.5`               |
| `bool`                | Literal             | `true`, `false`              |
| `str`                 | Unquoted or quoted  | `Alice`, `"Carol Smith"`     |
| `Optional[T]`         | Value or empty      | `hello` or _(blank)_ for nil |
| `list[T]`             | `[v1,v2,v3]`        | `[rust,go,python]`           |
| `dict[str, T]`        | `[(k1,v1),(k2,v2)]` | `[(age,30),(score,95)]`      |
| Nested dataclass      | `(field1,field2)`   | `(Engineering,500000)`       |
| `list[list[T]]`       | `[[v1,v2],[v3]]`    | `[[1,2],[3,4,5]]`            |
| `list[list[list[T]]]` | Triple-nested       | `[[[1,2],[3,4]]]`            |

### Dataclass Fields

The library uses `dataclasses` and `typing.get_type_hints()` to automatically discover field names and types:

```python
@dataclass
class Employee:
    id: int
    name: str
    dept: Department
    skills: list[str]
    active: bool
```

### Nested Structs

```python
@dataclass
class Dept:
    title: str

@dataclass
class Employee:
    name: str
    dept: Dept

# Schema reflects nesting:
# {name,dept}:(Alice,(Engineering))
```

### Optional Fields

```python
from typing import Optional

@dataclass
class Item:
    id: int
    label: Optional[str] = None

# With value:  {id,label}:(1,hello)
# With None:   {id,label}:(1,)
```

### Lists & Dicts

```python
@dataclass
class Tagged:
    name: str
    tags: list[str]
# {name,tags}:(Alice,[rust,go,python])

@dataclass
class Profile:
    name: str
    attrs: dict[str, int]
# {name,attrs}:(Alice,[(age,30),(score,95)])
```

### Type Annotations (Optional)

ASON schema supports **optional** type annotations. Both forms are fully equivalent — the loader handles them identically:

```text
// Without annotations (default output of dump / dump_slice)
{id,name,salary,active}:(1,Alice,5000.50,true)

// With annotations (output of dump_typed / dump_slice_typed)
{id:int,name:str,salary:float,active:bool}:(1,Alice,5000.50,true)
```

Annotations are **purely decorative metadata** — they do not affect parsing or loading behavior. The loader simply skips the `:type` portion when present.

**When to use annotations:**

- LLM prompts — helps models understand and generate correct data
- API documentation — self-describing schema without external docs
- Cross-language exchange — eliminates type ambiguity (is `42` an int or float?)
- Debugging — see data types at a glance

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

| Function                                     | Description                                               |
| -------------------------------------------- | --------------------------------------------------------- |
| `dump(obj) -> str`                           | Dump a dataclass → unannotated schema `{id,name}:`        |
| `dump_typed(obj) -> str`                     | Dump a dataclass → annotated schema `{id:int,name:str}:`  |
| `load(data, cls) -> obj`                     | Load a dataclass (accepts both annotated and unannotated) |
| `dump_slice(objs) -> str`                    | Dump a list → unannotated schema (schema-driven)          |
| `dump_slice_typed(objs, field_types) -> str` | Dump a list → annotated schema                            |
| `load_slice(data, cls) -> list`              | Load a list (accepts both annotated and unannotated)      |

## Performance

Benchmarked on Apple Silicon (M-series), Python 3.9, comparing against `json` module (C-based):

### Size Savings (identical to all ASON implementations)

| Scenario               | JSON   | ASON   | Saving  |
| ---------------------- | ------ | ------ | ------- |
| Flat struct × 100      | 13 KB  | 5.6 KB | **59%** |
| Flat struct × 1000     | 137 KB | 57 KB  | **59%** |
| 5-level deep × 10      | 49 KB  | 17 KB  | **66%** |
| 5-level deep × 50      | 249 KB | 87 KB  | **65%** |
| 100 countries (nested) | 215 KB | 73 KB  | **66%** |

### Speed vs json module

Python's `json` module is implemented in C (`_json` extension), making it inherently faster than any pure Python code. Despite this, ASON Python delivers:

- **59-66% smaller output** — significant for network transmission and LLM token savings
- **Schema-driven structure** — typed, self-describing format ideal for AI/LLM pipelines
- **Zero dependencies** — single-file, copy-and-use
- **Correct roundtrip** — full-precision floats, escaped strings, triple-nested arrays

For latency-critical paths, consider the [Rust](../rust/) or [Go](../go/) ASON implementations which outperform their respective JSON libraries.

### Why ASON is valuable even when slower than C-based json

1. **Network is the bottleneck** — ASON's 59-66% size reduction saves more time in network transmission than serialization overhead
2. **LLM token efficiency** — 30-70% fewer tokens = lower cost and faster responses
3. **Schema validation** — typed schemas prevent data errors
4. **Cross-language** — same format across Rust, Go, Python, and more

Run the benchmark yourself:

```bash
python examples/bench.py
```

## Implementation Highlights

- **Pure Python** — stdlib only (`dataclasses`, `typing`, `math`)
- **Byte-level parsing** — input is converted to `bytes` for fast indexed access
- **Lookup tables** — `bytearray(256)` for quoting decisions, `dict` for escape characters
- **Struct info caching** — type hints and field metadata cached per class
- **`__slots__`** — all internal classes use `__slots__` for minimal memory
- **Zerocopy where possible** — `bytes.decode()` directly from parsed ranges, no intermediate string construction when no escapes
- **Fast float formatting** — integer, 1-decimal, and 2-decimal fast paths; `repr()` for full precision

## Examples

```bash
# Basic usage (11 examples)
python examples/basic.py

# Comprehensive (17 examples — nested structs, 7-level nesting, large structures, edge cases)
python examples/complex.py

# Performance benchmark (ASON vs JSON, throughput, memory, 8 sections)
python examples/bench.py
```

## Tests

```bash
python test_ason.py
# 33 tests covering: roundtrip, typed/untyped schema, optional fields,
# nested structs, maps, arrays, triple-nested arrays, comments, edge cases
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
