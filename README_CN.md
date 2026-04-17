# ASUN — Array-Schema Unified Notation

[![License: MIT](https://img.shields.io/badge/license-MIT-blue.svg)](LICENSE)

**ASUN** 是一种紧凑的、Schema 优先的数据格式，适合 **LLM Prompt**、**结构化 API** 和 **大规模数据集**。它将 Schema 与数据分离，Key 只声明一次，数据行只保留值。

[English](README.md)

---

## 为什么选择 ASUN？

标准 JSON 会在每条记录里重复所有字段名。无论是发给 LLM、通过 API 传输，还是服务之间交换数据，这种重复都会浪费 Token、带宽和阅读成本：

```json
[
  { "id": 1, "name": "Alice", "active": true },
  { "id": 2, "name": "Bob", "active": false },
  { "id": 3, "name": "Carol", "active": true }
]
```

ASUN 只声明 **一次** Schema，数据以紧凑元组方式流式传输：

```
[{id@int, name@str, active@bool}]:(1,Alice,true),(2,Bob,false),(3,Carol,true)
```

**更少的 Token，更小的体积，更清晰的结构。**

---

## ASUN vs JSON

| 方面         | JSON         | ASUN                      |
| ------------ | ------------ | ------------------------- |
| Token 效率   | 100%（基准） | **30–70%** ✓              |
| Key 重复     | 每个对象都有 | 声明一次 ✓                |
| 字段绑定     | 无           | 内建 `@...` ✓             |
| 人类可读     | 是           | 是 ✓                      |
| 嵌套结构     | ✓            | ✓                         |
| 序列化       | 重复 Key     | Schema 一次，后面只写值 ✓ |
| 反序列化     | 通用对象扫描 | 基于 Schema 定向解码 ✓    |
| 数据体积     | 100%（基准） | **40–60%** ✓              |
| 二进制编解码 | ✗            | ✓                         |
| 基本类型提示 | 有限         | 可选 `@type` ✓            |

### Token 节省 —— 具体示例

```
JSON（100 tokens）：
{"users":[{"id":1,"name":"Alice","active":true},{"id":2,"name":"Bob","active":false}]}

ASUN（~35 tokens，节省 65%）：
[{id@int, name@str, active@bool}]:(1,Alice,true),(2,Bob,false)
```

Schema 头部也可以直接充当 LLM 和人类读者的内联提示：字段名、结构绑定和可选的基本类型提示先出现，不需要逐行反复扫描。

---

## ASUN vs TOON

[TOON（Token-Oriented Object Notation）](https://toonformat.dev) 是另一种旨在减少 LLM Prompt 中 Token 的格式。ASUN 和 TOON 都以消除对象数组中 Key 重复为核心理念，但在设计目标和适用范围上有显著不同。

### 语法对比

**TOON** —— 基于缩进，YAML 风格：

```
users[2]{id,name,active}:
  1,Alice,true
  2,Bob,false
```

**ASUN** —— 基于元组，Schema 显式/隐式：

```
[{id@int, name@str, active@bool}]:(1,Alice,true),(2,Bob,false)
```

### 对比表

| 方面          | TOON                       | ASUN                                                    |
| ------------- | -------------------------- | ------------------------------------------------------- |
| Schema 声明   | 编码时自动检测             | 显式声明，可复用 ✓                                      |
| 字段/类型绑定 | 无（仅 JSON 数据模型）     | 用 `@...` 显式绑定结构与可选基本类型提示 ✓              |
| 语法风格      | YAML 风格缩进              | 紧凑元组行                                              |
| 数组长度标记  | `[N]` —— 有助于检测截断    | Schema 头部定义结构 ✓                                   |
| 嵌套结构      | 退化为冗长的列表格式       | 原生递归支持 ✓                                          |
| 适用场景      | 仅 LLM 输入                | LLM + 序列化 + 存储 + 传输 ✓                            |
| 二进制编解码  | ✗                          | ✓                                                       |
| 语言实现      | 仅 TypeScript / JavaScript | **C、C++、C#、Go、Java、JS、Python、Rust、Zig、Dart** ✓ |
| 往返保真度    | 仅 JSON 数据模型           | 完整类型保真 ✓                                          |

### 何时选择 ASUN

- 希望在不丢失结构的前提下得到**更少的 Token**和**更小的 payload**
- 数据有**丰富结构** —— 可选字段、数组、嵌套结构体、键值条目列表
- 希望一个格式同时服务于 **LLM、API、存储、服务间传输**
- 数据具有**丰富结构** —— 可选字段、数组、嵌套结构体、键值条目列表
- 除文本格式外还需要**二进制编码**
- 跨**多种语言**工作，或需要与语言无关的线路格式
- 希望 Schema 作为面向 LLM Prompt 的**自文档化 API 契约**，通过 `@` 绑定字段与结构描述及可选基本类型提示

### TOON 足够用的场景

- 流程**仅为 LLM Prompt 输入**（不需要将结果解析回结构体）
- 数据是无类型约束的简单平表

---

## 格式概览

### 单个对象

```
{id@int, name@str, active@bool}:(42,Alice,true)
```

### 对象数组（Schema 驱动）

Schema 声明一次，每条数据为元组：

```
[{id@int, name@str, active@bool}]:(1,Alice,true),(2,Bob,false),(3,Carol,true)
```

### 嵌套结构体

```
{name@str, dept@{title@str}}:(Alice,(Engineering))
```

### 可选字段

```
{id@int, label@str}:(1,hello),(2,)
```

_（空值 = `None` / `null`）_

### 数组与键值条目

```
{name@str, scores@[int], attrs@[{key@str, value@int}]}:(Alice,[90,85,92],[(age,30),(score,95)])
```

---

## 各语言实现

| 语言                    | 仓库                                 | 状态 |
| ----------------------- | ------------------------------------ | ---- |
| C                       | [asun-c](asun-c/)                    | ✓    |
| C++                     | [asun-cpp](asun-cpp/)                | ✓    |
| C#                      | [asun-cs](asun-cs/)                  | ✓    |
| Go                      | [asun-go](asun-go/)                  | ✓    |
| Java / Kotlin           | [asun-java](asun-java/)              | ✓    |
| JavaScript / TypeScript | [asun-js](asun-js/) (`@athanx/asun`) | ✓    |
| Python                  | [asun-py](asun-py/)                  | ✓    |
| Rust                    | [asun-rs](asun-rs/)                  | ✓    |
| Zig                     | [asun-zig](asun-zig/)                | ✓    |
| Dart                    | [asun-dart](asun-dart/)              | ✓    |
| PHP                     | [asun-php](asun-php/)                | ✓    |
| Swift                   | [asun-swift](asun-swift/)            | ✓    |

## 插件

| IDE      | 仓库                                | 备注 |
| -------- | ----------------------------------- | ---- |
| VSCode   | [plugin_vscode](plugin_vscode/)     | ✓    |
| Jetbrain | [plugin_jetbrain](plugin_jetbrain/) | Todo |
| Zed      | [plugin_zed](plugin_zed/)           | Todo |

---

## 许可证

MIT
