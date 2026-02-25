# 语法参考

ASON 文本格式的完整参考文档。

## 文档结构

每个 ASON 文档的形式为：

```
schema ":" data
```

- **Schema** —— 用 `{ }` 包裹，定义字段名和可选类型提示
- **冒号** —— 分隔 Schema 与数据
- **Data** —— 一个或多个用 `( )` 包裹的元组

## Schema

```ason
{field1, field2, field3}
{field1:type1, field2:type2}
```

字段以逗号分隔，类型注解可选，字段名前后的空白被忽略。

### 支持的类型注解

| 注解 | 含义 |
|------|------|
| `int` | 整数（任意 `i64` 范围值） |
| `float` | IEEE 754 双精度浮点 |
| `str` | 字符串（带引号或不带引号） |
| `bool` | 布尔值（`true` / `false`） |
| `{...}` | 嵌套结构体 Schema |
| `[type]` | `type` 类型的数组 |

## 数据元组

```ason
(value1, value2, value3)
```

值以逗号分隔，必须与 Schema 顺序对应。值前后的空白被去除。多个元组以逗号分隔：

```ason
{id:int, name:str}:
  (1, Alice),
  (2, Bob),
  (3, Carol)
```

## 标量值

### 整数

```ason
42
-100
0
```

任意 `i64` 范围内的值，十进制表示，可有前置 `-`，禁止前置零。

### 浮点数

```ason
3.14
-0.5
1e10
```

IEEE 754 双精度，支持科学计数法，不支持 `NaN` 和 `Inf`（用空槽表示 null）。

### 布尔值

```ason
true
false
```

严格小写，区分大小写。

### Null / None

空槽 —— 两个逗号之间（或逗号与右括号之间）没有任何字符：

```ason
{id, name, score}:
  (1, Alice, 9.5),
  (2, Bob,      )
```

Bob 的 score 为 null / None。

### 不带引号的字符串

不包含 `,`、`(`、`)`、`[`、`]`、`\` 的 Unicode 字符序列。解析器自动去除首尾 ASCII 空白：

```ason
Alice Smith
hello world
```

### 带引号的字符串

用双引号包裹，可保留空白或包含保留字符：

```ason
"  leading spaces  "
"value with, comma"
"line\nnewline"
```

带引号字符串内的转义序列：

| 转义 | 含义 |
|------|------|
| `\\` | 字面反斜杠 |
| `\"` | 双引号 |
| `\n` | 换行 |
| `\t` | 制表符 |
| `\r` | 回车 |
| `\,` | 字面逗号（不带引号的上下文） |

## 嵌套结构

嵌套对象用嵌套 Schema 和嵌套元组表示：

```ason
{id:int, address:{city:str, zip:str}}:
  (1, (Berlin, 10115)),
  (2, (Paris,  75001))
```

## 数组

数组字段在 Schema 和数据中均使用 `[...]`：

```ason
{id:int, tags:[str]}:
  (1, [rust, go]),
  (2, [python, c++])
```

数组可以为空：

```ason
(1, [])
```

## 枚举

枚举变体使用 `变体::值` 形式：

```ason
Role::Admin
Status::Active
```

## 单结构体简写

序列化单个结构体（非列表）时，Schema 和数据可省略换行，写在同一行：

```ason
{id:int, name:str, active:bool}:(1, Alice, true)
```

## 空白规则

- Token 之间的空白被忽略
- 不带引号的字符串值首尾空白被去除
- 元组之间的换行可选，但推荐加上以提高可读性
