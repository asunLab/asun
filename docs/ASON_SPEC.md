# ASON Format Specification v1.0

> **ASON** = Array-Schema Object Notation  
> _"The efficiency of arrays, the structure of objects."_

ASON is a serialization format designed for large-scale data transmission and LLM (Large Language Model) interactions. By **separating schema from data**, it eliminates the repetitive key redundancy found in JSON and introduces a Markdown-inspired, row-oriented syntax that minimizes token consumption while maintaining excellent human readability.

---

## 1. Design Philosophy

- **Schema/Data Separation**: Structure is declared once; the data section carries only values.
- **Row-Oriented**: Supports multi-line format to enhance vertical visual flow, making it easy to read and stream.
- **Structural Harmony**: Uses `{}` to define the skeleton (Schema) and `()` to carry the body (Data) — symbols with clearly distinct semantics.
- **Token Optimization**: Compared to JSON, reduces token consumption by 30–70% in list scenarios.
- **Extreme Parse Performance**: Zero key-hashing and schema-driven parsing deliver deserialization speeds far exceeding JSON.

---

## 1.1 Why ASON?

### JSON vs ASON

```json
// JSON: 100 tokens
{
  "users": [
    { "id": 1, "name": "Alice", "active": true },
    { "id": 2, "name": "Bob", "active": false }
  ]
}
```

```ason
// ASON: ~35 tokens (65% token saving)
[{id@int, name@str, active@bool}]:
  (1, Alice, true),
  (2, Bob, false)
```

| Aspect              | JSON              | ASON              |
| ------------------- | ----------------- | ----------------- |
| **Token Efficiency**| 100%              | 30–70% ✓          |
| **Key Repetition**  | Every row         | Declared once ✓   |
| **Human Readable**  | High              | High ✓            |
| **Learning Curve**  | Zero              | Low               |
| **Nesting**         | Arbitrary depth   | Clear & bounded ✓ |
| **Streaming**       | Line-by-line      | Native support ✓  |

---

## 1.2 Extreme Parse Performance

Beyond token savings in LLM scenarios, ASON also dramatically outperforms JSON in traditional serialization/deserialization (serde) use cases — behaving much more like CSV or Protobuf:

1. **Zero Key-Hashing**:
   - **JSON**: When parsing an array of objects, every key in every row must be read, hashed, and matched against the target struct. 1,000 rows means 1,000 repeated hash lookups.
   - **ASON**: The parser first parses the schema to build a positional index (e.g., `0 → id`, `1 → name`). When parsing data rows, values are assigned via array index in O(1) — zero hash computation or string matching throughout.

2. **Schema-Driven Parsing**:
   - **JSON**: The parser must dynamically infer the type of each value by peeking at the next character (`"`, `t`, `f`, `[`, `{`, digit), causing frequent CPU branch mispredictions.
   - **ASON**: The schema provides structural information (e.g., `{id, name, active}`), so the parser reads each field value in a fixed, known order without dynamic type inference. In a serde framework the target struct's type is already known at compile time, and the parser calls `parse_int()` etc. directly. In schema text, `@` is the binding marker between a field and its following schema/type description: scalar hints like `{id@int}` are optional, while structural bindings such as `@{}` and `@[]` are required for complex fields.

3. **Low Memory Footprint**:
   - JSON builds a dynamic DOM tree that allocates memory for every key string in every object. ASON data rows are essentially a flat tuple array — all rows share a single schema reference, keeping memory overhead minimal.

---

## 1.3 Core Architecture Philosophy

ASON's design goes beyond just saving tokens; its deep architectural philosophy dictates its extreme performance and precise semantics:

1. **Physical Isolation via `Header : Body`**
   - ASON uses `:` to strictly divide the text into Schema (Blueprint) and Data (Body).
   - This physical isolation allows the parser to split the payload in **O(1)** time. The front-end parser can parse the Schema independently, pre-allocate structural memory, and then engage in high-speed or multi-threaded streaming of massive Data rows, completely discarding JSON's inefficient paradigm of continuously inferring object boundaries while parsing data.

2. **Tuple Semantics `()` vs Object Semantics `{}`**
   - In ASON, Schema uses `{}` to define the unordered key-value blueprint, but crucially, **data bodies are strictly wrapped in `()`**.
   - `()` represents a **Tuple** in programming languages — a strictly ordered, keyless, position-bound data structure. It sends a strong signal to humans and AI: data must perfectly align with the schema positions and cannot be shuffled like in JSON. This architectural "harmony of shapes" visually isolates structural definitions from pure data with extreme sharpness, drastically reducing structure-related parsing errors.

3. **Native Immunity to Key Collisions**
   - JSON syntax allows ambiguous key duplications (e.g., `{"age": 30, "age": 40}`), often leading to the latter covering the former.
   - By confining keys exclusively to the Schema header, ASON data regions consist solely of compact values (`(30), (40)`), completely insulating the format from key collisions at the syntactic source. By reducing Maps/Dictionaries to arrays of key-value tuples (e.g., `[{key, value}]: ((age,30))`), ASON perfectly extends this high-performance, collision-free design abstraction.

---

## 2. Core Syntax Preview

```ason
[{id@int, name@str, tags@[str]}]:
  (1, Alice, [rust, go]),
  (2, Bob, [python, c++])
```

- `{...}` **(Schema)**: Defines field names, nested objects `{}`, or array types `[...]`
- `:` **(Separator)**: Marks the boundary between schema and data
- `(...)` **(Data Tuple)**: Carries values in order, strictly aligned with the schema definition

## 3. Data Type Rules

| Type             | Example       | Description                                        |
| ---------------- | ------------- | -------------------------------------------------- |
| Integer          | `42`, `-100`  | Matches `-?[0-9]+`                                 |
| Float            | `3.14`, `-0.5`| Matches `-?[0-9]+\.[0-9]+`                         |
| Boolean          | `true`, `false` | Must be lowercase literals                       |
| Null             | *(blank)*     | Empty content between commas parses as `null`      |
| Empty string     | `""`          | Explicit empty string                              |
| Unquoted string  | `Hello World` | Leading/trailing spaces auto-trimmed; must escape `,()[]\` |
| Quoted string    | `" Space "`   | Spaces preserved as-is; supports `\"` escaping     |

### 3.1 String Rules

ASON supports two string forms:

| Form     | Example       | Behavior                               |
| -------- | ------------- | -------------------------------------- |
| Unquoted | `hello world` | Leading/trailing whitespace auto-trimmed |
| Quoted   | `" hello "`   | Content preserved verbatim (including spaces) |

**When to use quotes:**

- Preserve leading/trailing spaces: `" hello "`
- Leading zeros: `"001234"` (e.g., zip codes)
- Force string type: `"true"`, `"123"`
- Empty string: `""`

### 3.2 Field Binding and Optional Scalar Hints

**ASON v1.4 introduces the `@` binding syntax.** `@` is not merely a type-annotation symbol; it is the **field binding marker** between a field name and its following schema/type description.

> **Core principle: `@` carries both structural binding and optional scalar hints.** For terminal scalar fields, `@type` is an optional hint; for complex fields, `@{}` / `@[]` are mandatory structural bindings. Both of the following are layout-equivalent:
>
> ```ason
> // Without annotations
> {id,name,active}:(1,Alice,true)
>
> // With scalar hints — identical parse result
> {id@int,name@str,active@bool}:(1,Alice,true)
> ```

**Supported types:**

| Type    | Syntax                | Example               | Description                               |
| ------- | --------------------- | --------------------- | ----------------------------------------- |
| String  | `str`      | `name@str`            | Text data                                 |
| Integer | `int`      | `id@int`              | Signed integer                            |
| Float   | `float`    | `salary@float`        | Floating-point number                     |
| Boolean | `bool`     | `active@bool`         | Boolean value                             |

**Example with scalar hints:**

```ason
// Full scalar hints
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

```ason
// Annotate only key fields (recommended style)
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

| Position | Rule |
| --- | --- |
| Schema field name | `@` is part of the structural/type marker, such as `name@str` or `users@[{id@int}]`. Here, `@` is **not field-name content**. |
| Schema field name | If a field name contains spaces, starts with a digit, or contains special characters, it should be written as a quoted field name, such as `"id uuid"`, `"65"`, or `"{}[]@\\\""`. |
| Data value | In the data section, `@` is **not a type marker**; it is just an ordinary character. However, to avoid confusion with schema syntax, any value containing `@` should be written as a quoted string, for example `"@Alice"`. |
| Data value | Unquoted strings automatically trim leading and trailing spaces. To preserve outer whitespace, use quotes, for example `"  Alice  "`. |
| Data value | If a value contains delimiters or other potentially ambiguous special characters, prefer a quoted string. Do not apply schema field-name rules directly to data values. |

Example:

```text
{"id uuid"@str,"65"@bool,"{}[]@\\\""@str}:
("@Alice",true,"value@demo")
```

Explanation:

- `"id uuid"`, `"65"`, and `"{}[]@\\\""` are **schema field names**
- `"@Alice"` and `"value@demo"` are **data values**
- The same `@` character marks structure in schema, but is plain string content in value

---

## 4. Escape Rules

Escape special characters when they appear inside string values:

| Character | Escape   | Description          |
| --------- | -------- | -------------------- |
| `,`       | `\,`     | Comma                |
| `(`       | `\(`     | Left parenthesis     |
| `)`       | `\)`     | Right parenthesis    |
| `[`       | `\[`     | Left square bracket  |
| `]`       | `\]`     | Right square bracket |
| `"`       | `\"`     | Double quote         |
| `\`       | `\\`     | Backslash            |
| Newline   | `\n`     | Line feed            |
| Tab       | `\t`     | Horizontal tab       |
| Unicode   | `\uXXXX` | Unicode character (e.g., `\u4e2d`) |

**Notes:**

- ASON uses UTF-8 encoding by default.
- In unquoted strings, `,()[]` must be escaped.
- In quoted strings, at minimum `"` and `\` must be escaped; control characters may use `\n`, `\t`, `\r`, `\b`, and `\f`.

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

**Comment placement restrictions:**  
To ensure maximum streaming parse performance, comments **may only appear at the beginning or end of a line**, or in gaps between schema definitions. **Comments inside data tuples `(...)` are strictly forbidden.**

```text
✅ Correct:
/* username */ {name@str, age@int}:
  (Alice, 30) /* age */

❌ Incorrect (comments inside data are forbidden):
{name@str, age@int}:(Alice, /* age */ 30)
```

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
{name@str,age@int,email@str}:(Alice,30,)
```

→ `{name: "Alice", age: 30, email: null}`

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

### 6.10 Empty Object

```text
()
```

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

| Field | Alice row                                    | Bob row                                       |
| ----- | -------------------------------------------- | --------------------------------------------- |
| name  | `"Alice"` (unquoted, auto-trimmed)           | `"Bob"` (unquoted, auto-trimmed)              |
| city  | `"New York"` (unquoted, auto-trimmed)        | `"  Los Angeles  "` (quoted, spaces preserved)|
| zip   | `"001234"` (string, leading zero preserved)  | `90210` (all digits → parsed as integer)      |
| note  | `"hello world"` (unquoted, auto-trimmed)     | `"say \"hi\""` (quoted, escape supported)     |

### 6.14 Real-World Example: Database Query Results

```ason
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

**Token savings:** ASON ~65 tokens vs JSON ~180 tokens — **64% reduction** 🌟

---

## 7. Syntax Quick Reference

| Element              | Schema Syntax                        | Data Syntax           |
| -------------------- | ------------------------------------ | --------------------- |
| Single object        | `{field1@type,field2@type}:`         | `(val1,val2)`         |
| Array of objects     | `[{field1@type,field2@type}]:`       | `(v1,v2),(v3,v4)`     |
| Simple array field   | `field@[type]`                       | `[v1,v2,v3]`          |
| Array-of-objects field | `field@[{f1@type,f2@type}]`        | `[(v1,v2),(v3,v4)]`   |
| Nested object field  | `field@{f1@type,f2@type}`            | `(v1,(v3,v4))`        |
| Null value           | —                                    | *(blank)*             |
| Empty array          | —                                    | `[]`                  |
| Empty object         | —                                    | `()`                  |

## 8. Detailed Rules

### 8.1 Type Resolution Priority

When parsing a value, the following order is attempted:

1. **Blank** → `null`
2. **Boolean** → `true` or `false` (lowercase only)
3. **Integer** → matches `-?[0-9]+`
4. **Float** → matches `-?[0-9]+\.[0-9]+`
5. **String** → everything else

Examples:

| Value    | Parsed as          |
| -------- | ------------------ |
| *(blank)*| `null`             |
| `true`   | boolean `true`     |
| `123`    | integer `123`      |
| `3.14`   | float `3.14`       |
| `hello`  | string `"hello"`   |
| `123abc` | string `"123abc"`  |

### 8.2 Null vs Empty String

| ASON                               | Parse result             |
| ---------------------------------- | ------------------------ |
| `{name@str,age@int}:(Alice,)`      | `age = null`             |
| `{name@str,age@int}:(Alice,"")`    | `age = ""` (empty string)|

Example:

```text
{name@str,bio@str}:(Alice,)         /* bio = null */
{name@str,bio@str}:(Alice,"")       /* bio = "" (empty string) */
```

### 8.3 Top-Level Structure Detection

There are **three top-level forms**, determined by the first character(s):

| First char | Type                         | Example                                     |
| ---------- | ---------------------------- | ------------------------------------------- |
| `{`        | Single object with schema    | `{name@str,age@int}:(Alice,30)`             |
| `[{`       | Array of objects with schema | `[{id@int,name@str}]:(1,Alice),(2,Bob)`     |
| `[`        | Plain array                  | `[1,2,3]`                                   |
| Other      | Bare value (type inferred)   | `42`, `true`, `hello`                       |

**Key rules:**

- After `{schema}:` there can be **exactly one** `(...)` data tuple (single object).
- After `[{schema}]:` there can be **multiple** `(...)` data tuples, comma-separated (array of objects).
- The parser determines format from the first character — `{` vs `[` — no need for separate `_vec`-style APIs.
- A bare `()` at the top level is forbidden; `()` may only appear after a schema.

### 8.4 Field Name Rules

- **Allowed characters**: `a-z`, `A-Z`, `0-9`, `_`
- **May start with a digit**: `1st`, `2name` are valid
- **Forbidden characters**: `,`, `{`, `}`, `[`, `]`, `(`, `)`, `:`

### 8.5 Whitespace Handling

| Location     | Rule                                  |
| ------------ | ------------------------------------- |
| In schema    | All whitespace ignored                |
| In data      | Whitespace preserved (part of string) |
| After comma  | Leading whitespace ignored            |

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

| Input   | Parsed as         | Note                        |
| ------- | ----------------- | --------------------------- |
| `-123`  | integer `-123`    | ✓ Correct                   |
| `- 123` | string `"- 123"`  | ✗ Space → parsed as string  |
| `-3.14` | float `-3.14`     | ✓ Correct                   |
| `-0`    | integer `0`       | ✓ Special case              |

### 8.8 Data Alignment Rule (Strict Mode)

**The number of data items must exactly match the number of schema fields.**

| Schema                | Data         | Result                     |
| --------------------- | ------------ | -------------------------- |
| `{a@int,b@int,c@int}` | `(1,2,3)`    | ✓ Correct                  |
| `{a@int,b@int,c@int}` | `(1,2)`      | ✗ Error: missing field     |
| `{a@int,b@int,c@int}` | `(1,2,3,4)`  | ✗ Error: too many fields   |
| `{a@int,b@int,c@int}` | `(1,,3)`     | ✓ Correct: `b = null`      |

**Rationale**: ASON is a position-sensitive format. A field-count mismatch causes data misalignment and must be reported as a parse error.

### 8.9 Common Error Examples

| Incorrect                     | Reason                | Correct                                           |
| ----------------------------- | --------------------- | ------------------------------------------------- |
| `{a@int,b@int}:(1,2,3)`       | Too many values       | `{a@int,b@int,c@int}:(1,2,3)`                     |
| `{a@int,b@int}:(1)`           | Missing field (no null)| `{a@int,b@int}:(1,)`                             |
| `{a@int,b@int,c@int}:(1,2)`   | Insufficient data     | `{a@int,b@int,c@int}:(1,2,)`                      |
| `(1,2,3)`                     | Bare tuple needs schema| `{a@int,b@int,c@int}:(1,2,3)`                   |
| `{a@int,b@int}`               | Schema with no data   | `{a@int,b@int}:()` or `{a@int,b@int}:(1,2)`      |
| `{a@int,b@int}[1,2]`          | Missing colon after schema | `{a@int,b@int}:(1,2)`                       |

### 8.10 Trailing Commas

To facilitate version control diffs and LLM generation, ASON **permits** trailing commas in arrays and object data. Parsers must silently ignore them — they must **not** be interpreted as `null`.

| Input           | Parsed result         | Note                                               |
| --------------- | --------------------- | -------------------------------------------------- |
| `[1, 2, 3,]`   | `[1, 2, 3]`           | Trailing comma allowed                             |
| `(Alice, 30,)` | `("Alice", 30)`       | Trailing comma allowed (schema has 2 fields)       |
| `(Alice, 30,,)`| `("Alice", 30, null)` | Consecutive commas → null; final comma ignored     |

---

## 9. Implementation Notes

### 9.1 Parser Implementation Highlights

1. **Non-greedy matching**: When parsing `plain_str`, delimiters `,()[]` take priority over string content.
2. **Lookahead**: When encountering a potential delimiter, first check whether it is preceded by an escape character.
3. **Comment stripping**: Recommended to remove all `/* */` comments during the lexing phase.
4. **Streaming parse**: After parsing the schema, build a field index table; fill data fields by position.

### 9.2 Error Handling

| Error Type         | Example                  | Handling                       |
| ------------------ | ------------------------ | ------------------------------ |
| Field count mismatch | `{a@int,b@int}:(1,2,3)` | Throw error with location info |
| Unclosed quote     | `("hello)`               | Throw error                    |
| Unclosed bracket   | `{a@int,b@int}:(1,2`     | Throw error                    |
| Unclosed comment   | `/* comment`             | Throw error                    |
| Invalid escape     | `\x`                     | Throw error or keep verbatim   |

---

## 10. Complete Grammar BNF (Simplified)

```bnf
ason        ::= object_expr | array_expr | array | value

object_expr ::= schema ":" object
array_expr  ::= "[" schema "]" ":" object_list
schema      ::= "{" field_list "}"
field_list  ::= field ("," field)*
field       ::= identifier type_hint? | identifier "@" schema
type_hint   ::= "@" type_def
type_def    ::= base_type | array_type
base_type   ::= "int" | "float" | "str" | "bool"
array_type  ::= "[" type_def "]" | "[" schema "]"
identifier  ::= [a-zA-Z0-9_]+

object_list ::= object ("," object)*
object      ::= "(" value_list ")"
value_list  ::= element? ("," element?)*
element     ::= value | array

array       ::= "[" array_items? "]"
array_items ::= array_item ("," array_item)*
array_item  ::= value | object | array

value       ::= boolean | number | quoted_str | plain_str | null
boolean     ::= "true" | "false"
number      ::= integer | float
integer     ::= "-"?[0-9]+
float       ::= "-"?[0-9]+"."[0-9]+

quoted_str  ::= '"' (escaped_char | [^"\\])* '"'
plain_str   ::= (plain_char | escaped_char)+
plain_char  ::= [^,()[\]<>:"\\]
escaped_char::= "\\" | "\," | "\(" | "\)" | "\[" | "\]" | "\<" | "\>" | "\:" | "\"" | "\n" | "\t" | "\\u" [0-9a-fA-F]{4}

null        ::= (empty)
comment     ::= "/*" (any_char)* "*/"
```

**Notes:**

- Comments `/* ... */` may only appear at the beginning or end of a line; ignored during parsing.
- `plain_str` values are automatically trimmed of leading/trailing whitespace after parsing.
- `quoted_str` values are preserved verbatim (including spaces).
- Trailing commas are permitted and silently ignored.

---

## 11. ASON Binary Format Specification

In addition to the human-readable text format, ASON defines a compact binary wire format for high-performance serialization/deserialization.

### 11.1 Design Principles

- **Zero-copy decoding**: String fields return slice references directly into the input buffer — no new memory allocated.
- **Fixed-width scalars**: All numeric types use fixed-size little-endian encoding; no variable-length decoding required.
- **No schema header**: The binary format contains no field names — it relies entirely on compile-time type information for positional decoding.
- **Length-prefixed**: Variable-length data (strings, arrays) is preceded by a `u32` little-endian length.

### 11.2 Type Encoding Rules

| Type           | Bytes       | Encoding                                    |
| -------------- | ----------- | ------------------------------------------- |
| `bool`         | 1           | `0x00` = false, `0x01` = true               |
| `i8` / `u8`    | 1           | Raw byte                                    |
| `i16` / `u16`  | 2           | Little-endian                               |
| `i32` / `u32`  | 4           | Little-endian                               |
| `i64` / `u64`  | 8           | Little-endian                               |
| `f32`          | 4           | IEEE 754 bitcast, little-endian             |
| `f64`          | 8           | IEEE 754 bitcast, little-endian             |
| `str`          | 4 + N       | `u32 LE` byte length + N bytes UTF-8        |
| `Option<T>`    | 1 or 1+sizeof(T) | `u8` tag (`0x00` = null, `0x01` = some) + payload |
| `Array<T>`     | 4 + N×sizeof(T) | `u32 LE` element count + N encoded elements |
| `struct`       | Σ fields    | Fields encoded in declaration order; no padding or alignment |

### 11.3 Single Struct vs Struct Array

**Single struct**: Fields encoded in declaration order directly, no wrapper.

```text
struct User { id: i64, name: string, active: bool }

Encoding: [i64 LE][u32 LE + UTF-8 bytes][u8]
           8 bytes  4 + N bytes            1 byte
```

**Struct array**: Prefixed with `u32 LE` element count, followed by each element's encoding.

```text
Array<User>

Encoding: [u32 LE count][User₁][User₂]...[Userₙ]
           4 bytes        elements in sequence
```

> **Important**: Single structs and struct arrays use different binary layouts. A single struct has no count prefix; a struct array does. The decoder must know whether the target type is a single instance or an array.

### 11.4 Encoding Examples

```text
struct Point { x: f32, y: f32 }
Value: { x: 1.0, y: 2.0 }

Binary (8 bytes):
  00 00 80 3F   ← f32 LE: 1.0
  00 00 00 40   ← f32 LE: 2.0
```

```text
struct User { id: i64, name: string, active: bool }
Value: { id: 42, name: "Alice", active: true }

Binary (18 bytes):
  2A 00 00 00 00 00 00 00   ← i64 LE: 42
  05 00 00 00               ← u32 LE: string length 5
  41 6C 69 63 65            ← UTF-8: "Alice"
  01                        ← bool: true
```

```text
Array<Point> = [{ x: 1.0, y: 2.0 }, { x: 3.0, y: 4.0 }]

Binary (20 bytes):
  02 00 00 00               ← u32 LE: count = 2
  00 00 80 3F 00 00 00 40   ← Point₁: (1.0, 2.0)
  00 00 40 40 00 00 80 40   ← Point₂: (3.0, 4.0)
```

### 11.5 Correspondence with Text Format

| Text format                       | Binary format                           |
| --------------------------------- | --------------------------------------- |
| `{schema}:(data)` single object   | Fields encoded in order, no wrapper     |
| `[{schema}]:(d1),(d2),...` array  | `u32 LE` count + element sequence       |
| `[v1,v2,v3]` plain array         | `u32 LE` count + element sequence       |
| `true` / `false`                  | Single byte `0x01` / `0x00`             |
| Null / Option null                | Single byte `0x00`                      |
| Option some(v)                    | `0x01` + value encoding                 |

### 11.6 Performance Characteristics

| Characteristic      | Text Format                    | Binary Format                         |
| ------------------- | ------------------------------ | ------------------------------------- |
| Human readable      | ✓                              | ✗                                     |
| Encode speed        | Fast (1.5–2× JSON)             | Extremely fast (6–8× JSON)            |
| Decode speed        | Fast (1.5–2.3× JSON)           | Extremely fast (17–47× JSON)          |
| Size                | Compact (50–55% smaller than JSON) | More compact (39–55% smaller)     |
| Zero-copy decode    | ✗                              | ✓ (strings reference input buffer)    |
| Use cases           | API communication, LLM, debug  | RPC, IPC, high-frequency data pipelines |

### 11.7 Byte Order Conventions

- All multi-byte integers and floats use **little-endian** byte order.
- String content uses **UTF-8** encoding.
- No alignment padding.
- No metadata or magic numbers.

---

## 12. LLM Best Practices & Benchmarks

ASON is optimized for interaction with large language models. This section provides prompt templates, accuracy benchmarks, and corrections for common LLM errors.

### 12.1 Core Design Advantages

| Dimension         | ASON Advantage                                           |
| ----------------- | -------------------------------------------------------- |
| **Token efficiency** | 30–70% fewer tokens than JSON → more context headroom |
| **Structure clarity** | Schema declared once → lower model comprehension cost |
| **Generation pattern** | Row-oriented table format matches LLM training data   |
| **Streaming** | Native row-by-row streaming, no need to buffer full output |

### 12.2 Recommended System Prompt

```text
You are a data format conversion expert. Output data in ASON format (Array-Schema Object Notation).

Core ASON rules:
1. Single object: `{field1@type, field2@type, ...}:(val1, val2, ...)`
2. Array of objects: `[{field1@type, field2@type, ...}]:(val1, val2, ...),(val3, val4, ...)`
3. Type annotations are optional: `{field1@int, field2@str, ...}`
4. Supported types: `int`, `float`, `str`, `bool`, arrays, nested objects
5. Array fields: `name@[str]` is a string array with value `[item1, item2, ...]`; array-of-objects is `users@[{id@int}]`

You must follow:
- Single objects use `{schema}:` prefix; arrays of objects use `[{schema}]:` prefix
- Number of data items = number of schema fields (strict alignment)
- Strings generally need no quotes (unless they contain special characters)
- Null values are represented by blank content (empty between commas)
- Nested objects use parentheses: outer@{inner@type}:(val1,(nested_val))

Output ASON only, no additional explanation.
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

| Model              | Format Correct | Type Correct | Recommended  |
| ------------------ | -------------- | ------------ | ------------ |
| GPT-4 Turbo        | 99%+           | 97%+         | ⭐⭐⭐⭐⭐  |
| Claude 3.5 Sonnet  | 98%+           | 96%+         | ⭐⭐⭐⭐⭐  |
| GPT-4o             | 96%+           | 93%+         | ⭐⭐⭐⭐    |
| Llama-2 70B        | 90%            | 85%          | ⭐⭐⭐      |

### 12.5 Common Errors and Fixes

#### Error 1: Field Count Mismatch

```ason
❌ {id@int,name@str}:(1,Alice,true)
✅ {id@int,name@str,active@bool}:(1,Alice,true)
```

#### Error 2: Wrong Bracket Type for Nesting

```ason
❌ {user@{id@int,name@str}}:({1,Alice})      /* curly braces used */
✅ {user@{id@int,name@str}}:((1,Alice))      /* nested data uses parentheses */
```

#### Error 3: Wrong Bracket Type for Arrays

```ason
❌ {tags@[str]}:({python,rust})         /* curly braces */
✅ {tags@[str]}:([python,rust])         /* arrays use square brackets */
```

### 12.6 Token Cost Comparison

Test data: 500 user records, 8 fields each

| Format | Token count | Cost (GPT-4) | Savings  |
| ------ | ----------- | ------------ | -------- |
| JSON   | 12,450      | $0.55        | —        |
| ASON   | 4,280       | $0.19        | **65%**  |

### 12.7 Integration Workflow

```
Request → Build prompt (system role + few-shot + data)
  ↓
Call LLM (temperature = 0.1–0.3)
  ↓
Validate format (field count, alignment, types)
  ├─ ✓ Pass → return ASON
  └─ ✗ Fail → retry (max 3 times)
```

8. 📋 v2.0: Implement serde support
9. 📋 v2.0: Schema references and aliases
10. 📋 v2.0: Full data validation framework

---

## 13. Appendix A: Type System Reference

### A.1 Type Annotation Quick Reference

| Scenario                   | Schema                 | Data           | Note                  |
| -------------------------- | ---------------------- | -------------- | --------------------- |
| No annotations (v1.3 compat) | `{id,name}`          | `(1,Alice)`    | Implicit inference    |
| Explicit string            | `{name@str}`           | `(Alice)`      | Clear type            |
| Integer                    | `{id@int}`             | `(1)`          | Integer type          |
| Float                      | `{score@float}`        | `(95.5)`       | Floating-point        |
| Boolean                    | `{active@bool}`        | `(true)`       | Boolean               |
| Array                      | `{tags@[str]}`         | `([a,b,c])`    | Array type            |
| Nested                     | `{user@{id@int}}`      | `((1))`        | Nested object         |
| Mixed                      | `{id@int,bio@str}`     | `(1,Bio)`      | Partial annotations   |

### A.2 Type Assertion Examples

> **Note**: Strict scalar-hint assertions are an optional spec-level extension. In current implementations, parsers may skip these scalar hints and leave type validation to the target struct's type system at the serde layer. The examples below illustrate what a parser with strict checking enabled might do:

```ason
// Correctly typed data
[{id@int, score@float, pass@bool, name@str}]:
  (1, 95.5, true, Alice),      ✅ All types match
  (2, 87.0, false, Bob),       ✅ All types match
  (3, 76, true, "Carol Sue")   ✅ 76 auto-promoted to 76.0

// Type mismatches (parser may optionally enforce strict checking)
  ("1", 95.5, true, Alice),    ⚠️ id should be int, not string
  (1, "95.5", true, Alice)     ⚠️ score should be float, not string
```

---

## 14. Appendix B: Comparison with Other Formats

### B.1 ASON vs JSON vs CSV

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

```ason
// ASON (35 tokens, with type information)
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

```ason
// ASON — more compact
{employees@[{id@int, name@str, dept@{name@str, budget@int}}]}:
  ([(1, Alice, (Eng, 500000))])
```

---

## 15. Appendix C: Implementation Checklist

### C.1 Parser Checklist

- [ ] Lexer: recognize all symbols (`{`, `}`, `(`, `)`, `[`, `]`, `:`, `,`)
- [ ] Comment handling: support `/* */` block comments
- [ ] String parsing: support both quoted and unquoted forms
- [ ] Escape rules: handle `\,`, `\"`, `\\`, etc.
- [ ] Type annotation parsing: recognize `@int`, `@str`, etc.
- [ ] Schema parsing: recursively parse nested structures
- [ ] Alignment check: verify data field count matches schema field count
- [ ] Error reporting: location-aware error messages

### C.2 Serializer Checklist

- [ ] Value → ASON: serialize in-memory objects to text
- [ ] Indentation and alignment: optional column alignment
- [ ] Type annotation generation: infer and emit types from data
- [ ] Quoting logic: determine when quotes are required vs optional
- [ ] Performance optimization: streaming serialization for large datasets

### C.3 Validation Tool Checklist

- [ ] Format check: ASON syntactic correctness
- [ ] Type check: field types match data (optional strict mode)
- [ ] Data alignment: field count consistency
- [ ] Schema validation: field name legality
- [ ] LLM output validation: catch common LLM generation errors

---

## 16. Appendix D: FAQ

### Q1: Why doesn't ASON use a human-friendly format similar to JSON?

**A**: ASON is optimized for LLMs, which requires:

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

### Q3: Does ASON support large files?

**A**: Yes. The row-oriented format naturally supports:

- Streaming without loading the entire file
- Each row is independent, suitable for distributed processing
- For very large datasets, the ASONL format is recommended (similar to JSONL)

### Q4: How does ASON compare to MessagePack and Protocol Buffers?

**A**:

- **MessagePack**: Binary format, more compact but not human-readable
- **Protocol Buffers**: Requires `.proto` definition files, higher complexity
- **ASON**: Text format for LLM use, balancing token efficiency with readability

### Q5: What are the most common errors when LLMs generate ASON?

**A**: See Section 12.5 for details. The three most common:

1. Field count mismatch
2. Wrong bracket type for nesting
3. Wrong bracket type for arrays

### Q6: Is ASON's parse performance really faster than JSON?

**A**: Yes. When processing arrays of objects (list data), ASON's parse performance far exceeds JSON, due to **schema-driven parsing**:

1. **Zero key-hashing**: No need to repeatedly read key strings and compute hashes as JSON does.
2. **Excellent branch prediction**: The parser knows the type of the next value in advance, maximizing CPU instruction pipeline efficiency.
3. **Minimal memory usage**: All data rows share a single schema reference — no per-row key memory allocation.

---

**Document version**: v1.4.0  
**Last updated**: 2026-02-20  
**License**: MIT  
**GitHub**: https://github.com/ason-lab/ason
