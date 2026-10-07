# ASUN Format Specification v1.6

> **ASUN** = Array-Schema Unified Notation  
> _"The efficiency of arrays, the structure of objects."_

ASUN is a serialization format designed for large-scale data transmission and LLM (Large Language Model) interactions. By **separating schema from data**, it eliminates the repetitive key redundancy found in JSON and introduces a Markdown-inspired, row-oriented syntax that minimizes token consumption while maintaining excellent human readability.

---

## 1. Design Philosophy

- **Schema/Data Separation**: Structure is declared once; the data section carries only values.
- **Row-Oriented**: Supports multi-line format to enhance vertical visual flow, making it easy to read and stream.
- **Structural Harmony**: Uses `{}` to define the skeleton (Schema) and `()` to carry the body (Data) — symbols with clearly distinct semantics.
- **Token Optimization**: Compared to JSON, reduces token consumption by 30–70% in list scenarios.
- **Extreme Parse Performance**: Zero key-hashing and schema-driven parsing deliver deserialization speeds far exceeding JSON.

---

## 1.1 Why ASUN?

### JSON vs ASUN

```json
// JSON: 100 tokens
{
  "users": [
    { "id": 1, "name": "Alice", "active": true },
    { "id": 2, "name": "Bob", "active": false }
  ]
}
```

```asun
/* ASUN: ~35 tokens (65% token saving) */
[{id@int, name@str, active@bool}]:
  (1, Alice, true),
  (2, Bob, false)
```

| Aspect               | JSON            | ASUN              |
| -------------------- | --------------- | ----------------- |
| **Token Efficiency** | 100%            | 30–70% ✓          |
| **Key Repetition**   | Every row       | Declared once ✓   |
| **Human Readable**   | High            | High ✓            |
| **Learning Curve**   | Zero            | Low               |
| **Nesting**          | Arbitrary depth | Clear & bounded ✓ |
| **Streaming**        | Line-by-line    | Native support ✓  |

---

## 1.2 Extreme Parse Performance

Beyond token savings in LLM scenarios, ASUN also dramatically outperforms JSON in traditional serialization/deserialization (serde) use cases — behaving much more like CSV or Protobuf:

1. **Zero Key-Hashing**:
   - **JSON**: When parsing an array of objects, every key in every row must be read, hashed, and matched against the target struct. 1,000 rows means 1,000 repeated hash lookups.
   - **ASUN**: The parser first parses the schema to build a positional index (e.g., `0 → id`, `1 → name`). When parsing data rows, values are assigned via array index in O(1) — zero hash computation or string matching throughout.

2. **Schema-Driven Parsing**:
   - **JSON**: The parser must dynamically infer the type of each value by peeking at the next character (`"`, `t`, `f`, `[`, `{`, digit), causing frequent CPU branch mispredictions.
   - **ASUN**: The schema provides structural information (e.g., `{id, name, active}`), so the parser reads each field value in a fixed, known order without dynamic type inference. In a serde framework the target struct's type is already known at compile time, and the parser calls `parse_int()` etc. directly. In schema text, `@` is the binding marker between a field and its following schema/type description: scalar hints like `{id@int}` are optional, while structural bindings such as `@{}` and `@[]` are required for complex fields.

3. **Low Memory Footprint**:
   - JSON builds a dynamic DOM tree that allocates memory for every key string in every object. ASUN data rows are essentially a flat tuple array — all rows share a single schema reference, keeping memory overhead minimal.

---

## 1.3 Core Architecture Philosophy

ASUN's design goes beyond just saving tokens; its deep architectural philosophy dictates its extreme performance and precise semantics:

1. **Physical Isolation via `Header : Body`**
   - ASUN uses `:` to strictly divide the text into Schema (Blueprint) and Data (Body).
   - This physical isolation allows the parser to split the payload in **O(1)** time. The front-end parser can parse the Schema independently, pre-allocate structural memory, and then engage in high-speed or multi-threaded streaming of massive Data rows, completely discarding JSON's inefficient paradigm of continuously inferring object boundaries while parsing data.

2. **Tuple Semantics `()` vs Object Semantics `{}`**
   - In ASUN, Schema uses `{}` to define the unordered key-value blueprint, but crucially, **data bodies are strictly wrapped in `()`**.
   - `()` represents a **Tuple** in programming languages — a strictly ordered, keyless, position-bound data structure. It sends a strong signal to humans and AI: data must perfectly align with the schema positions and cannot be shuffled like in JSON. This architectural "harmony of shapes" visually isolates structural definitions from pure data with extreme sharpness, drastically reducing structure-related parsing errors.

3. **Native Immunity to Key Collisions**
   - JSON syntax allows ambiguous key duplications (e.g., `{"age": 30, "age": 40}`), often leading to the latter covering the former.
   - By confining keys exclusively to the Schema header, ASUN data regions consist solely of compact values (`(30), (40)`), completely insulating the format from key collisions at the syntactic source. Maps/Dictionaries are a first-class type declared in the schema (`attrs@[str:int]`) and written as `[age:30,score:95]`; duplicate map keys are an error too, so the format stays collision-free.

---

## 1.4 Changes in v1.6

| Area              | v1.5                                          | v1.6                                                              |
| ----------------- | --------------------------------------------- | ----------------------------------------------------------------- |
| Null              | Blank slot or `null`                          | `_`; decoders still accept `null`, encoders always emit `_`; `"_"` is a string |
| Empty slots       | `(a,)` = `a, null`; `(,)` = two nulls         | Error: every position holds a value; `()` is zero elements        |
| Plain strings     | Backslash escapes (`a\,b`)                    | No escapes; a string with `,()[]{}"\` or `/*` must be quoted       |
| Quoted strings    | JSON escapes + `\,` `\(` `\)` `\[` `\]` `\{` `\}` `\:` `\@` | JSON escapes only                                                 |
| Numbers           | `-?[0-9]+…` (leading zeros allowed)           | JSON number grammar: `007` is the string `"007"`                  |
| Maps              | Arrays of `{key,value}` tuples                | First-class: schema `attrs@[str:int]`, data `[age:30,score:95]`   |

---

## 2. Core Syntax Preview

```asun
[{id@int, name@str, tags@[str]}]:
  (1, Alice, [rust, go]),
  (2, Bob, [python, c++])
```

- `{...}` **(Schema)**: Defines field names, nested objects `{}`, or array types `[...]`
- `:` **(Separator)**: Marks the boundary between schema and data
- `(...)` **(Data Tuple)**: Carries values in order, strictly aligned with the schema definition

## 3. Data Type Rules

| Type            | Example         | Description                                                |
| --------------- | --------------- | ---------------------------------------------------------- |
| Integer         | `42`, `-100`    | JSON integer: `-?(0\|[1-9][0-9]*)`                          |
| Float           | `3.14`, `1e10`  | JSON number with a fraction and/or exponent                |
| Boolean         | `true`, `false` | Must be lowercase literals                                 |
| Null            | `_`             | Decoders also accept the keyword `null`                    |
| Empty string    | `""`            | Explicit empty string                                      |
| Unquoted string | `Hello World`   | Leading/trailing spaces auto-trimmed; no escapes           |
| Quoted string   | `" Space "`     | Spaces preserved as-is; JSON escaping rules                |

Keywords (`true`, `false`, `_`, `null`) and type names (`int`, `float`, `str`, `bool`) are **case-sensitive**: `TRUE`, `Null` and `@INT` are not keywords or types.

### 3.1 String Rules

ASUN supports two string forms:

| Form     | Example       | Behavior                                      |
| -------- | ------------- | --------------------------------------------- |
| Unquoted | `hello world` | Leading/trailing whitespace auto-trimmed      |
| Quoted   | `" hello "`   | Content preserved verbatim (including spaces) |

**When to use quotes:**

- Preserve leading/trailing spaces: `" hello "`
- Contains `,` `(` `)` `[` `]` `{` `}` `"` `\` or `/*`: `"a,b"`, `"f(x)"`
- Force string type: `"true"`, `"_"`, `"null"`, `"123"`

`001234` is not a JSON number, so it is already the string `"001234"` without quotes.
- Empty string: `""`

### 3.2 Field Binding and Optional Scalar Hints

**ASUN v1.4 introduces the `@` binding syntax.** `@` is not merely a type-annotation symbol; it is the **field binding marker** between a field name and its following schema/type description.

> **Core principle: `@` carries both structural binding and optional scalar hints.** For terminal scalar fields, `@type` is an optional hint; for complex fields, `@{}` / `@[]` are mandatory structural bindings. Both of the following are layout-equivalent:
>
> ```asun
> // Without annotations
> {id,name,active}:(1,Alice,true)
>
> // With scalar hints — identical parse result
> {id@int,name@str,active@bool}:(1,Alice,true)
> ```

**Supported types:**

| Type    | Syntax  | Example        | Description           |
| ------- | ------- | -------------- | --------------------- |
| String  | `str`   | `name@str`     | Text data             |
| Integer | `int`   | `id@int`       | Signed integer        |
| Float   | `float` | `salary@float` | Floating-point number |
| Boolean | `bool`  | `active@bool`  | Boolean value         |

Type names are lowercase and case-sensitive; `@INT` or `@Str` is an error.

**Hints are authoritative.** When a field carries a scalar hint, every non-null value must satisfy it, otherwise decoding fails:

| Hint     | Accepts                                   | Rejects                     |
| -------- | ----------------------------------------- | --------------------------- |
| `@int`   | integer literals                          | `1.5`, `1e3`, `007`, `hello` |
| `@float` | integer or float literals (`3` → `3.0`)   | `hello`, `true`             |
| `@bool`  | `true`, `false`                           | `True`, `1`, `yes`          |
| `@str`   | anything: unquoted `42` / `true` → string | — (`_` stays null)          |

A typed decoder must also reject a hint that contradicts the target type, e.g. `{age@str}` decoded into an integer field.
**Example with scalar hints:**

```asun
/* Full scalar hints */
[{id@int, name@str, salary@float, active@bool}]:
  (1, Alice, 5000.50, true),
  (2, Bob, 4500.00, false),
  (3, "Carol Smith", 6200.75, true)
```

**Key properties:**

- ✅ **Unified meaning**: `@` is the field binding marker for both scalar hints and complex structures
- ✅ **Optional scalar hints**: `@int`, `@str`, `@bool`, and `@float` can be omitted when you do not need extra type clarity
- ✅ **Required structural bindings**: `@{}` and `@[]` must stay for nested objects and arrays
- ✅ **Partial**: You can add scalar hints only to selected terminal fields
- ✅ **Negligible overhead**: Scalar hints only affect schema header parsing, not the data body. Vec scenarios <0.1%, single struct ~3% — constant overhead, does not grow with data volume

**Use cases:**

- **LLM prompts**: A schema with scalar hints can help the model understand and generate correct data
- **API documentation**: Self-describing structure without external docs
- **Cross-language exchange**: Eliminates type ambiguity (`42` — is it `int` or `float`?)
- **Debugging**: Field types visible at a glance

**Partial annotation example:**

```asun
/* Annotate only key fields (recommended style) */
[{id@int, name@str, email@str, age@int, bio@str}]:
  (1, Alice, alice@example.com, 30, "Engineer"),
  (2, Bob, bob@example.com, 28, "Designer")
```

**⚠️ CRITICAL WARNING: The `@` structural binding is mandatory for complex types!**

While `@type` for terminal scalar data (numbers, strings, etc.) is only an optional hint, for **complex type containers (nested objects, arrays)** the `@` followed by `{}` or `[]` acts as a crucial structural binding and **must absolutely not be omitted!**

- ✅ **Annotated nesting**: `dept@{title@str}`
- ✅ **Structural nesting without scalar hints**: `dept@{title}` (dropped `@str`, but the `@{}` must be kept to signal to the parser to enter the next object level)
- ✅ **Array without scalar hints**: `tags@[]` (kept `@[]` to indicate an array boundary)
- ❌ **Fatal omission**: `dept` (without the brackets, the parser will not know `dept` corresponds to a complex object, leading to an overall breakdown of stream reading)

### 3.3 Distinguishing Schema vs Value Character Rules

`schema` and `value` live in the same text, but characters do not mean the same thing in both places. Pay special attention to `@`, whitespace, and special characters:

| Position          | Rule                                                                                                                                                                                                                        |
| ----------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Schema field name | `@` is part of the structural/type marker, such as `name@str` or `users@[{id@int}]`. Here, `@` is **not field-name content**.                                                                                               |
| Schema field name | If a field name contains spaces, starts with a digit, or contains special characters, it should be written as a quoted field name, such as `"id uuid"`, `"65"`, or `"{}[]@\\\""`.                                           |
| Data value        | In the data section `@` and `:` have **no structural meaning**; they are ordinary characters and need no quotes: `alice@example.com`, `12:30`, `https://a.com/x`.                                                          |
| Data value        | Unquoted strings automatically trim leading and trailing spaces. To preserve outer whitespace, use quotes, for example `"  Alice  "`.                                                                                       |
| Data value        | An unquoted string must not contain raw `,` `(` `)` `[` `]` `{` `}` `"` `\`, control characters, or the sequence `/*` (it opens a comment). Quote the value instead; plain strings have no escapes.                                                  |

Example:

```text
{"id uuid"@str,"65"@bool,"{}[]@\\\""@str}:
(@Alice,true,value@demo)
```

Explanation:

- `"id uuid"`, `"65"`, and `"{}[]@\\\""` are **schema field names**
- `@Alice` and `value@demo` are **data values**
- The same `@` character marks structure in schema, but is plain string content in value

---

## 4. Escape Rules

Escapes exist **only inside quoted strings**, and they are exactly the JSON escapes. A plain (unquoted) string has no escapes: it is the source text as written, and a `\` outside quotes is an error. A value that contains a structural character is written as a quoted string.

| Character      | Escape            | Description                          |
| -------------- | ----------------- | ------------------------------------ |
| `"`            | `\"`              | Double quote                         |
| `\`            | `\\`              | Backslash                            |
| `/`            | `\/`              | Slash (optional, as in JSON)         |
| Control chars  | `\n` `\t` `\r` `\b` `\f` | Line feed, tab, CR, backspace, form feed |
| Unicode        | `\uXXXX`          | UTF-16 code unit (e.g., `\u4e2d`)    |

```text
{title@str,expr@str}:("Hello, World", "f(x) = [a, b]")
```

**Notes:**

- A document is UTF-8. A single leading BOM (U+FEFF) is skipped by decoders; encoders never write it.
- Unquoted strings must not contain raw `,()[]{}"\` or `/*`; quote such values.
- Quoted strings follow JSON (RFC 8259): `"` and `\` must be escaped, and so must every control character U+0000–U+001F (a raw newline or tab inside quotes is an error). DEL, C1 controls, U+2028/U+2029 may appear raw. Structural characters (`,()[]{}:@`) and `/*` are ordinary content inside quotes.
- Characters outside the BMP are written raw or as a surrogate pair (`\ud83d\ude00` → 😀). A lone surrogate is an error.
- Any other escape (e.g. `\,`, `\(`, `\x`, `\q`, `\ ` ) is an error.

## 5. Comments

Block comments are supported using `/* */` syntax:

```text
/* This is a user list */
[{name@str,age@int}]:(Alice,30),(Bob,25)
```

Multi-line comments:

```text
/*
  User data
  Last updated: 2024-01-01
*/
[{name@str,age@int}]:
  (Alice,30),
  (Bob,25)
```

**Comment placement:**  
A comment is layout, exactly like whitespace: it may appear between **any** two tokens — in the schema, around `:` and `,`, and inside data tuples and arrays. It may not appear inside a literal (quoted string, plain string, number, keyword). Comments do not nest.

```text
✅ Correct:
/* username */ {name@str, age@int /* years */}:
  (Alice /* name */, /* age */ 30)

[{id@int, tags@[str]}]:
  (1, [a, /* deprecated: b, */ c]),
  (2, [])  /* last row */

❌ Incorrect:
{name@str}:(Ali/* x */ce)       /* a comment cannot split a plain string */
```

Outside a quoted string, `/*` **always** opens a comment, so a plain string ends where `/*` begins: `(x /* note */)` is the string `x`. Write a literal `/*` inside quotes: `"a/*b"`. A lone `/` or `*` is ordinary content (`a/b*c`, `https://x`).

## 6. Syntax Rules

### 6.1 Single Object

```text
{name@str,age@int}:(Alice,30)
```

→ `{name: "Alice", age: 30}`

### 6.2 Array of Objects (same structure, multiple rows)

```text
[{name@str,age@int}]:(Alice,30),(Bob,25),(Charlie,35)
```

→ Array of 3 objects

> **Note:** An array of objects uses the `[{schema}]:` prefix, distinguishing it from a single object's `{schema}:`. The parser determines the format by the first character — `[` vs `{`.

### 6.3 Null / Optional Fields

```text
{name@str,age@int,email@str}:(Alice,30,_)
```

→ `{name: "Alice", age: 30, email: null}`

`_` stands for null in any position and for any type, including a whole nested object, array or map. A position is never left blank: `(Alice,30,)` is an error.

### 6.4 Nested Object

```text
{name@str,addr@{city@str,zip@int}}:(Alice,(NYC,10001))
```

→ `{name: "Alice", addr: {city: "NYC", zip: 10001}}`

### 6.5 Object with a Simple Array Field

```text
{name@str,scores@[int]}:(Alice,[90,85,92])
```

→ `{name: "Alice", scores: [90, 85, 92]}`

### 6.6 Object with an Array-of-Objects Field

```text
{team@str,users@[{id@int,name@str}]}:(Dev,[(1,Alice),(2,Bob)])
```

→ `{team: "Dev", users: [{id: 1, name: "Alice"}, {id: 2, name: "Bob"}]}`

### 6.7 Plain Array (no schema)

```text
[1,2,3]
```

→ `[1, 2, 3]`

### 6.8 Array of Arrays

```text
[[1,2],[3,4]]
```

→ `[[1, 2], [3, 4]]`

### 6.9 Empty Array

```text
[]
```

`[ ]` is also empty. An array holding a single null is `[_]`.

### 6.10 Empty Object

```text
{}:()
```

A zero-field schema matches `()`, the tuple with zero elements.

### 6.11 Mixed-Type Array

```text
[1,hello,true,3.14]
```

→ `[1, "hello", true, 3.14]`

### 6.12 Complex Nesting Example

```text
{company@str,employees@[{id@int,name@str,skills@[str]}],active@bool}:(ACME,[(1,Alice,[rust,go]),(2,Bob,[python])],true)
```

**Parsed result:**

- `company` = `"ACME"`
- `employees` = array of 2 objects
  - `{id: 1, name: "Alice", skills: ["rust", "go"]}`
  - `{id: 2, name: "Bob", skills: ["python"]}`
- `active` = `true`

### 6.13 String Handling Examples

```text
[{name@str,city@str,zip@str,note@str}]:
  (Alice, New York, "001234", hello world),
  (Bob, "  Los Angeles  ", 90210, "say \"hi\"")
```

**Parsed result:**

| Field | Alice row                                   | Bob row                                        |
| ----- | ------------------------------------------- | ---------------------------------------------- |
| name  | `"Alice"` (unquoted, auto-trimmed)          | `"Bob"` (unquoted, auto-trimmed)               |
| city  | `"New York"` (unquoted, auto-trimmed)       | `"  Los Angeles  "` (quoted, spaces preserved) |
| zip   | `"001234"` (quoted string)                  | `"90210"` (`@str` takes the token text)        |
| note  | `"hello world"` (unquoted, auto-trimmed)    | `"say \"hi\""` (quoted, escape supported)      |

### 6.14 Real-World Example: Database Query Results

```asun
[{id@int, name@str, dept@{title@str}, skills@[str], active@bool}]:
  (1, Alice, (Manager), [Rust, Go], true),
  (2, Bob, (Engineer), [Python, "C++"], false),
  (3, "Carol Smith", (Director), [Leadership, Strategy], true)
```

**Equivalent JSON:**

```json
[
  {
    "id": 1,
    "name": "Alice",
    "dept": { "title": "Manager" },
    "skills": ["Rust", "Go"],
    "active": true
  },
  {
    "id": 2,
    "name": "Bob",
    "dept": { "title": "Engineer" },
    "skills": ["Python", "C++"],
    "active": false
  },
  {
    "id": 3,
    "name": "Carol Smith",
    "dept": { "title": "Director" },
    "skills": ["Leadership", "Strategy"],
    "active": true
  }
]
```

**Token savings:** ASUN ~65 tokens vs JSON ~180 tokens — **64% reduction** 🌟

### 6.15 Map Field

A map is declared in the schema as `@[K:V]` and written as `[key:value, ...]`:

```text
{user@str,attrs@[str:int]}:(Alice,[age:30,score:95])
```

→ `{user: "Alice", attrs: {age: 30, score: 95}}`

| Schema              | Data                        | Meaning                              |
| ------------------- | --------------------------- | ------------------------------------ |
| `m@[str:int]`       | `[a:1,b:2]`                 | string keys, integer values          |
| `m@[int:str]`       | `[1:one,2:two]`             | integer keys                         |
| `m@[str:{x,y}]`     | `[a:(1,2),b:(3,4)]`         | struct values                        |
| `m@[str:[int]]`     | `[a:[1,2],b:[]]`            | array values                         |
| `m@[:]`             | `[a:1,b:x]`                 | untyped keys and values (resolved like any untyped value) |
| `m@[str:int]`       | `[]`                        | empty map                            |

**Rules:**

- A value is a map **only** when its binding is `@[K:V]`. Everywhere else `[` starts an array and `:` is ordinary content, so `[a:1,12:30]` with an array binding is `["a:1", "12:30"]`.
- The key type `K` is `str`, `int`, or omitted. Keys are never null, boolean or float.
- The first `:` ends the key. A key that contains `:` must be quoted (`["12:30":5]`); the value may contain `:` freely (`[start:12:30]` → `start` = `"12:30"`).
- Duplicate keys are an error. Entry order is preserved but has no meaning.
- A map is not a top-level form; wrap it in an object.

---

## 7. Syntax Quick Reference

| Element                | Schema Syntax                  | Data Syntax         |
| ---------------------- | ------------------------------ | ------------------- |
| Single object          | `{field1@type,field2@type}:`   | `(val1,val2)`       |
| Array of objects       | `[{field1@type,field2@type}]:` | `(v1,v2),(v3,v4)`   |
| Simple array field     | `field@[type]`                 | `[v1,v2,v3]`        |
| Array-of-objects field | `field@[{f1@type,f2@type}]`    | `[(v1,v2),(v3,v4)]` |
| Nested object field    | `field@{f1@type,f2@type}`      | `(v1,(v3,v4))`      |
| Map field              | `field@[ktype:vtype]`          | `[k1:v1,k2:v2]`     |
| Null value             | —                              | `_`                 |
| Empty array / map      | —                              | `[]`                |
| Empty object           | `{}:`                          | `()`                |

## 8. Detailed Rules

### 8.1 Type Resolution Priority

When parsing a value, the following order is attempted:

1. `_` or `null` → `null`
2. **Boolean** → `true` or `false` (lowercase only)
3. **Integer** → matches `-?(0|[1-9][0-9]*)`
4. **Float** → a JSON number with a fraction or an exponent: `-?(0|[1-9][0-9]*)(\.[0-9]+)?([eE][+-]?[0-9]+)?`
5. **String** → everything else

Quoted values are always strings. This order applies to values without a hint; a scalar hint overrides it (see §3.2).

A typed decoder that reads an unquoted, non-null value into a **string target** takes the token text as is, exactly as if the field had `@str`: `90210` → `"90210"`, `true` → `"true"`. `_` is still null, so it is an error for a non-optional string.

Examples:

| Value     | Parsed as         |
| --------- | ----------------- |
| `_`       | `null`            |
| `"_"`     | string `"_"`      |
| `true`    | boolean `true`    |
| `123`     | integer `123`     |
| `3.14`    | float `3.14`      |
| `hello`   | string `"hello"`  |
| `123abc`  | string `"123abc"` |
| `007`     | string `"007"`    |
| `null`    | `null` (accepted; encoders emit `_`) |
| `"null"`  | string `"null"`   |
| `TRUE`    | string `"TRUE"`   |

### 8.2 Null vs Empty String

| ASUN                            | Parse result              |
| ------------------------------- | ------------------------- |
| `{name@str,bio@str}:(Alice,_)`  | `bio = null`              |
| `{name@str,bio@str}:(Alice,"")` | `bio = ""` (empty string) |
| `{name@str,bio@str}:(Alice,"_")` | `bio = "_"` (string)     |
| `{name@str,bio@str}:(Alice,)`   | Error: blank position     |

Example:

```text
{name@str,bio@str}:(Alice,_)        /* bio = null */
{name@str,bio@str}:(Alice,"")       /* bio = "" (empty string) */
```

### 8.3 Top-Level Structure Detection

There are **three top-level forms**, determined by the first character(s):

| First char | Type                         | Example                                 |
| ---------- | ---------------------------- | --------------------------------------- |
| `{`        | Single object with schema    | `{name@str,age@int}:(Alice,30)`         |
| `[{`       | Array of objects with schema | `[{id@int,name@str}]:(1,Alice),(2,Bob)` |
| `[`        | Plain array                  | `[1,2,3]`                               |
| Other      | Bare value (type inferred)   | `42`, `true`, `hello`                   |

**Key rules:**

- After `{schema}:` there can be **exactly one** `(...)` data tuple (single object).
- After `[{schema}]:` there can be **zero or more** `(...)` data tuples, comma-separated (array of objects). Zero tuples is an empty array. Rows are never null and the row list has no trailing comma.
- The parser determines format from the first character — `{` vs `[` — no need for separate `_vec`-style APIs.
- A bare tuple `(...)` at the top level is forbidden; tuples may only appear after a schema or inside data.
- An empty document (only whitespace/comments) is **invalid**, as in JSON. A top-level null is written `_`.

### 8.4 Field Name Rules

- **Bare names**: `a-z`, `A-Z`, `0-9`, `_`; may start with a digit (`1st`, `2name`)
- **Anything else** (spaces, punctuation, non-ASCII) must use a quoted name: `"id uuid"`, `"中文"`, `"a,b"`
- A quoted name is compared after unescaping, as in JSON: `{"a"}` and `{a}` name the same field. No Unicode normalization is applied.
- **Duplicate names** in one schema (after unescaping) are an error.

### 8.5 Whitespace Handling

| Location             | Rule                                                        |
| -------------------- | ----------------------------------------------------------- |
| Between tokens       | Whitespace, newlines and comments are ignored               |
| Inside a bare name   | Not allowed: `{a b}` is an error, write `{"a b"}`           |
| Inside a plain value | Internal spaces/tabs preserved; leading/trailing trimmed    |
| Inside a quoted value | Preserved verbatim                                         |

Example:

```text
{name@str, age@int}:(Alice Smith, 30)
```

is equivalent to:

```text
{name@str,age@int}:(Alice Smith,30)
```

→ `{"name": "Alice Smith", "age": 30}`

### 8.6 Multi-Line Format

Newlines and indentation are treated as whitespace during parsing:

```text
[{name@str,age@int}]:
  (Alice,30),
  (Bob,25),
  (Charlie,35)
```

is equivalent to:

```text
[{name@str,age@int}]:(Alice,30),(Bob,25),(Charlie,35)
```

### 8.7 Negative Number Rules

The minus sign `-` must immediately precede the digit — no space allowed:

| Input   | Parsed as        | Note                       |
| ------- | ---------------- | -------------------------- |
| `-123`  | integer `-123`   | ✓ Correct                  |
| `- 123` | string `"- 123"` | ✗ Space → parsed as string |
| `-3.14` | float `-3.14`    | ✓ Correct                  |
| `-0`    | integer `0`      | ✓ Special case             |
| `007`   | string `"007"`   | Leading zeros: not a number |

### 8.8 Data Alignment Rule (Strict Mode)

**The number of data items must exactly match the number of schema fields.**

| Schema                | Data        | Result                   |
| --------------------- | ----------- | ------------------------ |
| `{a@int,b@int,c@int}` | `(1,2,3)`   | ✓ Correct                |
| `{a@int,b@int,c@int}` | `(1,2)`     | ✗ Error: missing field   |
| `{a@int,b@int,c@int}` | `(1,2,3,4)` | ✗ Error: too many fields |
| `{a@int,b@int,c@int}` | `(1,_,3)`   | ✓ Correct: `b = null`    |
| `{a@int,b@int,c@int}` | `(1,,3)`    | ✗ Error: blank position  |
| `{a@int,b@int,c@int}` | `(1,2,)`    | ✗ Error: blank position  |
| `{a@int,b@int}`       | `(_,_)`     | ✓ Correct: both null     |
| `{a@int,b@int}`       | `()`        | ✗ Error: 0 elements      |

**Rationale**: ASUN is a position-sensitive format. A field-count mismatch causes data misalignment and must be reported as a parse error. Decoders must never pad, truncate or silently skip values; a missing value is written `_`.

### 8.9 Common Error Examples

| Incorrect                   | Reasun                     | Correct                                     |
| --------------------------- | -------------------------- | ------------------------------------------- |
| `{a@int,b@int}:(1,2,3)`     | Too many values            | `{a@int,b@int,c@int}:(1,2,3)`               |
| `{a@int,b@int}:(1)`         | Missing field              | `{a@int,b@int}:(1,_)`                       |
| `{a@int,b@int,c@int}:(1,2)` | Insufficient data          | `{a@int,b@int,c@int}:(1,2,_)`               |
| `{a@int,b@int}:(1,)`        | Blank position             | `{a@int,b@int}:(1,_)`                       |
| `(1,2,3)`                   | Bare tuple needs schema    | `{a@int,b@int,c@int}:(1,2,3)`               |
| `{a@int,b@int}`             | Schema with no data        | `{a@int,b@int}:(_,_)` or `{a@int,b@int}:(1,2)` |
| `{a@int,b@int}:(1,2,)`      | Trailing comma             | `{a@int,b@int}:(1,2)`                       |
| `{a@str}:(a\,b)`            | No escapes in plain strings | `{a@str}:("a,b")`                          |
| `{a@int,b@int}[1,2]`        | Missing colon after schema | `{a@int,b@int}:(1,2)`                       |

### 8.10 No Blank Positions

A comma only separates two values. Every position in a tuple, array or map holds exactly one value, and null is written `_`. A leading, trailing or doubled comma is an error.

| Input           | Elements | Parsed result            |
| --------------- | -------- | ------------------------ |
| `()`            | 0        | `()` (matches `{}`)      |
| `(_)`           | 1        | `(null)`                 |
| `(Alice, 30, _)`| 3        | `("Alice", 30, null)`    |
| `[]` / `[ ]`    | 0        | `[]`                     |
| `[_]`           | 1        | `[null]`                 |
| `[1, _, 3]`     | 3        | `[1, null, 3]`           |
| `(a,)`, `(,)`, `[1,,3]`, `[1,2,]` | — | Error         |

Schema field lists and row lists have no blank positions either: `{a,b,}` and `[{a}]:(1),` are errors.

---

## 9. Implementation Notes

### 9.1 Parser Implementation Highlights

1. **Non-greedy matching**: When parsing `plain_str`, delimiters `,()[]{}"` and the sequence `/*` end the value (inside a map entry's key, `:` too). A `\` in a plain value is an error.
2. **Zero-copy plain values**: Plain strings have no escapes, so after trimming a plain value is exactly a slice of the input.
3. **Comments are layout**: Skip `/* */` wherever whitespace is skipped. Never strip them from inside quoted strings.
4. **Streaming parse**: After parsing the schema, build a field index table; fill data fields by position.
5. **Progress guarantee**: Every skip/recovery loop must consume input or fail; a malformed byte must produce an error, never a hang.
6. **Nesting limit**: A decoder may cap structural nesting (schemas, bindings, tuples, arrays) to bound stack use; the reference implementation uses 128. The cap applies equally to values that are skipped (unknown fields), so the same document is accepted or rejected regardless of the target type.

### 9.2 Error Handling

| Error Type           | Example                 | Handling                       |
| -------------------- | ----------------------- | ------------------------------ |
| Field count mismatch | `{a@int,b@int}:(1,2,3)` | Throw error with location info |
| Unclosed quote       | `("hello)`              | Throw error                    |
| Unclosed bracket     | `{a@int,b@int}:(1,2`    | Throw error                    |
| Unclosed comment     | `/* comment`            | Throw error                    |
| Invalid escape       | `"\x"`, `"\,"`, `a\,b`  | Throw error                    |
| Blank position       | `(1,,3)`, `[1,]`        | Throw error                    |
| Duplicate map key    | `[a:1,a:2]`             | Throw error                    |
| Hint mismatch        | `{a@int}:(1.5)`         | Throw error                    |
| Duplicate field      | `{a,a}:(1,2)`           | Throw error                    |
| Number out of range  | `1e309`, `{a@int}` into `u8` with `300` | Throw error    |
| Empty document       | _(empty)_               | Throw error                    |

---

## 10. Complete Grammar BNF (Simplified)

The authoritative grammar is [`conformance/GRAMMAR.abnf`](../conformance/GRAMMAR.abnf) (RFC 5234 + RFC 7405, loadable by standard ABNF tools); this is a readable summary.

```bnf
asun        ::= BOM? ows (object_expr | array_expr | array | scalar) ows

object_expr ::= schema ":" tuple
array_expr  ::= "[" schema "]" ":" (tuple ("," tuple)*)?
schema      ::= "{" (field ("," field)*)? "}"
field       ::= name ("@" binding)?
binding     ::= "int" | "float" | "str" | "bool" | schema
              | "[" binding? "]"                      /* array */
              | "[" ("str" | "int")? ":" binding? "]" /* map   */
name        ::= [a-zA-Z0-9_]+ | quoted_str

tuple       ::= "(" (element ("," element)*)? ")"
array       ::= "[" (element ("," element)*)? "]"
map         ::= "[" entry ("," entry)* "]"            /* only under a map binding */
entry       ::= key ":" element
key         ::= number | quoted_str | plain_str_without_colon
element     ::= scalar | tuple | array | map

scalar      ::= "true" | "false" | "_" | "null" | number | quoted_str | plain_str
number      ::= "-"? ("0" | [1-9][0-9]*) ("." [0-9]+)? ([eE] [+-]? [0-9]+)?
quoted_str  ::= '"' (char_no_ctrl_quote_bs | escape)* '"'
plain_str   ::= word ([ \t]+ word)*            /* no raw ,()[]{}"\ ctrl, no "/*", no escapes */
escape      ::= "\" ( '"' | "\" | "/" | "b" | "f" | "n" | "r" | "t" | "u" hex hex hex hex )

ows         ::= ( " " | "\t" | "\r" | "\n" | comment )*
comment     ::= "/*" ... "*/"                  /* no nesting */
```

**Notes:**

- All literals are case-sensitive.
- Layout (`ows`, including comments) is allowed between any two tokens.
- `plain_str` values never include surrounding whitespace, so they are trimmed by construction.
- Semantic rules (element alignment, hints, numeric range, surrogates, duplicate names, maps) are listed at the end of `GRAMMAR.abnf`.

---

## 11. ASUN Binary Format Specification

In addition to the human-readable text format, ASUN defines a compact binary wire format for high-performance serialization/deserialization. This section describes the format as implemented by `asun-rs`, the reference implementation.

### 11.1 Design Principles

- **No schema header**: The binary format contains no field names or type tags — it relies entirely on the target type for positional decoding. Both sides must agree on the type definition.
- **Variable-length integers**: Integers use LEB128 varints (signed integers are zigzag-encoded first), so small values take one byte.
- **Length-prefixed**: Strings, byte strings and sequences are preceded by a varint length or element count.
- **Zero-copy decoding**: String fields can borrow directly from the input buffer — no new memory allocated.

### 11.2 Varints

`uvarint` is unsigned LEB128: 7 payload bits per byte, least significant group first; the high bit (`0x80`) is set on every byte except the last. A `u64` takes at most 10 bytes.

`ivarint` maps a signed value to unsigned with zigzag, `(v << 1) ^ (v >> 63)` (`0 → 0`, `-1 → 1`, `1 → 2`, `-2 → 3`, …), then writes it as a `uvarint`.

Each value has exactly one valid encoding. A decoder rejects:

- padded encodings: a multi-byte varint whose last byte is `0x00` (e.g. `80 00` for 0);
- a tenth byte other than `0x01`, and any eleventh byte (more than 64 bits);
- a value too large for the target type (e.g. 65536 for `u16`, or a code point above U+10FFFF for `char`).

### 11.3 Type Encoding Rules

| Type                    | Encoding                                                                |
| ----------------------- | ----------------------------------------------------------------------- |
| `bool`                  | 1 byte: `0x00` = false, `0x01` = true; other values are an error        |
| `i8` / `u8`             | 1 raw byte                                                              |
| `i16` / `i32` / `i64`   | `ivarint`                                                               |
| `u16` / `u32` / `u64`   | `uvarint`                                                               |
| `f32` / `f64`           | IEEE 754 bits, 4 / 8 bytes little-endian (NaN payloads and `-0.0` kept) |
| `char`                  | `uvarint` of the Unicode scalar value (surrogates are an error)         |
| `str`                   | `uvarint` byte length + UTF-8 bytes (invalid UTF-8 is an error)         |
| `Option<T>`             | tag byte `0x00` = none, or `0x01` + encoding of `T`; other tags are an error |
| `Vec<T>` / array        | `uvarint` element count + each element encoded in order                 |
| map                     | `uvarint` entry count + each entry as key encoding then value encoding  |
| tuple                   | elements encoded in order, no prefix                                    |
| `struct`                | fields encoded in declaration order, no prefix, padding or alignment    |
| `enum`                  | `uvarint` variant index (declaration order, from 0) + the variant's fields in order |
| `()` / unit struct      | nothing (0 bytes)                                                       |

Fields skipped with `#[asun(skip)]`-style attributes are omitted on both sides, so encoder and decoder stay aligned.

### 11.4 Single Struct vs Struct Array

**Single struct**: Fields encoded in declaration order directly, no wrapper.

```text
struct User { id: i64, name: string, active: bool }

Encoding: [ivarint][uvarint len + UTF-8 bytes][u8]
```

**Struct array**: Prefixed with a `uvarint` element count, followed by each element's encoding.

```text
Array<User>

Encoding: [uvarint count][User₁][User₂]...[Userₙ]
```

> **Important**: Single structs and struct arrays use different binary layouts. A single struct has no count prefix; a struct array does. The decoder must know whether the target type is a single instance or an array.

### 11.5 Encoding Examples

```text
struct User { id: i64, name: string, active: bool }
Value: { id: 42, name: "Alice", active: true }

Binary (8 bytes):
  54                        ← ivarint: 42 (zigzag 84)
  05                        ← uvarint: string length 5
  41 6C 69 63 65            ← UTF-8: "Alice"
  01                        ← bool: true
```

```text
Array<Point> = [{ x: 1.0, y: 2.0 }, { x: 3.0, y: 4.0 }]   (x, y: f32)

Binary (17 bytes):
  02                        ← uvarint: count = 2
  00 00 80 3F 00 00 00 40   ← Point₁: (1.0, 2.0)
  00 00 40 40 00 00 80 40   ← Point₂: (3.0, 4.0)
```

```text
Integers and options:
  i64 300          → D8 04            (zigzag 600)
  i64 -1           → 01
  u64 300          → AC 02
  Option<str> "hi" → 01 02 68 69
  Option<str> none → 00

enum Shape { Unit, New(i32) }
  Shape::Unit      → 00
  Shape::New(-5)   → 01 09
```

### 11.6 Correspondence with Text Format

| Text format                      | Binary format                       |
| -------------------------------- | ----------------------------------- |
| `{schema}:(data)` single object  | Fields encoded in order, no wrapper |
| `[{schema}]:(d1),(d2),...` array | `uvarint` count + element sequence  |
| `[v1,v2,v3]` plain array         | `uvarint` count + element sequence  |
| `[k1:v1,k2:v2]` map              | `uvarint` count + key/value pairs   |
| `true` / `false`                 | Single byte `0x01` / `0x00`         |
| `_` / Option null                | Single byte `0x00`                  |
| Option some(v)                   | `0x01` + value encoding             |

### 11.7 Decoding Untrusted Input

A conforming decoder must fail cleanly — never crash, hang or exhaust memory — on malformed input:

- Truncated input, and lengths or counts larger than the remaining input, are errors.
- Sequence counts are capped (`asun-rs`: 16 Mi elements by default, configurable). The same budget bounds the total number of zero-sized elements in one input, since they consume no bytes.
- Memory reserved up front for a sequence is bounded (`asun-rs`: 1 MiB); a large claimed count must not allocate before its elements arrive.
- Sequence nesting is limited to 128 levels, so recursive types cannot overflow the stack.
- Trailing bytes after a value may be rejected (`asun-rs` `decode_binary_exact`) or left for the caller (`decode_binary`).

### 11.8 Performance Characteristics

| Characteristic   | Text Format                        | Binary Format                           |
| ---------------- | ---------------------------------- | --------------------------------------- |
| Human readable   | ✓                                  | ✗                                       |
| Encode speed     | Fast (1.5–2× JSON)                 | Extremely fast (6–8× JSON)              |
| Decode speed     | Fast (1.5–2.3× JSON)               | Extremely fast (17–47× JSON)            |
| Size             | Compact (50–55% smaller than JSON) | More compact (39–55% smaller)           |
| Zero-copy decode | ✗                                  | ✓ (strings reference input buffer)      |
| Use cases        | API communication, LLM, debug      | RPC, IPC, high-frequency data pipelines |

### 11.9 Byte Order and Layout

- Fixed-width floats use **little-endian** byte order; integers are varints and have no byte order.
- String content uses **UTF-8** encoding.
- No alignment padding.
- No metadata or magic numbers.

---

## 12. LLM Best Practices & Benchmarks

ASUN is optimized for interaction with large language models. This section provides prompt templates, accuracy benchmarks, and corrections for common LLM errors.

### 12.1 Core Design Advantages

| Dimension              | ASUN Advantage                                             |
| ---------------------- | ---------------------------------------------------------- |
| **Token efficiency**   | 30–70% fewer tokens than JSON → more context headroom      |
| **Structure clarity**  | Schema declared once → lower model comprehension cost      |
| **Generation pattern** | Row-oriented table format matches LLM training data        |
| **Streaming**          | Native row-by-row streaming, no need to buffer full output |

### 12.2 Recommended System Prompt

```text
You are a data format conversion expert. Output data in ASUN format (Array-Schema Unified Notation).

Core ASUN rules:
1. Single object: `{field1@type, field2@type, ...}:(val1, val2, ...)`
2. Array of objects: `[{field1@type, field2@type, ...}]:(val1, val2, ...),(val3, val4, ...)`
3. Type annotations are optional: `{field1@int, field2@str, ...}`
4. Supported types: `int`, `float`, `str`, `bool`, arrays, nested objects
5. Array fields: `name@[str]` is a string array with value `[item1, item2, ...]`; array-of-objects is `users@[{id@int}]`
6. Map fields: `attrs@[str:int]` with value `[key1:1, key2:2]`

You must follow:
- Single objects use `{schema}:` prefix; arrays of objects use `[{schema}]:` prefix
- Number of data items = number of schema fields (strict alignment)
- Strings generally need no quotes (unless they contain special characters)
- Null values are written `_`; never leave a position blank
- Quote any string containing , ( ) [ ] { } " or \
- Nested objects use parentheses: outer@{inner@type}:(val1,(nested_val))

Output ASUN only, no additional explanation.
```

### 12.3 Few-Shot Examples

```text
# Example 1: Multiple user records
Input: User Alice age 30, User Bob age 25
Output:
[{name@str, age@int}]:
  (Alice, 30),
  (Bob, 25)

# Example 2: Single user
Input: ID 1 Alice Email alice@example.com Active true
Output:
{id@int, name@str, email@str, active@bool}:
  (1, Alice, alice@example.com, true)

# Example 3: Nested data
Input: Company Acme, Department Engineering, Head Alice
Output:
{company@str, dept@{name@str, head@str}}:
  (Acme, (Engineering, Alice))
```

### 12.4 Accuracy Benchmarks

| Model             | Format Correct | Type Correct | Recommended |
| ----------------- | -------------- | ------------ | ----------- |
| GPT-4 Turbo       | 99%+           | 97%+         | ⭐⭐⭐⭐⭐  |
| Claude 3.5 Sonnet | 98%+           | 96%+         | ⭐⭐⭐⭐⭐  |
| GPT-4o            | 96%+           | 93%+         | ⭐⭐⭐⭐    |
| Llama-2 70B       | 90%            | 85%          | ⭐⭐⭐      |

### 12.5 Common Errors and Fixes

#### Error 1: Field Count Mismatch

```asun
❌ {id@int,name@str}:(1,Alice,true)
✅ {id@int,name@str,active@bool}:(1,Alice,true)
```

#### Error 2: Wrong Bracket Type for Nesting

```asun
❌ {user@{id@int,name@str}}:({1,Alice})      /* curly braces used */
✅ {user@{id@int,name@str}}:((1,Alice))      /* nested data uses parentheses */
```

#### Error 3: Wrong Bracket Type for Arrays

```asun
❌ {tags@[str]}:({python,rust})         /* curly braces */
✅ {tags@[str]}:([python,rust])         /* arrays use square brackets */
```

### 12.6 Token Cost Comparison

Test data: 500 user records, 8 fields each

| Format | Token count | Cost (GPT-4) | Savings |
| ------ | ----------- | ------------ | ------- |
| JSON   | 12,450      | $0.55        | —       |
| ASUN   | 4,280       | $0.19        | **65%** |

### 12.7 Integration Workflow

```
Request → Build prompt (system role + few-shot + data)
  ↓
Call LLM (temperature = 0.1–0.3)
  ↓
Validate format (field count, alignment, types)
  ├─ ✓ Pass → return ASUN
  └─ ✗ Fail → retry (max 3 times)
```

8. 📋 v2.0: Implement serde support
9. 📋 v2.0: Schema references and aliases
10. 📋 v2.0: Full data validation framework

---

## 13. Appendix A: Type System Reference

### A.1 Type Annotation Quick Reference

| Scenario                     | Schema             | Data        | Note                |
| ---------------------------- | ------------------ | ----------- | ------------------- |
| No annotations (v1.3 compat) | `{id,name}`        | `(1,Alice)` | Implicit inference  |
| Explicit string              | `{name@str}`       | `(Alice)`   | Clear type          |
| Integer                      | `{id@int}`         | `(1)`       | Integer type        |
| Float                        | `{score@float}`    | `(95.5)`    | Floating-point      |
| Boolean                      | `{active@bool}`    | `(true)`    | Boolean             |
| Array                        | `{tags@[str]}`     | `([a,b,c])` | Array type          |
| Nested                       | `{user@{id@int}}`  | `((1))`     | Nested object       |
| Map                          | `{m@[str:int]}`    | `([a:1])`   | Map type            |
| Null                         | `{id@int}`         | `(_)`       | Null value          |
| Mixed                        | `{id@int,bio@str}` | `(1,Bio)`   | Partial annotations |

### A.2 Type Assertion Examples

> **Note**: Scalar hints are authoritative (§3.2): a value that does not satisfy its hint is a decode error.

```asun
/* Correctly typed data */
[{id@int, score@float, pass@bool, name@str}]:
  (1, 95.5, true, Alice),      ✅ All types match
  (2, 87.0, false, Bob),       ✅ All types match
  (3, 76, true, "Carol Sue")   ✅ 76 auto-promoted to 76.0

/* Type mismatches — decoding fails */
  ("1", 95.5, true, Alice),    ❌ id must be int, not string
  (1, "95.5", true, Alice)     ❌ score must be float, not string
```

---

## 14. Appendix B: Comparison with Other Formats

### B.1 ASUN vs JSON vs CSV

```json
// JSON (100 tokens)
{
  "users": [
    { "id": 1, "name": "Alice", "role": "engineer", "active": true },
    { "id": 2, "name": "Bob", "role": "designer", "active": false }
  ]
}
```

```csv
// CSV (80 tokens, but no type information)
id,name,role,active
1,Alice,engineer,true
2,Bob,designer,false
```

```asun
/* ASUN (35 tokens, with type information) */
[{id@int, name@str, role@str, active@bool}]:
  (1, Alice, engineer, true),
  (2, Bob, designer, false)
```

### B.2 Nested Data Comparison

```json
// JSON
{
  "employees": [
    { "id": 1, "name": "Alice", "dept": { "name": "Eng", "budget": 500000 } }
  ]
}
```

```asun
/* ASUN — more compact */
{employees@[{id@int, name@str, dept@{name@str, budget@int}}]}:
  ([(1, Alice, (Eng, 500000))])
```

---

## 15. Appendix C: Implementation Checklist

### C.1 Parser Checklist

- [ ] Lexer: recognize all symbols (`{`, `}`, `(`, `)`, `[`, `]`, `:`, `,`)
- [ ] Comment handling: support `/* */` block comments
- [ ] String parsing: support both quoted and unquoted forms
- [ ] Escape rules: JSON escapes inside quoted strings only; reject `\` in plain strings
- [ ] Null and blanks: `_` / `null` are null; reject blank positions
- [ ] Maps: parse `[k:v,...]` under `@[K:V]` bindings; reject duplicate keys
- [ ] Type annotation parsing: recognize `@int`, `@str`, etc.
- [ ] Schema parsing: recursively parse nested structures
- [ ] Alignment check: verify data field count matches schema field count
- [ ] Error reporting: location-aware error messages

### C.2 Serializer Checklist

- [ ] Value → ASUN: serialize in-memory objects to text
- [ ] Indentation and alignment: optional column alignment
- [ ] Type annotation generation: infer and emit types from data
- [ ] Quoting logic: determine when quotes are required vs optional
- [ ] Performance optimization: streaming serialization for large datasets

### C.3 Validation Tool Checklist

- [ ] Format check: ASUN syntactic correctness
- [ ] Type check: field types match data (optional strict mode)
- [ ] Data alignment: field count consistency
- [ ] Schema validation: field name legality
- [ ] LLM output validation: catch common LLM generation errors

---

## 16. Appendix D: FAQ

### Q1: Why doesn't ASUN use a human-friendly format similar to JSON?

**A**: ASUN is optimized for LLMs, which requires:

1. Minimizing token consumption (every character counts)
2. Clear structural boundaries (easy to parse)
3. Natural support for row-by-row streaming

Compared to JSON's indented readability, token savings and LLM friendliness take priority.

### Q2: If implicit type inference is supported, why add scalar hints with `@type`?

**A**:

- **Unified binding model**: `@` consistently binds a field to its following schema or type description
- **Production safety**: Explicit scalar hints help prevent misinterpretation (e.g., `"123"` vs `123`)
- **LLM friendliness**: Models generate more accurate data when scalar hints are explicit
- **Optional for scalars**: Developers use them only when needed on terminal fields

### Q3: Does ASUN support large files?

**A**: Yes. The row-oriented format naturally supports:

- Streaming without loading the entire file
- Each row is independent, suitable for distributed processing
- For very large datasets, the ASUNL format is recommended (similar to JSONL)

### Q4: How does ASUN compare to MessagePack and Protocol Buffers?

**A**:

- **MessagePack**: Binary format, more compact but not human-readable
- **Protocol Buffers**: Requires `.proto` definition files, higher complexity
- **ASUN**: Text format for LLM use, balancing token efficiency with readability

### Q5: What are the most common errors when LLMs generate ASUN?

**A**: See Section 12.5 for details. The three most common:

1. Field count mismatch
2. Wrong bracket type for nesting
3. Wrong bracket type for arrays

### Q6: Is ASUN's parse performance really faster than JSON?

**A**: Yes. When processing arrays of objects (list data), ASUN's parse performance far exceeds JSON, due to **schema-driven parsing**:

1. **Zero key-hashing**: No need to repeatedly read key strings and compute hashes as JSON does.
2. **Excellent branch prediction**: The parser knows the type of the next value in advance, maximizing CPU instruction pipeline efficiency.
3. **Minimal memory usage**: All data rows share a single schema reference — no per-row key memory allocation.

---

**Document version**: v1.6.0  
**Last updated**: 2026-10-07  
**License**: MIT  
**GitHub**: https://github.com/asunLab/asun
