# ASON Python Library

A pure Python implementation of ASON (Array-Schema Object Notation).

## Installation

### From source

```bash
cd python
pip install .
```

### Development mode

```bash
pip install -e .
```

## Usage

```python
import ason

# Parse ASON string
result = ason.parse("{name,age}:(Alice,30)")
print(result["name"])  # "Alice"
print(result["age"])   # 30

# Parse multiple objects
result = ason.parse("{name,age}:(Alice,30),(Bob,25)")
for obj in result:
    print(f"{obj['name']}: {obj['age']}")

# Parse nested objects
result = ason.parse("{name,addr{city,zip}}:(Alice,(NYC,10001))")
print(result["addr"]["city"])  # "NYC"

# Parse arrays
result = ason.parse("{name,scores[]}:(Alice,[90,85,95])")
print(result["scores"])  # [90, 85, 95]

# Serialize Python dict to ASON
obj = {"name": "Alice", "age": 30}
print(ason.serialize(obj))  # (Alice,30)

# Serialize with schema
print(ason.serialize_with_schema(obj))  # {name,age}:(Alice,30)
```

## Testing

```bash
pip install pytest
pytest tests/ -v
```

## License

MIT License

