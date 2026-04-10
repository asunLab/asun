# Cross-Language Compatibility Matrix

This matrix defines the canonical text-format compatibility cases that
schema-driven language bindings should support when decoding ASON into a target struct/class
with fewer or reordered fields than the source payload.

The same cases should be reused across Go, Rust, Java, and any future
schema-driven bindings. Dynamic bindings such as JS/Python can still decode the raw payload,
but they are not the primary target for the "drop extra fields into a typed
target" scenarios below.

## Conventions

- `T` suffix: schema input with scalar hints
- `U` suffix: schema input without scalar hints
- "Target default" means the language's normal zero/default/null value for a
  missing field

## Core Cases

| Case ID | Category | Input shape | Target shape | Expected outcome |
| --- | --- | --- | --- | --- |
| `A1-T` | Exact match | single with scalar hints | exact fields | decode succeeds |
| `A1-U` | Exact match | single without scalar hints | exact fields | decode succeeds |
| `A2-T` | Source has extra field | single with scalar hints | fewer fields | extra field ignored |
| `A2-U` | Source has extra field | single without scalar hints | fewer fields | extra field ignored |
| `A3-T` | Target has extra field | single with scalar hints | more fields | missing field gets default |
| `A3-U` | Target has extra field | single without scalar hints | more fields | missing field gets default |
| `A4-T` | Field reorder | single with scalar hints | reordered fields | fields matched by name |
| `A4-U` | Field reorder | single without scalar hints | reordered fields | fields matched by name |
| `A5-T` | Source has extra field (vec) | vec with scalar hints | fewer fields | extra field ignored in each row |
| `A5-U` | Source has extra field (vec) | vec without scalar hints | fewer fields | extra field ignored in each row |
| `A6-T` | Target has extra field (vec) | vec with scalar hints | more fields | missing field gets default |
| `A6-U` | Target has extra field (vec) | vec without scalar hints | more fields | missing field gets default |
| `P1-T` | Partial overlap | single with scalar hints | target keeps a non-contiguous subset | only overlapping fields decoded |
| `P1-U` | Partial overlap | single without scalar hints | target keeps a non-contiguous subset | only overlapping fields decoded |
| `P2-T` | No overlap | single with scalar hints | target shares no field names | target fields stay default |
| `P2-U` | No overlap | single without scalar hints | target shares no field names | target fields stay default |

## Recommended Nested / Optional Follow-ups

| Case ID | Category | Input shape | Target shape | Expected outcome |
| --- | --- | --- | --- | --- |
| `N1-T` | Nested object drops field | typed single | nested target has fewer fields | nested extra field ignored |
| `N1-U` | Nested object drops field | untyped single | nested target has fewer fields | nested extra field ignored |
| `N2-T` | Nested vec element drops field | typed vec | nested element target has fewer fields | nested extra field ignored |
| `N2-U` | Nested vec element drops field | untyped vec | nested element target has fewer fields | nested extra field ignored |
| `N3-T` | Three-level nested object drops fields | typed single | deep target has fewer fields | extra fields ignored at each level |
| `N3-U` | Three-level nested object drops fields | untyped single | deep target has fewer fields | extra fields ignored at each level |
| `N4-T` | Nested optional fields | typed vec | nested target keeps optional subset | nested optionals preserved, trailing fields ignored |
| `N4-U` | Nested optional fields | untyped vec | nested target keeps optional subset | nested optionals preserved, trailing fields ignored |
| `O1-T` | Optional present/null | typed vec | optional field on target | present values decoded, blank -> null |
| `O1-U` | Optional present/null | untyped vec | optional field on target | untyped decode semantics preserved |

## Canonical Text Inputs

### `A2-T`

```text
{id@int,name@str,active@bool}:(42,Alice,true)
```

Target:

```text
Person { id, name }
```

Expected:

- `id = 42`
- `name = "Alice"`
- `active` ignored

### `A2-U`

```text
{id,name,active}:(42,Alice,true)
```

Target:

```text
Person { id, name }
```

Expected:

- `id = 42`
- `name = "Alice"`
- `active` ignored

### `A3-T`

```text
{id@int,name@str}:(42,Alice)
```

Target:

```text
PersonWithActive { id, name, active }
```

Expected:

- `id = 42`
- `name = "Alice"`
- `active = default`

### `A4-T`

```text
{active@bool,id@int,name@str}:(true,42,Alice)
```

Target:

```text
Person { id, name, active }
```

Expected:

- fields matched by name, not source order

### `A5-T`

```text
[{id@int,name@str,active@bool}]:(42,Alice,true),(7,Bob,false)
```

Target:

```text
[]Person { id, name }
```

Expected:

- both rows decode
- `active` ignored in each row

### `A6-T`

```text
[{id@int,name@str}]:(42,Alice),(7,Bob)
```

Target:

```text
[]PersonWithActive { id, name, active }
```

Expected:

- `active = default` for each row

### `A6-U`

```text
[{id,name}]:(42,Alice),(7,Bob)
```

Target:

```text
[]PersonWithActive { id, name, active }
```

Expected:

- `active = default` for each row

### `P1-T`

```text
{id@int,name@str,score@float,active@bool}:(42,Alice,9.5,true)
```

Target:

```text
PersonScore { id, score }
```

Expected:

- `id = 42`
- `score = 9.5`
- non-overlapping source fields ignored

### `P1-U`

```text
{id,name,score,active}:(42,Alice,9.5,true)
```

Target:

```text
PersonScore { id, score }
```

Expected:

- same structural behavior as `P1-T`

### `P2-T`

```text
{id@int,name@str}:(42,Alice)
```

Target:

```text
NoOverlap { foo, bar }
```

Expected:

- decode succeeds
- `foo` and `bar` remain default

### `P2-U`

```text
{id,name}:(42,Alice)
```

Target:

```text
NoOverlap { foo, bar }
```

Expected:

- same structural behavior as `P2-T`

### `N1-T`

```text
{name@str,inner@{x@int,y@int,z@float,w@bool},flag@bool}:(test,(10,20,3.14,true),true)
```

Target:

```text
OuterThin { name, inner { x, y } }
```

Expected:

- outer extra field `flag` ignored
- nested extra fields `z`, `w` ignored

### `N1-U`

```text
{name,inner@{x,y,z,w},flag}:(test,(10,20,3.14,true),true)
```

Target:

```text
OuterThin { name, inner { x, y } }
```

Expected:

- same structural behavior as `N1-T`

### `N2-T`

```text
[{name@str,tasks@[{title@str,done@bool,priority@int,weight@float}]}]:(Alpha,[(Design,true,1,0.5),(Code,false,2,0.8)]),(Beta,[(Test,false,3,1.0)])
```

Target:

```text
[]ProjectThin { name, tasks[] { title, done } }
```

Expected:

- nested task fields `priority`, `weight` ignored

### `N2-U`

```text
[{name,tasks@[{title,done,priority,weight}]}]:(Alpha,[(Design,true,1,0.5),(Code,false,2,0.8)]),(Beta,[(Test,false,3,1.0)])
```

Target:

```text
[]ProjectThin { name, tasks[] { title, done } }
```

Expected:

- same structural behavior as `N2-T`

### `N3-T`

```text
{id@int,child@{name@str,sub@{a@int,b@str,c@bool},code@int,tags@[str]},extra@str}:(7,(leaf,(11,hello,true),99,[x,y]),tail)
```

Target:

```text
L1Thin { id, child { name, sub { a } } }
```

Expected:

- `extra` ignored
- nested fields `b`, `c`, `code`, `tags` ignored

### `N3-U`

```text
{id,child@{name,sub@{a,b,c},code,tags},extra}:(7,(leaf,(11,hello,true),99,[x,y]),tail)
```

Target:

```text
L1Thin { id, child { name, sub { a } } }
```

Expected:

- same structural behavior as `N3-T`

### `N4-T`

```text
[{id@int,profile@{name@str,nick@str?,score@float?},active@bool}]:(1,(Alice,ally,9.5),true),(2,(Bob,,),false)
```

Target:

```text
[]UserWithNestedOptional { id, profile { name, nick? } }
```

Expected:

- row 0 nested `nick = "ally"`
- row 1 nested `nick = null/None/empty Optional`
- nested `score` and outer `active` ignored

### `N4-U`

```text
[{id,profile@{name,nick,score},active}]:(1,(Alice,ally,9.5),true),(2,(Bob,,),false)
```

Target:

```text
[]UserWithNestedOptional { id, profile { name, nick? } }
```

Expected:

- same structural behavior as `N4-T`

### `O1-T`

```text
[{id@int,label@str?,score@float?,flag@bool}]:(1,hello,95.5,true),(2,,,false)
```

Target:

```text
[]DstFewerOptionals { id, label? }
```

Expected:

- row 0: `label = Some("hello")`
- row 1: `label = null/None/empty Optional`
- trailing fields `score`, `flag` ignored

### `O1-U`

```text
[{id,label,score,flag}]:(1,hello,95.5,true),(2,,,false)
```

Target:

```text
[]DstFewerOptionals { id, label? }
```

Expected:

- row 0 keeps `label`
- row 1 gets empty optional/null
- trailing fields `score`, `flag` ignored

## Current Coverage Target

This repository should at minimum keep the following implementations aligned:

- `ason-go`
- `ason-rs`
- `ason-java`

These bindings already support typed target decoding and field-name matching, so
they are the best place to keep the canonical compatibility behavior locked in.
