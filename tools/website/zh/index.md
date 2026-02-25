---
layout: home

hero:
  name: "ASON"
  text: "数组模式对象符号"
  tagline: "比 JSON 少 65% Token。二进制序列化快 7–10 倍。人类可读。LLM 原生。"
  image:
    src: /logo.svg
    alt: ASON
  actions:
    - theme: brand
      text: 快速开始
      link: /zh/guide/getting-started
    - theme: alt
      text: 为什么选择 ASON？
      link: /zh/guide/why-ason
    - theme: alt
      text: 在 GitHub 查看
      link: https://github.com/your-org/ason

features:
  - icon: 🚀
    title: Token 节省 65%
    details: Schema 只声明一次，数据行只保留纯值，不再重复 Key。大幅降低 LLM 上下文消耗和 API 费用。
    link: /zh/guide/why-ason
    linkText: 查看对比

  - icon: ⚡
    title: 二进制快 7–10 倍
    details: ASON-BIN 采用零拷贝反序列化与 SIMD 加速，序列化比 JSON 快 7.7 倍，体积缩小 39%。
    link: /zh/reference/performance
    linkText: 查看基准测试

  - icon: 🧠
    title: LLM 原生设计
    details: 行导向语法 + Schema 分离，让大模型更容易生成正确的结构化输出，大幅减少幻觉和格式错误。
    link: /zh/guide/llm-integration
    linkText: LLM 集成方案

  - icon: 📖
    title: 人类可读
    details: 简洁的类 Markdown 语法，大多数值无需引号，行对齐清晰易扫描，信噪比极高。
    link: /zh/reference/syntax
    linkText: 语法参考

  - icon: 🌐
    title: 多语言支持
    details: 官方支持 Rust、Go、Python、C、C++、Java、Zig，跨语言 API 设计保持一致。
    link: /zh/languages/rust
    linkText: 语言指南

  - icon: 🔧
    title: 完整工具链
    details: VS Code 插件（语法高亮 + 实时预览）、LSP 服务器、跨语言兼容性测试套件。
    link: /zh/guide/getting-started
    linkText: 配置工具
---

<div class="vp-doc" style="max-width:900px;margin:0 auto;padding:2rem 1.5rem">

## ASON vs JSON — 一眼看懂

```json
// JSON — 约 100 Token
{
  "users": [
    { "id": 1, "name": "Alice", "active": true },
    { "id": 2, "name": "Bob",   "active": false },
    { "id": 3, "name": "Carol", "active": true  }
  ]
}
```

```ason
// ASON — 约 35 Token  (节省 65%)
{id:int, name:str, active:bool}:
  (1, Alice, true),
  (2, Bob,   false),
  (3, Carol, true)
```

Schema **只写一次**，`:` 之后的每行都是纯数据，逗号分隔的元组 `(...)`，没有重复 Key，没有多余引号，没有噪音。

## 快速安装

::: code-group

```toml [Rust]
# Cargo.toml
[dependencies]
ason = "0.1"
serde = { version = "1", features = ["derive"] }
```

```bash [Go]
go get github.com/your-org/ason/go
```

```bash [Python]
# 将 ason.py 复制到你的项目（零依赖，仅标准库）
```

:::

</div>
