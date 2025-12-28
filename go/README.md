# ASON Go Library

A Go implementation of the ASON (Array-Schema Object Notation) format.

## Installation

```bash
go get github.com/athxx/ason/go
```

## Usage

### Parsing

```go
import "github.com/athxx/ason/go"

// Parse a simple object
v, err := ason.Parse("{name,age}:(Alice,30)")
if err == nil {
    name := v.Field("name").AsString()  // "Alice"
    age := v.Field("age").AsInteger()   // 30
}

// Parse multiple objects
v, _ := ason.Parse("{name,age}:(Alice,30),(Bob,25)")
for i := 0; i < v.Len(); i++ {
    obj := v.Get(i)
    fmt.Println(obj.Field("name").AsString())
}

// Parse nested objects
v, _ := ason.Parse("{name,addr{city,zip}}:(Alice,(NYC,10001))")
city := v.Field("addr").Field("city").AsString()

// Parse arrays
v, _ := ason.Parse("{name,scores[]}:(Alice,[90,85,95])")
scores := v.Field("scores")
for i := 0; i < scores.Len(); i++ {
    fmt.Println(scores.Get(i).AsInteger())
}
```

### Serializing

```go
// Build an object
obj := ason.Object()
obj.Set("name", ason.String("Alice"))
obj.Set("age", ason.Integer(30))

// Serialize data only
fmt.Println(ason.Serialize(obj))  // (Alice,30)

// Serialize with schema
fmt.Println(ason.SerializeWithSchema(obj))  // {name,age}:(Alice,30)
```

### Value Types

```go
// Create values
null := ason.Null()
b := ason.Bool(true)
i := ason.Integer(42)
f := ason.Float(3.14)
s := ason.String("hello")
arr := ason.Array(ason.Integer(1), ason.Integer(2))
obj := ason.Object()

// Type checking
v.IsNull()
v.IsBool()
v.IsInteger()
v.IsFloat()
v.IsNumber()  // integer or float
v.IsString()
v.IsArray()
v.IsObject()

// Value access
v.AsBool()     // bool
v.AsInteger()  // int64
v.AsFloat()    // float64
v.AsString()   // string

// Array operations
arr.Len()
arr.Get(0)
arr.Push(item)
arr.Items()

// Object operations
obj.Set("key", value)
obj.Field("key")
obj.Keys()
obj.Len()

// Clone
copy := v.Clone()
```

## Testing

```bash
go test ./ason/... -v
```

## Example

```bash
go run ./examples/basic/
```

## License

MIT License

