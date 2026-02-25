# ASON 格式规范 v1.0

> **ASON** = Array-Schema Object Notation（数组模式对象符号）  
> _"The efficiency of arrays, the structure of objects."_

ASON 是一种专为大规模数据传输与 LLM 交互设计的序列化格式。本页为规范性文档，所有语言实现必须遵守此处定义的规则。

## 1. 设计原则

- **Schema 与数据分离** —— 结构只声明一次，数据行中不重复
- **行导向（Row-Oriented）** —— 多条记录以连续元组形式出现在单一 Schema Header 之后
- **方圆结合（Structural Harmony）** —— `{}` 定义骨架（Schema），`()` 承载数据（values）
- **Token 极致优化** —— 最少语法字符，比等价 JSON 少 30–70% Token
- **极致解析性能** —— 零哈希匹配、模式驱动解析、内存占用极小

## 2. 形式化语法

```abnf
document    = schema ":" SP data
schema      = "{" field-list "}"
field-list  = field *("," SP field)
field       = name [":" type]
name        = 1*(ALPHA / DIGIT / "_")
type        = primitive / schema / "[" type "]"
primitive   = "int" / "float" / "str" / "bool"

data        = tuple / tuple-list
tuple-list  = tuple *("," NL SP tuple)
tuple       = "(" value-list ")"
value-list  = value *("," value)
value       = scalar / tuple / array / ""
array       = "[" [value-list] "]"

scalar      = integer / float / boolean / quoted-str / unquoted-str
integer     = ["-"] 1*DIGIT
float       = ["-"] 1*DIGIT "." 1*DIGIT [("e"/"E") ["+"/"-"] 1*DIGIT]
boolean     = "true" / "false"
quoted-str  = DQUOTE *char DQUOTE
unquoted-str = 1*safe-char
safe-char   = 任意 Unicode，除 "," / "(" / ")" / "[" / "]" / "\"
```

## 3. 数据类型

### 3.1 整数
`i64` 范围（`-9223372036854775808` 到 `9223372036854775807`）内的任意值。十进制表示，可有前置 `-`，禁止前置零。

### 3.2 浮点数
IEEE 754 双精度。格式为 `[-]digits.digits`，可有 `e` / `E` 指数。`NaN` 和 `Inf` 不可在 ASON Text 中表示（改用空槽）。

### 3.3 布尔值
严格为 `true` 或 `false`，区分大小写，仅小写。

### 3.4 Null / None
空槽 —— 两个逗号之间或最后一个逗号与右括号 `)` 之间没有任何字符。所有实现必须将其映射到各语言原生的 null / option / 指针类型。

### 3.5 不带引号的字符串
不包含 `,`、`(`、`)`、`[`、`]`、`\` 的 Unicode 码点序列。解析器**去除**首尾 ASCII 空白（空格、制表符）。

### 3.6 带引号的字符串
用双引号 `"` 包裹的字符序列，支持转义序列：

| 转义 | Unicode |
|------|---------|
| `\\` | U+005C |
| `\"` | U+0022 |
| `\n` | U+000A |
| `\t` | U+0009 |
| `\r` | U+000D |
| `\,` | U+002C（仅限不带引号的上下文） |

### 3.7 数组
`[` `]` 内的逗号分隔值，可以为空（`[]`）。元素遵循相同的 `value` 规则。

### 3.8 嵌套结构体
嵌套结构体 Schema 内联出现在父 Schema 中：

```ason
{id:int, address:{city:str, zip:str}}
```

对应内层元组作为数据：

```ason
(1, (Berlin, 10115))
```

## 4. 规范化形式

合规工具生成的 ASON 文档**应当**：

1. 在 Schema 中包含类型注解（`{id:int, name:str}`）
2. 每个元组独占一行，缩进两个空格
3. `:` 之后换行，首个数据元组缩进出现

```ason
{id:int, name:str, active:bool}:
  (1, Alice, true),
  (2, Bob,   false)
```

非规范形式的空白对解析有效。

## 5. 单结构体简写

仅含一个元组的文档**可以**省略换行：

```ason
{id:int, name:str}:(1, Alice)
```

## 6. 二进制格式（ASON-BIN）

ASON-BIN 是独立的编码，不使用文本 Schema 语法。字段按声明顺序紧密排列，无填充。所有整数**小端序**。

完整线格式编码表见[二进制格式指南](/zh/guide/binary-format)。

## 7. 版本管理

本文档描述 **ASON v1.0**。未来版本：向后兼容的新增功能递增次版本号，破坏性变更递增主版本号。

## 8. 合规要求

合规实现**必须**：

1. 无错误地解析本规范中的所有示例
2. 生成可被任何其他合规实现解析的输出
3. 通过跨语言兼容性测试套件（`*/tests/cross_compat_test.*`）
4. 在文本编解码往返中保留所有非 null 标量值
