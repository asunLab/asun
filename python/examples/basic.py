"""ASON Basic Examples — matches Rust basic.rs."""

import sys
import os
sys.path.insert(0, os.path.join(os.path.dirname(__file__), ".."))

from dataclasses import dataclass
from typing import Optional
import ason


@dataclass
class User:
    id: int
    name: str
    active: bool


def main():
    print("=== ASON Basic Examples ===\n")

    # 1. Serialize a single struct
    user = User(id=1, name="Alice", active=True)
    ason_str = ason.dump(user)
    print("Serialize single struct:")
    print(f"  {ason_str}\n")

    # 2. Serialize with type annotations (dump_typed)
    typed_str = ason.dump_typed(user)
    print("Serialize with type annotations:")
    print(f"  {typed_str}\n")
    assert typed_str.startswith("{id:int,name:str,active:bool}:")

    # 3. Deserialize from ASON (accepts both annotated and unannotated)
    inp = "{id:int,name:str,active:bool}:(1,Alice,true)"
    user = ason.load(inp, User)
    print("Deserialize single struct:")
    print(f"  {user}\n")

    # 4. Serialize a list of structs (schema-driven)
    users = [
        User(id=1, name="Alice", active=True),
        User(id=2, name="Bob", active=False),
        User(id=3, name="Carol Smith", active=True),
    ]
    ason_vec = ason.dump_slice(users)
    print("Serialize list (schema-driven):")
    print(f"  {ason_vec}\n")

    # 5. Serialize list with type annotations (dump_slice_typed)
    typed_vec = ason.dump_slice_typed(users)
    print("Serialize list with type annotations:")
    print(f"  {typed_vec}\n")
    assert typed_vec.startswith("[{id:int,name:str,active:bool}]:")

    # 6. Deserialize list
    inp = '[{id:int,name:str,active:bool}]:(1,Alice,true),(2,Bob,false),(3,"Carol Smith",true)'
    users = ason.load_slice(inp, User)
    print("Deserialize list:")
    for u in users:
        print(f"  {u}")

    # 7. Multiline format
    print("\nMultiline format:")
    multiline = """[{id:int, name:str, active:bool}]:
  (1, Alice, true),
  (2, Bob, false),
  (3, "Carol Smith", true)"""
    users = ason.load_slice(multiline, User)
    for u in users:
        print(f"  {u}")

    # 8. Roundtrip
    print("\nRoundtrip test:")
    original = User(id=42, name="Test User", active=True)
    serialized = ason.dump(original)
    deserialized = ason.load(serialized, User)
    print(f"  original:     {original}")
    print(f"  serialized:   {serialized}")
    print(f"  deserialized: {deserialized}")
    assert original == deserialized
    print("  ✓ roundtrip OK")

    # 9. Optional fields
    print("\nOptional fields:")

    @dataclass
    class Item:
        id: int
        label: Optional[str] = None

    item = ason.load("{id,label}:(1,hello)", Item)
    print(f"  with value: {item}")

    item = ason.load("{id,label}:(2,)", Item)
    print(f"  with null:  {item}")

    # 10. Array fields
    print("\nArray fields:")

    @dataclass
    class Tagged:
        name: str
        tags: list[str]

    t = ason.load("{name,tags}:(Alice,[rust,go,python])", Tagged)
    print(f"  {t}")

    # 11. Comments
    print("\nWith comments:")
    inp = "/* user list */ {id,name,active}:(1,Alice,true)"
    user = ason.load(inp, User)
    print(f"  {user}")

    print("\n=== All examples passed! ===")


if __name__ == "__main__":
    main()
