# ason-c

[![C11](https://img.shields.io/badge/C-11-blue.svg)](https://en.cppreference.com/w/c/11)
[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

高性能 C11 [ASON](https://github.com/athxx/ason)（Array-Schema Object Notation）序列化/反序列化库 —— 面向 LLM 交互和大规模数据传输的高效序列化格式。SIMD 加速（NEON / SSE2），零拷贝解析，零外部依赖。

[English](README.md)

## 什么是 ASON？

ASON 将 **Schema** 与 **数据** 分离，消除了 JSON 中每个对象都重复出现 Key 的冗余。Schema 只声明一次，数据行仅保留纯值：

```text
JSON (100 tokens):
{"users":[{"id":1,"name":"Alice","active":true},{"id":2,"name":"Bob","active":false}]}

ASON (~35 tokens, 节省 65%):
[{id:int, name:str, active:bool}]:(1,Alice,true),(2,Bob,false)
```

| 方面         | JSON         | ASON             |
| ------------ | ------------ | ---------------- |
| Token 效率   | 100%         | 30–70% ✓         |
| Key 重复     | 每个对象都有 | 声明一次 ✓       |
| 人类可读     | 是           | 是 ✓             |
| 嵌套结构     | ✓            | ✓                |
| 类型注解     | 无           | 可选 ✓           |
| 序列化速度   | 1x           | **~1.7x 更快** ✓ |
| 反序列化速度 | 1x           | **~2.9x 更快** ✓ |
| 数据体积     | 100%         | **47–60%** ✓     |

## 特性

- **纯 C11** —— 无需 C++，任何 C11 兼容编译器均可编译
- **零依赖** —— 仅使用标准库，无第三方库
- **SIMD 加速** —— 支持 NEON（ARM64/Apple Silicon）和 SSE2（x86_64），带标量回退
- **零拷贝解析** —— 未转义字符串直接返回输入缓冲区指针，最小化内存分配
- **宏反射** —— `ASON_FIELDS` 宏注册结构体，编译期字段描述符
- **完整类型支持** —— bool、所有整数类型（i8–i64、u8–u64）、float、double、string、optional、vector、map、嵌套结构体、结构体数组
- **Schema 驱动** —— Schema 声明一次，数据行无限扩展
- **类型注解** —— 可选的类型提示，用于文档和 LLM prompt

## 快速开始

### 1. 添加到项目

```bash
# 将 include/ason.h 和 src/ason.c 复制到你的项目
cp include/ason.h /your/project/include/
cp src/ason.c /your/project/src/
```

### 2. 定义结构体

```c
#include "ason.h"

typedef struct {
    int64_t id;
    ason_string_t name;
    bool active;
} User;

ASON_FIELDS(User, 3,
    ASON_FIELD(User, id,     "id",     i64),
    ASON_FIELD(User, name,   "name",   str),
    ASON_FIELD(User, active, "active", bool))
```

### 3. 序列化与反序列化

```c
// 序列化
User user = {1, ason_string_from("Alice"), true};
ason_buf_t buf = ason_dump_User(&user);
// → "{id,name,active}:(1,Alice,true)"

// 带类型注解的序列化
ason_buf_t buf2 = ason_dump_typed_User(&user);
// → "{id:int,name:str,active:bool}:(1,Alice,true)"

// 反序列化
User user2 = {0};
ason_err_t err = ason_load_User(buf.data, buf.len, &user2);
assert(err == ASON_OK);
assert(user2.id == 1);

// 释放资源
ason_buf_free(&buf);
ason_buf_free(&buf2);
ason_string_free(&user.name);
ason_string_free(&user2.name);
```

### 4. Vec 序列化（Schema 驱动）

对于结构体数组，ASON 只写入一次 Schema，每个元素作为紧凑元组输出 —— 这是相比 JSON 的核心优势：

```c
User users[] = {
    {1, ason_string_from("Alice"), true},
    {2, ason_string_from("Bob"),   false},
    {3, ason_string_from("Carol Smith"), true},
};

// Schema 只写一次，数据为元组
ason_buf_t buf = ason_dump_vec_User(users, 3);
// → "[{id,name,active}]:(1,Alice,true),(2,Bob,false),(3,Carol Smith,true)"

// 反序列化
User* loaded = NULL;
size_t count = 0;
ason_err_t err = ason_load_vec_User(buf.data, buf.len, &loaded, &count);
assert(err == ASON_OK && count == 3);

// 释放
ason_buf_free(&buf);
for (size_t i = 0; i < count; i++) ason_string_free(&loaded[i].name);
free(loaded);
```

## 支持的类型

| 类型                  | C 类型             | ASON 字段类型           | 示例                            |
| --------------------- | ------------------ | ----------------------- | ------------------------------- |
| 整数 (i8–i64, u8–u64) | `int8_t`–`int64_t` | `i8`–`i64`, `u8`–`u64`  | `42`, `-100`                    |
| 浮点数                | `float`, `double`  | `f32`, `f64`            | `3.14`, `-0.5`                  |
| 布尔                  | `bool`             | `bool`                  | `true`, `false`                 |
| 字符串                | `ason_string_t`    | `str`                   | `Alice`, `"Carol Smith"`        |
| 可选整数              | `ason_opt_i64`     | `opt_i64`               | `42` 或 _（空白）_ 表示 None    |
| 可选字符串            | `ason_opt_str`     | `opt_str`               | `hello` 或 _（空白）_ 表示 None |
| 可选浮点              | `ason_opt_f64`     | `opt_f64`               | `3.14` 或 _（空白）_ 表示 None  |
| 整数向量              | `ason_vec_i64`     | `vec_i64`               | `[1,2,3]`                       |
| 字符串向量            | `ason_vec_str`     | `vec_str`               | `[rust,go,python]`              |
| 浮点向量              | `ason_vec_f64`     | `vec_f64`               | `[1.5,2.7]`                     |
| 布尔向量              | `ason_vec_bool`    | `vec_bool`              | `[true,false]`                  |
| 嵌套向量              | `ason_vec_vec_i64` | `vec_vec_i64`           | `[[1,2],[3,4]]`                 |
| 映射 (str→int)        | `ason_map_si`      | `map_si`                | `[(age,30),(score,95)]`         |
| 映射 (str→str)        | `ason_map_ss`      | `map_ss`                | `[(k1,v1),(k2,v2)]`             |
| 嵌套结构体            | 任意结构体         | `ASON_FIELD_STRUCT`     | `(field1,field2)`               |
| 结构体向量            | `ason_vec_T`       | `ASON_FIELD_VEC_STRUCT` | `[(v1,v2),(v3,v4)]`             |
| 字符                  | `char`             | `char`                  | `A`                             |

### 嵌套结构体

```c
typedef struct { ason_string_t title; } Dept;
ASON_FIELDS(Dept, 1, ASON_FIELD(Dept, title, "title", str))

typedef struct {
    ason_string_t name;
    Dept dept;
} Employee;
ASON_FIELDS(Employee, 2,
    ASON_FIELD(Employee, name, "name", str),
    ASON_FIELD_STRUCT(Employee, dept, "dept", &Dept_ason_desc))
// → {name,dept}:(Alice,(Engineering))
```

### 可选字段

```c
typedef struct {
    int64_t id;
    ason_opt_str label;
} Item;
ASON_FIELDS(Item, 2,
    ASON_FIELD(Item, id,    "id",    i64),
    ASON_FIELD(Item, label, "label", opt_str))
// 有值：   {id,label}:(1,hello)
// 无值：   {id,label}:(1,)
```

### 数组与映射

```c
typedef struct {
    ason_string_t name;
    ason_vec_str tags;
} Tagged;
ASON_FIELDS(Tagged, 2,
    ASON_FIELD(Tagged, name, "name", str),
    ASON_FIELD(Tagged, tags, "tags", vec_str))
// → {name,tags}:(Alice,[rust,go,python])

typedef struct {
    ason_string_t name;
    ason_map_si attrs;
} Profile;
ASON_FIELDS(Profile, 2,
    ASON_FIELD(Profile, name,  "name",  str),
    ASON_FIELD(Profile, attrs, "attrs", map_si))
// → {name,attrs}:(Alice,[(age,30),(score,95)])
```

### 结构体数组

```c
typedef struct { ason_string_t name; int64_t val; } Item;
ASON_FIELDS(Item, 2,
    ASON_FIELD(Item, name, "name", str),
    ASON_FIELD(Item, val,  "val",  i64))
ASON_VEC_STRUCT_DEFINE(Item)

typedef struct {
    ason_string_t label;
    ason_vec_Item items;
} Group;
ASON_FIELDS(Group, 2,
    ASON_FIELD(Group, label, "label", str),
    ASON_FIELD_VEC_STRUCT(Group, items, "items", Item))
// → {label,items}:(group1,[(a1,1),(a2,2)])
```

### 类型注解（可选）

带注解和不带注解的 Schema 完全等价 —— 反序列化器一视同仁：

```text
// 不带注解（dump / dump_vec 的默认输出）
{id,name,salary,active}:(1,Alice,5000.50,true)

// 带注解（dump_typed / dump_typed_vec 的输出）
{id:int,name:str,salary:float,active:bool}:(1,Alice,5000.50,true)
```

### 注释

```text
/* 用户列表 */
[{id:int, name:str, active:bool}]:(1,Alice,true),(2,Bob,false)
```

### 多行格式

```text
[{id:int, name:str, active:bool}]:
  (1, Alice, true),
  (2, Bob, false),
  (3, "Carol Smith", true)
```

## API 参考

| 函数                                  | 说明                                                |
| ------------------------------------- | --------------------------------------------------- |
| `ason_dump_T(obj)`                    | 序列化结构体 → 不带注解的 Schema `{id,name}:`       |
| `ason_dump_typed_T(obj)`              | 序列化结构体 → 带注解的 Schema `{id:int,name:str}:` |
| `ason_dump_vec_T(arr, count)`         | 序列化数组 → 不带注解的 Schema                      |
| `ason_dump_typed_vec_T(arr, count)`   | 序列化数组 → 带注解的 Schema                        |
| `ason_load_T(str, len, out)`          | 反序列化结构体（同时支持带/不带注解）               |
| `ason_load_vec_T(str, len, out, cnt)` | 反序列化数组（同时支持带/不带注解）                 |

### 反射宏

```c
// 注册结构体字段
ASON_FIELDS(StructType, field_count,
    ASON_FIELD(StructType, member, "schema_name", field_type),
    ASON_FIELD_STRUCT(StructType, member, "schema_name", &SubType_ason_desc),
    ASON_FIELD_VEC_STRUCT(StructType, member, "schema_name", ElemType))

// 定义结构体向量类型（在使用 ASON_FIELD_VEC_STRUCT 之前需要调用）
ASON_VEC_STRUCT_DEFINE(ElemType)
```

### 内存管理

C 语言需要显式内存管理。使用提供的释放辅助函数：

```c
ason_buf_free(&buf);          // 释放序列化缓冲区
ason_string_free(&str);       // 释放拥有的字符串
ason_vec_i64_free(&vec);      // 释放向量
// 对于加载的数组：先释放每个元素的字符串，再 free(arr)
```

## 性能

在 Apple Silicon M 系列（arm64，NEON SIMD）上以 `-O3` Release 模式进行基准测试，将 ASON 与最小化内联 JSON 序列化/反序列化器进行对比：

### 序列化（ASON 快 1.3–1.8x）

| 场景              | JSON     | ASON     | 加速比    |
| ----------------- | -------- | -------- | --------- |
| 扁平结构体 × 100  | 2.22 ms  | 1.21 ms  | **1.84x** |
| 扁平结构体 × 500  | 8.49 ms  | 4.59 ms  | **1.85x** |
| 扁平结构体 × 1000 | 12.12 ms | 7.26 ms  | **1.67x** |
| 扁平结构体 × 5000 | 59.92 ms | 35.99 ms | **1.67x** |
| 5 层嵌套 × 10     | 3.44 ms  | 2.55 ms  | **1.35x** |
| 5 层嵌套 × 50     | 17.95 ms | 13.63 ms | **1.32x** |
| 5 层嵌套 × 100    | 35.26 ms | 26.45 ms | **1.33x** |
| 大负载 (10k)      | 12.42 ms | 7.86 ms  | **1.58x** |

### 反序列化（ASON 快 2.9–3.3x）

| 场景              | JSON      | ASON     | 加速比    |
| ----------------- | --------- | -------- | --------- |
| 扁平结构体 × 100  | 8.98 ms   | 2.78 ms  | **3.23x** |
| 扁平结构体 × 500  | 31.91 ms  | 9.72 ms  | **3.28x** |
| 扁平结构体 × 1000 | 56.82 ms  | 19.04 ms | **2.98x** |
| 扁平结构体 × 5000 | 286.56 ms | 97.89 ms | **2.93x** |
| 大负载 (10k)      | 58.08 ms  | 20.16 ms | **2.88x** |

### 吞吐量

| 指标               | JSON           | ASON            | 加速比    |
| ------------------ | -------------- | --------------- | --------- |
| 序列化 records/s   | 8.2M records/s | 13.3M records/s | **1.63x** |
| 序列化 MB/s        | 947 MB/s       | 718 MB/s        | —         |
| 反序列化 records/s | 1.7M records/s | 5.0M records/s  | **2.92x** |

### 体积节省

| 场景              | JSON     | ASON     | 节省    |
| ----------------- | -------- | -------- | ------- |
| 扁平结构体 × 100  | 12,071 B | 5,612 B  | **54%** |
| 扁平结构体 × 1000 | 122 KB   | 57 KB    | **53%** |
| 5 层嵌套 × 10     | 43,244 B | 16,893 B | **61%** |
| 5 层嵌套 × 100    | 438 KB   | 175 KB   | **60%** |
| 大负载 (10k)      | 1.2 MB   | 0.6 MB   | **53%** |

### 为什么 ASON 更快？

1. **零 Key 哈希** —— Schema 只解析一次；数据字段按位置索引 O(1) 映射，无需逐行进行 Key 字符串哈希。
2. **Schema 驱动解析** —— 反序列化器已知每个字段的预期类型，直接解析而非运行时类型推断。CPU 分支预测命中率 ~100%。
3. **最小化内存分配** —— 所有数据行共享一个 Schema 引用。无重复 Key 字符串分配。
4. **SIMD 字符串扫描** —— NEON/SSE2 向量化字符扫描，用于引号检测、转义序列和引号字符串解析。每周期处理 16 字节。
5. **快速数字格式化** —— 整数序列化使用双数位查找表；浮点格式化对常见情况（整数、1–2 位小数）有快速路径。
6. **零拷贝解析** —— 未转义字符串直接返回输入缓冲区指针，避免字符串复制。

## 测试套件

36 个全面测试覆盖所有功能：

### 序列化与反序列化（4 个测试）

- 简单结构体往返
- 带类型注解的 Schema 往返
- Vec 往返（Schema 驱动）
- Vec 带类型注解往返

### 可选字段（4 个测试）

- 有值解析
- 无值（None）
- 可选字段序列化
- 逗号间空可选 `(1,,3)`

### 集合（5 个测试）

- 向量字段解析和往返
- 空向量处理
- 嵌套向量 `[[1,2],[3,4]]`
- 映射字段解析 `[(key,val)]`
- 映射往返

### 嵌套结构体（3 个测试）

- 嵌套结构体解析
- 嵌套结构体往返
- 3 层深度嵌套（DeepC > DeepB > DeepA，含结构体向量）

### 字符串处理（7 个测试）

- 引号字符串解析
- 完整转义序列支持（`\"`、`\\`、`\n`、`\t`、`\,`、`\(`、`\)`、`\[`、`\]`）
- 自动引号检测（SIMD 加速）
- 未引号字符串修剪
- 含内部空格的字符串
- 前导/尾随空格引号
- 反斜杠转义往返

### 数字处理（4 个测试）

- 浮点精度（double、float）
- 整数值浮点格式化（如 `42.0`）
- 负数（整数和浮点数）
- 大无符号整数（uint64 最大值）

### 解析特性（6 个测试）

- 布尔值（`true`/`false`）
- 块注释解析 `/* ... */`
- 空白处理
- 多行格式解析
- 带类型注解的 Schema 解析
- Schema 字段不匹配（多余/缺少字段）

### 边界情况（3 个测试）

- 错误处理（无效输入返回错误码）
- 空向量序列化 `[]`
- 前导/尾随空格引号往返

### 复杂示例（14 个测试）

- 带数组的嵌套结构体
- 带嵌套结构体的 Vec
- 映射/字典字段
- 转义字符串往返
- 全类型结构体（17 个字段：bool、i8–i64、u8–u64、float、double、string、optional、vectors、嵌套向量）
- 5 层深度嵌套（Country > Region > City > District > Street > Building）
- 7 层深度嵌套（Universe > Galaxy > SolarSystem > Planet > Continent > Nation > State）
- 复杂配置结构体（嵌套 + 映射 + 可选）
- 边界情况（空 vec、特殊字符、类布尔字符串、类数字字符串）
- 三重嵌套数组 `[[1,2],[3,4],[5,6,7]]`
- 真实数据中的注释解析

## 构建与运行

```bash
# 构建所有目标（Release 模式以获得最佳性能）
mkdir -p build && cd build
cmake .. -DCMAKE_BUILD_TYPE=Release
make -j$(nproc)

# 运行测试
./ason_test

# 运行示例
./basic              # 10 个基本示例
./complex_example    # 14 个复杂示例

# 运行基准测试
./bench              # ASON vs JSON 综合基准测试
```

## ASON 格式规范

完整规范请参见 [ASON Spec](https://github.com/athxx/ason/blob/main/docs/ASON_SPEC_CN.md)，包含语法规则、BNF 文法、转义规则、类型系统和 LLM 集成最佳实践。

### 语法快速参考

| 元素     | Schema                      | 数据                |
| -------- | --------------------------- | ------------------- |
| 对象     | `{field1:type,field2:type}` | `(val1,val2)`       |
| 数组     | `field:[type]`              | `[v1,v2,v3]`        |
| 对象数组 | `field:[{f1:type,f2:type}]` | `[(v1,v2),(v3,v4)]` |
| 映射     | `field:map[K,V]`            | `[(k1,v1),(k2,v2)]` |
| 嵌套对象 | `field:{f1:type,f2:type}`   | `(v1,(v3,v4))`      |
| 空值     | —                           | _（空白）_          |
| 空字符串 | —                           | `""`                |
| 注释     | —                           | `/* ... */`         |

## ASON 二进制格式

ASON-BIN 是一种紧凑的二进制序列化格式，与 ASON 文本共享相同的 Schema。采用小端固定宽度编码，无字段名称，纯位置式编码，追求极致吞吐量。

### 编码格式

| 类型       | 编码方式                              |
| ---------- | ------------------------------------- |
| `bool`     | 1 字节（0 或 1）                     |
| `i8/u8`    | 1 字节                                |
| `i16/u16`  | 2 字节小端                            |
| `i32/u32`  | 4 字节小端                            |
| `i64/u64`  | 8 字节小端                            |
| `f32`      | 4 字节小端                            |
| `f64`      | 8 字节小端                            |
| `str`      | u32 长度（小端）+ UTF-8 字节          |
| `vec<T>`   | u32 元素数（小端）+ N × 元素编码     |
| 结构体     | 各字段顺序编码                        |

### API 用法

在 `ASON_FIELDS(...)` 之后紧接 `ASON_FIELDS_BIN(结构体类型, 字段数)` 即可自动注入四个函数：

```c
#include "ason.h"

typedef struct { ason_string_t name; int64_t age; double score; bool active; } User;

ASON_FIELDS(User, 4,
    ASON_FIELD(User, name,   "name",   str),
    ASON_FIELD(User, age,    "age",    i64),
    ASON_FIELD(User, score,  "score",  f64),
    ASON_FIELD(User, active, "active", bool))

ASON_FIELDS_BIN(User, 4)   /* 自动注入：*/
/*
 *  ason_buf_t  ason_dump_bin_User(const User*)
 *  ason_buf_t  ason_dump_bin_vec_User(const User*, size_t count)
 *  ason_err_t  ason_load_bin_User(const char* data, size_t len, User* out)
 *  ason_err_t  ason_load_bin_vec_User(const char* data, size_t len,
 *                                     User** out_arr, size_t* out_count)
 */
```

#### 序列化

```c
User u = { .name = ason_string_from_len("Alice", 5), .age = 30,
           .score = 95.5, .active = true };

ason_buf_t buf = ason_dump_bin_User(&u);
/* 使用 buf.data / buf.len 进行网络/文件 I/O */
ason_buf_free(&buf);
ason_string_free(&u.name);
```

#### 反序列化

```c
ason_err_t err = ason_load_bin_User(buf.data, buf.len, &u2);
assert(err == ASON_OK);
/* 使用完毕后释放堆字符串 */
ason_string_free(&u2.name);
```

#### 批量（结构体数组）

```c
/* 序列化 */
ason_buf_t batch = ason_dump_bin_vec_User(users, count);

/* 反序列化 */
User* out = NULL;  size_t n = 0;
ason_load_bin_vec_User(batch.data, batch.len, &out, &n);
for (size_t i = 0; i < n; i++) ason_string_free(&out[i].name);
free(out);
ason_buf_free(&batch);
```

### 性能测试（macOS arm64 NEON，100 次迭代）

| 测试场景            | JSON 序列化 | BIN 序列化  | 序列化倍速    | JSON 反序列化 | BIN 反序列化 | 反序列化倍速  | 体积节省 |
| ------------------- | ----------- | ----------- | ------------- | ------------- | ------------ | ------------- | -------- |
| 扁平结构 ×100       | 2.07 ms     | 0.30 ms     | **6.97x**     | 8.86 ms       | 1.79 ms      | **4.96x**     | 38%      |
| 扁平结构 ×1000      | 12.51 ms    | 2.31 ms     | **5.40x**     | 58.84 ms      | 14.20 ms     | **4.14x**     | 39%      |
| 扁平结构 ×5000      | 64.04 ms    | 11.35 ms    | **5.64x**     | 295.65 ms     | 67.99 ms     | **4.35x**     | 39%      |
| 全类型结构 ×100     | 1.41 ms     | 0.56 ms     | **2.51x**     | 2.12 ms       | 1.72 ms      | **1.23x**     | 31%      |
| 五层嵌套 ×100       | 36.11 ms    | 26.11 ms    | **1.38x**     | 114.91 ms     | 58.34 ms     | **1.97x**     | 61%      |
| 大数据量 ×10000     | 12.74 ms    | 2.22 ms     | **5.73x**     | 58.11 ms      | 13.67 ms     | **4.25x**     | 39%      |

> BIN 格式对扁平/同构载荷最快。即使是五层深度嵌套结构，反序列化仍快 **约 2 倍**，体积减少 **61%**。

## 许可证

MIT
