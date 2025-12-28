# ASON C Library

A C11 implementation of the ASON (Array-Schema Object Notation) format.

## Features

- **Schema-Data Separation**: Define structure once, use compact data representation
- **Type Support**: Null, Bool, Integer, Float, String, Array, Object
- **Unicode Support**: Full UTF-8 support for strings
- **Comments**: Block comments (`/* ... */`) supported
- **Error Handling**: Error codes with descriptive messages
- **Memory Safe**: Clear ownership semantics

## Building

```bash
mkdir build && cd build
cmake ..
make
# test
./ason_c_test
# example
./basic_example_c
```

### Build Options

- `ASON_BUILD_TESTS=ON/OFF` - Build tests (default: ON)
- `ASON_BUILD_EXAMPLES=ON/OFF` - Build examples (default: ON)

## Usage

### Parsing

```c
#include <ason.h>

// Parse a simple object
ason_parse_result r = ason_parse("{name,age}:(Alice,30)");
if (r.error == ASON_OK) {
    const char* name = ason_value_get_string(ason_object_get(r.value, "name"));
    int64_t age = ason_value_get_integer(ason_object_get(r.value, "age"));
    printf("name=%s, age=%lld\n", name, (long long)age);
}
ason_parse_result_free(&r);

// Parse multiple objects
ason_parse_result r = ason_parse("{name,age}:(Alice,30),(Bob,25)");

// Parse nested objects
ason_parse_result r = ason_parse("{name,addr{city,zip}}:(Alice,(NYC,10001))");

// Parse arrays
ason_parse_result r = ason_parse("{name,scores[]}:(Alice,[90,85,95])");
```

### Serializing

```c
#include <ason.h>

ason_value* obj = ason_value_object();
ason_object_set(obj, "name", ason_value_string("Alice"));
ason_object_set(obj, "age", ason_value_integer(30));

char* str = ason_to_string(obj);
printf("%s\n", str);  // Output: (Alice,30)

free(str);
ason_value_free(obj);
```

### Value Types

```c
// Create values
ason_value* null_val = ason_value_null();
ason_value* bool_val = ason_value_bool(true);
ason_value* int_val = ason_value_integer(42);
ason_value* float_val = ason_value_float(3.14);
ason_value* str_val = ason_value_string("hello");
ason_value* arr_val = ason_value_array();
ason_value* obj_val = ason_value_object();

// Type checking
ason_value_is_null(v);
ason_value_is_bool(v);
ason_value_is_integer(v);
ason_value_is_float(v);
ason_value_is_string(v);
ason_value_is_array(v);
ason_value_is_object(v);

// Value access
bool b = ason_value_get_bool(v);
int64_t i = ason_value_get_integer(v);
double f = ason_value_get_float(v);
const char* s = ason_value_get_string(v);

// Array operations
ason_array_push(arr, item);
ason_value* item = ason_array_get(arr, index);
size_t len = ason_array_len(arr);

// Object operations
ason_object_set(obj, "key", value);
ason_value* val = ason_object_get(obj, "key");
size_t len = ason_object_len(obj);

// Memory management
ason_value_free(v);  // Frees value and all children
ason_value* copy = ason_value_clone(v);  // Deep copy
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

