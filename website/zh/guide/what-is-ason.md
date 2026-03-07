# ASON 是什么？

**ASON**（Array-Schema Object Notation，数组模式对象符号）是一种高性能序列化格式，专为两大场景设计：

1. **LLM 交互** —— 降低提示词和响应中结构化数据的 Token 消耗
2. **高吞吐系统** —— 以比 JSON 更快的编解码速度和更小的载荷替代 JSON

## 核心思想

ASON 将 **Schema（结构）** 与 **数据（值）** 分离。字段名只声明一次，数据行是纯粹的有序元组：

```ason
// Schema 只声明一次——数据不重复 Key
[{id:int, name:str, active:bool}]:
  (1, Alice, true),
  (2, Bob,   false),
  (3, Carol, true)
```

这一设计使列表密集型负载的 Token 开销降低 50–70%。

## 两种格式

| 格式 | 使用场景 | 核心优势 |
|------|----------|----------|
| **ASON Text（文本格式）** | LLM 提示词、API、配置 | 人类可读，比 JSON 小 53–61% |
| **ASON-BIN（二进制格式）** | 服务间通信、缓存、存储 | 序列化快 7–10 倍，零拷贝反序列化 |

两种格式共享同一套 Schema 模型，所有官方库均支持。

## 设计原则

- **Schema 与数据分离** —— 结构只声明一次，数据行不重复
- **行导向（Row-Oriented）** —— 多行格式增强纵向视觉流，便于人类和模型阅读
- **方圆结合（Structural Harmony）** —— `{}` 定义骨架（Schema），`()` 承载数据（value）
- **零哈希匹配（Zero Key-Hashing）** —— 解析器使用位置索引，反序列化期间无字符串比较
- **模式驱动解析（Schema-driven Parsing）** —— 解析时字段类型已知，无需动态推断

## 现状

ASON 拥有生产就绪的 **Rust、Go、Python、C、C++、Java、Zig** 实现，完整的[格式规范](/zh/spec)，以及 VS Code 插件和 LSP 服务器。所有实现共享跨语言兼容性测试套件。

## 下一步

- [为什么选择 ASON？](/zh/guide/why-ason) —— 与 JSON 的详细对比和基准测试
- [快速开始](/zh/guide/getting-started) —— 用你喜欢的语言五分钟上手
- [语法参考](/zh/reference/syntax) —— 完整格式规范
