# ASON — Array-Schema Object Notation

[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**ASON** is a compact, schema-first data format for **LLM prompts**, **structured APIs**, and **large datasets**. It separates schema from data, so keys are declared once and rows carry only values.

[中文文档](README_CN.md)

---

## Why ASON?

Standard JSON repeats every field name in every record. When you send structured data to an LLM, over an API, or across services, that repetition wastes tokens, bytes, and attention:

```json
[
  { "id": 1, "name": "Alice", "active": true },
  { "id": 2, "name": "Bob", "active": false },
  { "id": 3, "name": "Carol", "active": true }
]
```

ASON declares the schema **once** and streams data as compact tuples:

```
[{id@int, name@str, active@bool}]:(1,Alice,true),(2,Bob,false),(3,Carol,true)
```

**Fewer tokens. Smaller payloads. Clearer structure.**

---

## ASON vs JSON

| Aspect               | JSON                    | ASON                       |
| -------------------- | ----------------------- | -------------------------- |
| Token efficiency     | 100% (baseline)         | **30–70%** ✓               |
| Key repetition       | Every object            | Declared once ✓            |
| Field binding        | None                    | Built-in `@...` ✓          |
| Human readable       | Yes                     | Yes ✓                      |
| Nested structs       | ✓                       | ✓                          |
| Serialization        | Repeats keys            | Schema once, values only ✓ |
| Deserialization      | Generic object scanning | Schema-guided decoding ✓   |
| Data size            | 100% (baseline)         | **40–60%** ✓               |
| Binary codec         | ✗                       | ✓                          |
| Scalar hint support  | Limited                 | Optional via `@type` ✓     |

### Token Savings — A Concrete Example

```
JSON (100 tokens):
{"users":[{"id":1,"name":"Alice","active":true},{"id":2,"name":"Bob","active":false}]}

ASON (~35 tokens, 65% saving):
[{id@int, name@str, active@bool}]:(1,Alice,true),(2,Bob,false)
```

The schema header also acts as an inline hint for LLMs and humans: field names, structural bindings, and optional scalar hints are visible up front instead of being repeated row by row.

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
[{id@int, name@str, active@bool}]:(1,Alice,true),(2,Bob,false)
```

### Comparison Table

| Aspect                   | TOON                              | ASON                                                                            |
| ------------------------ | --------------------------------- | ------------------------------------------------------------------------------- |
| Schema declaration       | Auto-detected at encode time      | Explicit and reusable ✓                                                         |
| Field/type binding       | None (JSON data model only)       | Explicit `@...` binding for scalar hints and complex structures ✓                |
| Syntax style             | YAML-like indentation             | Compact tuple rows                                                              |
| Array length markers     | `[N]` — helps detect truncation   | Schema header defines structure ✓                                               |
| Nested structures        | Falls back to verbose list format | Native and recursive ✓                                                          |
| Use case                 | LLM input only                    | LLM + serialization + storage + transport ✓                                     |
| Binary codec             | ✗                                 | ✓                                                                               |
| Language implementations | TypeScript / JavaScript only      | **C, C++, C#, Go, Java, JS, Python, Rust, Zig, Dart** ✓                         |
| Round-trip fidelity      | JSON data model only              | Full type fidelity ✓                                                            |

### When to Choose ASON

- You want **fewer tokens and smaller payloads** without losing structure
- You need **rich structured data** — optional fields, arrays, nested structs, keyed entry lists
- You want one format to work across **LLMs, APIs, storage, and service-to-service transport**
- Your data has **rich structure** — optional fields, arrays, nested structs, keyed entry lists
- You need **binary encoding** alongside text
- You work in **multiple languages** or need a language-neutral wire format
- You want the schema to act as a **self-documenting API contract** for LLM prompts, with `@` binding fields to structural descriptions and optional scalar hints

### When TOON May Be Enough

- Your pipeline is **LLM prompt input only** (you don't parse it back into structs)
- Your data is simple flat tables with no type constraints

---

## Format Overview

### Single Object

```
{id@int, name@str, active@bool}:(42,Alice,true)
```

### Array of Objects (Schema-Driven)

Schema declared once, each row is a tuple:

```
[{id@int, name@str, active@bool}]:(1,Alice,true),(2,Bob,false),(3,Carol,true)
```

### Nested Structs

```
{name@str, dept@{title@str}}:(Alice,(Engineering))
```

### Optional Fields

```
{id@int, label@str}:(1,hello),(2,)
```

_(blank value = `None`/`null`)_

### Arrays and Keyed Entries

```
{name@str, scores@[int], attrs@[{key@str, value@int}]}:(Alice,[90,85,92],[(age,30),(score,95)])
```

---

## Implementations

| Language                | Repository                | Status |
| ----------------------- | ------------------------- | ------ |
| C                       | [ason-c](ason-c/)         | ✓      |
| C++                     | [ason-cpp](ason-cpp/)     | ✓      |
| C#                      | [ason-cs](ason-cs/)       | ✓      |
| Go                      | [ason-go](ason-go/)       | ✓      |
| Java / Kotlin           | [ason-java](ason-java/)   | ✓      |
| JavaScript / TypeScript | [ason-js](ason-js/)       | ✓      |
| Python                  | [ason-py](ason-py/)       | ✓      |
| Rust                    | [ason-rs](ason-rs/)       | ✓      |
| Zig                     | [ason-zig](ason-zig/)     | ✓      |
| Dart                    | [ason-dart](ason-dart/)   | ✓      |
| PHP                     | [ason-php](ason-php/)     | ✓      |
| Swift                   | [ason-swift](ason-swift/) | ✓      |

## Plugins

| IDE      | Repository                          | Notes |
| -------- | ----------------------------------- | ----- |
| VSCode   | [plugin_vscode](plugin_vscode/)     | ✓     |
| Jetbrain | [plugin_jetbrain](plugin_jetbrain/) | Todo  |
| Zed      | [plugin_zed](plugin_zed/)           | Todo  |

---

## License

MIT
