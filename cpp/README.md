# ASON C++ Library

A C++17 implementation of the ASON (Array-Schema Object Notation) format.

## Features

- **Schema-Data Separation**: Define structure once, use compact data representation
- **Type Support**: Null, Bool, Integer, Float, String, Array, Object
- **Unicode Support**: Full UTF-8 support for strings
- **Comments**: Block comments (`/* ... */`) supported
- **Error Handling**: Result type with detailed error information
- **Pretty Printing**: Optional formatted output

## Building

```bash
mkdir build && cd build
cmake ..
make
```

### Build Options

- `ASON_BUILD_TESTS=ON/OFF` - Build tests (default: ON)
- `ASON_BUILD_EXAMPLES=ON/OFF` - Build examples (default: ON)

## Usage

### Parsing

```cpp
#include <ason/ason.hpp>

// Parse a simple object
auto result = ason::parse("{name,age}:(Alice,30)");
if (result.ok()) {
    auto& value = result.value();
    std::cout << *value.get("name")->as_string() << std::endl;  // Alice
    std::cout << *value.get("age")->as_integer() << std::endl;  // 30
}

// Parse multiple objects
auto result = ason::parse("{name,age}:(Alice,30),(Bob,25)");

// Parse nested objects
auto result = ason::parse("{name,addr{city,zip}}:(Alice,(NYC,10001))");

// Parse arrays
auto result = ason::parse("{name,scores[]}:(Alice,[90,85,95])");
```

### Serializing

```cpp
#include <ason/ason.hpp>

ason::Object obj;
obj["name"] = ason::Value("Alice");
obj["age"] = ason::Value(30);

std::string output = ason::to_string(ason::Value(std::move(obj)));
// Output: (Alice,30)

// Pretty print
std::string pretty = ason::to_string_pretty(value);
```

### Value Types

```cpp
// Create values
ason::Value null_val;                    // Null
ason::Value bool_val(true);              // Bool
ason::Value int_val(42);                 // Integer
ason::Value float_val(3.14);             // Float
ason::Value str_val("hello");            // String
ason::Value arr_val(ason::Array{...});   // Array
ason::Value obj_val(ason::Object{...});  // Object

// Type checking
value.is_null();
value.is_bool();
value.is_integer();
value.is_float();
value.is_string();
value.is_array();
value.is_object();

// Value access
auto b = value.as_bool();      // std::optional<bool>
auto i = value.as_integer();   // std::optional<int64_t>
auto f = value.as_float();     // std::optional<double>
auto s = value.as_string();    // const std::string*
auto a = value.as_array();     // const Array*
auto o = value.as_object();    // const Object*

// Object/Array access
value.get("key");    // Get by key (returns const Value*)
value.get(0);        // Get by index (returns const Value*)
```

## ASON Format

ASON separates schema from data for compact representation:

```
{name,age}:(Alice,30)
```

Is equivalent to JSON:
```json
{"name": "Alice", "age": 30}
```

### Features

- **Nested Objects**: `{name,addr{city,zip}}:(Alice,(NYC,10001))`
- **Arrays**: `{scores[]}:([90,85,95])`
- **Object Arrays**: `{users[{id,name}]}:([(1,Alice),(2,Bob)])`
- **Multiple Records**: `{name,age}:(Alice,30),(Bob,25)`
- **Comments**: `/* comment */ {name}:(Alice)`
- **Unicode**: `{name}:(小明)`

## License

MIT License

