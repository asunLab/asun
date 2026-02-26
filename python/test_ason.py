"""Comprehensive tests for ASON Python library."""

import sys
import os
from dataclasses import dataclass, field
from typing import Optional

sys.path.insert(0, os.path.dirname(__file__))
import ason


# ---------------------------------------------------------------------------
# Test structs
# ---------------------------------------------------------------------------

@dataclass
class User:
    id: int
    name: str
    active: bool


@dataclass
class Score:
    id: int
    value: float


@dataclass
class Item:
    id: int
    label: Optional[str] = None


@dataclass
class Tagged:
    name: str
    tags: list[str] = field(default_factory=list)


@dataclass
class Note:
    text: str = ""


@dataclass
class Row:
    id: int = 0
    name: str = ""


@dataclass
class Department:
    title: str = ""


@dataclass
class Employee:
    id: int = 0
    name: str = ""
    dept: Department = field(default_factory=Department)
    skills: list[str] = field(default_factory=list)
    active: bool = True


@dataclass
class WithMap:
    name: str = ""
    attrs: dict[str, int] = field(default_factory=dict)


@dataclass
class Address:
    city: str = ""
    zip: int = 0


@dataclass
class Nested:
    name: str = ""
    addr: Address = field(default_factory=Address)


@dataclass
class WithVec:
    items: list[int] = field(default_factory=list)


@dataclass
class Special:
    val: str = ""


@dataclass
class Matrix3D:
    data: list[list[list[int]]] = field(default_factory=list)


# ---------------------------------------------------------------------------
# Test runner
# ---------------------------------------------------------------------------

_passed = 0
_failed = 0


def test(name: str, fn):
    global _passed, _failed
    try:
        fn()
        _passed += 1
        print(f"  ✓ {name}")
    except Exception as e:
        _failed += 1
        print(f"  ✗ {name}: {e}")


def assert_eq(a, b, msg=""):
    if a != b:
        raise AssertionError(f"{msg}: {a!r} != {b!r}" if msg else f"{a!r} != {b!r}")


class AssertionError(Exception):
    pass


# ---------------------------------------------------------------------------
# Basic tests
# ---------------------------------------------------------------------------

def test_dump_struct():
    u = User(id=1, name="Alice", active=True)
    got = ason.dump(u)
    assert_eq(got, "{id,name,active}:(1,Alice,true)")


def test_load_struct():
    u = ason.load("{id,name,active}:(1,Alice,true)", User)
    assert_eq(u.id, 1)
    assert_eq(u.name, "Alice")
    assert_eq(u.active, True)


def test_load_typed_schema():
    u = ason.load("{id:int,name:str,active:bool}:(1,Alice,true)", User)
    assert_eq(u.id, 1)
    assert_eq(u.name, "Alice")
    assert_eq(u.active, True)


def test_roundtrip():
    original = User(id=42, name="Bob", active=False)
    s = ason.dump(original)
    restored = ason.load(s, User)
    assert_eq(original, restored)


# ---------------------------------------------------------------------------
# Vec tests
# ---------------------------------------------------------------------------

def test_vec_load():
    data = "[{id:int,name:str,active:bool}]:(1,Alice,true),(2,Bob,false)"
    users = ason.load_slice(data, User)
    assert_eq(len(users), 2)
    assert_eq(users[0].name, "Alice")
    assert_eq(users[1].name, "Bob")


def test_vec_dump():
    users = [
        User(id=1, name="Alice", active=True),
        User(id=2, name="Bob", active=False),
    ]
    got = ason.dump_slice(users)
    assert_eq(got, "[{id,name,active}]:(1,Alice,true),(2,Bob,false)")


def test_vec_roundtrip():
    users = [
        User(id=1, name="Alice", active=True),
        User(id=2, name="Bob", active=False),
        User(id=3, name="Carol Smith", active=True),
    ]
    s = ason.dump_slice(users)
    restored = ason.load_slice(s, User)
    assert_eq(len(restored), 3)
    assert_eq(restored[0], users[0])
    assert_eq(restored[1], users[1])
    assert_eq(restored[2], users[2])


# ---------------------------------------------------------------------------
# Feature tests
# ---------------------------------------------------------------------------

def test_optional_with_value():
    item = ason.load("{id,label}:(1,hello)", Item)
    assert_eq(item.id, 1)
    assert_eq(item.label, "hello")


def test_optional_none():
    item = ason.load("{id,label}:(2,)", Item)
    assert_eq(item.id, 2)
    assert_eq(item.label, None)


def test_array_field():
    t = ason.load("{name,tags}:(Alice,[rust,go,python])", Tagged)
    assert_eq(t.name, "Alice")
    assert_eq(t.tags, ["rust", "go", "python"])


def test_nested_struct():
    emp = ason.load("{id,name,dept,skills,active}:(1,Alice,(Manager),[rust],true)", Employee)
    assert_eq(emp.id, 1)
    assert_eq(emp.name, "Alice")
    assert_eq(emp.dept.title, "Manager")
    assert_eq(emp.skills, ["rust"])
    assert_eq(emp.active, True)


def test_map_field():
    wm = ason.load("{name,attrs}:(Alice,[(age,30),(score,95)])", WithMap)
    assert_eq(wm.name, "Alice")
    assert_eq(wm.attrs["age"], 30)
    assert_eq(wm.attrs["score"], 95)


def test_escaped_string():
    note = Note(text='say "hi", then (wave)\tnewline\nend')
    s = ason.dump(note)
    restored = ason.load(s, Note)
    assert_eq(note.text, restored.text)


def test_comments():
    u = ason.load("/* user list */ {id,name,active}:(1,Alice,true)", User)
    assert_eq(u.id, 1)
    assert_eq(u.name, "Alice")


# ---------------------------------------------------------------------------
# Typed serialization tests
# ---------------------------------------------------------------------------

def test_dump_typed():
    u = User(id=1, name="Alice", active=True)
    got = ason.dump_typed(u)
    assert_eq(got, "{id:int,name:str,active:bool}:(1,Alice,true)")


def test_dump_typed_roundtrip():
    u = User(id=42, name="Test User", active=True)
    typed = ason.dump_typed(u)
    restored = ason.load(typed, User)
    assert_eq(u, restored)


def test_dump_slice_typed():
    users = [
        User(id=1, name="Alice", active=True),
        User(id=2, name="Bob", active=False),
    ]
    got = ason.dump_slice_typed(users, ["int", "str", "bool"])
    assert_eq(got, "[{id:int,name:str,active:bool}]:(1,Alice,true),(2,Bob,false)")


def test_dump_slice_typed_auto():
    users = [
        User(id=1, name="Alice", active=True),
    ]
    got = ason.dump_slice_typed(users)
    assert got.startswith("[{id:int,name:str,active:bool}]:")


def test_float_roundtrip():
    s = Score(id=1, value=3.14)
    data = ason.dump(s)
    restored = ason.load(data, Score)
    assert_eq(s.id, restored.id)
    assert abs(s.value - restored.value) < 1e-10


def test_float_integer_value():
    s = Score(id=2, value=95.0)
    data = ason.dump(s)
    assert "95.0" in data
    restored = ason.load(data, Score)
    assert_eq(restored.value, 95.0)


def test_negative_numbers():
    @dataclass
    class Nums:
        a: int
        b: float
        c: int

    n = Nums(a=-42, b=-3.14, c=-9223372036854775807)
    data = ason.dump(n)
    restored = ason.load(data, Nums)
    assert_eq(n.a, restored.a)
    assert abs(n.b - restored.b) < 1e-10
    assert_eq(n.c, restored.c)


# ---------------------------------------------------------------------------
# Annotated vs Unannotated tests
# ---------------------------------------------------------------------------

def test_load_annotated():
    data = "{id:int,name:str,active:bool}:(1,Alice,true)"
    u = ason.load(data, User)
    assert_eq(u.id, 1)
    assert_eq(u.name, "Alice")


def test_load_unannotated():
    data = "{id,name,active}:(1,Alice,true)"
    u = ason.load(data, User)
    assert_eq(u.id, 1)
    assert_eq(u.name, "Alice")


def test_load_annotated_nested():
    data = "{id:int,name:str,dept:{title:str},skills:[str],active:bool}:(1,Alice,(HR),[go],true)"
    emp = ason.load(data, Employee)
    assert_eq(emp.name, "Alice")
    assert_eq(emp.dept.title, "HR")


def test_load_annotated_vec():
    data = "[{id:int,name:str,active:bool}]:(1,Alice,true),(2,Bob,false)"
    users = ason.load_slice(data, User)
    assert_eq(len(users), 2)


def test_multiline():
    data = """[{id:int, name:str, active:bool}]:
  (1, Alice, true),
  (2, Bob, false),
  (3, "Carol Smith", true)"""
    users = ason.load_slice(data, User)
    assert_eq(len(users), 3)
    assert_eq(users[2].name, "Carol Smith")


# ---------------------------------------------------------------------------
# Edge cases
# ---------------------------------------------------------------------------

def test_empty_vec_field():
    wv = WithVec(items=[])
    s = ason.dump(wv)
    restored = ason.load(s, WithVec)
    assert_eq(restored.items, [])


def test_special_chars():
    sp = Special(val='tabs\there, newlines\nhere, quotes"and\\backslash')
    s = ason.dump(sp)
    restored = ason.load(s, Special)
    assert_eq(sp.val, restored.val)


def test_bool_like_string():
    sp = Special(val="true")
    s = ason.dump(sp)
    restored = ason.load(s, Special)
    assert_eq(sp.val, restored.val)


def test_number_like_string():
    sp = Special(val="12345")
    s = ason.dump(sp)
    restored = ason.load(s, Special)
    assert_eq(sp.val, restored.val)


def test_triple_nested_array():
    m = Matrix3D(data=[[[1, 2], [3, 4]], [[5, 6, 7], [8]]])
    s = ason.dump(m)
    restored = ason.load(s, Matrix3D)
    assert_eq(restored.data[0][0][0], 1)
    assert_eq(restored.data[1][0][2], 7)
    assert_eq(len(restored.data), 2)


def test_nested_roundtrip():
    n = Nested(name="Alice", addr=Address(city="NYC", zip=10001))
    s = ason.dump(n)
    restored = ason.load(s, Nested)
    assert_eq(n, restored)


def test_inline_comment():
    emp = ason.load(
        "{id,name,dept,skills,active}:/* inline */ (1,Alice,(HR),[rust],true)",
        Employee
    )
    assert_eq(emp.name, "Alice")


# ---------------------------------------------------------------------------
# Run all
# ---------------------------------------------------------------------------

def main():
    print("=== ASON Python Tests ===\n")

    tests = [
        # Basic
        ("dump struct", test_dump_struct),
        ("load struct", test_load_struct),
        ("load typed schema", test_load_typed_schema),
        ("roundtrip", test_roundtrip),
        # Vec
        ("vec load", test_vec_load),
        ("vec dump", test_vec_dump),
        ("vec roundtrip", test_vec_roundtrip),
        # Features
        ("optional with value", test_optional_with_value),
        ("optional none", test_optional_none),
        ("array field", test_array_field),
        ("nested struct", test_nested_struct),
        ("map field", test_map_field),
        ("escaped string", test_escaped_string),
        ("comments", test_comments),
        # Typed
        ("dump typed", test_dump_typed),
        ("dump typed roundtrip", test_dump_typed_roundtrip),
        ("dump slice typed", test_dump_slice_typed),
        ("dump slice typed auto", test_dump_slice_typed_auto),
        ("float roundtrip", test_float_roundtrip),
        ("float integer value", test_float_integer_value),
        ("negative numbers", test_negative_numbers),
        # Annotated
        ("load annotated", test_load_annotated),
        ("load unannotated", test_load_unannotated),
        ("load annotated nested", test_load_annotated_nested),
        ("load annotated vec", test_load_annotated_vec),
        ("multiline", test_multiline),
        # Edge cases
        ("empty vec field", test_empty_vec_field),
        ("special chars", test_special_chars),
        ("bool-like string", test_bool_like_string),
        ("number-like string", test_number_like_string),
        ("triple nested array", test_triple_nested_array),
        ("nested roundtrip", test_nested_roundtrip),
        ("inline comment", test_inline_comment),
    ]

    for name, fn in tests:
        test(name, fn)

    print(f"\n  {_passed} passed, {_failed} failed out of {_passed + _failed}")
    if _failed > 0:
        sys.exit(1)
    else:
        print("\n=== All tests passed! ===")


if __name__ == "__main__":
    main()
