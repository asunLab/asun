# ASON Kotlin Library

A pure Kotlin implementation of ASON (Array-Schema Object Notation).

## Requirements

- Kotlin 1.9+
- JDK 11+

## Building

```bash
./gradlew build
```

## Usage

```kotlin
import io.github.athxx.ason.*

// Parse ASON
val result = Ason.parse("{name,age}:(Alice,30)")
println(result["name"]?.asString()) // Alice
println(result["age"]?.asInteger()) // 30

// Parse multiple objects
val users = Ason.parse("{name,age}:(Alice,30),(Bob,25)")
users.asArray()?.items?.forEach { user ->
    println(user.asObject()?.get("name")?.asString())
}

// Serialize
val obj = Value.ofObject().apply {
    this["name"] = Value.ofString("Alice")
    this["age"] = Value.ofInteger(30)
}

println(Ason.serialize(obj))           // (Alice,30)
println(Ason.serializeWithSchema(obj)) // {name,age}:(Alice,30)
```

## API

### Ason object

- `Ason.parse(input: String): Value` - Parse ASON string
- `Ason.serialize(value: Value): String` - Serialize to ASON (data only)
- `Ason.serializeWithSchema(value: Value): String` - Serialize with schema

### Value sealed class

Types:
- `Value.Null` - Null value
- `Value.Bool(value: Boolean)` - Boolean value
- `Value.Integer(value: Long)` - Integer value
- `Value.Float(value: Double)` - Float value
- `Value.Str(value: String)` - String value
- `Value.Array(items: MutableList<Value>)` - Array value
- `Value.Object` - Object value

Factory methods:
- `Value.ofNull()`, `Value.ofBool()`, `Value.ofInteger()`, `Value.ofFloat()`, `Value.ofString()`, `Value.ofArray()`, `Value.ofObject()`

Type checks:
- `isNull`, `isBool`, `isInteger`, `isFloat`, `isNumber`, `isString`, `isArray`, `isObject`

Value getters:
- `asBool()`, `asInteger()`, `asFloat()`, `asString()`, `asArray()`, `asObject()`

## License

MIT License

