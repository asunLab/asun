# ASON — Array-Schema Object Notation

[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**ASON** 是一种紧凑的、Schema 驱动的数据格式，专为 **LLM 交互**与**高性能数据传输**而设计。它将 Schema 与数据分离 —— Key 只声明一次，数据行仅保留纯值。

[English](README.md)

---

## 为什么选择 ASON？

标准 JSON 在每条记录中都重复所有字段名。向 LLM 发送结构化数据集或在网络上传输时，这种冗余会大量消耗 Token 和带宽：

```json
[
  {"id": 1, "name": "Alice", "active": true},
  {"id": 2, "name": "Bob",   "active": false},
  {"id": 3, "name": "Carol", "active": true}
]
```

ASON 只声明 **一次** Schema，数据以紧凑元组方式流式传输：

```
[{id:int, name:str, active:bool}]:(1,Alice,true),(2,Bob,false),(3,Carol,true)
```

**减少约 65% 的 Token，信息量不变，Schema 机器可读。**

---

## ASON vs JSON

| 方面               | JSON              | ASON                      |
| ------------------ | ----------------- | ------------------------- |
| Token 效率         | 100%（基准）      | **30–70%** ✓              |
| Key 重复           | 每个对象都有      | 声明一次 ✓                |
| 类型注解           | 无                | 可选 ✓                    |
| 人类可读           | 是                | 是 ✓                      |
| 嵌套结构           | ✓                 | ✓                         |
| 序列化速度         | 1×                | **~1.7–2.4× 更快** ✓     |
| 反序列化速度       | 1×                | **~1.9–2.9× 更快** ✓     |
| 数据体积           | 100%（基准）      | **40–60%** ✓              |
| 二进制编解码       | ✗                 | ✓                         |
| 结构体反射         | ✗                 | 编译期 ✓                  |

### Token 节省 —— 具体示例

```
JSON（100 tokens）：
{"users":[{"id":1,"name":"Alice","active":true},{"id":2,"name":"Bob","active":false}]}

ASON（~35 tokens，节省 65%）：
[{id:int, name:str, active:bool}]:(1,Alice,true),(2,Bob,false)
```

Schema 头部同时充当 LLM 的内联提示 —— 字段名和可选类型一目了然，无需扫描每一行。

---

## ASON vs TOON

[TOON（Token-Oriented Object Notation）](https://toonformat.dev) 是另一种旨在减少 LLM Prompt 中 Token 的格式。ASON 和 TOON 都以消除对象数组中 Key 重复为核心理念，但在设计目标和适用范围上有显著不同。

### 语法对比

**TOON** —— 基于缩进，YAML 风格：
```
users[2]{id,name,active}:
  1,Alice,true
  2,Bob,false
```

**ASON** —— 基于元组，Schema 显式/隐式：
```
[{id:int, name:str, active:bool}]:(1,Alice,true),(2,Bob,false)
```

### 对比表

| 方面               | TOON                                     | ASON                                                           |
| ------------------ | ---------------------------------------- | -------------------------------------------------------------- |
| Schema 声明        | 编码时自动检测                           | 显式声明，可复用，语言级别 ✓                                   |
| 类型注解           | 无（仅 JSON 数据模型）                   | 丰富可选类型（`int`、`str`、`bool`、`f64`、`opt_*`、`vec_*`、`map_*`）✓ |
| 语法风格           | YAML 风格缩进                            | 紧凑元组行                                                     |
| 数组长度标记       | `[N]` —— 有助于检测截断                  | Schema 头部定义结构 ✓                                          |
| 嵌套结构           | 退化为冗长的列表格式                     | 原生高效，递归支持 ✓                                           |
| 适用场景           | 仅 LLM 输入（只读转换层）                | LLM + 序列化 + 数据传输 ✓                                      |
| 序列化性能         | 未测试（仅 JS）                          | SIMD 加速，约 1.7–2.9× vs JSON ✓                               |
| 反序列化性能       | 未测试（仅 JS）                          | 零拷贝解析 ✓                                                   |
| 二进制编解码       | ✗                                        | ✓                                                              |
| 语言实现           | 仅 TypeScript / JavaScript               | **C、C++、C#、Go、Java、JS、Python、Rust、Zig、Dart** ✓        |
| 结构体反射         | 动态（仅运行时）                         | 编译期（`ASON_FIELDS` 宏）✓                                    |
| 深层嵌套数据       | Token 开销显著增加                       | 任意嵌套层级均高效 ✓                                           |
| 往返保真度         | JSON 数据模型（无类型信息）              | 完整类型保真 ✓                                                 |

### 何时选择 ASON

- 需要在编译型语言（C、C++、Rust、Go……）中实现**高性能序列化**
- 数据具有**丰富类型** —— 可选字段、类型化向量、Map、嵌套结构体
- 除文本格式外还需要**二进制编码**
- 跨**多种语言**工作，或需要与语言无关的线路格式
- 希望 Schema 作为面向 LLM Prompt 的**自文档化 API 契约**

### TOON 足够用的场景

- 流程**仅为 LLM Prompt 输入**（不需要将结果解析回结构体）
- 数据是无类型约束的简单平表

---

## 格式概览

### 单个对象

```
{id:int, name:str, active:bool}:(42,Alice,true)
```

### 对象数组（Schema 驱动）

Schema 声明一次，每条数据为元组：

```
[{id:int, name:str, active:bool}]:(1,Alice,true),(2,Bob,false),(3,Carol,true)
```

### 嵌套结构体

```
{name:str, dept}:(Alice,(Engineering))
```

### 可选字段

```
{id:int, label:opt_str}:(1,hello),(2,)
```
*（空值 = `None` / `null`）*

### 类型化向量与 Map

```
{name:str, tags:vec_str, attrs:map_si}:(Alice,[rust,go,python],[(age,30),(score,95)])
```

---

## 各语言实现

| 语言       | 仓库                    | 备注                                 |
| ---------- | ----------------------- | ------------------------------------ |
| C          | [ason-c](ason-c/)       | C11, SIMD（NEON/SSE2）零拷贝  |
| C++        | [ason-cpp](ason-cpp/)   | C++17,Header-only, SIMD       |
| C#         | [ason-cs](ason-cs/)     | .NET, SIMD                    |
| Go         | [ason-go](ason-go/)     |                               |
| Java       | [ason-java](ason-java/) |                               |
| JavaScript | [ason-js](ason-js/)     |                               |
| Python     | [ason-py](ason-py/)     |                               |
| Rust       | [ason-rs](ason-rs/)     |                               |
| Zig        | [ason-zig](ason-zig/)   |                               |
| Dart       | [ason-dart](ason-dart/) |                               |
| Swift      | [ason-swift](ason-swift/) |   TODO                      |
| PHP       | [ason-php](ason-php/) |      TODO                        |
| Kotlin       | [ason-dart](ason-kt/) |    TODO                       |
| TypeScript  | [ason-dart](ason-ts/) |    TODO                        |

## 插件

| IDE       | 仓库    | 备注     |
| ---------- | ------ | ------ |
| VSCODE       | [plugin_vscode](plugin_vscode/)   | Done |
| JETBRAIN       | [plugin_idea](plugin_idea/)   | Todo |
| Zed       | [plugin_zed](plugin_zed/)   | Todo |

---

## 许可证

MIT
