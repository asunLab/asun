# Schema & Data

The most important concept in ASON is the separation of **schema** (structure) from **data** (values).

## Schema Syntax

A schema is delimited by `{` and `}` and contains a comma-separated list of field definitions:

```ason
{name, age, active}
```

Fields can optionally include type annotations:

```ason
{name:str, age:int, active:bool}
```

Type annotations are optional hints for humans and tooling — they do not change parsing behavior when the target type is already known from the host language's type system.

## Data Syntax

For a slice (list of structs), wrap the schema in `[` `]`:

```ason
[{name:str, age:int}]:
  (Alice, 30),
  (Bob,   25)
```

For a single struct (not a list), no wrapping needed:

```ason
{name, age}:(Alice, 30)
```

## Nested Schemas

Schemas can be nested to represent nested objects:

```ason
[{id:int, address:{city:str, zip:str}}]:
  (1, (Berlin, 10115)),
  (2, (Paris,  75001))
```

The inner schema `{city:str, zip:str}` is substituted by an inner tuple `(Berlin, 10115)`.

## Array Fields

Fields that contain lists use `[type]` notation:

```ason
[{id:int, tags:[str]}]:
  (1, [rust, go]),
  (2, [python, c++])
```

## Optional / Null Values

An empty slot between commas represents `null` / `None`:

```ason
[{id:int, name:str, score:float}]:
  (1, Alice, 9.5),
  (2, Bob,       )
```

`Bob`'s score is `None` (empty slot at the end).

## Full Grammar Summary

```
document    = single | slice
single      = schema ":" tuple
slice       = "[" schema "]" ":" rows
schema      = "{" fields "}"
fields      = field ("," field)*
field       = name (":" type)?
type        = "int" | "float" | "str" | "bool" | schema | "[" type "]"
rows        = tuple ("," tuple)*
tuple       = "(" values ")"
values      = value ("," value)*
value       = scalar | tuple | "[" values "]" | ""
scalar      = unquoted_string | quoted_string | number | bool | null
```

See the complete [Syntax Reference](/reference/syntax) for escape rules, whitespace handling, and edge cases.
