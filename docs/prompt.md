# ASON (Array-Schema Object Notation) Vibe Coding Prompt Guide

If you are an AI generating or processing ASON data, you MUST strictly adhere to the following architectural rules and syntax. ASON is designed for high-performance, token-efficient serialization, focusing on a clean separation of schema and data.

## 1. The Golden Rule: `Header : Body` Separation
ASON uses a colon `:` to physically isolate the Schema (Header) from the Data (Body). 
- **Everything before `:`** is the structure blueprint.
- **Everything after `:`** is pure, key-less data.
- **Example**: `{name@str, age@int}: (Alice, 30)`

## 2. Structural Harmony: `{}` vs `()`
- **`{}` (Object Schema)**: Used ONLY in the Header to define a set of field names.
- **`()` (Data Tuple)**: Used ONLY in the Body to hold values. Data must be strictly positional. ASON data bodies *never* use `{}` because they are positional tuples, not key-value objects. This prevents key collisions natively.

✅ **Correct**: `{name, age} : (Bob, 25)`
❌ **Incorrect**: `{name, age} : {Bob, 25}` (Never use curly braces for data)
❌ **Incorrect**: `{name, age} : (name: Bob, age: 25)` (Never repeat keys in data)

## 3. Top-level Structures
- **Single Object**: Starts with `{`. Example: `{id@int} : (1)`
- **Array of Objects**: Starts with `[{`. Example: `[{id@int}] : (1), (2)`
- **Plain Array**: `[1, 2, 3]`

## 4. The `@` Field Binding Marker
`@` is the field binding marker between a field name and its following schema/type description.
For scalar types (int, str, bool), the `@type` portion is an optional hint (`name@str` and `name` are layout-equivalent).
**HOWEVER**, for nested objects and arrays, the `@` followed by `{}` or `[]` is a required structural binding. You **MUST NOT** omit it.
- **Nested Object**: `dept@{title@str, level@int}` or `dept@{title, level}` (Notice `@` and `{}` must remain even without scalar types).
- **Array Field**: `tags@[str]` or `tags@[]`
- **Map/Dict Representation**: Maps are represented as arrays of key-value tuples: `attrs@[{key@str, value@int}] : ( [(age,30), (score,95)] )`

✅ **Correct Nested**: `{user, metadata@{created_at}} : (Alice, (2024))`
❌ **Fatal Error**: `{user, metadata} : (Alice, (2024))` (Parser will fail because it wasn't told `metadata` expects a nested tuple).

## 5. Value and Edge Case Rules
- Strings do not need quotes unless they need to preserve leading/trailing spaces, or contain structural characters like `, ( ) [ ] \ :`
- Boolean values must be lowercase (`true`, `false`).
- Null / Optional fields are represented by empty space between commas: `(Alice, , 30)` means the middle field is `null`.
- Trailing commas are allowed and ignored by the parser: `(1, 2,)` = `(1, 2)`.

## Summary Examples for AI Generation

**1. Multiple records with nested structures (With Scalar Hints)**
```ason
[{id@int, profile@{name@str, active@bool}}]:
  (1, (Alice, true)),
  (2, (Bob, false))
```

**2. Configuration object with array of objects (Minimal / Without Scalar Hints)**
```ason
{service, rules@[{path, secure}]}:
  (api, [(/auth, true), (/health, false)])
```

**CRITICAL REMINDER**: Never mix JSON habits into ASON. There are no keys in the data section. Strictly use `()` for data tuples, and always bind nested fields using `@{...}` or `@[...]` in the schema header.
