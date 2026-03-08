# ASON — Array-Schema Object Notation

[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**ASON** is a compact, schema-driven data format designed for **LLM interactions** and **high-performance data transmission**. It separates schema from data — keys are declared once, rows carry only values.

[中文文档](README_CN.md)

---

## Why ASON?

Standard JSON repeats every field name for every record. When sending structured datasets to an LLM or over a network, that redundancy burns tokens and bandwidth:

```json
[
  {"id": 1, "name": "Alice", "active": true},
  {"id": 2, "name": "Bob",   "active": false},
  {"id": 3, "name": "Carol", "active": true}
]
```

ASON declares the schema **once** and streams data as compact tuples:

```
[{id:int, name:str, active:bool}]:(1,Alice,true),(2,Bob,false),(3,Carol,true)
```

**~65% fewer tokens. Same information. Machine-readable schema.**

---

## ASON vs JSON

| Aspect                | JSON              | ASON                   |
| --------------------- | ----------------- | ---------------------- |
| Token efficiency      | 100% (baseline)   | **30–70%** ✓           |
| Key repetition        | Every object      | Declared once ✓        |
| Type annotations      | None              | Optional ✓             |
| Human readable        | Yes               | Yes ✓                  |
| Nested structs        | ✓                 | ✓                      |
| Serialization speed   | 1×                | **~1.7–2.4× faster** ✓ |
| Deserialization speed | 1×                | **~1.9–2.9× faster** ✓ |
| Data size             | 100% (baseline)   | **40–60%** ✓           |
| Binary codec          | ✗                 | ✓                      |
| Struct reflection     | ✗                 | Compile-time ✓         |

### Token Savings — A Concrete Example

```
JSON (100 tokens):
{"users":[{"id":1,"name":"Alice","active":true},{"id":2,"name":"Bob","active":false}]}

ASON (~35 tokens, 65% saving):
[{id:int, name:str, active:bool}]:(1,Alice,true),(2,Bob,false)
```

The schema header also serves as an inline hint for LLMs — field names and optional types are immediately visible without scanning every row.

---

## ASON vs TOON

[TOON (Token-Oriented Object Notation)](https://toonformat.dev) is another format designed for reducing tokens in LLM prompts. Both ASON and TOON share the core idea of eliminating key repetition for array-of-object data, but they differ significantly in design goals and scope.

### Syntax at a glance

**TOON** — indentation-based, YAML-inspired:
```
users[2]{id,name,active}:
  1,Alice,true
  2,Bob,false
```

**ASON** — tuple-based, schema-explicit:
```
[{id:int, name:str, active:bool}]:(1,Alice,true),(2,Bob,false)
```

### Comparison Table

| Aspect                    | TOON                                  | ASON                                       |
| ------------------------- | ------------------------------------- | ------------------------------------------ |
| Schema declaration        | Auto-detected at encode time          | Explicit, reusable, language-level ✓       |
| Type annotations          | None (JSON data model only)           | Optional rich types (`int`, `str`, `bool`, `f64`, `opt_*`, `vec_*`, `map_*`) ✓ |
| Syntax style              | YAML-like indentation                 | Compact tuple rows                         |
| Array length markers      | `[N]` — helps detect truncation       | Schema header defines structure ✓          |
| Nested structures         | Falls back to verbose list format     | Native, efficient, recursive ✓             |
| Use case                  | LLM input only (read-only layer)      | LLM + serialization + data transmission ✓  |
| Serialization speed       | Not measured (JS only)                | SIMD-accelerated, ~1.7–2.9× vs JSON ✓     |
| Deserialization speed     | Not measured (JS only)                | Zero-copy parsing ✓                        |
| Binary codec              | ✗                                     | ✓                                          |
| Language implementations  | TypeScript / JavaScript only          | **C, C++, C#, Go, Java, JS, Python, Rust, Zig, Dart** ✓ |
| Struct reflection         | Dynamic (runtime only)                | Compile-time (`ASON_FIELDS` macro) ✓       |
| Deeply nested data        | Token cost increases significantly    | Efficient at any nesting level ✓           |
| Round-trip fidelity       | JSON data model (no type info)        | Full type fidelity ✓                       |

### When to Choose ASON

- You need **high-performance serialization** in a compiled language (C, C++, Rust, Go, …)
- Your data has **rich types** — optional fields, typed vectors, maps, nested structs
- You need **binary encoding** alongside the text format
- You work in **multiple languages** or need a language-neutral wire format
- You want the schema to act as a **self-documenting API contract** for LLM prompts

### When TOON May Be Enough

- Your pipeline is **LLM prompt input only** (you don't parse it back into structs)
- Your data is simple flat tables with no type constraints

---

## Format Overview

### Single Object

```
{id:int, name:str, active:bool}:(42,Alice,true)
```

### Array of Objects (Schema-Driven)

Schema declared once, each row is a tuple:

```
[{id:int, name:str, active:bool}]:(1,Alice,true),(2,Bob,false),(3,Carol,true)
```

### Nested Structs

```
{name:str, dept}:(Alice,(Engineering))
```

### Optional Fields

```
{id:int, label:opt_str}:(1,hello),(2,)
```
*(blank value = `None`/`null`)*

### Typed Vectors and Maps

```
{name:str, tags:vec_str, attrs:map_si}:(Alice,[rust,go,python],[(age,30),(score,95)])
```

---

## Implementations

| Language   | Repository          | Notes                              |
| ---------- | ------------------- | ---------------------------------- |
| C          | [ason-c](ason-c/)   | C11, SIMD (NEON/SSE2), zero-copy   |
| C++        | [ason-cpp](ason-cpp/) | C++17, header-only, SIMD         |
| C#         | [ason-cs](ason-cs/) | .NET, SIMD                         |
| Go         | [ason-go](ason-go/) |                                    |
| Java       | [ason-java](ason-java/) |                                |
| JavaScript | [ason-js](ason-js/) |                                    |
| Python     | [ason-py](ason-py/) |                                    |
| Rust       | [ason-rs](ason-rs/) |                                    |
| Zig        | [ason-zig](ason-zig/) |                                  |
| Dart       | [ason-dart](ason-dart/) |                                |
| Swift      | [ason-swift](ason-swift/) |   TODO                      |
| PHP       | [ason-php](ason-php/) |      TODO                        |
| Kotlin       | [ason-dart](ason-kt/) |    TODO                       |
| TypeScript  | [ason-dart](ason-ts/) |    TODO                        |

## Plugins

| IDE       | Repository    | Notes     |
| ---------- | ------ | ------ |
| VSCode       | [plugin_vscode](plugin_vscode/)   | Done |
| Jetbrain       | [plugin_jetbrain](plugin_jetbrain/)   | Todo |
| Zed       | [plugin_zed](plugin_zed/)   | Todo |

---

## License

MIT
