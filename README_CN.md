# ASON — Array-Schema Object Notation

> _"The efficiency of arrays, the structure of objects."_
> _"数组的效率，对象的结构。"_

ASON 是一种为大规模数据传输与 LLM（大语言模型）交互专门设计的高性能序列化格式。
通过 **Schema 与数据分离** 的设计，ASON 彻底消除了 JSON 中冗余的 Key 重复，
引入类 Markdown 的行导向语法，在极致压缩 Token 的同时保持优秀的人类可读性。

---

## 为什么选择 ASON？

### Token 效率

```json
// JSON — 100 tokens
{
  "users": [
    { "id": 1, "name": "Alice", "active": true },
    { "id": 2, "name": "Bob", "active": false }
  ]
}
```

```ason
// ASON — ~35 tokens（节省 65%）
{id:int, name:str, active:bool}:
  (1, Alice, true),
  (2, Bob,   false)
```

### 性能对比 JSON

| 指标         | ASON 文本       | ASON-BIN 二进制 |
| ------------ | --------------- | --------------- |
| 序列化速度   | **快 2–4×**     | **快 7–10×**    |
| 反序列化速度 | **快 1.2–2.5×** | **快 2–2.5×**   |
| 数据大小     | **小 53–61%**   | **小 38–49%**   |

---

## 安装

在 `Cargo.toml` 中添加：

```toml
[dependencies]
ason = { path = "rust" }   # 本地引用
serde = { version = "1", features = ["derive"] }
```

---

## 文本格式 API

### 数据类型

| 类型         | 示例            | 说明                       |
| ------------ | --------------- | -------------------------- |
| 整数         | `42`, `-100`    | i64 范围                   |
| 浮点         | `3.14`, `-0.5`  | IEEE 754 双精度            |
| 布尔         | `true`, `false` | 必须小写                   |
| 空值         | _(空白)_        | 逗号间无内容 = null        |
| 无引号字符串 | `Alice Smith`   | 自动 trim，需转义 `,()[]\` |
| 有引号字符串 | `" 空格 "`      | 保留原样（含空格）         |
| 数组         | `[a, b, c]`     | 嵌套列表                   |
| 嵌套结构体   | `{...}:(...)`   | 递归 Schema                |

### `StructSchema` trait

ASON 使用编译期 Schema trait 让序列化器知道字段名和类型：

```rust
use ason::{Result, StructSchema, from_str, from_str_vec, to_string, to_string_vec};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, PartialEq)]
struct User {
    id: i64,
    name: String,
    active: bool,
}

impl StructSchema for User {
    fn field_names() -> &'static [&'static str] { &["id", "name", "active"] }
    fn field_types() -> &'static [&'static str] { &["int", "str", "bool"] }
    fn serialize_fields(&self, ser: &mut ason::serialize::Serializer) -> Result<()> {
        use serde::Serialize;
        self.id.serialize(&mut *ser)?;
        self.name.serialize(&mut *ser)?;
        self.active.serialize(&mut *ser)?;
        Ok(())
    }
}
```

### 序列化

```rust
// 单结构体  →  "{id,name,active}:(1,Alice,true)"
let s = to_string(&user)?;

// Vec<T>  →  "{id,name,active}:\n  (1,Alice,true),\n  (2,Bob,false)"
let s = to_string_vec(&users)?;

// 带类型注解
let s = to_string_typed(&user)?;       // "{id:int,name:str,active:bool}:..."
let s = to_string_vec_typed(&users)?;
```

### 反序列化

```rust
// 单结构体（接受带注解和不带注解的格式）
let user: User = from_str("{id,name,active}:(1,Alice,true)")?;

// Vec<T>
let users: Vec<User> = from_str_vec(&s)?;
```

---

## 二进制格式（ASON-BIN）

ASON-BIN 是对任意 `serde` 兼容类型的紧凑二进制编码。它提供最大幅度的性能提升，
非常适合内部服务通信、缓存和存储场景。

### API

```rust
use ason::{to_bin, to_bin_vec, from_bin, from_bin_vec};
```

| 函数           | 签名                                  | 说明                               |
| -------------- | ------------------------------------- | ---------------------------------- |
| `to_bin`       | `(value: &T) -> Result<Vec<u8>>`      | 将任意 `T: Serialize` 序列化为字节 |
| `from_bin`     | `(data: &'de [u8]) -> Result<T>`      | 零拷贝反序列化（字符串直接切片）   |
| `to_bin_vec`   | `(values: &[T]) -> Result<Vec<u8>>`   | 将 slice 序列化为字节              |
| `from_bin_vec` | `(data: &'de [u8]) -> Result<Vec<T>>` | 从字节反序列化序列                 |

### 使用示例

```rust
use ason::{from_bin, from_bin_vec, to_bin, to_bin_vec};
use serde::{Deserialize, Serialize};

#[derive(Debug, Serialize, Deserialize, PartialEq)]
struct User {
    id: i64,
    name: String,
    active: bool,
}

fn main() -> ason::Result<()> {
    // 单值
    let user = User { id: 1, name: "Alice".into(), active: true };
    let bytes = to_bin(&user)?;
    let restored: User = from_bin(&bytes)?;
    assert_eq!(user, restored);

    // Vec
    let users = vec![
        User { id: 1, name: "Alice".into(), active: true },
        User { id: 2, name: "Bob".into(),   active: false },
    ];
    let bytes = to_bin_vec(&users)?;
    let restored: Vec<User> = from_bin_vec(&bytes)?;
    assert_eq!(users, restored);

    Ok(())
}
```

### 零拷贝反序列化

当结构体字段使用 `'de` 生命周期时，`from_bin` 直接返回指向输入缓冲区的字符串切片——
无需为字符串字段分配堆内存：

```rust
#[derive(Debug, Deserialize)]
struct BorrowedUser<'a> {
    id: i64,
    name: &'a str,   // 零拷贝 — 直接指向输入字节
    active: bool,
}

let bytes = to_bin(&User { id: 1, name: "Alice".into(), active: true })?;
let u: BorrowedUser = from_bin(&bytes)?;
// u.name 是指向 `bytes` 的 &str 切片，无 String 分配
```

### 二进制格式规范

所有整数均为**小端序（LE）**。字符串使用长度前缀（u32 LE + UTF-8 字节）。

| 类型             | 编码格式                           |
| ---------------- | ---------------------------------- |
| `bool`           | 1 字节（0/1）                      |
| `i8` / `u8`      | 1 字节                             |
| `i16` / `u16`    | 2 字节 LE                          |
| `i32` / `u32`    | 4 字节 LE                          |
| `i64` / `u64`    | 8 字节 LE                          |
| `f32`            | 4 字节 IEEE 754 LE                 |
| `f64`            | 8 字节 IEEE 754 LE                 |
| `char`           | 4 字节（u32 码点 LE）              |
| `str` / `String` | `[u32 长度][UTF-8 字节]`           |
| `Option<T>`      | `[u8 标记: 0=None / 1=Some][载荷]` |
| `Vec<T>` / 序列  | `[u32 元素数][元素...]`            |
| `HashMap` / 映射 | `[u32 对数][键 值 对...]`          |
| `struct`         | 字段按声明顺序，无前缀             |
| `tuple`          | 元素按顺序，无前缀                 |
| `enum`           | `[u32 变体索引][载荷]`             |
| `unit`           | 0 字节                             |

### 性能测试数据（Apple M 系列芯片，release 构建）

**扁平结构体（8 字段）**

| 测试            | JSON      | ASON 文本 | ASON-BIN     | BIN vs JSON |
| --------------- | --------- | --------- | ------------ | ----------- |
| 序列化 × 1000   | 2.19 ms   | 1.07 ms   | **0.28 ms**  | **快 7.7×** |
| 反序列化 × 1000 | 6.05 ms   | 5.10 ms   | **2.96 ms**  | **快 2.0×** |
| 数据大小 × 1000 | 121,675 B | 56,716 B  | **74,454 B** | 小 39%      |

**深层嵌套结构体（5 层嵌套，× 100）**

| 测试           | JSON      | ASON 文本 | ASON-BIN      | BIN vs JSON |
| -------------- | --------- | --------- | ------------- | ----------- |
| 序列化 × 100   | 4.19 ms   | 2.67 ms   | **0.50 ms**   | **快 8.4×** |
| 反序列化 × 100 | 9.37 ms   | 8.55 ms   | **3.72 ms**   | **快 2.5×** |
| 数据大小 × 100 | 438,112 B | 174,611 B | **225,434 B** | 小 49%      |

**单次往返（User 结构体，× 100,000）**

| 格式      | 每次耗时 | vs JSON     |
| --------- | -------- | ----------- |
| ASON-BIN  | ~182 ns  | **快 2.1×** |
| JSON      | ~375 ns  | —           |
| ASON 文本 | ~552 ns  | 慢 0.7×     |

### SIMD 加速

`to_bin` 使用平台 SIMD 指令（ARM 上的 NEON，x86-64 上的 SSE2），在 `simd_bulk_extend`
函数中以 16 字节块批量复制 ≥ 32 字节的字符串载荷，避免标量逐字节循环。

---

## 运行示例

```bash
# 基础用法
cargo run --release --example basic

# 二进制格式（零拷贝演示、格式解析、性能测试）
cargo run --release --example binary

# 综合性能测试套件（共 9 个测试节）
cargo run --release --example bench

# 复杂嵌套结构
cargo run --release --example complex
```

---

## ASON 文本格式语法

```ason
// 单结构体
{id, name, active}:(1, Alice, true)

// 结构体数组（Schema 只声明一次）
{id:int, name:str, active:bool}:
  (1, Alice, true),
  (2, Bob,   false),
  (3, Carol, true)

// 嵌套结构体
{id:int, address:{city:str, zip:str}}:
  (1, (Berlin, 10115)),
  (2, (Paris,  75001))

// 嵌套数组
{id:int, tags:[str]}:
  (1, [rust, go]),
  (2, [python])

// 枚举值
Role::Admin

// Option（空槽 = None）
{id, name, score}:
  (1, Alice, 9.5),
  (2, Bob,      )   ← score 为 None
```

---

## 许可证

MIT
