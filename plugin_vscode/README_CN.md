# ASON Language Support — VS Code 插件使用指南

本文档说明如何构建、安装和使用 ASON 的 VS Code 插件。

---

## 目录

1. [环境准备](#1-环境准备)
2. [构建与打包](#2-构建与打包)
3. [安装插件到 VS Code](#3-安装插件到-vs-code)
4. [验证安装](#4-验证安装)
5. [功能说明与使用](#5-功能说明与使用)
6. [配置项](#6-配置项)
7. [常见问题](#7-常见问题)

---

## 1. 环境准备

开始之前，请确保以下工具已经安装：

| 工具 | 最低版本 | 用途 | 下载地址 |
| --- | --- | --- | --- |
| **VS Code** | 1.75+ | 编辑器 | [下载地址](https://code.visualstudio.com/) |
| **Node.js** | 18+ | 构建插件 | [下载地址](https://nodejs.org/) |
| **Zig** | 0.15+ | 构建语言服务器 | [下载地址](https://ziglang.org/download/) |
| **Git** | 较新版本 | 缺少 `lsp-ason` 时自动克隆仓库 | [下载地址](https://git-scm.com/) |

在终端中验证安装：

```bash
node --version
npm --version
zig version
code --version
git --version
```

如果终端里找不到 `code` 命令，可以在 VS Code 中执行 `Shell Command: Install 'code' command in PATH`。

---

## 2. 构建与打包

构建脚本会自动完成这些工作：

1. 安装插件的 npm 依赖
2. 编译 TypeScript 插件代码
3. 用 Zig 构建 `lsp-ason` 语言服务器
4. 把二进制文件打包进 `server/`
5. 生成平台特定的 `.vsix`

### 2.1 安装打包工具

```bash
npm install -g @vscode/vsce
```

### 2.2 一键构建当前平台

```bash
cd plugin_vscode
npm install
./scripts/build.sh current
```

构建完成后会在当前目录生成一个 `.vsix` 文件，例如：

```text
ason-vscode-darwin-arm64-0.1.0.vsix
```

### 2.3 构建所有支持的平台

```bash
./scripts/build.sh all
```

会生成 6 个 VSIX 安装包：

| 文件 | 平台 |
| --- | --- |
| `ason-vscode-darwin-arm64-0.1.0.vsix` | macOS Apple Silicon |
| `ason-vscode-darwin-x64-0.1.0.vsix` | macOS Intel |
| `ason-vscode-linux-x64-0.1.0.vsix` | Linux x86_64 |
| `ason-vscode-linux-arm64-0.1.0.vsix` | Linux ARM64 |
| `ason-vscode-win32-x64-0.1.0.vsix` | Windows x86_64 |
| `ason-vscode-win32-arm64-0.1.0.vsix` | Windows ARM64 |

### 2.4 构建指定平台

```bash
./scripts/build.sh darwin-arm64
./scripts/build.sh darwin-x64
./scripts/build.sh linux-x64
./scripts/build.sh linux-arm64
./scripts/build.sh win32-x64
./scripts/build.sh win32-arm64
```

### 2.5 工作原理

这套构建链路可以跨平台工作的原因是：

1. Zig 可以把 `lsp-ason` 交叉编译到多个目标平台
2. `vsce package --target ...` 会生成平台特定的 VSIX 包
3. 打包后的扩展会从 `server/` 目录加载语言服务器

如果本地不存在 `../lsp-ason`，`./scripts/build.sh` 会先自动从 GitHub 克隆它，再继续构建。

---

## 3. 安装插件到 VS Code

### 方法一：从 VSIX 文件安装

```bash
code --install-extension ason-vscode-darwin-arm64-0.1.0.vsix
```

或者在 VS Code 中：

1. 按 `Cmd+Shift+P`（macOS）或 `Ctrl+Shift+P`（Windows/Linux）打开命令面板
2. 运行 `Extensions: Install from VSIX...`
3. 选择与你平台对应的 `.vsix` 文件
4. 安装完成后重新加载 VS Code

### 方法二：开发模式运行

如果你暂时不想打包，可以直接以开发模式运行插件。

先构建语言服务器：

```bash
cd ../lsp-ason
zig build
```

再用 VS Code 打开插件目录：

```bash
code ../plugin_vscode
```

按 `F5` 后会启动一个新的 Extension Development Host 窗口。

开发模式下，插件会按下面的顺序查找语言服务器：

- `server/lsp-ason`
- `../lsp-ason/zig-out/bin/lsp-ason`
- `server/ason-lsp`
- `../ason-lsp/ason-lsp`
- `PATH`

其中 `ason-lsp` 是为了兼容旧实现保留的回退路径，当前主实现是 `lsp-ason`。

---

## 4. 验证安装

1. 创建一个名为 `test.ason` 的文件
2. 输入以下内容：

```ason
{name:str, age:int, active:bool}:
  (Alice, 30, true)
```

3. 检查下面几项是否正常：

- 已经有语法高亮
- 对合法内容没有报错诊断
- 元组值前面会显示字段名的内联提示

如果语法高亮正常，但诊断或格式化没有工作，请查看下面的 [常见问题](#7-常见问题)。

---

## 5. 功能说明与使用

### 5.1 语法高亮

打开 `.ason` 文件后会自动启用语法高亮。

高亮范围包括：

- 字段名
- 类型标注，例如 `:str`、`:int`、`:bool`、`:float`
- 字符串、数字和布尔值
- `{}` `()` `[]`
- `/* ... */` 形式的注释

### 5.2 Markdown 中的 ASON 代码块

Markdown 里的 ASON 围栏代码块也会高亮：

````markdown
```ason
{name:str, score:int}:(Alice, 100)
```
````

### 5.3 实时错误检查

编辑时会自动显示诊断信息。

例如下面的输入是无效的：

```ason
{name:str}:(Alice
```

VS Code 会在编辑器里标出错误，并在“问题”面板中显示对应信息。

### 5.4 格式化

把当前 ASON 文档格式化为易读的排版。

使用方式：

- `Shift+Option+F`（macOS）或 `Shift+Alt+F`（Windows/Linux）
- 命令面板运行 `ASON: Format (Beautify)`
- 编辑器右键菜单里的 `Format Document`

格式化前：

```ason
{name:str,age:int,addr:{city:str,zip:int}}:(Alice,30,(NYC,10001))
```

格式化后：

```ason
{name:str, age:int, addr:{city:str, zip:int}}:
  (Alice, 30, (NYC, 10001))
```

### 5.5 压缩

把当前 ASON 文档压缩成紧凑的单行格式。

在命令面板中运行 `ASON: Compress (Minify)` 即可。

压缩前：

```ason
{name:str, age:int}:
  (Alice, 30)
```

压缩后：

```ason
{name:str,age:int}:(Alice,30)
```

### 5.6 ASON 转 JSON

打开一个 `.ason` 文件，执行 `ASON: Convert to JSON`，插件会在新标签页里打开转换后的 JSON。

ASON 输入：

```ason
{name:str, age:int, active:bool}:
  (Alice, 30, true)
```

JSON 输出：

```json
{
  "active": true,
  "age": 30,
  "name": "Alice"
}
```

### 5.7 JSON 转 ASON

支持两种用法。

从 JSON 文件转换：

1. 打开一个 `.json` 文件
2. 执行 `ASON: Convert JSON to ASON`
3. 转换结果会在新标签页中打开

从输入框粘贴 JSON：

1. 执行 `ASON: Convert JSON to ASON`
2. 在输入框里粘贴 JSON
3. 回车后打开转换结果

JSON 输入：

```json
[
  { "name": "Alice", "score": 95 },
  { "name": "Bob", "score": 87 }
]
```

ASON 输出：

```ason
[{name:str,score:int}]:
  (Alice,95),
  (Bob,87)
```

### 5.8 智能提示

编辑 ASON 时按 `Ctrl+Space` 可以触发补全建议。

常见补全包括：

- 顶层模板片段
- 类型名，例如 `int`、`str`、`bool`、`float`
- 布尔值，例如 `true` 和 `false`

### 5.9 悬停提示

把鼠标停在字段、类型标注或值上，会显示当前节点的上下文信息，例如字段类型或节点种类。

### 5.10 Inlay Hints（内联提示）

插件可以在元组值前面显示字段名提示。

例如源码：

```ason
{name:str, age:int, city:str}:(Alice, 30, NYC)
```

编辑器里会显示成类似这样：

```text
{name:str, age:int, city:str}:(name: Alice, age: 30, city: NYC)
```

其中 `name:`、`age:`、`city:` 只是视觉提示，不会写入文件内容。

### 5.11 语义着色

插件会对不同类别的 token 提供语义级着色：

| 元素 | 语义类型 |
| --- | --- |
| `{}` `()` `[]` | keyword |
| `:int`、`:str` 等 | type |
| 字段名 | variable |
| 字符串值 | string |
| 数字 | number |
| 注释 | comment |
| `:` `,` | operator |
| `true` `false` | parameter |

---

## 6. 配置项

在 VS Code 设置中搜索 `ason`：

| 配置项 | 类型 | 默认值 | 说明 |
| --- | --- | --- | --- |
| `ason.lspPath` | string | `""` | 语言服务器二进制路径。留空时插件会自动查找 |
| `ason.inlayHints.enabled` | boolean | `true` | 是否显示字段名内联提示 |

### 手动指定 `ason.lspPath`

如果语言服务器不在默认位置，可以在 VS Code 设置里填绝对路径：

```json
{
  "ason.lspPath": "/Users/你的用户名/code/lsp-ason/zig-out/bin/lsp-ason"
}
```

---

## 7. 常见问题

### Q: 语法高亮正常，但没有错误检查或格式化不工作

通常是语言服务器没有正确启动。

可以按下面排查：

1. 先确认二进制是否存在：

   ```bash
   ls -la ../lsp-ason/zig-out/bin/lsp-ason
   ```

2. 如果不存在，重新构建：

   ```bash
   cd ../lsp-ason
   zig build
   ```

3. 打开 VS Code 的“输出”面板，选择 `ASON Language Server` 查看日志
4. 如果自动查找失败，手动设置 `ason.lspPath`

### Q: 提示 "ASON LSP binary not found"

这表示插件没有找到 `lsp-ason`，或者没有找到兼容回退用的 `ason-lsp`。

解决方法：

- 在 `lsp-ason` 目录运行 `zig build`
- 手动设置 `ason.lspPath`
- 把语言服务器二进制放到系统 `PATH` 中

### Q: 格式化快捷键没反应

可能是 VS Code 把别的扩展设成了 ASON 文件的默认格式化器。

可以在设置里加上：

```json
{
  "[ason]": {
    "editor.defaultFormatter": "ason.ason-vscode"
  }
}
```

### Q: 看不到 Inlay Hints

确认 VS Code 全局设置里已经开启内联提示：

```json
{
  "editor.inlayHints.enabled": "on"
}
```

### Q: 如何卸载插件

```bash
code --uninstall-extension ason.ason-vscode
```

也可以在 VS Code 的扩展面板中找到 `ASON Language Support` 后手动卸载。

---

## 快速参考卡片

| 功能 | 操作 |
| --- | --- |
| 格式化 | `Shift+Option+F` 或 `ASON: Format` |
| 压缩 | `ASON: Compress` |
| ASON → JSON | `ASON: Convert to JSON` |
| JSON → ASON | `ASON: Convert JSON to ASON` |
| 自动补全 | `Ctrl+Space` |
| 命令面板 | `Cmd+Shift+P`（macOS）或 `Ctrl+Shift+P`（Windows/Linux） |
