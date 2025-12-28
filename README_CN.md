# ASON - Array-Schema Object Notation

[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)

一个**高效节省 Token** 的数据格式，通过分离 Schema 和数据来最小化冗余。专为 LLM 交互和批量数据传输设计。

## 为什么选择 ASON？

在与 LLM 交互时，每个 token 都很重要。传统格式如 JSON 会为每条记录重复字段名，浪费 token 并增加成本。ASON 通过只声明一次 schema 并使用位置数据来解决这个问题。

```
JSON (62 tokens):
[{"name":"Alice","age":30},{"name":"Bob","age":25}]

ASON (~35 tokens):
{name,age}:(Alice,30),(Bob,25)
```

**批量数据减少约 45% 的 token！**

## 格式对比

### ASON vs JSON

| 特性 | JSON | ASON |
|------|------|------|
| 字段名 | 每条记录重复 | 在 schema 中声明一次 |
| 批量效率 | 差（冗余 key） | 优秀（schema 复用） |
| 单行友好 | 是 | 是 |
| 嵌套结构 | 支持 | 支持 |
| 类型推断 | 无（显式） | 有（自动检测） |
| LLM token 成本 | 高 | 低（减少 30-50%） |

**示例 - 100 条用户记录：**
```
JSON:  [{"id":1,"name":"Alice","email":"a@x.com"},{"id":2,"name":"Bob","email":"b@x.com"},...]
       → ~1500 tokens（字段名重复 100 次）

ASON:  {id,name,email}:(1,Alice,a@x.com),(2,Bob,b@x.com),...
       → ~800 tokens（字段名只声明一次）
```

### ASON vs CSV

| 特性 | CSV | ASON |
|------|-----|------|
| Schema 定义 | 标题行 | 显式 schema 语法 |
| 嵌套对象 | ❌ 不支持 | ✅ 支持 |
| 数组 | ❌ 不支持 | ✅ 支持 |
| 类型信息 | 无 | 有（bool, int, float, string, null） |
| 注释 | 无 | 有（`/* ... */`） |
| 单值 | 不自然 | 自然 |

**示例 - 嵌套数据：**
```
CSV 无法表示：
{name,addr{city,zip}}:(Alice,(NYC,10001))

→ {"name":"Alice","addr":{"city":"NYC","zip":10001}}
```

### ASON vs TOON

[TOON](https://github.com/BenoitZugmeyer/toon)（Token-Oriented Object Notation）是另一个 LLM 优化格式。主要区别：

| 特性 | TOON | ASON |
|------|------|------|
| Schema 位置 | 与数据混合 | 分离，前缀 schema |
| 批量记录 | 每条记录重复 schema | Schema 声明一次，多条记录 |
| 单行支持 | 面向多行 | **原生单行** |
| 压缩率 | 良好 | **批量更优** |

**TOON：**
```toon
name: Alice
age: 30
---
name: Bob
age: 25
```

**ASON：**
```ason
{name,age}:(Alice,30),(Bob,25)
```

#### ASON 的单行优势

ASON 从设计之初就是**单行友好**的：

1. **流式友好**：可以在一行内发送，无需换行
2. **嵌入友好**：易于嵌入到 prompts、日志或 URL 中
3. **复制友好**：单行更容易复制和分享
4. **紧凑 API 响应**：无格式化开销

```
# 多行格式（如 TOON）在 prompt 中需要转义或特殊处理：
"解析这个数据：\nname: Alice\nage: 30"

# ASON 天然支持单行：
"解析这个数据：{name,age}:(Alice,30)"
```

这在以下场景特别有用：
- **LLM Prompt 嵌入**：直接内联数据，无需转义换行
- **日志记录**：每条日志一行，结构化数据
- **URL 参数**：可直接编码在查询字符串中
- **WebSocket/SSE**：单行消息更简洁

## 语法速查

| 语法 | 示例 | 说明 |
|------|------|------|
| 对象 | `{name,age}:(Alice,30)` | Schema + 数据 |
| 多条记录 | `{id}:(1),(2),(3)` | 相同 schema，批量数据 |
| 嵌套对象 | `{user{name}}:((Alice))` | 嵌套结构 |
| 数组字段 | `{scores[]}:([90,85])` | 简单数组 |
| 对象数组 | `{users[{id}]}:([(1),(2)])` | 对象数组 |
| 纯数组 | `[1,2,3]` | 无需 schema |
| 空值 | `{a,b}:(1,)` | b = null |
| 空字符串 | `{a}:("")` | 显式空字符串 |
| 注释 | `/* 备注 */ {a}:(1)` | 块注释 |

## 实现

| 语言 | 路径 | 状态 |
|------|------|------|
| Rust | [`rust/`](rust/) | ✅ 完成 |

## 使用场景

- **LLM API 调用**：减少 30-50% 的 token 使用
- **批量数据传输**：高效发送多条记录
- **配置文件**：紧凑的配置格式
- **日志记录**：单行结构化日志
- **流式传输**：逐行数据流

## 规范

查看 [ASON 格式规范（中文）](docs/ASON_SPEC_CN.md) 或 [ASON Format Specification (English)](docs/ASON_SPEC.md) 获取完整语法和规则。

## 许可证

MIT

