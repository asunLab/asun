# 快速开始

按照你使用的语言选择对应的五分钟教程。

## Rust

### 安装

```toml
# Cargo.toml
[dependencies]
ason = "0.1"
serde = { version = "1", features = ["derive"] }
```

### 快速示例

```rust
use ason::{to_string_vec, from_str_vec, StructSchema, Result};
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
}

fn main() -> Result<()> {
    let users = vec![
        User { id: 1, name: "Alice".into(), active: true },
        User { id: 2, name: "Bob".into(),   active: false },
    ];

    let s = to_string_vec(&users)?;
    // "{id:int,name:str,active:bool}:\n  (1,Alice,true),\n  (2,Bob,false)"

    let restored: Vec<User> = from_str_vec(&s)?;
    assert_eq!(users, restored);
    Ok(())
}
```

查看完整 [Rust 指南](/zh/languages/rust)。

---

## Go

### 安装

```bash
go get github.com/your-org/ason/go
```

### 快速示例

```go
package main

import (
    "fmt"
    "github.com/your-org/ason/go"
)

type User struct {
    ID     int64  `ason:"id"`
    Name   string `ason:"name"`
    Active bool   `ason:"active"`
}

func main() {
    users := []User{
        {1, "Alice", true},
        {2, "Bob", false},
    }
    s, _ := ason.MarshalSlice(users)
    fmt.Println(s)

    var out []User
    ason.UnmarshalSlice(s, &out)
}
```

查看完整 [Go 指南](/zh/languages/go)。

---

## Python

### 安装

```bash
# 将 ason.py 复制到你的项目（零外部依赖）
cp /path/to/ason/python/ason.py .
```

### 快速示例

```python
from ason import encode_rows, decode_rows
from dataclasses import dataclass

@dataclass
class User:
    id: int
    name: str
    active: bool

users = [User(1, "Alice", True), User(2, "Bob", False)]

text = encode_rows(users)
print(text)
# {id,name,active}:
#   (1,Alice,true),
#   (2,Bob,false)

restored = decode_rows(text, User)
```

查看完整 [Python 指南](/zh/languages/python)。

---

## 其他语言

| 语言 | 指南 |
|------|------|
| C    | [C 指南](/zh/languages/c) |
| C++  | [C++ 指南](/zh/languages/cpp) |
| Java | [Java 指南](/zh/languages/java) |
| Zig  | [Zig 指南](/zh/languages/zig) |

---

## VS Code 插件

从插件市场安装 **ASON** 插件，获得：
- `.ason` 文件语法高亮
- 实时预览面板
- LSP 驱动的诊断信息

```bash
# 或直接使用内置 LSP 服务器
cd tools/ason-lsp && go build -o ason-lsp . && sudo mv ason-lsp /usr/local/bin/
```
