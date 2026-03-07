# C++ 指南

C++ ASON 库基于 C 核心，增加了模板类型安全包装、RAII 资源管理及可选的 STL 集成，需要 C++17 或更高版本。

## 安装

```cmake
# CMakeLists.txt
add_subdirectory(path/to/ason/cpp)
target_link_libraries(your_target ason_cpp)
```

## API 概览

```cpp
#include "ason.hpp"

// 编码
std::string ason::encode(const T &value);
std::string ason::encode_vec(const std::vector<T> &values);

// 解码
T              ason::decode<T>(std::string_view text);
std::vector<T> ason::decode_vec<T>(std::string_view text);

// 二进制格式
std::vector<uint8_t> ason::encode_bin(const T &value);
T                    ason::decode_bin<T>(std::span<const uint8_t> data);
```

## 示例

```cpp
#include <iostream>
#include <vector>
#include "ason.hpp"

struct User {
    int64_t     id;
    std::string name;
    bool        active;

    ASON_FIELDS(id, name, active)        // 宏生成 Schema 元数据
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

    // 二进制往返
    auto bytes     = ason::encode_bin_vec(users);
    auto restored2 = ason::decode_bin_vec<User>(bytes);
}
```

## 构建示例

```bash
cd cpp
cmake -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build
./build/examples/basic
./build/examples/bench
```

## 运行测试

```bash
cd cpp/build && ctest --output-on-failure
```
