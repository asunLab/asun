# C++ Guide

The C++ ASON library builds on the C core and adds template-based type-safe wrappers, RAII resource management, and optional STL integration.

## Installation

```cmake
# CMakeLists.txt
add_subdirectory(path/to/ason/cpp)
target_link_libraries(your_target ason_cpp)
```

Requires C++17 or later.

## API Overview

```cpp
#include "ason.hpp"

// Encoding
std::string ason::encode(const T &value);
std::string ason::encode_vec(const std::vector<T> &values);

// Decoding
T              ason::decode<T>(std::string_view text);
std::vector<T> ason::decode_vec<T>(std::string_view text);

// Binary
std::vector<uint8_t> ason::encode_bin(const T &value);
T                    ason::decode_bin<T>(std::span<const uint8_t> data);
```

## Example

```cpp
#include <iostream>
#include <vector>
#include "ason.hpp"

struct User {
    int64_t     id;
    std::string name;
    bool        active;

    ASON_FIELDS(id, name, active)          // macro generates schema metadata
    ASON_TYPES("int", "str", "bool")
};

int main() {
    std::vector<User> users = {
        {1, "Alice", true},
        {2, "Bob",   false},
    };

    std::string s = ason::encode_vec(users);
    std::cout << s << "\n";

    auto restored = ason::decode_vec<User>(s);
    // restored == users

    // Binary round-trip
    auto bytes    = ason::encode_bin_vec(users);
    auto restored2 = ason::decode_bin_vec<User>(bytes);
}
```

## Building Examples

```bash
cd cpp
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build
./build/examples/basic
./build/examples/bench
```

## Running Tests

```bash
cd cpp/build && ctest --output-on-failure
```
