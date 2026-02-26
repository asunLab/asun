"""ASON - Array-Schema Object Notation for Python.

High-performance, zero-copy where possible, schema-driven serialization.
Uses dataclasses and type hints for automatic field discovery.

Public API:
    dump(obj) -> str
    dump_typed(obj) -> str
    dump_slice(objs) -> str
    dump_slice_typed(objs, field_types) -> str
    load(data, cls) -> obj
    load_slice(data, cls) -> list[obj]
"""

from __future__ import annotations

import dataclasses
import math
from typing import Any, get_type_hints

__all__ = [
    "dump",
    "dump_typed",
    "dump_slice",
    "dump_slice_typed",
    "load",
    "load_slice",
    "AsonError",
    "DumpError",
    "LoadError",
]

# ---------------------------------------------------------------------------
# Errors
# ---------------------------------------------------------------------------

class AsonError(Exception):
    pass

class DumpError(AsonError):
    pass

class LoadError(AsonError):
    def __init__(self, pos: int, msg: str):
        self.pos = pos
        super().__init__(f"ason: load error at pos {pos}: {msg}")

# ---------------------------------------------------------------------------
# Lookup tables (module-level, built once)
# ---------------------------------------------------------------------------

_NEEDS_QUOTE = bytearray(256)
for _c in range(32):
    _NEEDS_QUOTE[_c] = 1
for _c in (ord(","), ord("("), ord(")"), ord("["), ord("]"), ord('"'), ord("\\")):
    _NEEDS_QUOTE[_c] = 1

_ESCAPE_MAP: dict[int, str] = {
    ord('"'): '\\"',
    ord("\\"): "\\\\",
    ord("\n"): "\\n",
    ord("\t"): "\\t",
}

_UNESCAPE_MAP: dict[int, int] = {
    ord('"'): ord('"'),
    ord("\\"): ord("\\"),
    ord("n"): ord("\n"),
    ord("t"): ord("\t"),
    ord(","): ord(","),
    ord("("): ord("("),
    ord(")"): ord(")"),
    ord("["): ord("["),
    ord("]"): ord("]"),
}

_DELIMS = frozenset((ord(","), ord(")"), ord("]")))
_WS = frozenset((ord(" "), ord("\t"), ord("\n"), ord("\r")))

# ---------------------------------------------------------------------------
# Struct info cache
# ---------------------------------------------------------------------------

class _FieldInfo:
    __slots__ = ("name", "type", "is_optional", "is_list", "is_dict",
                 "is_nested", "is_bool", "is_int", "is_float", "is_str",
                 "elem_type", "key_type", "val_type", "nested_cls",
                 "is_list_of_list", "inner_elem_type", "type_hint")

    def __init__(self, name: str, tp: type, resolved_hints: dict):
        self.name = name
        self.type = tp
        self.is_optional = False
        self.is_list = False
        self.is_dict = False
        self.is_nested = False
        self.is_bool = False
        self.is_int = False
        self.is_float = False
        self.is_str = False
        self.is_list_of_list = False
        self.elem_type = None
        self.inner_elem_type = None
        self.key_type = None
        self.val_type = None
        self.nested_cls = None
        self.type_hint = ""
        self._resolve(tp)

    def _resolve(self, tp):
        origin = getattr(tp, "__origin__", None)

        # Optional[X] = Union[X, None]
        if origin is type(None):
            return
        if _is_optional(tp):
            inner = _unwrap_optional(tp)
            self.is_optional = True
            self._resolve(inner)
            return

        if tp is bool:
            self.is_bool = True
            self.type_hint = "bool"
        elif tp is int:
            self.is_int = True
            self.type_hint = "int"
        elif tp is float:
            self.is_float = True
            self.type_hint = "float"
        elif tp is str:
            self.is_str = True
            self.type_hint = "str"
        elif origin is list:
            self.is_list = True
            args = getattr(tp, "__args__", None)
            if args:
                self.elem_type = args[0]
                # Check for list of list
                inner_origin = getattr(args[0], "__origin__", None)
                if inner_origin is list:
                    self.is_list_of_list = True
                    inner_args = getattr(args[0], "__args__", None)
                    if inner_args:
                        self.inner_elem_type = inner_args[0]
                # Check if elem is a dataclass (nested struct in list)
                if dataclasses.is_dataclass(args[0]) if isinstance(args[0], type) else False:
                    self.nested_cls = args[0]
        elif origin is dict:
            self.is_dict = True
            args = getattr(tp, "__args__", None)
            if args and len(args) == 2:
                self.key_type = args[0]
                self.val_type = args[1]
        elif isinstance(tp, type) and dataclasses.is_dataclass(tp):
            self.is_nested = True
            self.nested_cls = tp
        else:
            # Fallback: treat as string
            self.is_str = True
            self.type_hint = "str"


class _StructInfo:
    __slots__ = ("fields", "field_names", "cls")

    def __init__(self, cls: type):
        self.cls = cls
        self.field_names: list[str] = []
        self.fields: list[_FieldInfo] = []

        hints = get_type_hints(cls) if hasattr(cls, "__annotations__") else {}
        dc_fields = dataclasses.fields(cls) if dataclasses.is_dataclass(cls) else []

        for f in dc_fields:
            tp = hints.get(f.name, str)
            fi = _FieldInfo(f.name, tp, hints)
            self.fields.append(fi)
            self.field_names.append(f.name)


_struct_cache: dict[type, _StructInfo] = {}


def _get_struct_info(cls: type) -> _StructInfo:
    si = _struct_cache.get(cls)
    if si is not None:
        return si
    si = _StructInfo(cls)
    _struct_cache[cls] = si
    return si


def _is_optional(tp) -> bool:
    origin = getattr(tp, "__origin__", None)
    if origin is not None:
        import typing
        if origin is typing.Union:
            args = tp.__args__
            return len(args) == 2 and type(None) in args
    return False


def _unwrap_optional(tp):
    for a in tp.__args__:
        if a is not type(None):
            return a
    return str


# ---------------------------------------------------------------------------
# Dump (serialize)
# ---------------------------------------------------------------------------

def _string_needs_quoting(s: str) -> bool:
    if not s:
        return True
    if s[0] == " " or s[-1] == " ":
        return True
    if s == "true" or s == "false":
        return True
    b = s.encode("utf-8") if isinstance(s, str) else s
    could_be_number = True
    start = 1 if (len(b) > 0 and b[0] == ord("-")) else 0
    if start >= len(b):
        could_be_number = False
    for i in range(len(b)):
        if _NEEDS_QUOTE[b[i]]:
            return True
        if could_be_number and i >= start:
            c = b[i]
            if not (ord("0") <= c <= ord("9") or c == ord(".")):
                could_be_number = False
    return could_be_number and len(b) > start


def _append_escaped(parts: list, s: str):
    parts.append('"')
    for c in s:
        o = ord(c)
        esc = _ESCAPE_MAP.get(o)
        if esc:
            parts.append(esc)
        else:
            parts.append(c)
    parts.append('"')


def _append_str(parts: list, s: str):
    if _string_needs_quoting(s):
        _append_escaped(parts, s)
    else:
        parts.append(s)


def _append_float(parts: list, v: float):
    if math.isinf(v) or math.isnan(v):
        parts.append("0")
        return
    # Integer-valued float
    if v == int(v) and abs(v) < 1e18:
        parts.append(str(int(v)))
        parts.append(".0")
        return
    # Fast path: 1 decimal
    v10 = v * 10.0
    if v10 == int(v10) and abs(v10) < 1e18:
        iv = int(v10)
        sign = ""
        if iv < 0:
            sign = "-"
            iv = -iv
        int_part = iv // 10
        frac_part = iv % 10
        parts.append(f"{sign}{int_part}.{frac_part}")
        return
    # Fast path: 2 decimals
    v100 = v * 100.0
    if v100 == int(v100) and abs(v100) < 1e18:
        iv = int(v100)
        sign = ""
        if iv < 0:
            sign = "-"
            iv = -iv
        int_part = iv // 100
        frac_idx = iv % 100
        d1 = frac_idx // 10
        d2 = frac_idx % 10
        if d2 == 0:
            parts.append(f"{sign}{int_part}.{d1}")
        else:
            parts.append(f"{sign}{int_part}.{d1}{d2}")
        return
    # General case — repr gives full round-trip precision
    parts.append(repr(v))


def _dump_value(parts: list, val: Any, fi: _FieldInfo | None = None):
    """Serialize a single value to parts list."""
    if val is None:
        return  # empty = null

    if fi is not None:
        if fi.is_optional and val is None:
            return
        if fi.is_bool:
            parts.append("true" if val else "false")
            return
        if fi.is_int:
            parts.append(str(val))
            return
        if fi.is_float:
            _append_float(parts, val)
            return
        if fi.is_str:
            _append_str(parts, str(val))
            return
        if fi.is_list:
            _dump_list(parts, val, fi)
            return
        if fi.is_dict:
            _dump_dict(parts, val, fi)
            return
        if fi.is_nested:
            _dump_struct_data(parts, val)
            return

    # Runtime type dispatch
    _dump_any(parts, val)


def _dump_any(parts: list, val: Any):
    if val is None:
        return
    if isinstance(val, bool):
        parts.append("true" if val else "false")
    elif isinstance(val, int):
        parts.append(str(val))
    elif isinstance(val, float):
        _append_float(parts, val)
    elif isinstance(val, str):
        _append_str(parts, val)
    elif isinstance(val, (list, tuple)):
        parts.append("[")
        for i, item in enumerate(val):
            if i > 0:
                parts.append(",")
            _dump_any(parts, item)
        parts.append("]")
    elif isinstance(val, dict):
        parts.append("[")
        first = True
        for k, v in val.items():
            if not first:
                parts.append(",")
            first = False
            parts.append("(")
            _dump_any(parts, k)
            parts.append(",")
            _dump_any(parts, v)
            parts.append(")")
        parts.append("]")
    elif dataclasses.is_dataclass(val) and not isinstance(val, type):
        _dump_struct_data(parts, val)
    else:
        _append_str(parts, str(val))


def _dump_list(parts: list, lst: list, fi: _FieldInfo):
    parts.append("[")
    if fi.is_list_of_list:
        for i, inner in enumerate(lst):
            if i > 0:
                parts.append(",")
            parts.append("[")
            for j, item in enumerate(inner):
                if j > 0:
                    parts.append(",")
                _dump_any(parts, item)
            parts.append("]")
        parts.append("]")
        return
    if fi.nested_cls is not None:
        for i, item in enumerate(lst):
            if i > 0:
                parts.append(",")
            _dump_struct_data(parts, item)
        parts.append("]")
        return
    for i, item in enumerate(lst):
        if i > 0:
            parts.append(",")
        _dump_any(parts, item)
    parts.append("]")


def _dump_dict(parts: list, d: dict, fi: _FieldInfo):
    parts.append("[")
    first = True
    for k, v in d.items():
        if not first:
            parts.append(",")
        first = False
        parts.append("(")
        _dump_any(parts, k)
        parts.append(",")
        _dump_any(parts, v)
        parts.append(")")
    parts.append("]")


def _dump_struct_data(parts: list, obj: Any):
    si = _get_struct_info(type(obj))
    parts.append("(")
    for i, fi in enumerate(si.fields):
        if i > 0:
            parts.append(",")
        val = getattr(obj, fi.name)
        _dump_value(parts, val, fi)
    parts.append(")")


def dump(obj: Any) -> str:
    """Serialize a dataclass instance to ASON string.

    Output: {field1,field2,...}:(val1,val2,...)
    """
    if not dataclasses.is_dataclass(obj) or isinstance(obj, type):
        raise DumpError("dump requires a dataclass instance")

    si = _get_struct_info(type(obj))
    parts: list[str] = []

    # Data
    data_parts: list[str] = []
    data_parts.append("(")
    type_hints: list[str] = []
    for i, fi in enumerate(si.fields):
        if i > 0:
            data_parts.append(",")
        val = getattr(obj, fi.name)
        _dump_value(data_parts, val, fi)
        type_hints.append(fi.type_hint)
    data_parts.append(")")

    # Schema header
    parts.append("{")
    for i, fi in enumerate(si.fields):
        if i > 0:
            parts.append(",")
        parts.append(fi.name)
    parts.append("}:")

    parts.extend(data_parts)
    return "".join(parts)


def dump_typed(obj: Any) -> str:
    """Serialize a dataclass instance to ASON with type annotations.

    Output: {field1:type1,field2:type2,...}:(val1,val2,...)
    """
    if not dataclasses.is_dataclass(obj) or isinstance(obj, type):
        raise DumpError("dump_typed requires a dataclass instance")

    si = _get_struct_info(type(obj))
    parts: list[str] = []

    # Data first (to capture type hints for nested)
    data_parts: list[str] = []
    data_parts.append("(")
    type_hints: list[str] = []
    for i, fi in enumerate(si.fields):
        if i > 0:
            data_parts.append(",")
        val = getattr(obj, fi.name)
        _dump_value(data_parts, val, fi)
        type_hints.append(fi.type_hint)
    data_parts.append(")")

    # Schema header with types
    parts.append("{")
    for i, fi in enumerate(si.fields):
        if i > 0:
            parts.append(",")
        parts.append(fi.name)
        hint = type_hints[i] if i < len(type_hints) else ""
        if hint:
            parts.append(":")
            parts.append(hint)
    parts.append("}:")

    parts.extend(data_parts)
    return "".join(parts)


def dump_slice(objs: list) -> str:
    """Serialize a list of dataclass instances to ASON.

    Output: {field1,field2,...}:(v1,v2,...),(v3,v4,...)
    """
    if not objs:
        raise DumpError("dump_slice requires a non-empty list")
    first = objs[0]
    if not dataclasses.is_dataclass(first) or isinstance(first, type):
        raise DumpError("dump_slice requires dataclass instances")

    si = _get_struct_info(type(first))
    parts: list[str] = []

    # Schema header
    parts.append("[{")
    for i, fi in enumerate(si.fields):
        if i > 0:
            parts.append(",")
        parts.append(fi.name)
    parts.append("}]:")

    # Data rows
    for idx, obj in enumerate(objs):
        if idx > 0:
            parts.append(",")
        _dump_struct_data(parts, obj)
    return "".join(parts)


def dump_slice_typed(objs: list, field_types: list[str] | None = None) -> str:
    """Serialize a list of dataclass instances to ASON with type annotations.

    Output: [{field1:type1,field2:type2,...}]:(v1,v2,...),(v3,v4,...)
    """
    if not objs:
        raise DumpError("dump_slice_typed requires a non-empty list")
    first = objs[0]
    if not dataclasses.is_dataclass(first) or isinstance(first, type):
        raise DumpError("dump_slice_typed requires dataclass instances")

    si = _get_struct_info(type(first))
    parts: list[str] = []

    # Schema header with types
    parts.append("[{")
    for i, fi in enumerate(si.fields):
        if i > 0:
            parts.append(",")
        parts.append(fi.name)
        if field_types and i < len(field_types) and field_types[i]:
            parts.append(":")
            parts.append(field_types[i])
        elif fi.type_hint:
            parts.append(":")
            parts.append(fi.type_hint)
    parts.append("}]:")

    # Data rows
    for idx, obj in enumerate(objs):
        if idx > 0:
            parts.append(",")
        _dump_struct_data(parts, obj)
    return "".join(parts)


# ---------------------------------------------------------------------------
# Load (deserialize)
# ---------------------------------------------------------------------------

class _Decoder:
    __slots__ = ("data", "pos", "length")

    def __init__(self, data: bytes | str):
        if isinstance(data, str):
            self.data: bytes = data.encode("utf-8")
        else:
            self.data = data
        self.pos: int = 0
        self.length: int = len(self.data)

    def error(self, msg: str) -> LoadError:
        return LoadError(self.pos, msg)

    # -- Whitespace / comments --

    def skip_ws(self):
        pos = self.pos
        data = self.data
        length = self.length
        while pos < length and data[pos] in _WS:
            pos += 1
        self.pos = pos

    def skip_ws_comments(self):
        while True:
            self.skip_ws()
            if (self.pos + 1 < self.length
                    and self.data[self.pos] == ord("/")
                    and self.data[self.pos + 1] == ord("*")):
                self.pos += 2
                while self.pos + 1 < self.length:
                    if self.data[self.pos] == ord("*") and self.data[self.pos + 1] == ord("/"):
                        self.pos += 2
                        break
                    self.pos += 1
            else:
                return

    # -- Schema parsing --

    def parse_schema(self) -> list[str]:
        if self.data[self.pos] != ord("{"):
            raise self.error("expected '{'")
        self.pos += 1

        fields: list[str] = []
        while True:
            self.skip_ws()
            if self.pos >= self.length:
                raise self.error("unexpected EOF in schema")
            if self.data[self.pos] == ord("}"):
                self.pos += 1
                break

            if fields:
                if self.data[self.pos] != ord(","):
                    raise self.error("expected ','")
                self.pos += 1
                self.skip_ws()

            # Field name
            start = self.pos
            while self.pos < self.length:
                b = self.data[self.pos]
                if b in (ord(","), ord("}"), ord(":"), ord(" "), ord("\t")):
                    break
                self.pos += 1
            name = self.data[start:self.pos].decode("utf-8")
            self.skip_ws()

            # Skip optional type annotation
            if self.pos < self.length and self.data[self.pos] == ord(":"):
                self.pos += 1
                self.skip_ws()
                if self.pos < self.length and self.data[self.pos] == ord("{"):
                    self._skip_balanced(ord("{"), ord("}"))
                elif self.pos < self.length and self.data[self.pos] == ord("["):
                    self._skip_balanced(ord("["), ord("]"))
                elif (self.pos + 3 <= self.length
                      and self.data[self.pos:self.pos + 3] == b"map"):
                    self.pos += 3
                    if self.pos < self.length and self.data[self.pos] == ord("["):
                        self._skip_balanced(ord("["), ord("]"))
                else:
                    while self.pos < self.length:
                        b = self.data[self.pos]
                        if b in (ord(","), ord("}"), ord(" "), ord("\t")):
                            break
                        self.pos += 1

            fields.append(name)
        return fields

    def _skip_balanced(self, open_b: int, close_b: int):
        depth = 0
        while self.pos < self.length:
            b = self.data[self.pos]
            self.pos += 1
            if b == open_b:
                depth += 1
            elif b == close_b:
                depth -= 1
                if depth == 0:
                    return
        raise self.error("unbalanced brackets")

    # -- Value end check --

    def at_value_end(self) -> bool:
        if self.pos >= self.length:
            return True
        return self.data[self.pos] in _DELIMS

    # -- Primitive parsing --

    def parse_bool(self) -> bool:
        self.skip_ws_comments()
        if (self.pos + 4 <= self.length
                and self.data[self.pos:self.pos + 4] == b"true"):
            if self.pos + 4 >= self.length or self._is_delim(self.data[self.pos + 4]):
                self.pos += 4
                return True
        if (self.pos + 5 <= self.length
                and self.data[self.pos:self.pos + 5] == b"false"):
            if self.pos + 5 >= self.length or self._is_delim(self.data[self.pos + 5]):
                self.pos += 5
                return False
        raise self.error("invalid bool")

    @staticmethod
    def _is_delim(b: int) -> bool:
        return b in (ord(","), ord(")"), ord("]"), ord(" "), ord("\t"), ord("\n"), ord("\r"))

    def parse_int(self) -> int:
        self.skip_ws_comments()
        neg = False
        if self.pos < self.length and self.data[self.pos] == ord("-"):
            neg = True
            self.pos += 1
        val = 0
        digits = 0
        while self.pos < self.length and ord("0") <= self.data[self.pos] <= ord("9"):
            val = val * 10 + (self.data[self.pos] - ord("0"))
            self.pos += 1
            digits += 1
        if digits == 0:
            raise self.error("invalid number")
        return -val if neg else val

    def parse_float(self) -> float:
        self.skip_ws_comments()
        start = self.pos
        if self.pos < self.length and self.data[self.pos] == ord("-"):
            self.pos += 1
        while self.pos < self.length and ord("0") <= self.data[self.pos] <= ord("9"):
            self.pos += 1
        if self.pos < self.length and self.data[self.pos] == ord("."):
            self.pos += 1
            while self.pos < self.length and ord("0") <= self.data[self.pos] <= ord("9"):
                self.pos += 1
        # Handle scientific notation
        if self.pos < self.length and self.data[self.pos] in (ord("e"), ord("E")):
            self.pos += 1
            if self.pos < self.length and self.data[self.pos] in (ord("+"), ord("-")):
                self.pos += 1
            while self.pos < self.length and ord("0") <= self.data[self.pos] <= ord("9"):
                self.pos += 1
        if self.pos == start:
            raise self.error("invalid float")
        s = self.data[start:self.pos]
        try:
            return float(s)
        except ValueError:
            raise self.error(f"invalid float: {s!r}")

    def parse_string(self) -> str:
        self.skip_ws_comments()
        if self.pos >= self.length:
            return ""
        if self.data[self.pos] == ord('"'):
            return self._parse_quoted()
        return self._parse_plain()

    def _parse_plain(self) -> str:
        start = self.pos
        has_escape = False
        while self.pos < self.length:
            b = self.data[self.pos]
            if b in _DELIMS:
                break
            if b == ord("\\"):
                has_escape = True
                self.pos += 2
                continue
            self.pos += 1
        raw = self.data[start:self.pos]
        # Trim whitespace
        raw = raw.strip()
        if has_escape:
            return self._unescape(raw)
        # Zerocopy-ish: decode directly
        return raw.decode("utf-8")

    def _parse_quoted(self) -> str:
        self.pos += 1  # skip "
        start = self.pos

        # Fast scan for closing quote without escapes
        scan = self.pos
        while scan < self.length:
            if self.data[scan] == ord('"'):
                # No escapes — direct decode
                s = self.data[start:scan].decode("utf-8")
                self.pos = scan + 1
                return s
            if self.data[scan] == ord("\\"):
                break
            scan += 1

        # Slow path: has escapes
        parts: list[int] = []
        if scan > start:
            parts.extend(self.data[start:scan])
        self.pos = scan

        while self.pos < self.length:
            b = self.data[self.pos]
            if b == ord('"'):
                self.pos += 1
                return bytes(parts).decode("utf-8")
            if b == ord("\\"):
                self.pos += 1
                if self.pos >= self.length:
                    raise self.error("unclosed string")
                esc = self.data[self.pos]
                self.pos += 1
                mapped = _UNESCAPE_MAP.get(esc)
                if mapped is not None:
                    parts.append(mapped)
                elif esc == ord("u"):
                    if self.pos + 4 > self.length:
                        raise self.error("invalid unicode escape")
                    hex_str = self.data[self.pos:self.pos + 4].decode("ascii")
                    cp = int(hex_str, 16)
                    parts.extend(chr(cp).encode("utf-8"))
                    self.pos += 4
                else:
                    raise self.error(f"invalid escape: \\{chr(esc)}")
            else:
                parts.append(b)
                self.pos += 1
        raise self.error("unclosed string")

    def _unescape(self, raw: bytes) -> str:
        parts: list[int] = []
        i = 0
        while i < len(raw):
            if raw[i] == ord("\\"):
                i += 1
                if i >= len(raw):
                    raise self.error("unexpected EOF in escape")
                mapped = _UNESCAPE_MAP.get(raw[i])
                if mapped is not None:
                    parts.append(mapped)
                elif raw[i] == ord("u"):
                    if i + 4 >= len(raw):
                        raise self.error("invalid unicode escape")
                    hex_str = raw[i + 1:i + 5].decode("ascii")
                    cp = int(hex_str, 16)
                    parts.extend(chr(cp).encode("utf-8"))
                    i += 4
                else:
                    raise self.error(f"invalid escape: \\{chr(raw[i])}")
            else:
                parts.append(raw[i])
            i += 1
        return bytes(parts).decode("utf-8")

    # -- Struct loading --

    def load_top(self, cls: type) -> Any:
        si = _get_struct_info(cls)

        if self.pos >= self.length or self.data[self.pos] != ord("{"):
            raise self.error("expected '{'")
        schema_fields = self.parse_schema()

        self.skip_ws_comments()
        if self.pos >= self.length or self.data[self.pos] != ord(":"):
            raise self.error("expected ':'")
        self.pos += 1
        self.skip_ws_comments()

        field_map = self._build_field_map(si, schema_fields)
        return self._load_tuple(cls, si, field_map)

    def _build_field_map(self, si: _StructInfo, schema_fields: list[str]) -> list[int]:
        name_to_idx: dict[str, int] = {}
        for i, fi in enumerate(si.fields):
            name_to_idx[fi.name] = i
        return [name_to_idx.get(name, -1) for name in schema_fields]

    def _load_tuple(self, cls: type, si: _StructInfo, field_map: list[int]) -> Any:
        self.skip_ws_comments()
        if self.pos >= self.length or self.data[self.pos] != ord("("):
            raise self.error("expected '('")
        self.pos += 1

        values: dict[str, Any] = {}
        for i in range(len(field_map)):
            self.skip_ws_comments()
            if self.pos >= self.length:
                raise self.error("unexpected EOF in tuple")
            if self.data[self.pos] == ord(")"):
                break
            if i > 0:
                if self.data[self.pos] == ord(","):
                    self.pos += 1
                    self.skip_ws_comments()
                    if self.pos < self.length and self.data[self.pos] == ord(")"):
                        break
                elif self.data[self.pos] == ord(")"):
                    break
                else:
                    raise self.error("expected ',' or ')'")

            fi_idx = field_map[i]
            if fi_idx < 0:
                self._skip_value()
                continue

            fi = si.fields[fi_idx]
            val = self._load_field_value(fi)
            values[fi.name] = val

        self.skip_ws_comments()
        if self.pos < self.length and self.data[self.pos] == ord(")"):
            self.pos += 1

        # Construct object
        return cls(**values)

    def _load_field_value(self, fi: _FieldInfo) -> Any:
        self.skip_ws_comments()
        if self.pos >= self.length:
            return None

        # Optional: check for empty value
        if fi.is_optional:
            if self.at_value_end():
                return None
            # Recurse with the inner type info
            inner_fi = _FieldInfo.__new__(_FieldInfo)
            inner_fi.name = fi.name
            inner_fi.is_optional = False
            inner_fi.is_bool = fi.is_bool
            inner_fi.is_int = fi.is_int
            inner_fi.is_float = fi.is_float
            inner_fi.is_str = fi.is_str
            inner_fi.is_list = fi.is_list
            inner_fi.is_dict = fi.is_dict
            inner_fi.is_nested = fi.is_nested
            inner_fi.is_list_of_list = fi.is_list_of_list
            inner_fi.elem_type = fi.elem_type
            inner_fi.inner_elem_type = fi.inner_elem_type
            inner_fi.key_type = fi.key_type
            inner_fi.val_type = fi.val_type
            inner_fi.nested_cls = fi.nested_cls
            inner_fi.type_hint = fi.type_hint
            return self._load_field_value(inner_fi)

        if fi.is_bool:
            return self.parse_bool()
        if fi.is_int:
            return self.parse_int()
        if fi.is_float:
            return self.parse_float()
        if fi.is_str:
            return self.parse_string()
        if fi.is_list:
            return self._load_list(fi)
        if fi.is_dict:
            return self._load_dict(fi)
        if fi.is_nested:
            return self._load_nested_struct(fi.nested_cls)

        # Fallback
        return self.parse_string()

    def _load_list(self, fi: _FieldInfo) -> list:
        self.skip_ws_comments()
        if self.pos >= self.length or self.data[self.pos] != ord("["):
            raise self.error("expected '['")
        self.pos += 1

        result: list = []
        first = True
        while True:
            self.skip_ws_comments()
            if self.pos >= self.length or self.data[self.pos] == ord("]"):
                self.pos += 1
                break
            if not first:
                if self.data[self.pos] == ord(","):
                    self.pos += 1
                    self.skip_ws_comments()
                    if self.pos < self.length and self.data[self.pos] == ord("]"):
                        self.pos += 1
                        break
                else:
                    break
            first = False

            if fi.is_list_of_list:
                # Parse inner list
                inner = self._load_inner_list(fi.inner_elem_type)
                result.append(inner)
            elif fi.nested_cls is not None:
                elem = self._load_nested_struct(fi.nested_cls)
                result.append(elem)
            elif fi.elem_type is not None:
                elem = self._load_typed_value(fi.elem_type)
                result.append(elem)
            else:
                elem = self._parse_any_value()
                result.append(elem)
        return result

    def _load_inner_list(self, elem_type) -> list:
        self.skip_ws_comments()
        if self.pos >= self.length or self.data[self.pos] != ord("["):
            raise self.error("expected '['")
        self.pos += 1
        result: list = []
        first = True
        while True:
            self.skip_ws_comments()
            if self.pos >= self.length or self.data[self.pos] == ord("]"):
                self.pos += 1
                break
            if not first:
                if self.data[self.pos] == ord(","):
                    self.pos += 1
                    self.skip_ws_comments()
                    if self.pos < self.length and self.data[self.pos] == ord("]"):
                        self.pos += 1
                        break
                else:
                    break
            first = False

            # Check for another nested list
            inner_origin = getattr(elem_type, "__origin__", None)
            if inner_origin is list:
                inner_args = getattr(elem_type, "__args__", None)
                inner_elem = inner_args[0] if inner_args else None
                elem = self._load_inner_list(inner_elem)
            else:
                elem = self._load_typed_value(elem_type)
            result.append(elem)
        return result

    def _load_dict(self, fi: _FieldInfo) -> dict:
        self.skip_ws_comments()
        if self.pos >= self.length or self.data[self.pos] != ord("["):
            raise self.error("expected '['")
        self.pos += 1

        result: dict = {}
        first = True
        while True:
            self.skip_ws_comments()
            if self.pos >= self.length or self.data[self.pos] == ord("]"):
                self.pos += 1
                break
            if not first:
                if self.data[self.pos] == ord(","):
                    self.pos += 1
                    self.skip_ws_comments()
                    if self.pos < self.length and self.data[self.pos] == ord("]"):
                        self.pos += 1
                        break
                else:
                    break
            first = False

            if self.pos >= self.length or self.data[self.pos] != ord("("):
                raise self.error("expected '(' in map entry")
            self.pos += 1

            key = self._load_typed_value(fi.key_type) if fi.key_type else self.parse_string()
            self.skip_ws_comments()
            if self.pos < self.length and self.data[self.pos] == ord(","):
                self.pos += 1
            val = self._load_typed_value(fi.val_type) if fi.val_type else self.parse_string()

            self.skip_ws_comments()
            if self.pos < self.length and self.data[self.pos] == ord(")"):
                self.pos += 1

            result[key] = val
        return result

    def _load_nested_struct(self, cls: type) -> Any:
        self.skip_ws_comments()
        if self.pos >= self.length:
            raise self.error("unexpected EOF")

        if self.data[self.pos] == ord("{"):
            # Inline schema
            si = _get_struct_info(cls)
            fields = self.parse_schema()
            self.skip_ws_comments()
            if self.pos >= self.length or self.data[self.pos] != ord(":"):
                raise self.error("expected ':'")
            self.pos += 1
            self.skip_ws_comments()
            field_map = self._build_field_map(si, fields)
            return self._load_tuple(cls, si, field_map)

        if self.data[self.pos] == ord("("):
            # Positional
            si = _get_struct_info(cls)
            field_map = list(range(len(si.fields)))
            return self._load_tuple(cls, si, field_map)

        raise self.error("expected '{' or '(' for struct")

    def _load_typed_value(self, tp) -> Any:
        if tp is bool:
            return self.parse_bool()
        if tp is int:
            return self.parse_int()
        if tp is float:
            return self.parse_float()
        if tp is str:
            return self.parse_string()
        if isinstance(tp, type) and dataclasses.is_dataclass(tp):
            return self._load_nested_struct(tp)
        # Check for list/dict origins
        origin = getattr(tp, "__origin__", None)
        if origin is list:
            args = getattr(tp, "__args__", None)
            fi = _FieldInfo.__new__(_FieldInfo)
            fi.is_list = True
            fi.is_list_of_list = False
            fi.nested_cls = None
            fi.elem_type = args[0] if args else None
            fi.inner_elem_type = None
            if args:
                inner_origin = getattr(args[0], "__origin__", None)
                if inner_origin is list:
                    fi.is_list_of_list = True
                    inner_args = getattr(args[0], "__args__", None)
                    fi.inner_elem_type = inner_args[0] if inner_args else None
                elif isinstance(args[0], type) and dataclasses.is_dataclass(args[0]):
                    fi.nested_cls = args[0]
            return self._load_list(fi)
        if origin is dict:
            args = getattr(tp, "__args__", None)
            fi = _FieldInfo.__new__(_FieldInfo)
            fi.is_dict = True
            fi.key_type = args[0] if args else str
            fi.val_type = args[1] if args and len(args) > 1 else str
            return self._load_dict(fi)
        return self.parse_string()

    def _parse_any_value(self) -> Any:
        self.skip_ws_comments()
        if self.pos >= self.length or self.at_value_end():
            return None
        b = self.data[self.pos]
        if b == ord('"'):
            return self._parse_quoted()
        if b == ord("["):
            self.pos += 1
            arr: list = []
            first = True
            while True:
                self.skip_ws_comments()
                if self.pos >= self.length or self.data[self.pos] == ord("]"):
                    self.pos += 1
                    break
                if not first:
                    if self.data[self.pos] == ord(","):
                        self.pos += 1
                        self.skip_ws_comments()
                        if self.pos < self.length and self.data[self.pos] == ord("]"):
                            self.pos += 1
                            break
                first = False
                arr.append(self._parse_any_value())
            return arr
        if b == ord("("):
            self.pos += 1
            arr = []
            first = True
            while True:
                self.skip_ws_comments()
                if self.pos >= self.length or self.data[self.pos] == ord(")"):
                    self.pos += 1
                    break
                if not first:
                    if self.data[self.pos] == ord(","):
                        self.pos += 1
                        self.skip_ws_comments()
                        if self.pos < self.length and self.data[self.pos] == ord(")"):
                            self.pos += 1
                            break
                first = False
                arr.append(self._parse_any_value())
            return arr
        if b in (ord("t"), ord("f")):
            try:
                return self.parse_bool()
            except LoadError:
                return self._parse_plain()
        if b == ord("-") and self.pos + 1 < self.length and ord("0") <= self.data[self.pos + 1] <= ord("9"):
            return self._parse_number_any()
        if ord("0") <= b <= ord("9"):
            return self._parse_number_any()
        return self._parse_plain()

    def _parse_number_any(self) -> int | float:
        start = self.pos
        if self.pos < self.length and self.data[self.pos] == ord("-"):
            self.pos += 1
        while self.pos < self.length and ord("0") <= self.data[self.pos] <= ord("9"):
            self.pos += 1
        is_float = False
        if self.pos < self.length and self.data[self.pos] == ord("."):
            is_float = True
            self.pos += 1
            while self.pos < self.length and ord("0") <= self.data[self.pos] <= ord("9"):
                self.pos += 1
        s = self.data[start:self.pos].decode("ascii")
        if is_float:
            return float(s)
        return int(s)

    def _skip_value(self):
        self.skip_ws_comments()
        if self.pos >= self.length:
            return
        b = self.data[self.pos]
        if b == ord('"'):
            self.pos += 1
            while self.pos < self.length:
                if self.data[self.pos] == ord('"'):
                    self.pos += 1
                    return
                if self.data[self.pos] == ord("\\"):
                    self.pos += 1
                self.pos += 1
            raise self.error("unclosed string")
        if b == ord("("):
            self._skip_balanced(ord("("), ord(")"))
            return
        if b == ord("["):
            self._skip_balanced(ord("["), ord("]"))
            return
        while self.pos < self.length:
            if self.data[self.pos] in _DELIMS:
                return
            self.pos += 1


# ---------------------------------------------------------------------------
# Public load API
# ---------------------------------------------------------------------------

def load(data: str | bytes, cls: type) -> Any:
    """Deserialize an ASON string into a dataclass instance.

    Input: {field1,field2,...}:(val1,val2,...)
    """
    d = _Decoder(data)
    d.skip_ws_comments()
    return d.load_top(cls)


def load_slice(data: str | bytes, cls: type) -> list:
    """Deserialize an ASON string into a list of dataclass instances.

    Input: [{field1,field2,...}]:(v1,v2,...),(v3,v4,...)
    """
    d = _Decoder(data)
    d.skip_ws_comments()

    if d.pos >= d.length or d.data[d.pos] != ord("["):
        raise d.error("expected '['")
    d.pos += 1
    d.skip_ws_comments()

    if d.pos >= d.length or d.data[d.pos] != ord("{"):
        raise d.error("expected '{'")
    schema_fields = d.parse_schema()

    d.skip_ws_comments()
    if d.pos >= d.length or d.data[d.pos] != ord("]"):
        raise d.error("expected ']'")
    d.pos += 1

    d.skip_ws_comments()
    if d.pos >= d.length or d.data[d.pos] != ord(":"):
        raise d.error("expected ':'")
    d.pos += 1

    si = _get_struct_info(cls)
    field_map = d._build_field_map(si, schema_fields)

    result: list = []
    while True:
        d.skip_ws_comments()
        if d.pos >= d.length:
            break
        if d.data[d.pos] != ord("("):
            break

        obj = d._load_tuple(cls, si, field_map)
        result.append(obj)

        d.skip_ws_comments()
        if d.pos < d.length and d.data[d.pos] == ord(","):
            d.pos += 1
            d.skip_ws_comments()
            if d.pos >= d.length or d.data[d.pos] != ord("("):
                break
    return result
