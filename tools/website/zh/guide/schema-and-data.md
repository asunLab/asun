# Schema 与数据

ASON 最重要的概念是 **Schema（结构）** 与 **数据（值）** 的分离。

## Schema 语法

Schema 以 `{` 和 `}` 包裹，内含逗号分隔的字段定义：

```ason
{name, age, active}
```

字段可选地包含类型注解：

```ason
{name:str, age:int, active:bool}
```

类型注解是为人类和工具提供的可选提示 —— 当目标类型已由宿主语言类型系统确定时，它们不影响解析行为。

## 数据语法

数据跟在 Schema 和 `:` 分隔符之后，每行是一个用 `(` `)` 包裹的元组：

```ason
{name:str, age:int}:
  (Alice, 30),
  (Bob,   25)
```

对于单个结构体（非列表），Schema 和数据可以写在同一行：

```ason
{name, age}:(Alice, 30)
```

## 嵌套 Schema

Schema 可以嵌套，表示嵌套对象：

```ason
{id:int, address:{city:str, zip:str}}:
  (1, (Berlin, 10115)),
  (2, (Paris,  75001))
```

内层 Schema `{city:str, zip:str}` 对应内层元组 `(Berlin, 10115)`。

## 数组字段

包含列表的字段使用 `[type]` 记法：

```ason
{id:int, tags:[str]}:
  (1, [rust, go]),
  (2, [python, c++])
```

## 可选值 / Null

两个逗号之间的空槽（什么都没有）表示 `null` / `None`：

```ason
{id:int, name:str, score:float}:
  (1, Alice, 9.5),
  (2, Bob,       )
```

`Bob` 的 score 是 `None`（末尾空槽）。

## 语法概要

```
document    = schema ":" rows
schema      = "{" fields "}"
fields      = field ("," field)*
field       = name (":" type)?
type        = "int" | "float" | "str" | "bool" | schema | "[" type "]"
rows        = tuple | (tuple ",")* tuple
tuple       = "(" values ")"
values      = value ("," value)*
value       = scalar | tuple | "[" values "]" | ""
scalar      = unquoted_string | quoted_string | number | bool | null
```

查看完整 [语法参考](/zh/reference/syntax) 了解转义规则、空白处理和边界情况。
