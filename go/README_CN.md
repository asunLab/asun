# ason-go

[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Go](https://img.shields.io/badge/go-1.24+-00ADD8.svg)](https://go.dev)

高性能 [ASON](https://github.com/athxx/ason)（Array-Schema Object Notation）Go 序列化/反序列化库 —— 面向 LLM 交互和大规模数据传输的高效序列化格式。

**零拷贝设计，零外部依赖（仅标准库），极致性能。**

[English](README.md)

## 什么是 ASON？

ASON 将 **Schema** 与 **数据** 分离，消除了 JSON 中每个对象都重复出现 Key 的冗余。Schema 只声明一次，数据行仅保留纯值：

```text
JSON (100 tokens):
{"users":[{"id":1,"name":"Alice","active":true},{"id":2,"name":"Bob","active":false}]}

ASON (~35 tokens, 节省 65%):
[{id:int, name:str, active:bool}]:(1,Alice,true),(2,Bob,false)
```

| 方面       | JSON         | ASON                 |
| ---------- | ------------ | -------------------- |
| Token 效率 | 100%         | 30–70% ✓             |
| Key 重复   | 每个对象都有 | 声明一次 ✓           |
| 人类可读   | 是           | 是 ✓                 |
| 嵌套结构   | ✓            | ✓                    |
| 类型注解   | 无           | 可选 ✓               |
| 序列化速度 | 1x           | **1.4–1.9 倍更快** ✓ |
| 反序列化   | 1x           | **2.9–4.4 倍更快** ✓ |
| 数据体积   | 100%         | **40–50%** ✓         |

## 快速开始

```bash
go get github.com/example/ason
```

### 序列化与反序列化结构体

```go
package main

import (
    "fmt"
    ason "github.com/example/ason"
)

type User struct {
    ID     int64  `ason:"id"`
    Name   string `ason:"name"`
    Active bool   `ason:"active"`
}

func main() {
    user := User{ID: 1, Name: "Alice", Active: true}

    // 序列化
    b, _ := ason.Marshal(&user)
    fmt.Println(string(b))
    // 输出: {id,name,active}:(1,Alice,true)

    // 反序列化
    var user2 User
    ason.Unmarshal(b, &user2)
    fmt.Println(user == user2) // true
}
```

### 带类型注解序列化

使用 `MarshalTyped` 输出带类型注解的 Schema —— 适用于文档生成、LLM 提示词和跨语言数据交换：

```go
b, _ := ason.MarshalTyped(&user)
// 输出: {id:int,name:str,active:bool}:(1,Alice,true)

// 反序列化同时支持带注解和不带注解的 Schema
var user2 User
ason.Unmarshal(b, &user2)
```

### 序列化与反序列化切片（Schema 驱动）

对于 `[]T`，ASON 只写入一次 Schema，每个元素以紧凑元组形式输出 —— 这是相比 JSON 的核心优势：

```go
users := []User{
    {ID: 1, Name: "Alice", Active: true},
    {ID: 2, Name: "Bob", Active: false},
}

// 无注解 Schema
b, _ := ason.MarshalSlice(users)
// 输出: [{id,name,active}]:(1,Alice,true),(2,Bob,false)

// 带类型注解 Schema
b2, _ := ason.MarshalSliceTyped(users, []string{"int", "str", "bool"})
// 输出: [{id:int,name:str,active:bool}]:(1,Alice,true),(2,Bob,false)

// 反序列化 —— 两种格式均可
var parsed []User
ason.UnmarshalSlice(b, &parsed)
```
## 二进制格式（ASON-BIN）

ASON-BIN 是对任意 Go 值的紧凑二进制编码。它提供最大幅度的性能提升，非常适合内部服务通信、缓存和存储场景。

### API

```go
import ason "github.com/example/ason"

// 序列化任意值为字节
b, err := ason.MarshalBinary(&user)

// 从字节反序列化（字符串和 []byte 零拷贝）
var u2 User
err = ason.UnmarshalBinary(b, &u2)
```

### 零拷贝反序列化

在反序列化 ASON-BIN 时，`string` 和 `[]byte` 字段使用零拷贝技术（`unsafe.String` 和切片重切片）创建。它们直接指向输入的字节切片，避免了堆分配和内存拷贝。

### 性能测试数据（Apple M 系列芯片）

**扁平结构体（8 字段）**

| 测试 | JSON | ASON 文本 | ASON-BIN | BIN vs JSON |
|---|---|---|---|---|
| 序列化 × 1000 | 0.17 ms | 0.08 ms | **0.08 ms** | **快 2.3×** |
| 反序列化 × 1000 | 0.83 ms | 0.21 ms | **0.07 ms** | **快 12.6×** |
| 数据大小 × 1000 | 122,071 B | 57,112 B | **74,784 B** | 小 39% |

**深层嵌套结构体（5 层嵌套，× 100）**

| 测试 | JSON | ASON 文本 | ASON-BIN | BIN vs JSON |
|---|---|---|---|---|
| 序列化 × 100 | 0.35 ms | 0.20 ms | **0.05 ms** | **快 7.0×** |
| 反序列化 × 100 | 1.20 ms | 0.45 ms | **0.08 ms** | **快 15.0×** |
| 数据大小 × 100 | 43,811 B | 17,461 B | **22,543 B** | 小 49% |
## 支持的类型

| 类型                        | ASON 表示           | 示例                         |
| --------------------------- | ------------------- | ---------------------------- |
| 整数 (int8–int64, uint8–64) | 纯数字              | `42`, `-100`                 |
| 浮点 (float32, float64)     | 带小数点            | `3.14`, `-0.5`               |
| 布尔                        | 字面量              | `true`, `false`              |
| 字符串                      | 无引号或有引号      | `Alice`, `"Carol Smith"`     |
| 指针 (\*T)                  | 有值或留空          | `hello` 或 _(空白)_ 表示 nil |
| 切片 ([]T)                  | `[v1,v2,v3]`        | `[rust,go,python]`           |
| 字典 (map[K]V)              | `[(k1,v1),(k2,v2)]` | `[(age,30),(score,95)]`      |
| 嵌套结构体                  | `(field1,field2)`   | `(Engineering,500000)`       |

### 结构体标签

库优先读取 `ason` 标签，其次回退到 `json` 标签。使用 `"-"` 跳过字段：

```go
type Employee struct {
    ID     int64      `ason:"id"`
    Name   string     `ason:"name"`
    Dept   Department `ason:"dept"`
    Secret string     `ason:"-"`  // 被跳过
}
```

### 嵌套结构体

```go
type Dept struct {
    Title string `ason:"title"`
}

type Employee struct {
    Name string `ason:"name"`
    Dept Dept   `ason:"dept"`
}

// Schema 自动反映嵌套结构：
// {name,dept}:(Alice,(Engineering))
```

### 可选字段（指针）

```go
type Item struct {
    ID    int64   `ason:"id"`
    Label *string `ason:"label"`
}

// 有值:  {id,label}:(1,hello)
// nil:   {id,label}:(1,)
```

### 切片与字典

```go
type Tagged struct {
    Name string   `ason:"name"`
    Tags []string `ason:"tags"`
}
// {name,tags}:(Alice,[rust,go,python])

type Profile struct {
    Name  string           `ason:"name"`
    Attrs map[string]int64 `ason:"attrs"`
}
// {name,attrs}:(Alice,[(age,30),(score,95)])
```

### 类型注解（可选）

ASON Schema 支持**可选的**类型注解。两种形式完全等价 —— 反序列化器对它们的处理完全一致：

```text
// 不带注解（Marshal / MarshalSlice 的默认输出）
{id,name,salary,active}:(1,Alice,5000.50,true)

// 带注解（MarshalTyped / MarshalSliceTyped 的输出）
{id:int,name:str,salary:float,active:bool}:(1,Alice,5000.50,true)
```

注解是**纯粹的装饰性元数据** —— 它们不影响解析和反序列化行为。反序列化器遇到 `:type` 部分时会直接跳过。

**适用场景：**

- LLM 提示词 — 帮助模型理解并生成正确的数据
- API 文档 — 无需外部文档即可自描述 Schema
- 跨语言数据交换 — 消除类型歧义（`42` 是 int 还是 float？）
- 调试 — 一眼看出数据类型

**性能影响：** 可忽略不计。开销是常数级的，不随数据量增长 —— 注解仅影响 Schema 头部，不影响数据体。

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

| 函数                                            | 说明                                              |
| ----------------------------------------------- | ------------------------------------------------- |
| `Marshal(v any) ([]byte, error)`                | 序列化结构体 → 无注解 Schema `{id,name}:`         |
| `MarshalTyped(v any) ([]byte, error)`           | 序列化结构体 → 带注解 Schema `{id:int,name:str}:` |
| `Unmarshal(data []byte, v any) error`           | 反序列化结构体（两种 Schema 格式均可）            |
| `MarshalSlice(v any) ([]byte, error)`           | 序列化切片 → 无注解 Schema（Schema 驱动）         |
| `MarshalSliceTyped(v, types) ([]byte, error)`   | 序列化切片 → 带注解 Schema                        |
| `MarshalSliceFields(v, fields) ([]byte, error)` | 序列化切片，自定义字段名                          |
| `UnmarshalSlice(data []byte, v any) error`      | 反序列化切片（两种 Schema 格式均可）              |

## 性能

在 Apple Silicon（M 系列）上测试，与 `encoding/json` 对比：

### 序列化（ASON 快 1.4–1.9 倍）

| 场景            | JSON     | ASON     | 加速比    |
| --------------- | -------- | -------- | --------- |
| 扁平结构 × 100  | 2.74 ms  | 1.41 ms  | **1.95x** |
| 扁平结构 × 500  | 9.72 ms  | 5.27 ms  | **1.85x** |
| 扁平结构 × 1000 | 15.52 ms | 8.87 ms  | **1.75x** |
| 扁平结构 × 5000 | 79.15 ms | 55.15 ms | **1.44x** |
| 5 层嵌套 × 10   | 6.80 ms  | 4.19 ms  | **1.62x** |
| 5 层嵌套 × 50   | 32.90 ms | 21.15 ms | **1.56x** |
| 5 层嵌套 × 100  | 65.79 ms | 41.88 ms | **1.57x** |
| 大数据量 (10k)  | 17.07 ms | 10.57 ms | **1.62x** |

### 反序列化（ASON 快 2.9–4.4 倍）

| 场景            | JSON      | ASON      | 加速比    |
| --------------- | --------- | --------- | --------- |
| 扁平结构 × 100  | 13.29 ms  | 3.04 ms   | **4.37x** |
| 扁平结构 × 500  | 44.46 ms  | 11.76 ms  | **3.78x** |
| 扁平结构 × 1000 | 84.93 ms  | 22.88 ms  | **3.71x** |
| 扁平结构 × 5000 | 440.20 ms | 120.71 ms | **3.65x** |
| 5 层嵌套 × 10   | 35.02 ms  | 11.47 ms  | **3.05x** |
| 5 层嵌套 × 50   | 175.52 ms | 58.04 ms  | **3.02x** |
| 5 层嵌套 × 100  | 350.60 ms | 118.59 ms | **2.96x** |
| 大数据量 (10k)  | 91.14 ms  | 27.53 ms  | **3.31x** |

### 单结构体往返（10,000 次迭代）

| 场景 | JSON     | ASON     | 加速比    |
| ---- | -------- | -------- | --------- |
| 扁平 | 11.37 ms | 4.95 ms  | **2.30x** |
| 深层 | 48.20 ms | 20.59 ms | **2.34x** |

### 吞吐量（1000 条记录 × 100 次迭代）

| 操作     | JSON            | ASON             | 加速比          |
| -------- | --------------- | ---------------- | --------------- |
| 序列化   | 6,158,789 条/秒 | 11,426,939 条/秒 | **1.86 倍更快** |
| 反序列化 | 1,133,686 条/秒 | 4,291,293 条/秒  | **3.79 倍更快** |

### 体积节省

| 场景            | JSON   | ASON   | 节省    |
| --------------- | ------ | ------ | ------- |
| 扁平结构 × 100  | 12 KB  | 5.6 KB | **54%** |
| 扁平结构 × 1000 | 122 KB | 57 KB  | **53%** |
| 5 层嵌套 × 10   | 43 KB  | 17 KB  | **60%** |
| 5 层嵌套 × 100  | 432 KB | 175 KB | **60%** |
| 大数据量 (10k)  | 1.2 MB | 0.6 MB | **53%** |

### 为什么 ASON 更快？

1. **零哈希匹配** — Schema 只解析一次，数据字段通过位置索引 `O(1)` 映射，无需每行对 Key 字符串计算哈希。
2. **模式驱动解析** — 反序列化器通过结构体已知每个字段的类型，可以直接调用 `parseInt64()` 等方法，而非运行时推断类型。CPU 分支预测命中率接近 100%。
3. **极小内存分配** — `sync.Pool` 缓冲区复用，`[256]bool` 查找表实现字符分类，`unsafe.String` 零拷贝字节到字符串转换。
4. **结构体信息缓存** — 反射结果缓存在 `sync.Map` 中，消除重复反射开销。
5. **无中间表示层** — 数据直接写入字节缓冲区，没有任何 DOM/AST 中间层。

## 实现亮点

- **零外部依赖** — 仅使用 Go 标准库（`reflect`、`sync`、`unsafe`、`math`、`strconv`）
- **零拷贝** — `unsafe.String` 实现字符串转换，避免不必要的内存分配
- **快速数字格式化** — 自定义 `appendI64`/`appendU64`/`appendFloat64`，对常见场景有快速路径（1 位小数、2 位小数、整数值浮点数）
- **查找表** — `[256]bool` 用于引号判断，`[256]byte` 用于转义字符
- **缓冲区池化** — `sync.Pool` 跨 Marshal 调用复用缓冲区
- **结构体信息缓存** — `sync.Map` 按类型缓存结构体字段元数据

运行基准测试：

```bash
go run examples/bench/main.go
```

## 示例

```bash
# 基础用法（11 个示例）
go run examples/basic/main.go

# 全面测试（17 个示例 — 嵌套结构、7 层嵌套、大型结构、边界用例）
go run examples/complex/main.go

# 性能基准（ASON vs JSON，吞吐量，内存占用，8 大测试板块）
go run examples/bench/main.go
```

## ASON 格式规范

完整的 [ASON 规范](https://github.com/athxx/ason/blob/main/docs/ASON_SPEC_CN.md) 包含语法规则、BNF 文法、转义规则、类型系统及 LLM 集成最佳实践。

### 语法速查表

| 元素     | Schema 语法                 | 数据语法            |
| -------- | --------------------------- | ------------------- |
| 对象     | `{field1:type,field2:type}` | `(val1,val2)`       |
| 简单数组 | `field:[type]`              | `[v1,v2,v3]`        |
| 对象数组 | `field:[{f1:type,f2:type}]` | `[(v1,v2),(v3,v4)]` |
| 字典     | `field:map[K,V]`            | `[(k1,v1),(k2,v2)]` |
| 嵌套对象 | `field:{f1:type,f2:type}`   | `(v1,(v3,v4))`      |
| 空值     | —                           | _(空白)_            |
| 空字符串 | —                           | `""`                |
| 注释     | —                           | `/* ... */`         |

## 许可证

MIT
