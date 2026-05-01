# ASUN Conformance Suite

This directory is the **single source of truth** for ASUN textual syntax across
all language implementations (`asun-rs`, `asun-c`, `asun-cpp`, `asun-go`,
`asun-java`, `asun-py`, `asun-js`, `asun-cs`, `asun-swift`, `asun-zig`,
`asun-dart`, `asun-php`).

## Files

| Path                  | Purpose                                                          |
| --------------------- | ---------------------------------------------------------------- |
| `GRAMMAR.abnf`        | Formal grammar (RFC 5234 ABNF). Authoritative.                   |
| `cases.json`          | Generated input/output test vectors (264 cases as of v1).        |
| `encode-cases.json`   | Encode round-trip vectors: `decode(encode(value)) == value`.      |
| `generate.py`         | Generator for `cases.json`. Re-run after editing.                |
| `runners/rust/`       | Reference runner against `asun-rs`.                              |
| `runners/<lang>/`     | (To be added) Per-language runner that loads `cases.json`.       |

## Test-case schema

Each entry in `cases.json[].cases` has:

```jsonc
{
  "id":           "objects.single.basic",       // unique dotted id
  "category":     "objects.single",
  "desc":         "basic 3-field object",
  "input":        "{id,name,active}:(1,Alice,true)",
  "kind":         "ok",                          // "ok" | "error"
  "schemaDriven": true,                          // true if input begins with `{` or `[{`
  "expected":     { "id": 1, "name": "Alice", "active": true }   // only when kind=ok
  // "errorHint": "parse.field_count"            // only when kind=error
}
```

## Runner contract

A conforming runner MUST:

1. Load every case from `cases.json`.
2. For each case, attempt to decode `input` with the language's ASUN decoder.
3. **kind == "ok"**: result must deep-equal `expected` (with tolerant numeric
   comparison: integers compare exactly; floats compare with relative epsilon).
4. **kind == "error"**: decoder must return / throw / raise an error. The
   specific error category (`errorHint`) is advisory only — runners SHOULD NOT
   match on it.
5. Exit non-zero when any case is mishandled.

## Encode runner contract

`encode-cases.json` validates encoder safety rather than a canonical byte
layout. A conforming encode runner MUST:

1. Load every case from `encode-cases.json`.
2. Encode `value` with the language's ASUN encoder.
3. Decode the encoded ASUN document back into a dynamic value.
4. Deep-compare the decoded value with the original `value` using tolerant
   numeric comparison.
5. Exit non-zero when any case fails.

The Zig encode runner is available at:

```bash
cd conformance/runners/zig
zig build run
```

## Schema-driven vs untyped cases

ASUN is **schema-driven**: the parser uses the schema header (`{...}:` or
`[{...}]:`) to position-decode data into a typed target. Some implementations
can also offer an **untyped / dynamic** decode path that returns a
language-native dynamic value (e.g. `serde_json::Value`, Python `dict`, JS
object).

Cases are flagged with `schemaDriven`:

- `schemaDriven: false` — top-level form is a bare scalar or plain `[...]`
  array. These cases MUST work with both untyped and typed decoders.
- `schemaDriven: true` — top-level form starts with `{` or `[{`. These cases
  MAY require a typed Rust/Go/Java/etc. struct target. Runners that lack a
  typed harness for a given case can mark it "skipped (needs typed harness)";
  this is not a conformance failure of the format.

## Reference runner: asun-rs

```bash
cd conformance/runners/rust
cargo run --release
```

Last run against `asun-rs` v1.0.1 (264 cases):

| Bucket                        | Count |
| ----------------------------- | ----: |
| Total cases                   |   264 |
| Skipped (need typed harness)  |   193 |
| Untyped ok-cases passed       |    67 |
| Untyped ok-cases failed       |     2 |
| Untyped error-cases passed    |     2 |
| Untyped error-cases failed    |     0 |
| **Untyped pass rate**         | **97.2 % (69/71)** |

### Findings vs current asun-rs

Two SPEC deviations the suite caught:

1. **`123abc` should be a plain string** (SPEC §8.1, type-priority cascade
   item 5). `asun-rs` greedily parses `123` as int and then errors with
   "trailing characters".
   - Case id: `primitives.bare.string.startsDigit`
2. **`\r` is a valid escape inside quoted strings** (SPEC §4 escape table).
   `asun-rs` rejects it with "invalid escape: \r".
   - Case id: `strings.escape.bs.cr`

Both are real format-level conformance gaps, not test-suite issues.

The 193 schema-driven cases are not yet exercised because `asun-rs`'s
untyped `deserialize_any` path does not recognise the `{...}:` schema header
(`peek_value_type` returns `String` for `{`). A typed harness covering a
representative subset is the next deliverable.

## Adding a runner for a new language

1. Read `cases.json` (UTF-8).
2. For each case where `schemaDriven == false`, decode into the language's
   dynamic value type and compare to `expected`.
3. For schema-driven cases, register typed structs corresponding to the
   `expected` JSON shape and decode through them. (A code-generation step
   from `expected` is recommended.)
4. Print a final pass/fail summary. Exit non-zero on any failure.

## Updating the suite

1. Edit `generate.py` (preferred) or hand-craft cases there. Keep IDs unique.
2. Run `python3 generate.py` to regenerate `cases.json`.
3. Re-run the Rust runner to catch any regressions.
4. Update SPEC and this README if grammar changes.

## Versioning

`cases.json` carries a `version` integer. Bump it on any backward-incompatible
change. Implementations SHOULD declare which version they are tested against.
