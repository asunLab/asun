"""ASON Serializer."""

from typing import Any, Dict, List, Union

AsonValue = Union[None, bool, int, float, str, List[Any], Dict[str, Any]]


def serialize(obj: AsonValue) -> str:
    """Serialize a Python object to ASON format (data only)."""
    return _serialize_value(obj)


def serialize_with_schema(obj: AsonValue) -> str:
    """Serialize a Python object to ASON format with schema."""
    if obj is None:
        return "null"
    if not isinstance(obj, (dict, list)):
        return _serialize_value(obj)
    
    # Handle list of objects
    if isinstance(obj, list) and obj and isinstance(obj[0], dict):
        first = obj[0]
        schema = _build_schema(first)
        data_parts = [_serialize_object_data(item) for item in obj]
        return f"{schema}:{','.join(data_parts)}"
    
    # Single object
    if isinstance(obj, dict):
        schema = _build_schema(obj)
        data = _serialize_object_data(obj)
        return f"{schema}:{data}"

    return _serialize_value(obj)


def _serialize_value(v: AsonValue) -> str:
    if v is None:
        return "null"
    if isinstance(v, bool):
        return "true" if v else "false"
    if isinstance(v, int):
        return str(v)
    if isinstance(v, float):
        return str(v)
    if isinstance(v, str):
        return _serialize_string(v)
    if isinstance(v, list):
        items = [_serialize_value(item) for item in v]
        return f"[{','.join(items)}]"
    if isinstance(v, dict):
        items = [_serialize_value(val) for val in v.values()]
        return f"({','.join(items)})"
    return "null"


def _serialize_string(s: str) -> str:
    # Check if quoting is needed
    if not s:
        return '""'

    needs_quote = False
    for ch in s:
        if ch in '"\\()[]{},:' or ch.isspace():
            needs_quote = True
            break

    # Check for keywords and numbers
    if not needs_quote:
        if s in ("null", "true", "false"):
            needs_quote = True
        elif s and (s[0] in "+-" or s[0].isdigit()):
            needs_quote = True

    if not needs_quote:
        return s

    # Quote and escape
    result = ['"']
    for ch in s:
        if ch == '"':
            result.append('\\"')
        elif ch == '\\':
            result.append('\\\\')
        elif ch == '\n':
            result.append('\\n')
        elif ch == '\r':
            result.append('\\r')
        elif ch == '\t':
            result.append('\\t')
        else:
            result.append(ch)
    result.append('"')
    return ''.join(result)


def _build_schema(obj: Dict[str, Any]) -> str:
    parts = []
    for key, val in obj.items():
        if isinstance(val, list):
            if val and isinstance(val[0], dict):
                inner_schema = _build_schema(val[0])
                parts.append(f"{key}[{inner_schema}]")
            else:
                parts.append(f"{key}[]")
        elif isinstance(val, dict):
            inner_schema = _build_schema(val)
            parts.append(f"{key}{inner_schema}")
        else:
            parts.append(key)
    return "{" + ",".join(parts) + "}"


def _serialize_object_data(obj: Dict[str, Any]) -> str:
    parts = []
    for val in obj.values():
        if isinstance(val, dict):
            parts.append(_serialize_object_data(val))
        elif isinstance(val, list):
            parts.append(_serialize_array_data(val))
        else:
            parts.append(_serialize_value(val))
    return f"({','.join(parts)})"


def _serialize_array_data(arr: List[Any]) -> str:
    parts = []
    for item in arr:
        if isinstance(item, dict):
            parts.append(_serialize_object_data(item))
        else:
            parts.append(_serialize_value(item))
    return f"[{','.join(parts)}]"

