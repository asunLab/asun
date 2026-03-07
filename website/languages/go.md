# Go Guide

The Go ASON library provides idiomatic, reflection-based encoding and decoding for ASON Text and ASON-BIN formats.

## Installation

```bash
go get github.com/your-org/ason/go
```

## Text Format

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
    // Serialize single struct
    user := User{ID: 1, Name: "Alice", Active: true}
    s, err := ason.Marshal(user)

    // Serialize slice
    users := []User{{1, "Alice", true}, {2, "Bob", false}}
    s, err = ason.MarshalSlice(users)
    fmt.Println(s)
    // {id:str,name:str,active:bool}:
    //   (1,Alice,true),
    //   (2,Bob,false)

    // Deserialize slice
    var out []User
    err = ason.UnmarshalSlice(s, &out)

    // Deserialize single
    var u User
    err = ason.Unmarshal(s, &u)
    _ = err
}
```

## Binary Format

```go
// Serialize
data, err := ason.MarshalBin(user)
data, err  = ason.MarshalBinSlice(users)

// Deserialize
var u User
err = ason.UnmarshalBin(data, &u)

var us []User
err = ason.UnmarshalBinSlice(data, &us)
```

## Field Tags

Use `ason:"fieldname"` struct tags to control field naming:

```go
type Product struct {
    ProductID   int64   `ason:"id"`
    ProductName string  `ason:"name"`
    Price       float64 `ason:"price"`
}
```

Fields without an `ason` tag use the lowercase struct field name.

## Running Examples

```bash
cd go/examples/basic    && go run main.go
cd go/examples/bench    && go run main.go
cd go/examples/complex  && go run main.go
```

## Running Tests

```bash
cd go && go test ./...
```
