# ASON - Array-Schema Object Notation

[![Rust](https://img.shields.io/badge/rust-1.85+-orange.svg)](https://www.rust-lang.org)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

一个 **Token 高效** 的数据格式，通过分离 Schema 和数据来减少冗余。专为 LLM 交互优化。

## 特性

- 🚀 **Token 高效** - 字段名只声明一次，减少 30-50% token
- 📦 **Schema 分离** - 结构与数据分离，批量数据更紧凑
- 🔧 **Zero-Copy 解析** - 可选的零拷贝模式，高性能场景
- 🦀 **Serde 支持** - 与 Rust 生态无缝集成

## 快速开始

```toml
[dependencies]
ason = "0.1"
```

## 基本用法

### 解析 ASON

```rust
use ason::parse;

// 单个对象
let value = parse("{name,age}:(Alice,30)").unwrap();
assert_eq!(value.get("name").unwrap().as_str(), Some("Alice"));
assert_eq!(value.get("age").unwrap().as_i64(), Some(30));

// 多条记录
let users = parse("{name,age}:(Alice,30),(Bob,25)").unwrap();
let arr = users.as_array().unwrap();
assert_eq!(arr.len(), 2);
```

### 序列化

```rust
use ason::{Value, to_string, to_string_pretty};
use indexmap::IndexMap;

let mut obj = IndexMap::new();
obj.insert("name".to_string(), Value::String("Alice".to_string()));
obj.insert("age".to_string(), Value::Integer(30));
let value = Value::Object(obj);

assert_eq!(to_string(&value), "(Alice,30)");
```

### Serde 集成

```rust
use ason::{parse, from_value, to_value};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize)]
struct User {
    name: String,
    age: i64,
}

// ASON → Rust struct
let value = parse("{name,age}:(Alice,30)").unwrap();
let user: User = from_value(&value).unwrap();

// Rust struct → ASON Value
let value = to_value(&user).unwrap();
```

### Zero-Copy 解析 (高性能)

```rust
use ason::zero_copy;

let input = "{name,age}:(Alice,30)";
let value = zero_copy::parse(input).unwrap();

// 字符串直接借用输入，无内存分配
assert_eq!(value.get("name").unwrap().as_str(), Some("Alice"));

// 需要独立存在时，转换为 owned
let owned = value.into_owned();
```

## ASON 语法速览

| 语法 | 示例 | 说明 |
|------|------|------|
| 对象 | `{name,age}:(Alice,30)` | Schema + 数据 |
| 多记录 | `{id}:(1),(2),(3)` | 同结构批量数据 |
| 嵌套 | `{user{name}}:((Alice))` | 嵌套对象 |
| 数组字段 | `{scores[]}:([90,85])` | 简单数组 |
| 对象数组 | `{users[{id}]}:([(1),(2)])` | 对象数组 |
| 纯数组 | `[1,2,3]` | 无 Schema |
| 空值 | `{a,b}:(1,)` | b = null |

### 与 JSON 对比

```
JSON (62 tokens):
[{"name":"Alice","age":30},{"name":"Bob","age":25}]

ASON (约 35 tokens):
{name,age}:(Alice,30),(Bob,25)
```

## API 概览

### 解析

| 函数 | 说明 |
|------|------|
| `parse(input)` | 解析 ASON 字符串 |
| `zero_copy::parse(input)` | 零拷贝解析 |

### 序列化

| 函数 | 说明 |
|------|------|
| `to_string(&value)` | 紧凑输出 |
| `to_string_pretty(&value)` | 格式化输出 |

### Serde

| 函数 | 说明 |
|------|------|
| `to_value(&T)` | Rust 值 → ASON Value |
| `from_value(Value)` | ASON Value → Rust 值 |

## 运行示例

```bash
cargo run --example basic
```

## License

MIT

