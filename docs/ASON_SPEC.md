# ASON Format Specification v1.0

> **ASON** = Array-Schema Object Notation

## 1. Design Philosophy

- **Schema-Data Separation**: Declare structure first, then fill in data
- **Minimize Redundancy**: Field names declared once, no repeated keys in data
- **Single-Line Friendly**: All formats naturally support compression to one line
- **Token Efficient**: Designed specifically to reduce LLM token consumption

## 2. Basic Syntax

```
schema:data
```

- `schema`: Defines data structure (field names, nesting relationships)
- `data`: Pure data values, arranged in schema order

## 3. Basic Types

| Type | Example | Description |
|------|---------|-------------|
| Integer | `42`, `-100` | Numbers without decimal point |
| Float | `3.14`, `-0.5` | Numbers with decimal point |
| String | `hello`, `"world"` | Optional quotes |
| Boolean | `true`, `false` | Lowercase |
| Null | _(blank)_ | Empty field position |
| Empty String | `""` | Explicit empty string |

### 3.1 String Rules

ASON supports two string forms:

| Form | Example | Behavior |
|------|---------|----------|
| Unquoted | `hello world` | Auto-trim leading/trailing spaces |
| Quoted | `" hello "` | Preserved as-is (including spaces) |

**Use quotes when:**
- Preserving leading/trailing spaces: `" hello "`
- Leading zeros: `"001234"` (zip codes)
- Forcing string type: `"true"`, `"123"`
- Empty string: `""`

## 4. Escape Rules

Escape special characters in string values:

| Character | Escape Form | Description |
|-----------|-------------|-------------|
| `,` | `\,` | Comma |
| `(` | `\(` | Left parenthesis |
| `)` | `\)` | Right parenthesis |
| `[` | `\[` | Left bracket |
| `]` | `\]` | Right bracket |
| `"` | `\"` | Double quote |
| `\` | `\\` | Backslash |
| Newline | `\n` | Newline character |
| Tab | `\t` | Tab character |

**Notes:**
- In unquoted strings, must escape `,()[]`
- In quoted strings, only need to escape `"` and `\`

## 5. Comments

Block comments supported using `/* */` syntax:

```
/* This is a user list */
{name,age}:(Alice,30),(Bob,25)
```

Multi-line comments:
```
/*
  User data
  Last updated: 2024-01-01
*/
{name,age}:
  (Alice,30),
  (Bob,25)
```

Comments can appear anywhere (before, within, or after schema and data):
```
{name /* username */, age}:(Alice, /* age */ 30)
```

## 6. Syntax Rules

### 6.1 Single Object

```
{name,age}:(Alice,30)
```
→ `{name: "Alice", age: 30}`

### 6.2 Object Array (Multiple Records with Same Structure)

```
{name,age}:(Alice,30),(Bob,25),(Charlie,35)
```
→ Array of 3 objects

### 6.3 Null/Optional Fields

```
{name,age,email}:(Alice,30,)
```
→ `{name: "Alice", age: 30, email: null}`

### 6.4 Nested Objects

```
{name,addr{city,zip}}:(Alice,(NYC,10001))
```
→ `{name: "Alice", addr: {city: "NYC", zip: 10001}}`

### 6.5 Object with Simple Array

```
{name,scores[]}:(Alice,[90,85,92])
```
→ `{name: "Alice", scores: [90, 85, 92]}`

### 6.6 Object with Object Array

```
{team,users[{id,name}]}:(Dev,[(1,Alice),(2,Bob)])
```
→ `{team: "Dev", users: [{id: 1, name: "Alice"}, {id: 2, name: "Bob"}]}`

### 6.7 Plain Array (No Schema)

```
[1,2,3]
```
→ `[1, 2, 3]`

### 6.8 Nested Arrays

```
[[1,2],[3,4]]
```
→ `[[1, 2], [3, 4]]`

### 6.9 Empty Array

```
[]
```

### 6.10 Empty Object

```
()
```

### 6.11 Mixed Type Array

```
[1,hello,true,3.14]
```
→ `[1, "hello", true, 3.14]`

### 6.12 Complex Nested Example

```
{company,employees[{id,name,skills[]}],active}:(ACME,[(1,Alice,[rust,go]),(2,Bob,[python])],true)
```

**Parsed result:**
- `company` = `"ACME"`
- `employees` = Array containing 2 objects
  - `{id: 1, name: "Alice", skills: ["rust", "go"]}`
  - `{id: 2, name: "Bob", skills: ["python"]}`
- `active` = `true`

### 6.13 String Processing Example

```
{name,city,zip,note}:
  (Alice, New York, "001234", hello world),
  (Bob, "  Los Angeles  ", 90210, "say \"hi\"")
```

**Parsed result:**
| Field | Alice row | Bob row |
|-------|-----------|---------|
| name | `"Alice"` (trimmed) | `"Bob"` (trimmed) |
| city | `"New York"` (trimmed) | `"  Los Angeles  "` (quoted, preserved) |
| zip | `"001234"` (quoted, leading zeros) | `90210` (integer) |
| note | `"hello world"` (trimmed) | `"say \"hi\""` (quoted + escaped) |

## 7. Syntax Quick Reference

| Element | Schema Syntax | Data Syntax |
|---------|---------------|-------------|
| Object | `{field1,field2}` | `(val1,val2)` |
| Simple Array | `[]` | `[v1,v2,v3]` |
| Object Array | `[{f1,f2}]` | `[(v1,v2),(v3,v4)]` |
| Nested Object | `{f1,f2{f3,f4}}` | `(v1,(v3,v4))` |
| Null | - | _(blank)_ |
| Empty Array | - | `[]` |
| Empty Object | - | `()` |

## 8. Detailed Rules

### 8.1 Type Parsing Priority

When parsing values, try in this order:

1. **Blank** → `null`
2. **Boolean** → `true` or `false` (lowercase)
3. **Integer** → Matches `-?[0-9]+`
4. **Float** → Matches `-?[0-9]+\.[0-9]+`
5. **String** → All other cases

Examples:
| Value | Parsed as |
|-------|-----------|
| _(blank)_ | `null` |
| `true` | Boolean `true` |
| `123` | Integer `123` |
| `3.14` | Float `3.14` |
| `hello` | String `"hello"` |
| `123abc` | String `"123abc"` |

### 8.2 Null vs Empty String

| ASON Syntax | Parsed Result |
|-------------|---------------|
| `{name,age}:(Alice,)` | `age = null` |
| `{name,age}:(Alice,"")` | `age = ""` (empty string) |

Examples:
```
{name,bio}:(Alice,)         /* bio = null */
{name,bio}:(Alice,"")       /* bio = "" (empty string) */
```

### 8.3 Top-Level Structure Recognition

Top level has only **three forms**, determined by first character:

| First Char | Type | Example |
|------------|------|---------|
| `{` | Object/Object array with Schema | `{name,age}:(Alice,30)` |
| `[` | Plain array | `[1,2,3]` |
| Other | Plain value (auto-inferred type) | `42`, `true`, `hello` |

**Note**: Top-level `()` is forbidden; `()` can only appear after schema in data section.

### 8.4 Field Name Rules

- **Allowed characters**: `a-z`, `A-Z`, `0-9`, `_`
- **Can start with digit**: `1st`, `2name` are valid
- **Forbidden characters**: `,`, `{`, `}`, `[`, `]`, `(`, `)`, `:`

### 8.5 Whitespace Handling Rules

| Location | Rule |
|----------|------|
| In Schema | Ignore all whitespace |
| In Data | Preserve whitespace (as part of string) |
| After comma | Ignore leading whitespace |

Example:
```
{name, age}:(Alice Smith, 30)
```
Equivalent to:
```
{name,age}:(Alice Smith,30)
```
→ `{"name": "Alice Smith", "age": 30}`

### 8.6 Multi-Line Format

Supports newlines and indentation, treated as whitespace during parsing:

```
{name,age}:
  (Alice,30),
  (Bob,25),
  (Charlie,35)
```

Equivalent to:
```
{name,age}:(Alice,30),(Bob,25),(Charlie,35)
```

### 8.7 Negative Number Rules

Minus sign `-` must immediately precede digit, no space allowed:

| Syntax | Parsed as |
|--------|-----------|
| `-123` | Integer `-123` ✓ |
| `- 123` | String `"- 123"` ✗ |
| `-3.14` | Float `-3.14` ✓ |

### 8.8 Data Alignment Rules (Strict Mode)

**Number of data items must exactly match number of schema fields.**

| Schema | Data | Result |
|--------|------|--------|
| `{a,b,c}` | `(1,2,3)` | ✓ Correct |
| `{a,b,c}` | `(1,2)` | ✗ Error: missing field |
| `{a,b,c}` | `(1,2,3,4)` | ✗ Error: too many fields |
| `{a,b,c}` | `(1,,3)` | ✓ Correct: b = null |

**Reason**: ASON is position-sensitive format; field count mismatch causes data misalignment, must error during parsing.

## 9. Complete BNF Grammar (Simplified)

```bnf
ason        ::= object_expr | array | value

object_expr ::= schema ":" object_list
schema      ::= "{" field_list "}"
field_list  ::= field ("," field)*
field       ::= identifier | identifier schema | identifier "[]" | identifier "[" schema "]"
identifier  ::= [a-zA-Z0-9_]+

object_list ::= object ("," object)*
object      ::= "(" value_list ")"
value_list  ::= value? ("," value?)*

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
plain_char  ::= [^,()[\]"\\]                   /* excludes delimiters and quotes */
escaped_char::= "\\" | "\," | "\(" | "\)" | "\[" | "\]" | "\"" | "\n" | "\t"

null        ::= (empty)
comment     ::= "/*" (any_char)* "*/"
```

**Notes:**
- Comments `/* ... */` can appear anywhere, ignored during parsing
- `plain_str` is auto-trimmed after parsing
- `quoted_str` is preserved as-is (including spaces)

## 10. Implementation Notes

### 10.1 Parser Implementation Points

1. **Non-greedy matching**: When parsing `plain_str`, delimiters `,()[]` take priority over string content
2. **Lookahead**: When encountering potential delimiter, first check if it's an escape sequence
3. **Comment stripping**: Recommend removing all `/* */` comments during lexical analysis
4. **Streaming parsing**: Build Access Map after schema parsing, fill data section by index

### 10.2 Error Handling

| Error Type | Example | Handling |
|------------|---------|----------|
| Field count mismatch | `{a,b}:(1,2,3)` | Throw error, report position |
| Unclosed quote | `("hello)` | Throw error |
| Unclosed parenthesis | `{a,b}:(1,2` | Throw error |
| Unclosed comment | `/* comment` | Throw error |
| Invalid escape | `\x` | Throw error or preserve as-is |

## 11. Future Features (v2.0 Consideration)

The following features are not in v1.0 scope, but worth considering for future:

### 11.1 Schema References

```
$user = {id,name};
$book = {title,author};

$user:(1,Alice),(2,Bob)
$book:(ASON Guide,Alice)
```

### 11.2 Type Annotations

```
{name:str, age:int, zip:str}:(Alice,30,"001234")
```

### 11.3 Data Validation

```
{name,age}:(Alice,30),(Bob,25) /* @count:2 */
```

## 12. Next Steps

1. ✅ Confirm format specification
2. ✅ Implement Rust data structures (AST)
3. ✅ Implement Lexer
4. ✅ Implement Parser
5. ✅ Implement Serializer (Value → ASON string)
6. ✅ Implement serde support

