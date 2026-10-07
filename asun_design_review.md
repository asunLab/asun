# ASUN 设计评审与优化建议

> 目标：在不破坏 ASUN 现有“简单、紧凑、人类可读”特性的前提下，评估它作为长期通用序列化格式还能如何进一步优化。

---

## 1. 总体判断

ASUN 目前最有价值的地方，不是简单地把 JSON 的语法压缩得更短，而是从数据表示模型上做了一个更重要的改变：

> **Schema 只描述一次，重复记录只传 positional values。**

例如：

```asun
[{id@int,name@str,active@bool}]:(1,Alice,true),(2,Bob,false)
```

相比 JSON：

```json
[
  {"id":1,"name":"Alice","active":true},
  {"id":2,"name":"Bob","active":false}
]
```

ASUN 避免了在每条记录中重复传输 `id`、`name`、`active` 等字段名。

这使它天然适合：

- API 批量数据
- RPC
- 数据库查询结果
- Telemetry / Event Stream
- 日志
- LLM structured context
- Dataset
- 高密度结构化数据传输

ASUN 当前最宝贵的资产，是它仍然保持了类似 JSON 的简单感。

同时需要明确一个非常重要的现状：

> **ASUN Binary 当前不携带 Schema，只编码 values；解码端必须事先拥有对应 Schema。**

这不是一个需要“修复”的缺陷，反而可以成为 ASUN Binary 的核心优势：Schema 已经决定 wire layout，因此 Binary 可以避免重复携带字段名、字段类型、field tag 等元数据，把 payload 压缩为纯值编码。

因此后续设计原则应该是：

> **尽量冻结现有文本语法，不继续堆积表面 feature；把精力投入到语义、Canonical、Schema Evolution，以及如何让 values-only Binary 最大程度利用 Schema。**

---

# 2. 建议的总体架构

建议把 ASUN 明确分成两个层次：

```text
ASUN Core
    ↓
ASUN Protocol
```

## 2.1 ASUN Core

只定义最核心的数据模型和编码规则：

- Scalar
- Struct
- Array
- Map
- Null / Optional
- Schema
- Text Encoding
- Binary Encoding

ASUN Core 应该始终保持非常小。

## 2.2 ASUN Protocol

负责高级协议能力：

- Schema ID
- Schema Version
- Schema Negotiation
- Schema Registry
- Stream Framing
- RPC Transport
- Compression
- Compatibility / Evolution

这样未来即使增加高级功能，也不会污染核心文本语法。

一个重要原则是：

> **复杂性应该尽量进入工具链和协议层，而不是进入 wire syntax。**

---

# 3. 建议补齐一等 Map / Dictionary

如果目标是长期替代 JSON，那么 Map 应该成为核心数据模型的一等类型。

JSON 的对象实际上同时承担两种完全不同的角色：

1. 固定结构的 Struct
2. 动态 key-value 的 Map

例如：

```json
{
  "id": 1,
  "name": "Alice"
}
```

这里 `id` / `name` 是 Schema。

但：

```json
{
  "headers": {
    "Content-Type": "text/plain",
    "X-Foo": "bar"
  }
}
```

这里 `Content-Type` / `X-Foo` 是数据，不是 Schema。

建议明确区分 Struct 和 Map。

一种可能的简洁语法：

```asun
[T]
```

表示 Array：

```text
Array<T>
```

而：

```asun
[K:V]
```

表示 Map：

```text
Map<K,V>
```

例如：

```asun
{user@str,attrs@[str:int]}
```

数据：

```asun
(Alice,[age:30,score:95])
```

这样核心数据模型可以完整覆盖：

```text
scalar
struct
array
map
null
```

---

# 4. 建议重新审视 Null 表达

如果使用空位代表 null，例如：

```asun
{id,name,email}:(1,Alice,)
```

优点是极其紧凑。

但问题是，当连续 null 较多时，人类可读性会快速下降：

```asun
(1,,,,true,)
```

肉眼很容易看错字段位置。

建议考虑：

```text
_ = null
```

例如：

```asun
{id,name,email}:(1,Alice,_)
```

数组：

```asun
[1,_,3]
```

优点：

- 只有 1 byte
- 视觉明显
- Parser 简单
- 不容易与普通字符串冲突
- 对 LLM 也容易理解

相比 `null`，它更紧凑；相比空位，它更可靠。

---

# 5. 建议区分 Null 和 Absent

对于 API，特别是 PATCH / Partial Update，下面两个语义经常不同：

```json
{}
```

和：

```json
{"email": null}
```

前者可能表示：

> 不修改 email

后者可能表示：

> 明确清空 email

因此建议 ASUN 在 Schema / Protocol 层允许区分：

```text
value   = 有值
_       = explicit null
empty   = absent / missing
```

例如：

```asun
{id@int,email@str?}
```

```asun
(1,_)
```

表示 email 明确为 null。

而：

```asun
(1,)
```

表示 email 未提供。

如果最终不希望文本层承担这一语义，也可以只在协议层定义，但这个问题应该尽早明确。

Binary 编码中可以天然使用：

```text
presence bitmap
null bitmap
values
```

高效表达。

---

# 6. 必须定义 Canonical ASUN

这是我认为必须尽早完成的能力。

同一份数据可以允许多种“人类友好”的写法，例如：

```asun
{id,name}:(1,Alice)
```

和：

```asun
{
  id,
  name
}:
(
  1,
  Alice
)
```

都可以合法。

但系统必须存在唯一的：

> **Canonical ASUN**

例如固定输出为：

```asun
{id,name}:(1,Alice)
```

Canonical Form 应明确规定：

- UTF-8
- 换行规则
- Whitespace 规则
- String escaping
- Float formatting
- Integer formatting
- Decimal formatting
- Null 表达
- Field ordering
- Map ordering 是否 canonicalized
- Unicode normalization 是否处理

Canonical ASUN 的价值非常大：

```text
ASUN Value
    ↓
canonical encode
    ↓
identical bytes
```

于是不同语言：

```text
Rust
Go
Java
C#
Python
JavaScript
C/C++
```

对同一数据都可以产生完全一致的 byte sequence。

这直接支持：

- Hash
- Digital Signature
- Cache Key
- Schema Fingerprint
- Content Addressing
- Golden Test
- Cross-language Conformance Test

---

# 7. Number 类型必须定义得比 JSON 更严谨

JSON 只有一个模糊的 `number`。

这会导致：

```text
9007199254740993
```

在不同语言、不同 runtime 中产生不同精度行为。

既然 ASUN 已经有类型提示，并且还支持 Binary，就不应该继承 JSON 这个问题。

建议至少明确：

```text
int
uint
float
decimal
```

或者更严格：

```text
i64
u64
f64
dec
```

例如：

```asun
{id@u64,balance@dec,score@f64}
```

但仍然允许普通用户完全不写类型：

```asun
{id,name}
```

核心原则：

> **Type system 是增强能力，不应该成为使用 ASUN 的门槛。**

---

# 8. String Grammar 应保持极简

Bare String 是 ASUN 相比 JSON 很有价值的一点：

```asun
Alice
Bob
Engineering
```

而不是：

```json
"Alice"
"Bob"
"Engineering"
```

但 Bare String 的 grammar 必须非常明确。

建议类似：

Bare String 不允许包含：

```text
whitespace
, ( ) [ ] { } : "
```

同时不能与：

```text
boolean
number
null marker
reserved token
```

产生歧义。

有歧义时统一使用 quoted string：

```asun
"Hello world"
"123"
"true"
"A,B"
```

这样 parser 可以非常高效地通过 delimiter scan / SIMD 寻找 token 边界。

---

# 9. Text 和 Binary 必须共享同一个 Semantic Model

这是 Binary 能否成为 ASUN 真正优势的关键。

首先需要明确当前 ASUN Binary 的正确定位：

```text
                 ASUN Schema
                 /        \
                /          \
        ASUN Text        ASUN Binary
        schema + values    values only
```

也就是说：

- ASUN Text 可以携带 Schema 与 values，因此适合人类阅读、调试、LLM、通用交换。
- ASUN Binary 本身只编码 values，不携带 Schema。
- Binary Decoder 必须在解码前已经获得正确的 Schema。

这是一个很重要的设计选择，不建议为了“self-describing”而把完整 Schema 重新塞回 Binary Core。

Binary 的价值恰恰在于：

> **Schema 决定 layout，Binary 只传 values。**

Text 和 Binary 虽然 wire representation 不同，但必须共享唯一的语义模型：

```text
                 ASUN Semantic Model
                        │
              ┌─────────┴─────────┐
              │                   │
         ASUN Text           ASUN Binary
       schema + values     values + schema context
```

API 语义应更接近：

```text
decodeText(text)
      ↓
   ASUN Value
      ↑
decodeBinary(schema, bytes)
```

以及：

```text
encodeText(schema, value)
encodeBinary(schema, value)
```

因此 Binary 的 round-trip requirement 也必须显式包含 Schema：

```text
value
==
decodeBinary(
    schema,
    encodeBinary(schema, value)
)
```

如果从 Text 转 Binary：

```text
text
  ↓ decodeText
(schema, value)
  ↓ encodeBinary(schema, value)
binary values
```

反向则是：

```text
binary values + schema
  ↓ decodeBinary
value
  ↓ encodeText(schema, value)
canonical text
```

核心原则：

> **Text is self-describing. Binary is schema-bound. Both represent the same semantic value.**

中文可以概括为：

> **文本自描述，二进制绑定 Schema；两者共享同一个语义模型。**

---

# 10. Binary 应坚持 Values-only，并充分利用 Schema-first

例如 Text：

```asun
[{id@u64,name@str,active@bool}]:(1,Alice,true),(2,Bob,false)
```

Schema 已经告诉 Decoder：

```text
field 0 = u64
field 1 = string
field 2 = bool
```

因此 Raw ASUN Binary 没有必要再次携带：

```text
id
name
active
u64
str
bool
field tag
type tag
```

Binary payload 可以只编码值，例如概念上：

```text
[varint 1]
[length 5][Alice]
[bool 1]

[varint 2]
[length 3][Bob]
[bool 0]
```

Decoder 已经拿到了 Schema，因此解析路径可以非常直接：

```text
read_u64()
read_string()
read_bool()
```

理想情况下甚至可以通过 codegen 生成固定 decoder，避免：

- field-name parsing
- key hashing
- field lookup
- type dispatch
- per-value type tags
- dynamic schema lookup

所以 ASUN Binary 的目标不应该是模仿 MessagePack 那种“任何 payload 单独拿出来都能解释”，而更像：

> **一个由 Schema 预先编译好的 wire layout。**

这是 values-only Binary 最值得保留的设计特征。

---

# 11. Binary 可以进一步使用 Bitmap / Varint / Bit Packing

因为 Binary 已经由外部 Schema 决定类型和布局，所以很多优化可以比 self-describing binary format 更激进。

## Boolean

8 个 bool：

```text
true,false,true,true,false,false,true,false
```

可以压成：

```text
10110010
```

一个 byte。

## Optional / Null

大量 Optional Field 可以使用：

```text
presence bitmap
null bitmap
payload
```

避免为每个字段单独写 null marker。

## Integer

适合时可使用：

```text
varint
zigzag
```

减少小整数成本。

由于 Schema 已经在 Binary 外部，payload 不需要重复写类型信息，因此 Binary 的优化目标可以非常明确：

> **更少 bytes、更少分支、更少 tokenization、更少字段 lookup、更高 decode throughput。**

---

# 12. 可以考虑 ASUN Binary Block Mode

这是一个可选的 P2 / P3 方向，不建议过早实现，但非常有潜力。

ASUN 天然擅长：

```text
Array<Struct>
```

例如：

```asun
[{id,name,age,active}]
```

Text 可以保持 Row-oriented：

```asun
(1,Alice,30,true)
(2,Bob,28,false)
```

但 Binary 不一定也要 Row-oriented。

可以增加一个 Block / Columnar Encoding：

```text
IDs:
1 2 3 4 5 ...

Names:
Alice Bob Carol ...

Ages:
30 28 35 ...

Active:
101101...
```

潜在优势：

- 更好的 Compression
- 更好的 Cache Locality
- SIMD Decode
- Numeric Delta Encoding
- Boolean Bit Packing
- 只读取部分字段

这有一点 Arrow / Parquet 的思路，但不需要让 ASUN 本身变成分析型存储格式。

可以理解为：

```text
ASUN Binary Row
ASUN Binary Block
```

两者都仍然是 **values-only encoding**，都依赖外部 Schema，并最终 decode 成同一个 ASUN Semantic Model。

---

# 13. Schema Identity / Fingerprint 应放在 Protocol 层，而不是 Binary Core

既然 Raw ASUN Binary 本身不携带 Schema，那么跨进程、跨机器或持久化时，一个自然的问题是：

> **接收端如何确定这段 values 应该用哪个 Schema 解码？**

这里建议引入 Schema Identity，但不要改变 Raw ASUN Binary。

例如 Schema：

```asun
{id@u64,name@str,active@bool}
```

首先定义 Canonical Schema，再计算稳定 fingerprint：

```text
canonical schema
      ↓
   hash
      ↓
91A72C...
```

然后把 schema identity 放在 ASUN Core 外面的 Envelope / Protocol：

```text
┌──────────────────────┐
│ schema fingerprint   │
│ payload length       │
│ flags / version      │
├──────────────────────┤
│ ASUN binary values   │
└──────────────────────┘
```

这里最重要的是：

```text
ASUN Binary payload = values only
```

保持不变。

可以形成三个清晰的使用模式：

## ASUN Text

```text
[schema + values]
```

适合：

- 人类阅读
- Debug
- 通用 API
- LLM
- 可移植文本文件

## Raw ASUN Binary

```text
[values only]
```

Schema 由调用方、生成代码、API contract 或本地上下文提供。

适合：

- IPC
- 高性能 RPC
- 进程内 / 进程间固定协议
- 已知类型的存储

## ASUN Envelope / Protocol

```text
[schema-id][binary values]
```

注意 `schema-id` 属于外层协议，不属于 Raw ASUN Binary。

适合：

- 网络消息
- Message Bus
- Kafka
- Schema Registry
- 需要长期持久化且必须识别 Schema 的场景

如果连接已经提前完成 Schema Negotiation，协议层甚至可以省略 schema-id，直接传 Raw Binary：

```text
[data only]
```

因此更准确的设计哲学是：

> **Text is self-describing. Binary is schema-bound. Protocol may be schema-referencing.**

即：

> **文本自描述，二进制只传值，协议层按需引用 Schema。**

Schema Fingerprint 因此很重要，但它应该是 ASUN ecosystem / protocol 的一等能力，而不是强迫每个 Binary payload 自带的 header。

---

# 14. Schema Evolution 必须尽早定义

Serialization format 真正进入生产环境后，最难的问题往往不是 parsing，而是 schema evolution。

例如：

V1：

```asun
{id,name}
```

V2：

```asun
{id,name,email}
```

除了新增字段，还需要考虑：

```text
rename
reorder
remove
change type
optional -> required
required -> optional
```

建议正式定义兼容规则，例如：

| 修改 | 建议兼容性 |
|---|---|
| append optional field | Safe |
| remove optional field | Usually safe |
| reorder field | Schema changed, but mappable |
| rename field | Breaking unless alias / stable ID |
| i32 -> i64 | Widening safe |
| i64 -> i32 | Breaking |
| optional -> required | Breaking |
| required -> optional | Usually safe |

其中非常重要的一点是：

> **Field identity 不应该永久绑定 positional index。**

Position 可以是当前 wire layout，但 Schema Evolution 最好存在稳定身份，例如：

```text
field name
stable field ID
schema-level alias
```

不过这些能力应该尽量存在于 Schema / Protocol 层，不要污染普通 ASUN Text。

---

# 15. Stream Framing 应单独定义

单个 ASUN Document 很简单。

但如果使用在：

- TCP
- WebSocket
- Kafka
- Append-only File
- Streaming API

就必须明确：

> 一条 ASUN Message 在哪里结束？

建议区分：

```text
ASUN Document
ASUN Stream
```

Binary 可以采用：

```text
[length][payload]
[length][payload]
```

Text Stream 可以定义自己的 canonical framing，例如：

- Record Separator
- Length Prefix
- 明确的 Line-delimited 变体

最好在生态形成之前就确定，而不是事后产生多个不兼容扩展。

---

# 16. 应严格抵抗 Feature Creep

ASUN 最大的潜在优势之一，是 Parser 可以非常简单。

因此不建议在 Core 中逐步加入：

```text
inheritance
anchors
references
macros
expressions
variables
functions
conditional syntax
automatic date parsing
implicit environment variables
object merging
```

每个 feature 单独看都可能“很方便”，但长期累积会让格式从 Serialization Format 演化成 Programming Language。

应该坚持：

> **ASUN 是 serialization format，不是 configuration language，也不是 programming language。**

任何新 syntax 都应该满足：

```text
收益 >> parser complexity + semantic complexity
```

否则不加入 Core。

---

# 17. 一个可能的理想 ASUN 1.0 Core

如果重新整理成一个极小核心，我会希望大致只有：

## Scalar

```text
str
bool
int
uint
float
decimal
bytes
```

## Containers

```text
[T]       array
[K:V]     map
{a,b,c}   struct
```

## Schema

```text
field
field@type
field@schema
```

## Data

```text
(...)     struct value
[...]     collection
```

## Special

```text
_         null
empty     absent
"..."     quoted string
```

例如：

```asun
[
  {
    id@u64,
    name@str,
    email@str?,
    scores@[int],
    attrs@[str:str]
  }
]:
(
  1,
  Alice,
  _,
  [98,95,91],
  [country:NZ,lang:zh]
),
(
  2,
  Bob,
  bob@example.com,
  [87,92],
  [country:AU,lang:en]
)
```

Canonical / Minified：

```asun
[{id@u64,name@str,email@str?,scores@[int],attrs@[str:str]}]:(1,Alice,_,[98,95,91],[country:NZ,lang:zh]),(2,Bob,bob@example.com,[87,92],[country:AU,lang:en])
```

核心仍然应该让第一次看到 ASUN 的开发者在几十秒内理解。

---

# 18. 推荐优先级

| 优先级 | 建议 |
|---|---|
| **P0** | 冻结 Core Grammar |
| **P0** | 定义 Canonical ASUN |
| **P0** | 明确 Raw Binary = values only + external Schema |
| **P0** | Text / Binary 共享唯一 Semantic Model |
| **P0** | 精确定义 Number / String / Null |
| **P0** | 一等 Map / Dictionary |
| **P1** | Null vs Absent |
| **P1** | Schema Evolution |
| **P1** | Canonical Schema + Schema Fingerprint |
| **P1** | 可选 Envelope / Protocol（schema-id 不进入 Binary Core） |
| **P1** | Stream Framing |
| **P2** | Binary Varint / Bitmap / Bit Packing |
| **P2/P3** | Binary Columnar Block Mode |
| **不要做** | 把 ASUN 变成复杂 Schema Language / Programming Language |

---

# 19. 最终建议

我认为 ASUN 已经过了“继续发明更聪明文本语法”的阶段。

继续微调：

```text
{}
:
@
,
()
```

收益已经不会特别大。

真正决定 ASUN 最终只是：

> 一个不错的 JSON Alternative

还是发展成：

> 一个长期通用的 Serialization System

关键是下面四件事：

## 1. Semantic Precision

所有语言对同一个 ASUN 数据必须有完全一致的解释。

## 2. Canonical Representation

同一个 Value 必须可以生成唯一 canonical bytes。

## 3. Schema Evolution

格式必须能够支撑长期运行的真实 API / RPC / Storage 系统。

## 4. Binary Optimization

Binary 应继续坚持 **values-only + external Schema**，最大程度利用 ASUN 的 schema-first 特性，而不是为了 self-describing 再把 Schema、字段名或类型标签塞回 payload，也不是简单做一个文本 AST 的二进制版本。

---

# 20. 一句话总结

如果继续发展 ASUN，我最建议遵循的原则是：

> **Keep the syntax small. Make the semantics rigorous. Keep raw binary values-only. Let binary exploit the schema. Put complexity into tooling and protocol layers, not into the core wire format.**

中文可以概括为：

> **保持语法简单，把语义做到严格；Raw Binary 只传值，让 Binary 吃尽 Schema-first 的性能红利，把复杂性放进工具链和协议层，而不是塞回核心格式。**

