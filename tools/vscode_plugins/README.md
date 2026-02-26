# ASON Language Support — VS Code 插件使用指南

本文档将从零开始教你如何构建、安装和使用 ASON VS Code 插件。

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

在开始之前，请确保你的电脑上已安装以下工具：

| 工具        | 最低版本 | 用途           | 安装方式                                   |
| ----------- | -------- | -------------- | ------------------------------------------ |
| **VS Code** | 1.75+    | 编辑器         | [下载地址](https://code.visualstudio.com/) |
| **Node.js** | 18+      | 构建插件       | [下载地址](https://nodejs.org/)            |
| **Go**      | 1.21+    | 构建语言服务器 | [下载地址](https://go.dev/dl/)             |

打开终端，验证安装：

```bash
# 检查 Node.js
node --version
# 应该输出类似 v18.x.x 或更高

# 检查 npm（随 Node.js 一起安装）
npm --version

# 检查 Go
go version
# 应该输出类似 go1.21.x 或更高

# 检查 VS Code 命令行工具
code --version
# 如果提示找不到 code 命令，打开 VS Code，按 Cmd+Shift+P，
# 输入 "Shell Command: Install 'code' command in PATH" 并执行
```

---

## 2. 构建与打包

构建脚本会自动完成以下工作：
1. 编译 TypeScript 插件代码
2. 交叉编译 ason-lsp 语言服务器（Go）
3. 将二进制文件打包到 `server/` 目录内
4. 生成对应平台的 `.vsix` 安装包

### 2.1 安装打包工具（首次需要）

```bash
npm install -g @vscode/vsce
```

### 2.2 一键构建当前平台（最简单）

```bash
# 进入插件目录
cd tools/vscode_ason_plugins

# 安装依赖（首次需要）
npm install

# 构建当前平台
./scripts/build.sh current
```

完成后会在当前目录生成一个 `.vsix` 文件，例如：`ason-vscode-0.1.0@darwin-arm64.vsix`

### 2.3 构建所有平台

```bash
./scripts/build.sh all
```

会生成 6 个 `.vsix` 文件，覆盖所有支持的平台：

| 文件 | 平台 |
|------|------|
| `ason-vscode-0.1.0@darwin-arm64.vsix` | macOS Apple Silicon (M1/M2/M3/M4) |
| `ason-vscode-0.1.0@darwin-x64.vsix` | macOS Intel |
| `ason-vscode-0.1.0@linux-x64.vsix` | Linux x86_64 |
| `ason-vscode-0.1.0@linux-arm64.vsix` | Linux ARM64 |
| `ason-vscode-0.1.0@win32-x64.vsix` | Windows x86_64 |
| `ason-vscode-0.1.0@win32-arm64.vsix` | Windows ARM64 |

### 2.4 构建指定平台

```bash
./scripts/build.sh darwin-arm64    # macOS Apple Silicon
./scripts/build.sh darwin-x64      # macOS Intel
./scripts/build.sh linux-x64       # Linux x86_64
./scripts/build.sh linux-arm64     # Linux ARM64
./scripts/build.sh win32-x64       # Windows x86_64
./scripts/build.sh win32-arm64     # Windows ARM64
```

### 2.5 工作原理

跨平台的秘密在于：

1. **Go 天生支持交叉编译**。只需设置 `GOOS` 和 `GOARCH` 环境变量，就能在任何平台上编译出其他平台的二进制文件（不需要安装交叉编译工具链）
2. **VS Code 支持平台特定插件**。`vsce package --target darwin-arm64` 会在 vsix 包的 `package.json` 中标记目标平台。VS Code Marketplace 会根据用户平台自动推送对应的版本
3. 打包时，ason-lsp 二进制文件被放在插件内部的 `server/ason-lsp` 路径（就像 rust-analyzer 插件放在 `server/rust-analyzer` 一样）。插件启动时自动从这个路径加载

> **注意**：交叉编译只需要本机安装 Go 即可，不需要目标平台的机器。

---

## 3. 安装插件到 VS Code

### 方法一：通过 .vsix 文件安装（推荐）

```bash
# 安装当前平台的 vsix（文件名中会包含平台标识）
code --install-extension ason-vscode-0.1.0@darwin-arm64.vsix
```

或者在 VS Code 中：

1. 按 `Cmd+Shift+P`（macOS）或 `Ctrl+Shift+P`（Windows/Linux）打开命令面板
2. 输入 **Extensions: Install from VSIX...**
3. 选择对应你平台的 `.vsix` 文件
4. 安装完成后重新加载 VS Code

### 方法二：开发模式运行（调试用）

如果你不想打包，可以直接在开发模式下运行。需要先手动构建 ason-lsp：

```bash
# 1. 构建语言服务器
cd tools/ason-lsp
go build -o ason-lsp .

# 2. 用 VS Code 打开插件目录
code tools/vscode_ason_plugins
```

然后在 VS Code 中按 `F5` 启动调试，会打开一个新的 VS Code 窗口，插件在其中已激活。
开发模式下插件会自动在相邻的 `../ason-lsp/` 目录中查找二进制文件。

---

## 4. 验证安装

1. 创建一个文件，命名为 `test.ason`
2. 输入以下内容：

```ason
{name:str, age:int, active:bool}:
  (Alice, 30, true)
```

3. 如果你看到：
   - **语法高亮**（不同颜色区分 schema 和 data）→ ✅ 插件语法部分正常
   - **没有红色波浪线错误提示** → ✅ 语言服务器正常连接
   - 在 `(Alice, 30, true)` 前面显示灰色的字段名提示（`name:` `age:` `active:`）→ ✅ Inlay Hints 正常

如果语法高亮正常但没有错误检查，请查看 [常见问题](#7-常见问题) 部分。

---

## 5. 功能说明与使用

### 5.1 语法高亮

**自动生效**，打开 `.ason` 文件即可看到。

以下元素会被不同颜色高亮：

- 字段名（如 `name`, `age`）
- 类型标注（如 `:str`, `:int`, `:bool`, `:float`）
- 字符串值、数字、布尔值
- 大括号 `{}`、小括号 `()`、方括号 `[]`
- 注释 `/* ... */`

### 5.2 Markdown 中的 ASON 代码块

在 Markdown 文件中使用 ` ```ason ` 围栏代码块也会自动高亮：

````markdown
下面是一个 ASON 示例：

```ason
{name:str, score:int}:(Alice, 100)
```
````

### 5.3 实时错误检查（Diagnostics）

**自动生效**，保存或编辑时实时检查。

如果有语法错误，会用红色波浪线标注并在 "问题" 面板中显示。

例如，缺少右括号：

```ason
{name:str}:(Alice
```

会在底部看到错误提示。

### 5.4 格式化（Format / Beautify）

将 ASON 代码格式化为易读的缩进格式。

**使用方式（3 种任选其一）：**

| 方式     | 操作                                                |
| -------- | --------------------------------------------------- |
| 快捷键   | `Shift+Option+F`（macOS）/ `Shift+Alt+F`（Windows） |
| 命令面板 | `Cmd+Shift+P` → 输入 **ASON: Format (Beautify)**    |
| 右键菜单 | 在编辑器中右键 → **Format Document**                |

**效果示例：**

格式化前：

```ason
{name:str,age:int,addr:{city:str,zip:int}}:(Alice,30,(NYC,10001))
```

格式化后：

```ason
{name:str, age:int, addr:{city:str, zip:int}}:
  (Alice, 30, (NYC, 10001))
```

### 5.5 压缩（Compress / Minify）

将 ASON 代码压缩为单行，去掉所有多余空格和换行。

**使用方式：**

`Cmd+Shift+P` → 输入 **ASON: Compress (Minify)**

**效果示例：**

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

将当前 ASON 文件转换为 JSON 格式，并在新标签页中打开。

**使用方式：**

1. 打开一个 `.ason` 文件
2. `Cmd+Shift+P` → 输入 **ASON: Convert to JSON**
3. 会自动打开一个新标签页显示转换后的 JSON

**效果示例：**

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

将 JSON 内容转换为 ASON 格式。

**使用方式（2 种）：**

**方式一：从 JSON 文件转换**

1. 打开一个 `.json` 文件
2. `Cmd+Shift+P` → 输入 **ASON: Convert JSON to ASON**
3. 自动在新标签页中打开转换后的 ASON

**方式二：手动输入 JSON**

1. 在任意非 JSON 文件中（或不打开文件）
2. `Cmd+Shift+P` → 输入 **ASON: Convert JSON to ASON**
3. 在弹出的输入框中粘贴 JSON 内容
4. 回车后在新标签页中打开转换后的 ASON

**效果示例：**

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

### 5.8 智能提示（Completion）

在编写 ASON 时按 `Ctrl+Space` 触发自动补全：

- 在 `{` 后提示 `map` 关键字
- 在 `:` 后提示类型（`int`, `str`, `bool`, `float`）
- 在 `(` 或 `,` 后提示对应的字段名

### 5.9 悬停提示（Hover）

将鼠标悬停在 ASON 代码的不同部分上，会显示该位置的上下文信息，例如字段类型、节点种类等。

### 5.10 Inlay Hints（内联提示）

**自动生效**。在数据元组的每个值前面，会用灰色文字显示对应的字段名，帮助你快速识别每个值对应 schema 中的哪个字段。

例如：

```
{name:str, age:int, city:str}:(Alice, 30, NYC)
```

编辑器中会显示为：

```
{name:str, age:int, city:str}:(name: Alice, age: 30, city: NYC)
```

（其中 `name:` `age:` `city:` 是灰色的内联提示，不属于实际代码）

### 5.11 语义着色（Semantic Tokens）

插件提供精确的语义级着色，不同元素会使用不同的颜色主题：

| 元素             | 语义类型  |
| ---------------- | --------- |
| `{}` `()` `[]`   | keyword   |
| `:int` `:str` 等 | type      |
| 字段名           | variable  |
| 字符串值         | string    |
| 数字             | number    |
| `/* 注释 */`     | comment   |
| `:` `,`          | operator  |
| `true` `false`   | parameter |

---

## 6. 配置项

在 VS Code 设置中（`Cmd+,`）搜索 `ason`：

| 配置项                    | 类型    | 默认值 | 说明                                      |
| ------------------------- | ------- | ------ | ----------------------------------------- |
| `ason.lspPath`            | string  | `""`   | ason-lsp 可执行文件的路径。留空则自动查找 |
| `ason.inlayHints.enabled` | boolean | `true` | 是否显示字段名内联提示                    |

### 手动指定 ason-lsp 路径

如果 ason-lsp 不在默认位置，打开 VS Code 设置（`Cmd+,`），搜索 `ason.lspPath`，填入完整路径：

```json
{
  "ason.lspPath": "/Users/你的用户名/code/ason/tools/ason-lsp/ason-lsp"
}
```

---

## 7. 常见问题

### Q: 语法高亮正常，但没有错误检查 / 格式化不工作

**原因**：语言服务器 (ason-lsp) 没有正确启动。

**解决方法**：

1. 确认 ason-lsp 已经构建：
   ```bash
   ls -la tools/ason-lsp/ason-lsp
   ```
2. 如果文件不存在，回到 [第 2 步](#2-构建-ason-语言服务器-ason-lsp) 重新构建
3. 在 VS Code 底部状态栏查看 "输出" 面板 → 选择 "ASON Language Server"，检查日志
4. 尝试在设置中手动指定 `ason.lspPath`

### Q: 提示 "ASON LSP binary not found"

**原因**：插件找不到 ason-lsp 可执行文件。

**解决方法**：

- 运行 `go build -o ason-lsp .` 构建语言服务器
- 或在设置中指定路径：`"ason.lspPath": "/完整路径/ason-lsp"`
- 或将 ason-lsp 复制到系统 PATH 中

### Q: 格式化快捷键没反应

**可能原因**：VS Code 将默认格式化器设置为了其他插件。

**解决方法**：在设置中添加：

```json
{
  "[ason]": {
    "editor.defaultFormatter": "ason.ason-vscode"
  }
}
```

### Q: Inlay Hints 看不到

**解决方法**：确认 VS Code 全局设置中 inlay hints 是开启的：

```json
{
  "editor.inlayHints.enabled": "on"
}
```

### Q: 如何卸载插件

```bash
code --uninstall-extension ason.ason-vscode
```

或在 VS Code 扩展面板中找到 "ASON Language Support"，点击卸载。

---

## 快速参考卡片

| 功能        | 操作                                              |
| ----------- | ------------------------------------------------- |
| 格式化      | `Shift+Option+F` 或命令面板 → `ASON: Format`      |
| 压缩        | 命令面板 → `ASON: Compress`                       |
| ASON → JSON | 命令面板 → `ASON: Convert to JSON`                |
| JSON → ASON | 命令面板 → `ASON: Convert JSON to ASON`           |
| 自动补全    | `Ctrl+Space`                                      |
| 命令面板    | `Cmd+Shift+P`（macOS）/ `Ctrl+Shift+P`（Windows） |
