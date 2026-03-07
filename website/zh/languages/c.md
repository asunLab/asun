# C 指南

C ASON 库提供轻量级实现，可嵌入 C 项目中使用，除 C 标准库外无任何外部依赖。

## 安装

将源文件添加到你的项目：

```cmake
# CMakeLists.txt
add_subdirectory(path/to/ason/c)
target_link_libraries(your_target ason)
```

或直接复制 `include/ason.h` 和 `src/ason.c`。

## API 概览

```c
#include "ason.h"

// 编码
AsonBuffer *ason_encode(const AsonSchema *schema, const void *data, size_t count);

// 解码
int ason_decode(const char *text, const AsonSchema *schema, void *out, size_t *count);

// 二进制格式
AsonBuffer *ason_encode_bin(const AsonSchema *schema, const void *data, size_t count);
int ason_decode_bin(const uint8_t *data, size_t len,
                    const AsonSchema *schema, void *out, size_t *count);

// 释放资源
void ason_buffer_free(AsonBuffer *buf);
```

## 示例

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
    .fields      = user_fields,
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

## 构建示例

```bash
cd c
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build
./build/examples/basic
./build/examples/bench
```

## 运行测试

```bash
cd c/build && ctest --output-on-failure
```
