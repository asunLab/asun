# What is ASON?

**ASON** (Array-Schema Object Notation) is a high-performance serialization format designed for two primary use cases:

1. **LLM interaction** — reducing the token cost of structured data in prompts and responses
2. **High-throughput systems** — replacing JSON with a format that is faster to encode/decode and produces smaller payloads

## Core Idea

ASON separates **schema** from **data**. Instead of repeating field names in every row like JSON does, the schema is declared once and data rows are plain ordered tuples:

```ason
// Schema declared once — data repeated without keys
[{id:int, name:str, active:bool}]:
  (1, Alice, true),
  (2, Bob,   false),
  (3, Carol, true)
```

This single design decision eliminates 50–70% of token overhead in list-heavy workloads.

## Two Formats

| Format | Use case | Key advantage |
|--------|----------|---------------|
| **ASON Text** | LLM prompts, APIs, config | Human-readable, ~53–61% smaller than JSON |
| **ASON-BIN** | Services, caches, storage | 7–10× faster serialize, zero-copy deserialize |

Both formats share the same schema model and are supported by all official libraries.

## Design Principles

- **Schema & data separation** — structure declared once, never repeated
- **Row-oriented** — vertical alignment makes data scannable by humans and models alike
- **Structural harmony** — `{}` defines the skeleton (schema), `()` holds the values (data)
- **Zero key-hashing** — parser uses positional indexing; no string comparisons during deserialization
- **Schema-driven parsing** — field types are known at parse time, eliminating dynamic type inference

## Status

ASON has production-ready libraries in **Rust, Go, Python, C, C++, Java, and Zig**, a formal [specification](/spec), a VS Code extension, and an LSP server. All implementations share a cross-language compatibility test suite.

## Next Steps

- [Why ASON over JSON?](/guide/why-ason) — detailed comparison with benchmarks
- [Getting Started](/guide/getting-started) — five-minute quickstart in your language
- [Syntax Reference](/reference/syntax) — complete format specification
