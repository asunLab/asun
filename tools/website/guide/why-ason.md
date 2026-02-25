# Why ASON over JSON?

ASON was built to solve real problems that JSON cannot solve without compromising on either readability or efficiency. This page provides a detailed comparison.

## 1. Token Efficiency

JSON repeats field names for every object in an array. With 1 000 rows and 5 fields, that's 5 000 redundant key strings flying across the wire (or into an LLM context window).

```json
// JSON — 100 tokens
{
  "users": [
    { "id": 1, "name": "Alice", "active": true },
    { "id": 2, "name": "Bob",   "active": false }
  ]
}
```

```ason
// ASON — ~35 tokens (65% reduction)
{id:int, name:str, active:bool}:
  (1, Alice, true),
  (2, Bob,   false)
```

### Measured savings across payload sizes

| Payload | JSON size | ASON Text | ASON-BIN | Text savings |
|---------|-----------|-----------|----------|-------------|
| 1 000 flat structs (8 fields) | 121 675 B | 56 716 B | 74 454 B | **53%** |
| 100 deep nested structs (5 levels) | 438 112 B | 174 611 B | 225 434 B | **60%** |

## 2. Parse Performance

### Zero Key-Hashing

JSON parsers must hash every field name and look it up in the target struct on every row. With 1 000 rows × 5 fields = 5 000 redundant hash operations.

ASON parsers build a **positional index** once from the schema (`field 0 → id`, `field 1 → name`, …), then assign values by array index — `O(1)`, no hashing.

### Schema-Driven Parsing

JSON parsers dynamically infer types by peeking at the next character (`"`, `t`, `f`, `[`, `{`, digit). This causes CPU branch mispredictions.

ASON parsers know field types from the schema: they call `parse_int()` or `parse_str()` directly — no dynamic inference, no mispredictions.

### Benchmark (Apple M-series, `--release`)

**Flat struct (8 fields, × 1 000)**

| Test | JSON | ASON Text | ASON-BIN | BIN vs JSON |
|------|------|-----------|----------|-------------|
| Serialize | 2.19 ms | 1.07 ms | **0.28 ms** | **7.7×** |
| Deserialize | 6.05 ms | 5.10 ms | **2.96 ms** | **2.0×** |

**Deep struct (5-level nested, × 100)**

| Test | JSON | ASON Text | ASON-BIN | BIN vs JSON |
|------|------|-----------|----------|-------------|
| Serialize | 4.19 ms | 2.67 ms | **0.50 ms** | **8.4×** |
| Deserialize | 9.37 ms | 8.55 ms | **3.72 ms** | **2.5×** |

## 3. LLM Output Quality

LLMs generating JSON frequently hallucinate mismatched braces, trailing commas, and incorrectly quoted strings. ASON's syntax is more forgiving:

- Most string values are **unquoted** — the parser auto-trims whitespace
- Structure is **visually linear** — one tuple per line, aligned columns
- Schema is **declarative** — the model only needs to learn the header once per session

In practice this means fewer retries, fewer validation errors, and lower token cost per structured response.

## 4. Human Readability

Despite its compression, ASON remains highly readable:

```ason
{id:int, name:str, role:str, score:float}:
  (1, Alice, admin,   9.5),
  (2, Bob,   viewer,  7.2),
  (3, Carol, editor,  8.8)
```

Columns line up naturally when padded. The schema acts as a header row — familiar to anyone who has read a CSV or a markdown table.

## 5. Comparison Table

| Feature | JSON | ASON Text | ASON-BIN |
|---------|------|-----------|----------|
| Human-readable | ✅ | ✅ | ❌ |
| Token-efficient | ❌ | ✅ | N/A |
| Schema separation | ❌ | ✅ | ✅ |
| Fast serialize | ✗ baseline | 2× | **7–10×** |
| Fast deserialize | ✗ baseline | 1.2–2.5× | **2–2.5×** |
| Zero-copy strings | ❌ | ❌ | ✅ |
| SIMD acceleration | ❌ | ❌ | ✅ |
| LLM-friendly | ⚠️ | ✅ | ❌ |
| Universal tooling | ✅ | growing | growing |

## When to Keep JSON

ASON is not a universal JSON replacement. Stick with JSON when:

- You need ad-hoc object structures (not arrays of uniform structs)
- You are integrating with external APIs that require JSON
- The overhead is irrelevant (tiny payloads, low frequency)

ASON shines with **arrays of homogeneous structs** — which is the dominant pattern in data pipelines, LLM prompts, and internal service communication.
