# 数据类型

ASON 支持丰富的标量和复合类型。本页记录它们的文本格式表示及跨语言等价类型。

## 标量类型

| ASON 类型 | 示例 | 说明 |
|-----------|------|------|
| `int` | `42`, `-100`, `0` | 任意 `i64` 范围整数 |
| `float` | `3.14`, `-0.5`, `1e10` | IEEE 754 双精度 |
| `bool` | `true`, `false` | 严格小写 |
| `str`（不带引号） | `Alice Smith` | 自动去除首尾空白，`\,` 转义逗号 |
| `str`（带引号） | `" spaces "` | 保留空白，支持 `\"` `\n` `\t` |
| null / None | _（空槽）_ | 两个逗号之间没有字符 |

## 复合类型

### 嵌套结构体

```ason
// Schema
{id:int, address:{city:str, zip:str}}

// 数据
(1, (Berlin, 10115))
```

### 数组 / 列表

```ason
// Schema
{id:int, tags:[str]}

// 数据
(1, [rust, go, wasm])
```

数组可以嵌套：

```ason
[{matrix:[[int]]}]:
  ([[1, 2], [3, 4]]),
  ([[5, 6], [7, 8]])
```

### 枚举

```ason
Role::Admin
Status::Active
Color::Rgb
```

枚举变体使用 `::` 命名空间分隔。

### 可选值 / 可空

空槽（逗号间没有任何字符）表示 `None` / `null`：

```ason
[{id:int, score:float}]:
  (1, 9.5),
  (2,    )
```

## 跨语言类型映射

| ASON 类型 | Rust | Go | Python | Java | C |
|-----------|------|----|--------|------|---|
| `int` | `i64` | `int64` | `int` | `long` | `int64_t` |
| `float` | `f64` | `float64` | `float` | `double` | `double` |
| `bool` | `bool` | `bool` | `bool` | `boolean` | `bool` |
| `str` | `String` / `&str` | `string` | `str` | `String` | `char*` |
| null | `Option<T>` | 指针 / `*T` | `Optional` | `Optional<T>` | 可空指针 |
| `[T]` | `Vec<T>` | `[]T` | `list` | `List<T>` | `T[]` |
| 嵌套结构体 | struct | struct | dataclass | class | struct |

## 整数大小（ASON-BIN）

在二进制格式中，整数使用其原生大小：

| 类型 | 字节数 |
|------|--------|
| `i8` / `u8` | 1 |
| `i16` / `u16` | 2 |
| `i32` / `u32` | 4 |
| `i64` / `u64` | 8 |

所有整数在 ASON-BIN 中均为**小端序**。
