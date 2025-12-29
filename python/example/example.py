#!/usr/bin/env python3
"""ASON Python Example"""

import sys
import os
sys.path.insert(0, os.path.dirname(os.path.dirname(os.path.abspath(__file__))))

from parser import parse
from serializer import serialize, serialize_with_schema

def main():
    print("=== ASON Python Example ===\n")
    
    # 1. Parse simple object
    print("1. Parse simple object:")
    result = parse("{name,age}:(Alice,30)")
    print("   Input:  {name,age}:(Alice,30)")
    print(f"   Output: {result}")
    print(f"   name = {result['name']}, age = {result['age']}")
    print()
    
    # 2. Parse multiple objects
    print("2. Parse multiple objects:")
    users = parse("{name,age}:(Alice,30),(Bob,25),(Carol,28)")
    print("   Input:  {name,age}:(Alice,30),(Bob,25),(Carol,28)")
    for i, user in enumerate(users):
        print(f"   [{i}] {user}")
    print()
    
    # 3. Parse nested object
    print("3. Parse nested object:")
    result = parse("{name,addr{city,zip}}:(Alice,(NYC,10001))")
    print("   Input:  {name,addr{city,zip}}:(Alice,(NYC,10001))")
    print(f"   Output: {result}")
    print(f"   addr.city = {result['addr']['city']}")
    print()
    
    # 4. Parse array field
    print("4. Parse array field:")
    result = parse("{name,scores[]}:(Alice,[90,85,95])")
    print("   Input:  {name,scores[]}:(Alice,[90,85,95])")
    print(f"   Output: {result}")
    print(f"   scores = {result['scores']}")
    print()
    
    # 5. Parse array of objects
    print("5. Parse array of objects:")
    result = parse("{name,items[{id,qty}]}:(Order1,[(A,2),(B,3)])")
    print("   Input:  {name,items[{id,qty}]}:(Order1,[(A,2),(B,3)])")
    print(f"   Output: {result}")
    print()
    
    # 6. Unicode support
    print("6. Unicode support:")
    result = parse("{name,city}:(小明,北京)")
    print("   Input:  {name,city}:(小明,北京)")
    print(f"   Output: {result}")
    print()
    
    # 7. Serialize object
    print("7. Serialize object:")
    obj = {"name": "Alice", "age": 30, "active": True}
    print(f"   Input:  {obj}")
    print(f"   serialize():           {serialize(obj)}")
    print(f"   serialize_with_schema(): {serialize_with_schema(obj)}")
    print()
    
    # 8. Serialize array of objects
    print("8. Serialize array of objects:")
    users = [
        {"name": "Alice", "age": 30},
        {"name": "Bob", "age": 25}
    ]
    print(f"   Input:  {users}")
    print(f"   Output: {serialize_with_schema(users)}")
    print()
    
    print("=== Done ===")

if __name__ == "__main__":
    main()

