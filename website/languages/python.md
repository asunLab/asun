# Python Guide

The Python ASON implementation is a single-file library (`ason.py`) with zero external dependencies — works with the Python standard library only.

## Installation

Copy `ason.py` into your project:

```bash
cp /path/to/ason/python/ason.py ./ason.py
```

Or install from the repo:

```bash
pip install git+https://github.com/your-org/ason.git#subdirectory=python
```

## Encoding

```python
from ason import encode, encode_rows
from dataclasses import dataclass

@dataclass
class User:
    id: int
    name: str
    active: bool

# Single struct
user = User(1, "Alice", True)
text = encode(user)
# → "{id,name,active}:(1,Alice,true)"

# List
users = [User(1, "Alice", True), User(2, "Bob", False)]
text = encode_rows(users)
# → "{id,name,active}:\n  (1,Alice,true),\n  (2,Bob,false)"
```

## Decoding

```python
from ason import decode, decode_rows

# Single struct
user = decode(text, User)

# List
users = decode_rows(text, User)
```

## Type Support

| Python type | ASON type |
|-------------|-----------|
| `int` | `int` |
| `float` | `float` |
| `bool` | `bool` |
| `str` | `str` |
| `None` | empty slot |
| `list` | `[type]` |
| `dataclass` | nested schema |
| `Optional[T]` | nullable slot |

## Working with Dicts

If you don't have a dataclass, use `encode_dict_rows`:

```python
from ason import encode_dict_rows, decode_dict_rows

rows = [
    {"id": 1, "name": "Alice", "active": True},
    {"id": 2, "name": "Bob",   "active": False},
]
text = encode_dict_rows(rows)
back = decode_dict_rows(text)
```

## Running Tests

```bash
cd python && python -m pytest test_ason.py -v
```

## Running Examples

```bash
cd python/examples
python basic.py
python bench.py
python complex.py
```
