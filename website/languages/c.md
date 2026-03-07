# C Guide

The C ASON library provides a lightweight implementation for embedding in C projects. It has no external dependencies beyond the C standard library.

## Installation

Add the source files to your project:

```cmake
# CMakeLists.txt
add_subdirectory(path/to/ason/c)
target_link_libraries(your_target ason)
```

Or copy `include/ason.h` and `src/ason.c` directly.

## API Overview

```c
#include "ason.h"

// Encoding
AsonBuffer *ason_encode(const AsonSchema *schema, const void *data, size_t count);

// Decoding
int ason_decode(const char *text, const AsonSchema *schema, void *out, size_t *count);

// Binary
AsonBuffer *ason_encode_bin(const AsonSchema *schema, const void *data, size_t count);
int ason_decode_bin(const uint8_t *data, size_t len,
                    const AsonSchema *schema, void *out, size_t *count);

// Cleanup
void ason_buffer_free(AsonBuffer *buf);
```

## Example

```c
#include <stdio.h>
#include "ason.h"

typedef struct {
    int64_t id;
    char    name[64];
    bool    active;
} User;

static const AsonField user_fields[] = {
    { "id",     ASON_INT,  offsetof(User, id)     },
    { "name",   ASON_STR,  offsetof(User, name)   },
    { "active", ASON_BOOL, offsetof(User, active) },
};
static const AsonSchema user_schema = {
    .fields     = user_fields,
    .field_count = 3,
    .struct_size = sizeof(User),
};

int main(void) {
    User users[] = {
        { 1, "Alice", true  },
        { 2, "Bob",   false },
    };

    AsonBuffer *buf = ason_encode(&user_schema, users, 2);
    printf("%s\n", (char *)buf->data);
    ason_buffer_free(buf);
    return 0;
}
```

## Building Examples

```bash
cd c
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build
./build/examples/basic
./build/examples/bench
```

## Running Tests

```bash
cd c/build && ctest --output-on-failure
```
