# ASON Java Library

A pure Java implementation of ASON (Array-Schema Object Notation).

## Requirements

- Java 11+
- Maven 3.6+

## Building

```bash
mvn compile
```

## Usage

```java
import io.github.athxx.ason.*;

// Parse ASON
Value result = Ason.parse("{name,age}:(Alice,30)");
System.out.println(result.get("name").asString()); // Alice
System.out.println(result.get("age").asInteger()); // 30

// Parse multiple objects
Value users = Ason.parse("{name,age}:(Alice,30),(Bob,25)");
for (Value user : users.items()) {
    System.out.println(user.get("name").asString());
}

// Serialize
Value obj = Value.ofObject();
obj.set("name", Value.ofString("Alice"));
obj.set("age", Value.ofInteger(30));

System.out.println(Ason.serialize(obj));           // (Alice,30)
System.out.println(Ason.serializeWithSchema(obj)); // {name,age}:(Alice,30)
```

## API

### Ason class

- `Ason.parse(String input)` - Parse ASON string
- `Ason.serialize(Value value)` - Serialize to ASON (data only)
- `Ason.serializeWithSchema(Value value)` - Serialize with schema

### Value class

Factory methods:
- `Value.ofNull()` - Create null value
- `Value.ofBool(boolean)` - Create boolean value
- `Value.ofInteger(long)` - Create integer value
- `Value.ofFloat(double)` - Create float value
- `Value.ofString(String)` - Create string value
- `Value.ofArray()` - Create empty array
- `Value.ofObject()` - Create empty object

Type checks:
- `isNull()`, `isBool()`, `isInteger()`, `isFloat()`, `isString()`, `isArray()`, `isObject()`

Value getters:
- `asBool()`, `asInteger()`, `asFloat()`, `asString()`

Array operations:
- `size()` - Get array/object size
- `get(int index)` - Get array element
- `push(Value)` - Add to array
- `items()` - Get all array items

Object operations:
- `get(String key)` - Get object field
- `set(String key, Value)` - Set object field
- `keys()` - Get all keys

## License

MIT License

