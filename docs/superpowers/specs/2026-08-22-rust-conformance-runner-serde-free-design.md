# Rust Conformance Runner — Serde-Free Redesign

**Date:** 2026-08-22
**Status:** Approved (pending spec review)

## Problem

`conformance/runners/rust/` no longer compiles. asun-rs dropped its serde
dependency and now uses a self-authored `#[derive(AsunEncode, AsunDecode)]`.
The runner calls `asun::decode::<serde_json::Value>(...)` and
`asun::encode(&serde_json::Value)`, which require
`serde_json::Value: AsunDecode / AsunEncode` — impls that only ever existed
via the now-removed serde bridge. Error: `the trait bound Value: AsunDecode<'_>
is not satisfied` at `main.rs:217`.

### Root cause (not "missing types")

asun-rs's public `AsunDecode` is **static/type-driven**: blanket impls exist
only for concrete types (ints, floats, `String`, `&str`, `Option<T>`,
`Vec<T>`, `()`), and everything else comes from `#[derive]` on user structs.
The public `Decoder` API exposes only *typed* primitives
(`decode_i64`, `decode_string`, `decode_vec::<T>`, `begin_struct_decode`) —
there is **no peek / "what type comes next" capability**. Untyped/dynamic
decode (parse arbitrary ASUN into a generic `Value` without knowing the type
ahead) is therefore **not expressible** through the current public API. The
Go runner decodes into `any` because asun-go supports dynamic decode; Rust
has no equivalent. So this cannot be fixed by adding a few type impls.

## Scope decisions (confirmed with user)

1. **serde removal applies to the ASUN format only.** The runner MAY use
   `serde` / `serde_json` freely for reading the JSON *case files*.
2. **Bridge approach A:** the runner carries its own independent ASUN
   text parser/encoder producing/consuming `serde_json::Value`. It does NOT
   go through asun-rs's typed `decode`/`encode`. (Mirrors the Go runner's
   "independent implementation checked against the spec" model, and avoids
   touching the library's public API or the just-verified perf changes.)
3. **Full coverage:** support all 277 cases, including the 193 schemaDriven
   ones. The old runner skipped all 193 and ran only 84. Since the runner now
   parses ASUN itself and the schema carries field names + `@`types, it can
   materialise schemaDriven inputs into JSON objects/arrays directly.

## Design

### File layout

```
conformance/runners/rust/
  Cargo.toml          # asun dep DROPPED; keep serde + serde_json
  src/
    main.rs           # harness: read cases, run, report, exit code (unchanged shape)
    asun.rs           # NEW: independent ASUN text -> serde_json::Value parser
    encode.rs         # NEW: serde_json::Value -> ASUN text encoder (round-trip)
```

`Cargo.toml` after change:

```toml
[dependencies]
serde = { version = "1", features = ["derive"] }
serde_json = "1"
# asun dependency removed entirely — the runner is an independent oracle.
```

The `asun = { path = "../../../asun-rs" }` line is removed. The runner no
longer links the library at all; it validates the *spec* via its own
implementation, exactly like the Go runner validates against `interface{}`.

### `asun.rs` — decoder (ASUN text → `serde_json::Value`)

A small recursive-descent parser over `&str`, faithful to `GRAMMAR.abnf`.

**Lexing/whitespace:** `ows` = spaces, tabs, newlines, and `/* ... */`
comments, skipped between tokens. Comments inside data tuples `(...)` or
array literals `[...]` are an error (SPEC §5 / grammar S4).

**Top form dispatch** (after leading `ows`):
- starts with `[{`  → array-of-objects-form
- starts with `{`   → object-form
- starts with `[`   → plain-array
- otherwise         → bare-value

**Schema parse** (`{ field-list }`): each `field` is
`field-name [ @ binding-target ]`.
- `field-name` = bare (`[A-Za-z0-9_]+`) or quoted-string.
- `binding-target` = `int|float|str|bool`, or nested `{...}` schema, or
  `[ inner ]` array-hint (`inner` = base-type | schema | nested array-hint).
- Represented as an enum `Hint { None, Int, Float, Str, Bool, Obj(Schema),
  Arr(Box<Hint>) }` plus the field name.

**Tuple parse** (`( elems )`): comma-separated `value-or-empty`. Semantic
rules S1/S5: element count MUST equal schema field count; a *trailing* comma
is absorbed (does not add null); two consecutive commas DO add a null. A
tuple element may itself be a nested tuple `(...)` (nested object data) or an
array literal `[...]`.

**Applying a schema to a tuple** → JSON object: zip field names with tuple
elements; coerce each element by its `Hint`:
- `Int`   → parse i64 (JSON number); non-integer literal → type error.
- `Float` → parse f64.
- `Str`   → string (quoted content, or trimmed plain text).
- `Bool`  → `true`/`false`; else error.
- `Obj(s)`→ element must be a nested tuple; recurse with sub-schema.
- `Arr(h)`→ element must be an array literal; each item coerced by `h`.
- `None`  → untyped: apply S3 type-resolution priority.

**Untyped value resolution (S3)** for elements/bare values without a hint:
1. empty → `null`
2. `true`/`false` → bool
3. `^-?[0-9]+$` → i64 integer
4. matches float regex AND contains `.` or `e`/`E` → f64 float
5. else → string. Tokens like `.5`, `5.`, `+5`, `1e`, `1e+` fall through to
   string (verified against cases: `'.5' => ".5"`, `'+5' => "+5"`).

**Strings:** quoted strings honour escapes
(`\" \\ \n \t \r \b \f \, \( \) \[ \] \{ \} \: \@` and `\uXXXX` incl.
surrogate pairs); NOT trimmed. Plain (unquoted) strings are trimmed of
leading/trailing ASCII whitespace (S2), internal whitespace preserved, and
honour the same escapes.

**Plain array** `[ ... ]`: comma-separated `array-element` (value / nested
array / tuple); trailing comma allowed. Elements resolved untyped (S3).

**Errors** are a flat `Err(String)` (kind = "error" cases only check that we
reject; the message is for the failure report). No panics on malformed input.

### `encode.rs` — encoder (`serde_json::Value` → ASUN text)

Used only by the encode round-trip suite (`encode-cases.json`): encode a
`Value`, decode it back with `asun.rs`, compare. Chosen encoding, spec-legal
and round-trip-stable:
- object → `{k1,k2,...}:(v1,v2,...)` (bare field name if it matches
  `[A-Za-z0-9_]+`, else quoted).
- array of objects with identical string keys → `[{...}]:(...),(...)`
  (optional optimisation; the plain `[ ... ]` form below always works, so
  MVP may emit arrays as plain arrays and rely on round-trip equality).
- array → `[e1,e2,...]`.
- string → quoted with minimal escaping (always safe).
- int/float/bool/null → bare literal (`null` only valid inside tuple/array;
  a top-level `null` value is not representable — encode-cases are expected
  not to contain a bare top-level null; if they do, emit `""`-in-tuple is not
  applicable, so treat as documented limitation).
- **Round-trip is the only correctness bar** — any spec-legal encoding that
  decodes back equal passes. We do not need to match asun-rs byte output.

### `main.rs` — harness (shape preserved)

- `Manifest` / `Case` structs keep `#[derive(Deserialize)]` (serde_json still
  reads the case files — allowed).
- Replace `asun::decode(&case.input)` with `asun::decode_value(&case.input)`
  (our `asun.rs`), returning `Result<serde_json::Value, String>`.
- **Remove the `schema_driven` skip** — run all cases. Drop the
  `skipped_typed` counter (or keep at 0).
- `values_equal` (tolerant numeric/array/object compare) is unchanged.
- Encode suite: `asun::encode(&value)` → our `encode.rs`;
  decode-after-encode → our `asun.rs`.
- Exit 1 on any failure; report first N failures. Unchanged.

## Testing

1. `cargo build` in `conformance/runners/rust/` succeeds with no asun dep.
2. `cargo run` executes all 277 cases + encode cases.
3. Target: all `ok` cases match `expected`; all `error` cases are rejected;
   all encode round-trips hold. Any residual mismatch is investigated case by
   case (either a parser bug or a genuine spec-corner to document) — we will
   report the exact pass/fail counts, not hand-wave "it passes".
4. Cross-check a sample of schemaDriven expectations (e.g.
   `{name@str,addr@{city@str,zip@int}}:(Alice,(NYC,10001))` →
   `{"name":"Alice","addr":{"city":"NYC","zip":10001}}`).

## Non-goals

- Not modifying asun-rs (public API, derive, or the verified perf changes).
- Not adding a dynamic `Value` type to the library.
- Not matching asun-rs's exact encoder byte output (round-trip equality only).
```