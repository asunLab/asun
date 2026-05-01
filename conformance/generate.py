#!/usr/bin/env python3
"""
ASUN conformance case generator.

Emits conformance/cases.json: an array of test vectors that all language
implementations must satisfy.

Schema of each case:
  {
    "id":        "<unique dotted id>",
    "category":  "primitives | strings | objects | arrays | nested | comments | whitespace | errors | ...",
    "desc":      "<short human description>",
    "input":     "<raw ASUN text>",
    "kind":      "ok" | "error",
    // when kind == "ok":
    "expected":  <canonical JSON value>,   // result of decoding
    // when kind == "error":
    "errorHint": "<rough category, advisory only>"
  }

Decoders are checked by:
    decode(input) deep-equals expected     (ok)
    decode(input) raises a parse/format error (error)

The canonical JSON value uses these conventions:
  - integers   -> JSON number (must round-trip exactly)
  - floats     -> JSON number with a decimal point
  - booleans   -> true / false
  - null       -> null
  - strings    -> JSON string (UTF-8)
  - objects    -> JSON object {fieldName: value, ...} (order = schema order)
  - arrays     -> JSON array

Implementations whose native dynamic type cannot represent something
(e.g. "int" vs "float" distinction in JS) MAY normalise to a numeric value
that compares equal; conformance runners are responsible for tolerant
numeric comparison.
"""
from __future__ import annotations
import json, os, sys
from typing import Any

CASES: list[dict] = []
ENC_CASES: list[dict] = []

def add(id_: str, category: str, desc: str, input_: str, expected: Any) -> None:
    CASES.append({
        "id": id_,
        "category": category,
        "desc": desc,
        "input": input_,
        "kind": "ok",
        "schemaDriven": _is_schema_driven(input_),
        "expected": expected,
    })

def err(id_: str, category: str, desc: str, input_: str, hint: str) -> None:
    CASES.append({
        "id": id_,
        "category": category,
        "desc": desc,
        "input": input_,
        "kind": "error",
        "schemaDriven": _is_schema_driven(input_),
        "errorHint": hint,
    })


def add_enc(id_: str, category: str, desc: str, value: Any, *, note: str | None = None) -> None:
    """Add a round-trip encode case.

    Semantics: `decode(encode(value))` must deep-equal `value`. Implementations
    are free to choose any quoting / escaping strategy that preserves the
    value; we do NOT require byte-exact output across implementations.
    """
    case = {
        "id": id_,
        "category": category,
        "desc": desc,
        "value": value,
        "kind": "round-trip",
    }
    if note is not None:
        case["note"] = note
    ENC_CASES.append(case)


def _is_schema_driven(input_: str) -> bool:
    """An input is schema-driven if its top-level form starts with `{` or `[{`
    (after stripping leading whitespace and block comments). Such inputs require
    a typed decoder; untyped/dynamic decoders need not support them.
    """
    s = input_.lstrip(" \t\r\n")
    # strip leading block comments
    while s.startswith("/*"):
        end = s.find("*/")
        if end < 0:
            break
        s = s[end + 2 :].lstrip(" \t\r\n")
    if s.startswith("{"):
        return True
    if s.startswith("[") and len(s) > 1 and s.lstrip("[ \t\r\n").startswith("{"):
        return True
    return False

# ---------------------------------------------------------------------------
# 1. Bare values (top-level scalars)
# ---------------------------------------------------------------------------

def gen_bare_values():
    cat = "primitives.bare"
    add(f"{cat}.int.zero",        cat, "bare int 0",         "0",          0)
    add(f"{cat}.int.pos",         cat, "positive int",       "42",         42)
    add(f"{cat}.int.neg",         cat, "negative int",       "-7",         -7)
    add(f"{cat}.int.large",       cat, "i64-range int",      "9223372036854775807", 9223372036854775807)
    add(f"{cat}.int.minusZero",   cat, "negative zero int",  "-0",         0)
    add(f"{cat}.float.basic",     cat, "float 3.14",         "3.14",       3.14)
    add(f"{cat}.float.neg",       cat, "negative float",     "-2.5",       -2.5)
    add(f"{cat}.float.tiny",      cat, "small float",        "0.1",        0.1)
    # ABNF requires both integer-part and fractional-part to have >= 1 digit;
    # leading "+" is forbidden. The following tokens are therefore plain strings.
    add(f"{cat}.float.dotPrefix",     cat, "leading-dot is NOT a number per ABNF; treated as plain-string", ".5",   ".5")
    add(f"{cat}.float.dotPrefixNeg",  cat, "leading -. is NOT a number per ABNF; treated as plain-string", "-.5", "-.5")
    add(f"{cat}.float.dotSuffix",     cat, "trailing-dot is NOT a number per ABNF; treated as plain-string", "5.",   "5.")
    add(f"{cat}.float.plusPrefix",    cat, "leading + is NOT a number per ABNF; treated as plain-string",   "+5",   "+5")
    add(f"{cat}.float.plusDotPrefix", cat, "leading +. is NOT a number per ABNF; treated as plain-string",  "+.5", "+.5")
    add(f"{cat}.float.dotPrefixExp",  cat, "leading-dot with exponent is NOT a number; treated as plain-string", ".5e2", ".5e2")
    add(f"{cat}.float.expEmpty",      cat, "exponent without digits is NOT a number; treated as plain-string", "1e",   "1e")
    add(f"{cat}.float.expSignOnly",   cat, "exponent with sign but no digits is NOT a number; treated as plain-string", "1e+", "1e+")
    add(f"{cat}.float.exp",           cat, "scientific notation is a number",                "1e10",   1e10)
    add(f"{cat}.float.expSigned",     cat, "scientific notation with signed exponent",       "1.5e-3", 1.5e-3)
    add(f"{cat}.float.expIntOnly",    cat, "scientific notation without decimal part (lowercase)", "1e6",  1e6)
    add(f"{cat}.float.expIntOnlyCap", cat, "scientific notation without decimal part (uppercase)", "1E6",  1e6)
    add(f"{cat}.float.expBig",        cat, "scientific notation with positive sign",          "-2.0E+10", -2.0e10)
    add(f"{cat}.bool.true",       cat, "boolean true",       "true",       True)
    add(f"{cat}.bool.false",      cat, "boolean false",      "false",      False)
    add(f"{cat}.string.plain",    cat, "bare plain string",  "hello",      "hello")
    add(f"{cat}.string.alphanum", cat, "alphanumeric",       "abc123",     "abc123")
    add(f"{cat}.string.startsDigit", cat, "starts with digit (mixed)", "123abc", "123abc")
    add(f"{cat}.string.quoted.empty",  cat, "quoted empty",  '""',         "")
    add(f"{cat}.string.quoted.basic",  cat, "quoted basic",  '"hello"',    "hello")
    add(f"{cat}.string.quoted.spaces", cat, "preserves spaces", '"  hi  "', "  hi  ")
    add(f"{cat}.string.quoted.unicode",cat, "unicode literal", '"中文"',    "中文")

    # Whitespace handling on bare plain strings (trimmed)
    add(f"{cat}.string.trim.leading",  cat, "leading ws trimmed",  "  hello",   "hello")
    add(f"{cat}.string.trim.trailing", cat, "trailing ws trimmed", "hello   ",  "hello")
    add(f"{cat}.string.trim.both",     cat, "both sides trimmed",  "  hello  ", "hello")
    add(f"{cat}.string.internalWs",    cat, "internal ws kept",    "hello world", "hello world")

    # Negative-number rules
    add(f"{cat}.minus.notNumber",      cat, "space after minus -> string", "- 5", "- 5")
    add(f"{cat}.minus.dashWord",       cat, "lone dash word",     "-foo",  "-foo")

# ---------------------------------------------------------------------------
# 2. Quoted-string escape coverage
# ---------------------------------------------------------------------------

def gen_escapes():
    cat = "strings.escape"
    add(f"{cat}.bs.quote",     cat, "escaped quote",      r'"a\"b"',     'a"b')
    add(f"{cat}.bs.backslash", cat, "escaped backslash",  r'"a\\b"',     'a\\b')
    add(f"{cat}.bs.newline",   cat, "newline escape",     r'"a\nb"',     'a\nb')
    add(f"{cat}.bs.tab",       cat, "tab escape",         r'"a\tb"',     'a\tb')
    add(f"{cat}.bs.cr",        cat, "carriage return",    r'"a\rb"',     'a\rb')
    add(f"{cat}.bs.unicode.bmp",   cat, "unicode BMP",    r'"\u4e2d\u6587"', "中文")
    add(f"{cat}.bs.unicode.ascii", cat, "unicode ascii",  r'"\u0041"',   "A")
    add(f"{cat}.bs.delim.comma",   cat, "escaped comma in quoted",  r'"a\,b"', "a,b")
    add(f"{cat}.bs.delim.lparen",  cat, "escaped (",      r'"a\(b"',     "a(b")
    add(f"{cat}.bs.delim.rparen",  cat, "escaped )",      r'"a\)b"',     "a)b")
    add(f"{cat}.bs.delim.lbrack",  cat, "escaped [",      r'"a\[b"',     "a[b")
    add(f"{cat}.bs.delim.rbrack",  cat, "escaped ]",      r'"a\]b"',     "a]b")

    # Plain-string escapes
    cat2 = "strings.plainEscape"
    add(f"{cat2}.comma",    cat2, "plain string \\,",   r"a\,b",   "a,b")
    add(f"{cat2}.lparen",   cat2, "plain string \\(",   r"a\(b",   "a(b")
    add(f"{cat2}.rparen",   cat2, "plain string \\)",   r"a\)b",   "a)b")
    add(f"{cat2}.lbrack",   cat2, "plain string \\[",   r"a\[b",   "a[b")
    add(f"{cat2}.rbrack",   cat2, "plain string \\]",   r"a\]b",   "a]b")
    add(f"{cat2}.bsbs",     cat2, "plain string \\\\",  r"a\\b",   "a\\b")

# ---------------------------------------------------------------------------
# 3. Single object
# ---------------------------------------------------------------------------

def gen_single_object():
    cat = "objects.single"
    add(f"{cat}.basic",
        cat, "basic 3-field",
        "{id,name,active}:(1,Alice,true)",
        {"id": 1, "name": "Alice", "active": True})
    add(f"{cat}.typed",
        cat, "with @-types",
        "{id@int,name@str,active@bool}:(1,Alice,true)",
        {"id": 1, "name": "Alice", "active": True})
    add(f"{cat}.float",
        cat, "with float field",
        "{x@float,y@float}:(1.5,-2.5)",
        {"x": 1.5, "y": -2.5})
    add(f"{cat}.nullField",
        cat, "missing -> null",
        "{name@str,age@int}:(Alice,)",
        {"name": "Alice", "age": None})
    add(f"{cat}.emptyString",
        cat, "explicit empty string",
        '{name@str,bio@str}:(Alice,"")',
        {"name": "Alice", "bio": ""})
    add(f"{cat}.singleField",
        cat, "single field",
        "{name@str}:(Alice)",
        {"name": "Alice"})
    add(f"{cat}.partialHints",
        cat, "partial hints",
        "{id@int,name,active@bool}:(1,Alice,true)",
        {"id": 1, "name": "Alice", "active": True})
    add(f"{cat}.numericLikeStringQuoted",
        cat, "leading-zero kept in quoted",
        '{zip@str}:("001234")',
        {"zip": "001234"})
    add(f"{cat}.spaceAroundComma",
        cat, "schema with spaces",
        "{ id @ int , name @ str }:(1,Alice)",
        {"id": 1, "name": "Alice"})
    add(f"{cat}.boolForcedString",
        cat, "quoted bool stays string",
        '{flag@str}:("true")',
        {"flag": "true"})
    add(f"{cat}.fieldNameWithDigit",
        cat, "field name starts with digit",
        "{1st@int,2nd@int}:(10,20)",
        {"1st": 10, "2nd": 20})
    add(f"{cat}.fieldNameUnderscore",
        cat, "field name with underscore",
        "{_id@int,user_name@str}:(7,bob)",
        {"_id": 7, "user_name": "bob"})

# ---------------------------------------------------------------------------
# 4. Array of objects
# ---------------------------------------------------------------------------

def gen_array_of_objects():
    cat = "objects.array"
    add(f"{cat}.basic",
        cat, "two rows",
        "[{id@int,name@str,active@bool}]:(1,Alice,true),(2,Bob,false)",
        [{"id": 1, "name": "Alice", "active": True},
         {"id": 2, "name": "Bob", "active": False}])
    add(f"{cat}.empty",
        cat, "empty array of objects",
        "[{id@int,name@str}]:",
        [])
    add(f"{cat}.singleRow",
        cat, "single row in array form",
        "[{id@int,name@str}]:(1,Alice)",
        [{"id": 1, "name": "Alice"}])
    add(f"{cat}.trailingComma",
        cat, "trailing comma between rows",
        "[{id@int,name@str}]:(1,Alice),(2,Bob),",
        [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}])
    add(f"{cat}.multilineIndented",
        cat, "indented multi-line",
        "[{id@int,name@str,active@bool}]:\n  (1, Alice, true),\n  (2, Bob, false)",
        [{"id": 1, "name": "Alice", "active": True},
         {"id": 2, "name": "Bob", "active": False}])

    # Auto-generate 50 rows
    rows = ",".join(f"({i},user_{i:03d},{('true' if i%2 else 'false')})" for i in range(50))
    add(f"{cat}.fifty.rows",
        cat, "50 rows",
        f"[{{id@int,name@str,active@bool}}]:{rows}",
        [{"id": i, "name": f"user_{i:03d}", "active": (i % 2 == 1)} for i in range(50)])

# ---------------------------------------------------------------------------
# 5. Nested structures
# ---------------------------------------------------------------------------

def gen_nested():
    cat = "nested"
    add(f"{cat}.objectInObject",
        cat, "addr@{...}",
        "{name@str,addr@{city@str,zip@int}}:(Alice,(NYC,10001))",
        {"name": "Alice", "addr": {"city": "NYC", "zip": 10001}})
    add(f"{cat}.arrayInObject",
        cat, "scores@[int]",
        "{name@str,scores@[int]}:(Alice,[90,85,92])",
        {"name": "Alice", "scores": [90, 85, 92]})
    add(f"{cat}.arrayObjectsInObject",
        cat, "users@[{...}]",
        "{team@str,users@[{id@int,name@str}]}:(Dev,[(1,Alice),(2,Bob)])",
        {"team": "Dev", "users": [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}]})
    add(f"{cat}.deep",
        cat, "deeply nested",
        "{a@{b@{c@{d@int}}}}:(((((42))))" + ")",
        {"a": {"b": {"c": {"d": 42}}}})
    add(f"{cat}.complexBig",
        cat, "company example",
        "{company@str,employees@[{id@int,name@str,skills@[str]}],active@bool}:(ACME,[(1,Alice,[rust,go]),(2,Bob,[python])],true)",
        {"company": "ACME",
         "employees": [{"id": 1, "name": "Alice", "skills": ["rust", "go"]},
                       {"id": 2, "name": "Bob",   "skills": ["python"]}],
         "active": True})
    add(f"{cat}.arrayOfArrays",
        cat, "array of arrays in field",
        "{m@[[int]]}:([[1,2],[3,4]])",
        {"m": [[1, 2], [3, 4]]})

# ---------------------------------------------------------------------------
# 6. Plain arrays
# ---------------------------------------------------------------------------

def gen_plain_arrays():
    cat = "arrays.plain"
    add(f"{cat}.empty",   cat, "empty array",       "[]",           [])
    add(f"{cat}.ints",    cat, "int array",         "[1,2,3]",      [1, 2, 3])
    add(f"{cat}.floats",  cat, "float array",       "[1.0,2.5,3.25]", [1.0, 2.5, 3.25])
    add(f"{cat}.strings", cat, "plain strings",     "[a,b,c]",      ["a", "b", "c"])
    add(f"{cat}.mixed",   cat, "mixed types",       "[1,hello,true,3.14]", [1, "hello", True, 3.14])
    add(f"{cat}.nested",  cat, "nested arrays",     "[[1,2],[3,4]]", [[1, 2], [3, 4]])
    add(f"{cat}.trailingComma", cat, "trailing comma", "[1,2,3,]",   [1, 2, 3])
    add(f"{cat}.singletonString", cat, "single string", "[hello]",   ["hello"])
    add(f"{cat}.boolArray", cat, "bool array",      "[true,false,true]", [True, False, True])
    add(f"{cat}.sparseNull", cat, "sparse with null",  "[1,,3]",     [1, None, 3])
    add(f"{cat}.size200",  cat, "200-element ints",
        "[" + ",".join(str(i) for i in range(200)) + "]",
        list(range(200)))

# ---------------------------------------------------------------------------
# 7. Whitespace / multi-line / formatting
# ---------------------------------------------------------------------------

def gen_whitespace():
    cat = "whitespace"
    add(f"{cat}.schemaSpaces",
        cat, "spaces inside schema",
        "{ id @ int , name @ str }:(1, Alice)",
        {"id": 1, "name": "Alice"})
    add(f"{cat}.dataSpaceTrim",
        cat, "spaces around plain string trimmed",
        "{name@str}:(   Alice   )",
        {"name": "Alice"})
    add(f"{cat}.dataSpaceQuoted",
        cat, "spaces preserved in quoted",
        '{name@str}:("   Alice   ")',
        {"name": "   Alice   "})
    add(f"{cat}.crlf",
        cat, "CRLF line endings",
        "[{id@int,name@str}]:\r\n  (1,Alice),\r\n  (2,Bob)",
        [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}])
    add(f"{cat}.tabIndent",
        cat, "tab indentation",
        "[{id@int,name@str}]:\n\t(1,Alice),\n\t(2,Bob)",
        [{"id": 1, "name": "Alice"}, {"id": 2, "name": "Bob"}])
    add(f"{cat}.leadingDocWs",
        cat, "leading document whitespace",
        "   \n  {id@int}:(1)",
        {"id": 1})
    add(f"{cat}.trailingDocWs",
        cat, "trailing document whitespace",
        "{id@int}:(1)\n   \n",
        {"id": 1})

# ---------------------------------------------------------------------------
# 8. Comments
# ---------------------------------------------------------------------------

def gen_comments():
    cat = "comments"
    add(f"{cat}.leading",
        cat, "leading block comment",
        "/* hello */ {id@int}:(1)",
        {"id": 1})
    add(f"{cat}.trailing",
        cat, "trailing block comment",
        "{id@int}:(1) /* tail */",
        {"id": 1})
    add(f"{cat}.between",
        cat, "comment in schema",
        "{id@int /* pk */, name@str}:(1,Alice)",
        {"id": 1, "name": "Alice"})
    add(f"{cat}.multiline",
        cat, "multiline comment",
        "/*\n  doc\n*/\n{id@int}:(1)",
        {"id": 1})
    add(f"{cat}.betweenRows",
        cat, "comment between rows",
        "[{id@int}]:(1) /* row1 */ ,(2)",
        [{"id": 1}, {"id": 2}])

# ---------------------------------------------------------------------------
# 9. Trailing-comma vs null semantics
# ---------------------------------------------------------------------------

def gen_commas():
    cat = "commas"
    add(f"{cat}.trail",
        cat, "trailing comma absorbed",
        "{a@int,b@int}:(1,2,)",
        {"a": 1, "b": 2})
    add(f"{cat}.consec",
        cat, "consecutive commas -> null",
        "{a@int,b@int,c@int}:(1,,3)",
        {"a": 1, "b": None, "c": 3})
    add(f"{cat}.consecAndTrail",
        cat, "null then trailing",
        "{a@int,b@int,c@int}:(1,2,,)",
        {"a": 1, "b": 2, "c": None})
    add(f"{cat}.allNull",
        cat, "all nulls",
        "{a@int,b@int,c@int}:(,,)",
        {"a": None, "b": None, "c": None})

# ---------------------------------------------------------------------------
# 10. Errors
# ---------------------------------------------------------------------------

def gen_errors():
    cat = "errors"
    err(f"{cat}.unclosedQuote",
        cat, "unterminated quoted string",
        '{a@str}:("hello)',
        "lex.unterminated_string")
    err(f"{cat}.unclosedTuple",
        cat, "missing close paren",
        "{a@int,b@int}:(1,2",
        "parse.unclosed_tuple")
    err(f"{cat}.unclosedSchema",
        cat, "missing close brace",
        "{a@int,b@int:(1,2)",
        "parse.unclosed_schema")
    err(f"{cat}.unclosedComment",
        cat, "unterminated comment",
        "/* nope {a@int}:(1)",
        "lex.unterminated_comment")
    err(f"{cat}.tooMany",
        cat, "too many values",
        "{a@int,b@int}:(1,2,3)",
        "parse.field_count")
    err(f"{cat}.tooFew",
        cat, "too few values (no trailing comma)",
        "{a@int,b@int}:(1)",
        "parse.field_count")
    err(f"{cat}.bareTuple",
        cat, "bare tuple at top",
        "(1,2,3)",
        "parse.bare_tuple")
    err(f"{cat}.missingColon",
        cat, "missing : after schema",
        "{a@int,b@int}(1,2)",
        "parse.missing_colon")
    err(f"{cat}.invalidEscape",
        cat, "invalid escape sequence",
        r'{a@str}:("a\xb")',
        "lex.bad_escape")
    err(f"{cat}.commentInsideTuple",
        cat, "comment inside data tuple is forbidden",
        "{a@int,b@int}:(1,/*x*/2)",
        "parse.comment_in_data")
    err(f"{cat}.typeMismatch",
        cat, "string in @int field",
        "{a@int}:(hello)",
        "type.coercion")
    err(f"{cat}.boolWrongCase",
        cat, "True (capitalised) is a string, not boolean",
        "{a@bool}:(True)",
        "type.coercion")

# ---------------------------------------------------------------------------
# 11. Parameterised int/float ranges
# ---------------------------------------------------------------------------

def gen_numeric_sweep():
    cat = "primitives.numeric"
    int_values = [0, 1, -1, 10, -10, 127, -128, 32767, -32768, 2147483647, -2147483648,
                  9223372036854775806, -9223372036854775807, 100, 1000, 100000, 1000000]
    for v in int_values:
        add(f"{cat}.int.{v}", cat, f"int {v} in tuple",
            f"{{x@int}}:({v})", {"x": v})
    floats = [0.0, 1.0, -1.0, 0.5, -0.5, 3.14, -3.14, 2.5, 100.25, -1000.125, 0.125, 1.5, 99.99]
    for v in floats:
        add(f"{cat}.float.{str(v).replace('.','_').replace('-','m')}",
            cat, f"float {v} in tuple",
            f"{{x@float}}:({v})", {"x": v})

# ---------------------------------------------------------------------------
# 12. String-shape sweep
# ---------------------------------------------------------------------------

def gen_string_shapes():
    cat = "strings.shape"
    cases = [
        ("ascii_lower", "hello",           "hello"),
        ("ascii_upper", "HELLO",           "HELLO"),
        ("ascii_mixed", "Hello",           "Hello"),
        ("with_dash",   "a-b-c",           "a-b-c"),
        ("with_dot",    "file.ext",        "file.ext"),
        ("with_slash",  "path/to/file",    "path/to/file"),
        ("with_eq",     "a=b",             "a=b"),
        ("with_amp",    "a&b",             "a&b"),
        ("with_pct",    "100%",            "100%"),
        ("with_plus",   "a+b",             "a+b"),
        ("with_star",   "a*b",             "a*b"),
        ("with_q",      "a?b",             "a?b"),
        ("with_excl",   "wow!",            "wow!"),
        ("with_hash",   "#tag",            "#tag"),
        ("with_dollar", "$amount",         "$amount"),
        ("emoji",       '"🚀"',             "🚀"),
        ("cjk_kr",      '"한국어"',          "한국어"),
        ("cjk_jp",      '"日本語"',          "日本語"),
        ("rtl_ar",      '"العربية"',       "العربية"),
        ("mixed_unicode", '"Hi 🌍 안녕"',    "Hi 🌍 안녕"),
    ]
    for name, lit, val in cases:
        add(f"{cat}.{name}", cat, f"string shape {name}",
            f"{{x@str}}:({lit})", {"x": val})

# ---------------------------------------------------------------------------
# 13. Optional / nullable sweep
# ---------------------------------------------------------------------------

def gen_optionals():
    cat = "optionals"
    add(f"{cat}.someInt",  cat, "Some(int)",  "{x@int}:(7)",   {"x": 7})
    add(f"{cat}.noneInt",  cat, "None(int)",  "{x@int}:()",    {"x": None})
    add(f"{cat}.someStr",  cat, "Some(str)",  "{x@str}:(hi)",  {"x": "hi"})
    add(f"{cat}.noneStr",  cat, "None(str)",  "{x@str}:()",    {"x": None})
    add(f"{cat}.emptyStrSome", cat, "explicit empty string is some",
        '{x@str}:("")', {"x": ""})
    add(f"{cat}.middleNone", cat, "middle field is null",
        "{a@int,b@int,c@int}:(1,,3)",
        {"a": 1, "b": None, "c": 3})

# ---------------------------------------------------------------------------
# main
# ---------------------------------------------------------------------------

# ---------------------------------------------------------------------------
# 14. Bulk row sweep — array-of-objects with N in {1..30}
# ---------------------------------------------------------------------------

def gen_bulk_rows():
    cat = "objects.array.bulk"
    for n in range(1, 31):
        rows = ",".join(f"({i},u{i})" for i in range(n))
        add(f"{cat}.n{n:02d}", cat, f"{n} rows",
            f"[{{id@int,name@str}}]:{rows}",
            [{"id": i, "name": f"u{i}"} for i in range(n)])

# ---------------------------------------------------------------------------
# 15. Plain-array length sweep
# ---------------------------------------------------------------------------

def gen_array_lengths():
    cat = "arrays.length"
    for n in [1, 2, 3, 5, 8, 13, 21, 34, 55, 89, 100]:
        items = list(range(n))
        text = "[" + ",".join(str(i) for i in items) + "]"
        add(f"{cat}.int.n{n}", cat, f"int array len={n}", text, items)
    for n in [1, 2, 4, 8, 16, 32]:
        items = [f"s{i}" for i in range(n)]
        text = "[" + ",".join(items) + "]"
        add(f"{cat}.str.n{n}", cat, f"str array len={n}", text, items)

# ---------------------------------------------------------------------------
# 16. Schema width sweep — object with k fields
# ---------------------------------------------------------------------------

def gen_object_widths():
    cat = "objects.width"
    for k in [1, 2, 3, 4, 5, 8, 12, 16, 24, 32]:
        fields = ",".join(f"f{i}@int" for i in range(k))
        values = ",".join(str(i * 7) for i in range(k))
        add(f"{cat}.k{k:02d}", cat, f"{k}-field object",
            f"{{{fields}}}:({values})",
            {f"f{i}": i * 7 for i in range(k)})

# ---------------------------------------------------------------------------
# 17. Nesting-depth sweep
# ---------------------------------------------------------------------------

def gen_depth():
    cat = "nested.depth"
    for d in range(1, 9):
        schema = "{a@" + "{a@" * (d - 1) + "int" + "}" * d
        data = "(" * d + "1" + ")" * d
        text = f"{schema}:{data}"
        # build expected
        v: Any = 1
        for _ in range(d):
            v = {"a": v}
        add(f"{cat}.d{d}", cat, f"depth={d}", text, v)

# ---------------------------------------------------------------------------
# 18. Quoted vs unquoted equivalence
# ---------------------------------------------------------------------------

def gen_quoted_equiv():
    cat = "strings.quotedEquiv"
    pairs = [
        ("hello", '"hello"', "hello"),
        ("hello world", '"hello world"', "hello world"),
        ("abc", '"abc"', "abc"),
        ("with-dash", '"with-dash"', "with-dash"),
    ]
    for i, (plain, quoted, expected) in enumerate(pairs):
        add(f"{cat}.unquoted.{i}", cat, f"unquoted {plain!r}",
            f"{{x@str}}:({plain})", {"x": expected})
        add(f"{cat}.quoted.{i}",   cat, f"quoted {quoted!r}",
            f"{{x@str}}:({quoted})", {"x": expected})

# ---------------------------------------------------------------------------
# 19. Mixed null / value patterns over 5-field schema
# ---------------------------------------------------------------------------

def gen_null_patterns():
    cat = "objects.nullPatterns"
    for mask in range(1, 32):  # skip 0 (would be ",,,," — all null, covered)
        bits = [(mask >> i) & 1 for i in range(5)]
        vals = []
        expected = {}
        for i, bit in enumerate(bits):
            field = f"f{i}"
            if bit:
                vals.append(str(i + 1))
                expected[field] = i + 1
            else:
                vals.append("")
                expected[field] = None
        schema = ",".join(f"f{i}@int" for i in range(5))
        text = f"{{{schema}}}:({','.join(vals)})"
        add(f"{cat}.m{mask:02d}", cat, f"null-mask {mask:05b}", text, expected)


# =============================================================================
# Encode (round-trip) cases
# =============================================================================
#
# Each case provides a JSON-serialisable value. The runner must verify that
# `decode(encode(value))` deep-equals `value`. Implementations are free to
# choose any quoting / escaping strategy that preserves the value; we do NOT
# require byte-exact output across implementations.
#
# Why this catches real bugs that the decode-only suite cannot:
#
#   * If an encoder emits the string "42" without quoting, a downstream
#     decoder will read the integer 42 — round-trip BROKEN.
#   * If an encoder emits the string "{x}" without quoting, an untyped
#     decoder will see a schema header (or a bare-tuple error) — round-trip
#     BROKEN. (asun-rs has this exact bug; see simd.rs special-char list.)
#   * If an encoder emits the string "" as nothing, the decoder reads
#     null (or array element absence) — round-trip BROKEN.
#   * Strings containing structural chars (`,`, `(`, `)`, `[`, `]`, `@`,
#     `"`, `\`) inside a plain array become parser ambiguities.

def gen_encode_strings_lookalike():
    """Strings that, if emitted bare, would be parsed as a different type."""
    cat = "encode.string.lookalike"
    cases: list[tuple[str, str, str]] = [
        # int-like
        ("zero",            "string '0'",      "0"),
        ("posInt",          "string '1'",      "1"),
        ("smallPos",        "string '42'",     "42"),
        ("negInt",          "string '-7'",     "-7"),
        ("negZero",         "string '-0'",     "-0"),
        ("longInt",         "20-digit int str","12345678901234567890"),
        ("plusSign",        "string '+5' (not a valid number, but starts oddly)", "+5"),
        ("leadingZero",     "string '007'",    "007"),
        # float-like
        ("zeroFloat",       "string '0.0'",    "0.0"),
        ("negZeroFloat",    "string '-0.0'",   "-0.0"),
        ("pi",              "string '3.14'",   "3.14"),
        ("scientific",      "string '1e10'",   "1e10"),
        ("scientificCap",   "string '1E10'",   "1E10"),
        ("scientificNeg",   "string '1.5e-3'", "1.5e-3"),
        ("noLeadDigit",     "string '.5'",     ".5"),
        ("trailingDot",     "string '5.'",     "5."),
        ("twoDots",         "string '1.2.3'",  "1.2.3"),
        # bool / null lookalikes
        ("bareTrue",        "string 'true'",   "true"),
        ("bareFalse",       "string 'false'",  "false"),
        ("capTrue",         "string 'True'",   "True"),
        ("upTrue",          "string 'TRUE'",   "TRUE"),
        ("yes",             "string 'yes'",    "yes"),
        ("no",              "string 'no'",     "no"),
        ("bareNull",        "string 'null'",   "null"),
        ("capNull",         "string 'Null'",   "Null"),
        ("nilStr",          "string 'nil'",    "nil"),
        ("none",            "string 'None'",   "None"),
        # decoder-fallback edge cases (from the existing decode suite)
        ("dashSpaceDigit",  "string '- 5'",    "- 5"),
        ("dashWord",        "string '-foo'",   "-foo"),
        ("digitsThenWord",  "string '123abc'", "123abc"),
        ("digitsThenSpace", "string '123 abc'","123 abc"),
    ]
    for sub, desc, val in cases:
        add_enc(f"{cat}.{sub}", cat, desc, val)


def gen_encode_strings_whitespace():
    """Strings made entirely of, or surrounded by, whitespace.

    SPEC §S2: unquoted plain strings are trimmed. So a string like " hi "
    can only round-trip if the encoder QUOTES it.
    """
    cat = "encode.string.whitespace"
    cases: list[tuple[str, str, str]] = [
        ("empty",         "empty string",                 ""),
        ("spaceOnly",     "single space",                 " "),
        ("spaces",        "two spaces",                   "  "),
        ("tabOnly",       "single tab",                   "\t"),
        ("newlineOnly",   "single newline",               "\n"),
        ("crOnly",        "single carriage return",       "\r"),
        ("leadingSpace",  "leading space",                " hi"),
        ("trailingSpace", "trailing space",               "hi "),
        ("bothSpace",     "leading and trailing space",   " hi "),
        ("doubleBoth",    "double leading/trailing",      "  hi  "),
        ("internalTab",   "internal tab",                 "a\tb"),
        ("internalNL",    "internal newline",             "a\nb"),
        ("internalCR",    "internal CR",                  "a\rb"),
        ("internalSpace", "internal space (bare-OK)",     "a b"),
    ]
    for sub, desc, val in cases:
        add_enc(f"{cat}.{sub}", cat, desc, val)


def gen_encode_strings_structural():
    """Strings containing ASUN structural chars (`,()[]{}@"\\:<>`)."""
    cat = "encode.string.structural"
    cases: list[tuple[str, str, str]] = [
        # single structural chars
        ("comma",       "string ','",          ","),
        ("lparen",      "string '('",          "("),
        ("rparen",      "string ')'",          ")"),
        ("lbracket",    "string '['",          "["),
        ("rbracket",    "string ']'",          "]"),
        ("lbrace",      "string '{'",          "{"),
        ("rbrace",      "string '}'",          "}"),
        ("at",          "string '@'",          "@"),
        ("colon",       "string ':'",          ":"),
        ("lt",          "string '<'",          "<"),
        ("gt",          "string '>'",          ">"),
        ("dquote",      "string '\"'",         '"'),
        ("backslash",   "string '\\\\'",       "\\"),
        ("slashStar",   "string '/*'",         "/*"),
        ("starSlash",   "string '*/'",         "*/"),
        # combinations
        ("commaInWord", "string 'a,b'",        "a,b"),
        ("parenInWord", "string 'a(b)c'",      "a(b)c"),
        ("brackInWord", "string 'a[b]c'",      "a[b]c"),
        ("braceInWord", "string 'a{b}c'",      "a{b}c"),
        ("atInWord",    "string 'x@int'",      "x@int"),
        # whole-string mimicry of structural forms
        ("looksLikeBareTuple",     "string '(1,2,3)'",      "(1,2,3)"),
        ("looksLikeArray",         "string '[1,2,3]'",      "[1,2,3]"),
        ("looksLikeSchemaHead",    "string '{x}'",          "{x}"),
        ("looksLikeStruct",        "string '{a}:(1)'",      "{a}:(1)"),
        ("looksLikeVecStruct",     "string '[{a}]:(1)'",    "[{a}]:(1)"),
        ("looksLikeSchemaWithAt",  "string '{x@int}'",      "{x@int}"),
        ("emptyParens",            "string '()'",           "()"),
        ("emptyBrackets",          "string '[]'",           "[]"),
        ("emptyBraces",            "string '{}'",           "{}"),
        # would-be comments
        ("commentOpen",            "string starts with /*", "/* not a comment"),
        ("blockComment",           "string '/* x */'",      "/* x */"),
    ]
    for sub, desc, val in cases:
        add_enc(f"{cat}.{sub}", cat, desc, val)


def gen_encode_strings_escape():
    """Strings whose bytes require escape sequences to survive a round-trip."""
    cat = "encode.string.escape"
    cases: list[tuple[str, str, str]] = [
        ("backslash",        "lone backslash",           "\\"),
        ("backslashN",       "literal '\\n' (2 chars)",  "\\n"),
        ("backslashTimes2",  "two backslashes",          "\\\\"),
        ("quote",            "lone double quote",        '"'),
        ("escSeq",           "string '\\\"'",            '\\"'),
        ("nlInside",         "embedded newline",         "line1\nline2"),
        ("crlf",             "embedded CRLF",            "line1\r\nline2"),
        ("tabInside",        "embedded tab",             "col1\tcol2"),
        ("formFeed",         "embedded form feed",       "a\x0cb"),
        ("verticalTab",      "embedded vertical tab",    "a\x0bb"),
        ("backspace",        "embedded backspace",       "a\x08b"),
        ("nullByte",         "embedded NUL",             "a\x00b"),
        ("ctrl01",           "control char U+0001",      "a\x01b"),
        ("ctrl1F",           "control char U+001F",      "a\x1fb"),
        ("del",              "DEL U+007F",               "a\x7fb"),
        # decoder will need this string verbatim:
        ("plainEscaped",     r"string 'a\,b' literal",   "a\\,b"),
    ]
    for sub, desc, val in cases:
        add_enc(f"{cat}.{sub}", cat, desc, val)


def gen_encode_strings_unicode():
    cat = "encode.string.unicode"
    cases: list[tuple[str, str, str]] = [
        ("ascii",         "plain ASCII",            "hello world"),
        ("latin1",        "Latin-1 char 'café'",    "café"),
        ("cjk",           "CJK '中文'",             "中文"),
        ("japanese",      "Japanese '日本語'",      "日本語"),
        ("emoji",         "single emoji",           "🎉"),
        ("emojiZWJ",      "ZWJ family emoji",       "👨\u200d👩\u200d👧"),
        ("flag",          "regional indicator flag","🇺🇸"),
        ("rtl",           "RTL Arabic",             "مرحبا"),
        ("combining",     "combining mark",         "e\u0301"),  # é via combining
        ("bom",           "BOM at start",           "\ufeffhello"),
        ("u0080",         "boundary U+0080",        "\u0080"),
        ("u07FF",         "boundary U+07FF",        "\u07ff"),
        ("u0800",         "boundary U+0800",        "\u0800"),
        ("uFFFF",         "boundary U+FFFF",        "\uffff"),
        ("supplementary", "U+1F600 grinning",       "😀"),
    ]
    for sub, desc, val in cases:
        add_enc(f"{cat}.{sub}", cat, desc, val)


def gen_encode_numbers():
    cat = "encode.number"
    cases: list[tuple[str, str, Any]] = [
        ("intZero",       "0",                       0),
        ("intOne",        "1",                       1),
        ("intNegOne",     "-1",                      -1),
        ("intMaxI32",     "i32 max",                 2147483647),
        ("intMinI32",     "i32 min",                 -2147483648),
        ("intMaxI64",     "i64 max",                 9223372036854775807),
        ("intMinI64",     "i64 min",                 -9223372036854775808),
        ("floatZero",     "0.0",                     0.0),
        ("floatHalf",     "0.5",                     0.5),
        ("floatPi",       "3.141592653589793",       3.141592653589793),
        ("floatNeg",      "-2.5",                    -2.5),
        ("floatBig",      "1e100",                   1e100),
        ("floatSmall",    "1e-100",                  1e-100),
        ("floatTiny",     "5e-324 subnormal",        5e-324),
    ]
    for sub, desc, val in cases:
        add_enc(f"{cat}.{sub}", cat, desc, val)


def gen_encode_bool_null():
    add_enc("encode.bool.true",  "encode.bool", "true",  True)
    add_enc("encode.bool.false", "encode.bool", "false", False)
    add_enc("encode.null.bare",  "encode.null", "null",  None)


def gen_encode_arrays():
    cat = "encode.array"
    add_enc(f"{cat}.empty",       cat, "empty array",        [])
    add_enc(f"{cat}.singleInt",   cat, "[1]",                [1])
    add_enc(f"{cat}.ints",        cat, "[1,2,3]",            [1, 2, 3])
    add_enc(f"{cat}.floats",      cat, "[1.5,2.5]",          [1.5, 2.5])
    add_enc(f"{cat}.bools",       cat, "[true,false]",       [True, False])
    add_enc(f"{cat}.strings",     cat, "['a','b','c']",      ["a", "b", "c"])
    add_enc(f"{cat}.singleNull",  cat, "[null]",             [None])
    add_enc(f"{cat}.nulls",       cat, "[null,null,null]",   [None, None, None])
    add_enc(f"{cat}.mixed",       cat, "[1,'two',true,null]",[1, "two", True, None])
    add_enc(f"{cat}.singleEmpty", cat, "['']",               [""])
    add_enc(f"{cat}.severalEmpty",cat, "['','','']",         ["", "", ""])


def gen_encode_array_of_lookalikes():
    """Arrays containing strings that, if not quoted, would alias other types.

    This is the highest-value cross-language test set: it directly exercises
    the encoder's quoting-discipline contract.
    """
    cat = "encode.array.lookalike"
    cases: list[tuple[str, str, list]] = [
        ("intStrings",      "['1','2','3']",                 ["1", "2", "3"]),
        ("floatStrings",    "['1.5','2.5']",                 ["1.5", "2.5"]),
        ("boolStrings",     "['true','false']",              ["true", "false"]),
        ("nullStrings",     "['null','null']",               ["null", "null"]),
        ("emptyStrings",    "['','','']",                    ["", "", ""]),
        ("commaStrings",    "['a,b','c,d']",                 ["a,b", "c,d"]),
        ("parenStrings",    "['(1)','(2)']",                 ["(1)", "(2)"]),
        ("bracketStrings",  "['[1]','[2]']",                 ["[1]", "[2]"]),
        ("braceStrings",    "['{a}','{b}']",                 ["{a}", "{b}"]),
        ("atStrings",       "['x@int','y@str']",             ["x@int", "y@str"]),
        ("backslashStrs",   "['a\\\\b','c\\\\d']",           ["a\\b", "c\\d"]),
        ("quoteStrs",       "['a\"b']",                      ['a"b']),
        ("nlStrings",       "['line1\\nline2']",             ["line1\nline2"]),
        ("wsStrings",       "[' a ',' b ',' ']",             [" a ", " b ", " "]),
        ("dashStrings",     "['-foo','- 5']",                ["-foo", "- 5"]),
        ("digitWord",       "['123abc','7q']",               ["123abc", "7q"]),
        ("commentLooking",  "['/* x */','y']",               ["/* x */", "y"]),
        ("mixed",           "[1,'1',1.5,'1.5',true,'true']", [1, "1", 1.5, "1.5", True, "true"]),
    ]
    for sub, desc, val in cases:
        add_enc(f"{cat}.{sub}", cat, desc, val)


def gen_encode_nested_arrays():
    cat = "encode.array.nested"
    add_enc(f"{cat}.depth2",     cat, "[[1,2],[3,4]]",         [[1, 2], [3, 4]])
    add_enc(f"{cat}.depth3",     cat, "[[[1]]]",                [[[1]]])
    add_enc(f"{cat}.empties",    cat, "[[],[]]",                [[], []])
    add_enc(f"{cat}.mixedDepth", cat, "[1,[2,3],[],[[4]]]",     [1, [2, 3], [], [[4]]])
    add_enc(f"{cat}.strNested",  cat, "[['a,b'],['c)d']]",      [["a,b"], ["c)d"]])


def main():
    gen_bare_values()
    gen_escapes()
    gen_single_object()
    gen_array_of_objects()
    gen_nested()
    gen_plain_arrays()
    gen_whitespace()
    gen_comments()
    gen_commas()
    gen_errors()
    gen_numeric_sweep()
    gen_string_shapes()
    gen_optionals()
    gen_bulk_rows()
    gen_array_lengths()
    gen_object_widths()
    gen_depth()
    gen_quoted_equiv()
    gen_null_patterns()

    # Encode (round-trip) cases.
    gen_encode_strings_lookalike()
    gen_encode_strings_whitespace()
    gen_encode_strings_structural()
    gen_encode_strings_escape()
    gen_encode_strings_unicode()
    gen_encode_numbers()
    gen_encode_bool_null()
    gen_encode_arrays()
    gen_encode_array_of_lookalikes()
    gen_encode_nested_arrays()

    # Sanity: unique IDs
    seen = set()
    for c in CASES:
        if c["id"] in seen:
            raise SystemExit(f"duplicate id: {c['id']}")
        seen.add(c["id"])

    out_path = os.path.join(os.path.dirname(__file__), "cases.json")
    manifest = {
        "version": 1,
        "spec":    "../docs/SPEC.md",
        "grammar": "GRAMMAR.abnf",
        "count":   len(CASES),
        "cases":   CASES,
    }
    with open(out_path, "w", encoding="utf-8") as f:
        json.dump(manifest, f, ensure_ascii=False, indent=2)
        f.write("\n")
    print(f"wrote {out_path} ({len(CASES)} cases)")

    # Encode cases — separate manifest, round-trip semantics.
    enc_seen: set[str] = set()
    for c in ENC_CASES:
        if c["id"] in enc_seen:
            raise SystemExit(f"duplicate encode id: {c['id']}")
        enc_seen.add(c["id"])
    enc_path = os.path.join(os.path.dirname(__file__), "encode-cases.json")
    enc_manifest = {
        "version": 1,
        "spec":    "../docs/SPEC.md",
        "grammar": "GRAMMAR.abnf",
        "count":   len(ENC_CASES),
        "semantics": "round-trip",
        "rule":    "decode(encode(value)) deep-equals value",
        "cases":   ENC_CASES,
    }
    with open(enc_path, "w", encoding="utf-8") as f:
        json.dump(enc_manifest, f, ensure_ascii=False, indent=2)
        f.write("\n")
    print(f"wrote {enc_path} ({len(ENC_CASES)} encode cases)")

if __name__ == "__main__":
    main()
