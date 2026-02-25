# Rust 指南

Rust 实现是 ASON 的参考实现，使用 `serde` 进行序列化，支持包括 SIMD 加速二进制编码在内的全部 ASON 特性。

## 安装

```toml
# Cargo.toml
[dependencies]
ason = "0.1"
serde = { version = "1", features = ["derive"] }
```

## StructSchema Trait

ASON 要求结构体实现 `StructSchema`，以便序列化器在运行时获知字段名和类型：

```rust
use ason::{StructSchema, Result};
use serde::{Serialize, Deserialize};

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

## 文本格式 API

```rust
use ason::{to_string, to_string_vec, to_string_typed, to_string_vec_typed,
           from_str, from_str_vec};

// 序列化单个结构体
let s = to_string(&user)?;
// → "{id,name,active}:(1,Alice,true)"

// 序列化 Vec<T>
let s = to_string_vec(&users)?;
// → "{id,name,active}:\n  (1,Alice,true),\n  (2,Bob,false)"

// 带类型注解
let s = to_string_typed(&user)?;
// → "{id:int,name:str,active:bool}:(1,Alice,true)"

// 反序列化单个结构体
let user: User = from_str("{id,name,active}:(1,Alice,true)")?;

// 反序列化 Vec<T>
let users: Vec<User> = from_str_vec(&s)?;
```

## 二进制格式 API

```rust
use ason::{to_bin, to_bin_vec, from_bin, from_bin_vec};

// 单个值
let bytes: Vec<u8> = to_bin(&user)?;
let user: User      = from_bin(&bytes)?;

// 切片
let bytes: Vec<u8>  = to_bin_vec(&users)?;
let users: Vec<User> = from_bin_vec(&bytes)?;
```

### 零拷贝二进制反序列化

```rust
#[derive(Deserialize)]
struct BorrowedUser<'a> {
    id: i64,
    name: &'a str,   // 零拷贝，直接切片输入字节
    active: bool,
}

let bytes = to_bin(&user)?;
let u: BorrowedUser = from_bin(&bytes)?;
```

## 美化打印

```rust
use ason::to_string_pretty;

let pretty = to_string_pretty(&user)?;
```

## 运行示例

```bash
cargo run --release --example basic
cargo run --release --example bench
cargo run --release --example complex
cargo run --release --example cross_compat
```

## 运行测试

```bash
cargo test
```
