---
layout: home

hero:
  name: "ASON"
  text: "Array-Schema Object Notation"
  tagline: "65% fewer tokens than JSON. 7–10× faster binary. Human-readable. LLM-native."
  image:
    src: /logo.svg
    alt: ASON
  actions:
    - theme: brand
      text: Get Started
      link: /guide/getting-started
    - theme: alt
      text: Why ASON?
      link: /guide/why-ason
    - theme: alt
      text: View on GitHub
      link: https://github.com/your-org/ason

features:
  - icon: 🚀
    title: 65% Token Reduction
    details: Schema is declared once. Data rows carry only values — no repeated keys. Ideal for LLM context windows and API cost reduction.
    link: /guide/why-ason
    linkText: See comparison

  - icon: ⚡
    title: 7–10× Faster Binary
    details: ASON-BIN uses zero-copy deserialization and SIMD acceleration. Serialization is up to 7.7× faster than JSON with 39% smaller payloads.
    link: /reference/performance
    linkText: Benchmark results

  - icon: 🧠
    title: LLM-Native Design
    details: Row-oriented syntax with schema separation is easy for language models to generate and parse. Reduces hallucination in structured output.
    link: /guide/llm-integration
    linkText: LLM integration

  - icon: 📖
    title: Human-Readable
    details: Clean Markdown-like syntax. No quotes around most values. Vertically aligned rows for easy scanning. High signal-to-noise ratio.
    link: /reference/syntax
    linkText: Syntax reference

  - icon: 🌐
    title: Multi-Language
    details: Official libraries for Rust, Go, Python, C, C++, Java, and Zig. Consistent API design across all implementations.
    link: /languages/rust
    linkText: Language guides

  - icon: 🔧
    title: Drop-in Tooling
    details: VS Code extension with syntax highlighting and live preview. LSP server for editor integration. Spec-compliant cross-language test suite.
    link: /guide/getting-started
    linkText: Setup tools
---

<div class="vp-doc" style="max-width:900px;margin:0 auto;padding:2rem 1.5rem">

## ASON vs JSON — at a glance

```json
// JSON — ~100 tokens
{
  "users": [
    { "id": 1, "name": "Alice", "active": true },
    { "id": 2, "name": "Bob",   "active": false },
    { "id": 3, "name": "Carol", "active": true  }
  ]
}
```

```ason
// ASON — ~35 tokens  (65% savings)
[{id:int, name:str, active:bool}]:
  (1, Alice, true),
  (2, Bob,   false),
  (3, Carol, true)
```

Schema is written **once**. Every row after `:` is pure data — comma-separated values in a tuple `(...)`. No repeated keys, no extra quotes, no noise.

## Quick Install

::: code-group

```toml [Rust]
# Cargo.toml
[dependencies]
ason = "0.1"
serde = { version = "1", features = ["derive"] }
```

```bash [Go]
go get github.com/your-org/ason/go
```

```bash [Python]
# copy ason.py into your project (stdlib only, zero dependencies)
```

:::

</div>
