# ason-python

[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)
[![Python](https://img.shields.io/badge/python-3.9+-3776AB.svg)](https://python.org)

高性能的 [ASON](https://github.com/athxx/ason)（Array-Schema Object Notation）Python dump/load 库 — 一种面向 LLM 交互和大规模数据传输的 token 高效、模式驱动数据格式。

**纯 Python 实现，零依赖（仅标准库），基于 dataclass 驱动。**

[English Documentation](README.md)

## 什么是 ASON？

ASON 将**模式**与**数据**分离，消除 JSON 中重复的键名。模式只声明一次，数据行只携带值：

```text
JSON（100 tokens）：
{"users":[{"id":1,"name":"Alice","active":true},{"id":2,"name":"Bob","active":false}]}

ASON（约 35 tokens，节省 65%）：
[{id:int, name:str, active:bool}]:(1,Alice,true),(2,Bob,false)
```

| 方面       | JSON     | ASON         |
| ---------- | -------- | ------------ |
| Token 效率 | 100%     | 30–70% ✓     |
| 键名重复   | 每个对象 | 只声明一次 ✓ |
| 可读性     | 是       | 是 ✓         |
| 嵌套结构   | ✓        | ✓            |
| 类型注解   | 无       | 可选 ✓       |
| 数据大小   | 100%     | **40–50%** ✓ |

## 快速开始

将 `ason.py` 复制到你的项目中即可使用 — 无需安装。

### 序列化与反序列化结构体

```python
from dataclasses import dataclass
import ason

@dataclass
class User:
    id: int
    name: str
    active: bool

user = User(id=1, name="Alice", active=True)

# 序列化
s = ason.dump(user)
print(s)
# 输出: {id,name,active}:(1,Alice,true)

# 反序列化
user2 = ason.load(s, User)
print(user == user2)  # True
```

### 带类型注解的序列化

使用 `dump_typed` 输出带类型注解的模式 — 适用于文档、LLM 提示和跨语言交换：

```python
s = ason.dump_typed(user)
# 输出: {id:int,name:str,active:bool}:(1,Alice,true)

# load 同时接受带注解和不带注解的模式
user2 = ason.load(s, User)
```

### 列表的序列化与反序列化（模式驱动）

对于 `list[T]`，ASON 只写一次模式，每个元素作为紧凑的元组 — 这是相比 JSON 的关键优势：

```python
users = [
    User(id=1, name="Alice", active=True),
    User(id=2, name="Bob", active=False),
]

# 不带注解的模式
s = ason.dump_slice(users)
# 输出: [{id,name,active}]:(1,Alice,true),(2,Bob,false)

# 带类型注解的模式
s2 = ason.dump_slice_typed(users)
# 输出: [{id:int,name:str,active:bool}]:(1,Alice,true),(2,Bob,false)

# 反序列化 — 两种形式都支持
parsed = ason.load_slice(s, User)
```

## 支持的类型

| 类型                  | ASON 表示           | 示例                          |
| --------------------- | ------------------- | ----------------------------- |
| `int`                 | 纯数字              | `42`, `-100`                  |
| `float`               | 小数                | `3.14`, `-0.5`                |
| `bool`                | 字面量              | `true`, `false`               |
| `str`                 | 无引号或带引号      | `Alice`, `"Carol Smith"`      |
| `Optional[T]`         | 值或空              | `hello` 或 _(空白)_ 表示 None |
| `list[T]`             | `[v1,v2,v3]`        | `[rust,go,python]`            |
| `dict[str, T]`        | `[(k1,v1),(k2,v2)]` | `[(age,30),(score,95)]`       |
| 嵌套 dataclass        | `(field1,field2)`   | `(Engineering,500000)`        |
| `list[list[T]]`       | `[[v1,v2],[v3]]`    | `[[1,2],[3,4,5]]`             |
| `list[list[list[T]]]` | 三层嵌套            | `[[[1,2],[3,4]]]`             |

### 数据类字段

库使用 `dataclasses` 和 `typing.get_type_hints()` 自动发现字段名和类型：

```python
@dataclass
class Employee:
    id: int
    name: str
    dept: Department
    skills: list[str]
    active: bool
```

### 嵌套结构体

```python
@dataclass
class Dept:
    title: str

@dataclass
class Employee:
    name: str
    dept: Dept

# 模式反映嵌套：
# {name,dept}:(Alice,(Engineering))
```

### 可选字段

```python
from typing import Optional

@dataclass
class Item:
    id: int
    label: Optional[str] = None

# 有值:  {id,label}:(1,hello)
# 为 None: {id,label}:(1,)
```

### 列表与字典

```python
@dataclass
class Tagged:
    name: str
    tags: list[str]
# {name,tags}:(Alice,[rust,go,python])

@dataclass
class Profile:
    name: str
    attrs: dict[str, int]
# {name,attrs}:(Alice,[(age,30),(score,95)])
```

### 类型注解（可选）

ASON 模式支持**可选**的类型注解。两种形式完全等价 — 加载器对它们的处理完全相同：

```text
// 不带注解（dump / dump_slice 的默认输出）
{id,name,salary,active}:(1,Alice,5000.50,true)

// 带注解（dump_typed / dump_slice_typed 的输出）
{id:int,name:str,salary:float,active:bool}:(1,Alice,5000.50,true)
```

注解是**纯粹的装饰性元数据** — 不影响解析或加载行为。加载器在遇到 `:type` 部分时直接跳过。

**何时使用注解：**

- LLM 提示 — 帮助模型理解和生成正确数据
- API 文档 — 无需外部文档的自描述模式
- 跨语言交换 — 消除类型歧义（`42` 是 int 还是 float？）
- 调试 — 一目了然地查看数据类型

### 注释

```text
/* 用户列表 */
[{id:int, name:str, active:bool}]:(1,Alice,true),(2,Bob,false)
```

### 多行格式

```text
[{id:int, name:str, active:bool}]:
  (1, Alice, true),
  (2, Bob, false),
  (3, "Carol Smith", true)
```

## API 参考

| 函数                                         | 描述                                           |
| -------------------------------------------- | ---------------------------------------------- |
| `dump(obj) -> str`                           | 序列化 dataclass → 不带注解 `{id,name}:`       |
| `dump_typed(obj) -> str`                     | 序列化 dataclass → 带注解 `{id:int,name:str}:` |
| `load(data, cls) -> obj`                     | 反序列化 dataclass（支持带注解和不带注解）     |
| `dump_slice(objs) -> str`                    | 序列化列表 → 不带注解（模式驱动）              |
| `dump_slice_typed(objs, field_types) -> str` | 序列化列表 → 带注解                            |
| `load_slice(data, cls) -> list`              | 反序列化列表（支持带注解和不带注解）           |

## 性能

在 Apple Silicon（M 系列）上基准测试，Python 3.9，与 `json` 模块（C 实现）对比：

### 体积节省（与所有 ASON 实现一致）

| 场景               | JSON   | ASON   | 节省    |
| ------------------ | ------ | ------ | ------- |
| 扁平结构 × 100     | 13 KB  | 5.6 KB | **59%** |
| 扁平结构 × 1000    | 137 KB | 57 KB  | **59%** |
| 5 层深度 × 10      | 49 KB  | 17 KB  | **66%** |
| 5 层深度 × 50      | 249 KB | 87 KB  | **65%** |
| 100 个国家（嵌套） | 215 KB | 73 KB  | **66%** |

### 速度 vs json 模块

Python 的 `json` 模块使用 C 实现（`_json` 扩展），因此在原始速度上必然快于任何纯 Python 代码。尽管如此，ASON Python 提供：

- **59-66% 更小的输出** — 显著节省网络传输和 LLM token 成本
- **模式驱动结构** — 类型化、自描述的格式，非常适合 AI/LLM 管道
- **零依赖** — 单文件，复制即用
- **完整往返** — 全精度浮点数、转义字符串、三层嵌套数组

对于延迟敏感的场景，建议使用 [Rust](../rust/) 或 [Go](../go/) ASON 实现，它们在速度上超越各自的 JSON 库。

### 为什么 ASON 即使慢于 C 实现的 json 仍然有价值

1. **网络是瓶颈** — ASON 59-66% 的体积缩减节省的网络传输时间远超序列化开销
2. **LLM token 效率** — 减少 30-70% 的 token = 更低成本和更快响应
3. **模式验证** — 类型化模式防止数据错误
4. **跨语言** — 同一格式跨 Rust、Go、Python 等语言使用

运行基准测试：

```bash
python examples/bench.py
```

## 实现亮点

- **纯 Python** — 仅标准库（`dataclasses`、`typing`、`math`）
- **字节级解析** — 输入转换为 `bytes` 进行快速索引访问
- **查找表** — `bytearray(256)` 用于引号判断，`dict` 用于转义字符
- **结构信息缓存** — 类型提示和字段元数据按类缓存
- **`__slots__`** — 所有内部类使用 `__slots__` 最小化内存
- **尽可能零拷贝** — `bytes.decode()` 直接从解析范围转换，无转义时不构造中间字符串
- **快速浮点格式化** — 整数、1 位小数、2 位小数快速路径；`repr()` 保证完整精度

## 示例

```bash
# 基本用法（11 个示例）
python examples/basic.py

# 综合示例（17 个 — 嵌套结构、7 层嵌套、大型结构、边界条件）
python examples/complex.py

# 性能基准测试（ASON vs JSON，吞吐量，内存，8 个部分）
python examples/bench.py
```

## 测试

```bash
python test_ason.py
# 33 个测试覆盖：往返、类型/非类型模式、可选字段、
# 嵌套结构、字典、数组、三层嵌套数组、注释、边界条件
```

## ASON 格式规范

完整的 [ASON 规范](https://github.com/athxx/ason/blob/main/docs/ASON_SPEC_CN.md) 包含语法规则、BNF 语法、转义规则、类型系统和 LLM 集成最佳实践。

### 语法快速参考

| 元素     | 模式                        | 数据                |
| -------- | --------------------------- | ------------------- |
| 对象     | `{field1:type,field2:type}` | `(val1,val2)`       |
| 数组     | `field:[type]`              | `[v1,v2,v3]`        |
| 对象数组 | `field:[{f1:type,f2:type}]` | `[(v1,v2),(v3,v4)]` |
| 映射     | `field:map[K,V]`            | `[(k1,v1),(k2,v2)]` |
| 嵌套对象 | `field:{f1:type,f2:type}`   | `(v1,(v3,v4))`      |
| 空值     | —                           | _(空白)_            |
| 空字符串 | —                           | `""`                |
| 注释     | —                           | `/* ... */`         |

## 许可证

MIT
