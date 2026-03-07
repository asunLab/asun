# Getting Started

Pick your language below for a five-minute quickstart.

## Rust

### Installation

```toml
# Cargo.toml
[dependencies]
ason = "0.1"
serde = { version = "1", features = ["derive"] }
```

### Quickstart

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

See the full [Rust guide](/languages/rust).

---

## Go

### Installation

```bash
go get github.com/your-org/ason/go
```

### Quickstart

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

See the full [Go guide](/languages/go).

---

## Python

### Installation

```bash
# copy ason.py into your project (zero external dependencies)
cp /path/to/ason/python/ason.py .
```

### Quickstart

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

See the full [Python guide](/languages/python).

---

## Other Languages

| Language | Guide |
|----------|-------|
| C        | [C guide](/languages/c) |
| C++      | [C++ guide](/languages/cpp) |
| Java     | [Java guide](/languages/java) |
| Zig      | [Zig guide](/languages/zig) |

---

## VS Code Extension

Install the **ASON** extension from the marketplace for:
- Syntax highlighting for `.ason` files
- Live preview panel
- LSP-powered diagnostics

```bash
# Or use the built-in LSP server
cd tools/ason-lsp && go build -o ason-lsp . && sudo mv ason-lsp /usr/local/bin/
```
