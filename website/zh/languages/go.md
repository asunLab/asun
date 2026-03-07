# Go 指南

Go ASON 库通过反射提供惯用的编解码 API，支持 ASON Text 和 ASON-BIN 格式。

## 安装

```bash
go get github.com/your-org/ason/go
```

## 文本格式

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
    // 序列化单个结构体
    user := User{ID: 1, Name: "Alice", Active: true}
    s, err := ason.Marshal(user)

    // 序列化切片
    users := []User{{1, "Alice", true}, {2, "Bob", false}}
    s, err = ason.MarshalSlice(users)
    fmt.Println(s)
    // {id:str,name:str,active:bool}:
    //   (1,Alice,true),
    //   (2,Bob,false)

    // 反序列化切片
    var out []User
    err = ason.UnmarshalSlice(s, &out)

    // 反序列化单个
    var u User
    err = ason.Unmarshal(s, &u)
    _ = err
}
```

## 二进制格式

```go
// 序列化
data, err := ason.MarshalBin(user)
data, err  = ason.MarshalBinSlice(users)

// 反序列化
var u User
err = ason.UnmarshalBin(data, &u)

var us []User
err = ason.UnmarshalBinSlice(data, &us)
```

## 字段标签

使用 `ason:"fieldname"` 结构体标签控制字段命名：

```go
type Product struct {
    ProductID   int64   `ason:"id"`
    ProductName string  `ason:"name"`
    Price       float64 `ason:"price"`
}
```

没有 `ason` 标签的字段使用结构体字段名的小写形式。

## 运行示例

```bash
cd go/examples/basic    && go run main.go
cd go/examples/bench    && go run main.go
cd go/examples/complex  && go run main.go
```

## 运行测试

```bash
cd go && go test ./...
```
