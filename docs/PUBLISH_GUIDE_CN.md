# ASUN 多语言发布手册（中文）

更新时间：2026-03-11

这份文档用于说明 ASUN 各语言仓库应该发布到哪里、首次发布怎么做、后续版本怎么更新。

适用范围：

- `asun-js`（npm 包名：`@athanx/asun`）
- `asun-py`
- `asun-cs`
- `asun-dart`
- `asun-rs`
- `asun-java`
- `asun-go`
- `asun-swift`
- `asun-zig`
- `asun-c`
- `asun-cpp`
- `asun-php`
- `plugin_vscode`

不包含独立 LSP 和网站。

## 先看结论

| 仓库                        | 推荐发布渠道                               | 是否需要单独平台注册 | 备注                                         |
| --------------------------- | ------------------------------------------ | -------------------- | -------------------------------------------- |
| `asun-js`（`@athanx/asun`） | npm                                        | 是                   | 标准 JS/TS 包发布渠道                        |
| `asun-py`                   | PyPI                                       | 是                   | 标准 Python 包发布渠道                       |
| `asun-cs`                   | NuGet.org                                  | 是                   | 标准 .NET 包发布渠道                         |
| `asun-dart`                 | pub.dev                                    | 是                   | 标准 Dart/Flutter 包发布渠道                 |
| `asun-rs`                   | crates.io                                  | 是                   | 标准 Rust crate 发布渠道                     |
| `asun-java`                 | Maven Central                              | 是                   | 标准 Java/JVM 库发布渠道                     |
| `asun-go`                   | GitHub 仓库 + 语义化 tag                   | 否（只要 GitHub）    | Go 没有必须单独上传的中心仓库                |
| `asun-swift`                | GitHub 仓库 + 语义化 tag                   | 否（只要 GitHub）    | SwiftPM 通常直接消费 Git tag                 |
| `asun-zig`                  | GitHub 仓库 + tag/release                  | 否（只要 GitHub）    | Zig 当前没有官方中心仓库                     |
| `asun-c`                    | GitHub Releases                            | 否（只要 GitHub）    | 可后续再补 vcpkg/conan/homebrew              |
| `asun-cpp`                  | GitHub Releases + Conan / vcpkg / Homebrew | 否（只要 GitHub）    | 现在已具备包管理器接入文件，可按需要逐步发布 |
| `asun-php`                  | GitHub Releases + PIE 生态                 | 否（建议先 GitHub）  | 新 PECL 包当前不建议作为首发渠道             |
| `plugin_vscode`             | VS Code Marketplace + Open VSX（可选）     | 是                   | 推荐先发 VS Code Marketplace，再决定是否补 Open VSX |

### 哪些语言需要“多平台二进制发布”

- 不需要额外做多平台二进制矩阵：`asun-js`、`asun-cs`、`asun-dart`、`asun-rs`、`asun-java`
  这些发布到仓库/中心后，分发的是源码包、IL 包或 JVM 包，本身不按操作系统拆分上传。
- 通常也不需要上传多平台二进制：`asun-go`、`asun-swift`、`asun-zig`
  这三类目前推荐通过源码仓库 + tag/release 分发；用户侧按源码构建消费。
- 需要单独考虑多平台产物：`asun-py`、`asun-c`、`asun-cpp`、`asun-php`
  这些包含原生扩展或原生代码。如果你想给最终用户“开箱即用”的安装体验，就应该准备多平台构建矩阵。

### 各语言版本号目前在哪里改

- `asun-js`: `asun-js/package.json`
- `asun-py`: `asun-py/pyproject.toml` 和 `asun-py/setup.py`
- `asun-cs`: `asun-cs/src/Asun/Asun.csproj`
- `asun-dart`: `asun-dart/pubspec.yaml`
- `asun-rs`: `asun-rs/Cargo.toml`
- `asun-java`: `asun-java/build.gradle`
- `asun-go`: Git tag（`go.mod` 只定义 module path，不保存发布版本号）
- `asun-swift`: Git tag（`Package.swift` 不单独声明发布版本）
- `asun-zig`: `asun-zig/build.zig.zon`
- `asun-c`: Git tag / GitHub Release
- `asun-cpp`: `asun-cpp/CMakeLists.txt`
- `asun-php`: Git tag / GitHub Release
- `plugin_vscode`: `plugin_vscode/package.json`

## 当前建议的发布顺序

建议先发这些“标准包管理器”语言：

1. `asun-js`（发布名 `@athanx/asun`）
2. `asun-py`
3. `asun-cs`
4. `asun-dart`
5. `asun-rs`
6. `asun-java`

然后再发这些“基于 Git tag / Release 分发”的语言：

1. `asun-go`
2. `asun-swift`
3. `asun-zig`
4. `asun-c`
5. `asun-cpp`
6. `asun-php`
7. `plugin_vscode`

原因很简单：前一组用户通常会直接通过包管理器安装，发布收益最高；后一组更多是源码依赖或 Release 分发。

## 编辑器插件补充

目前仓库里最适合单独发布的编辑器插件是：

- `plugin_vscode`

`plugin_vscode` 不是语言运行时库，而是一个带内置 `lsp-asun` 二进制的 VS Code 扩展。  
它的发布流程与普通 npm 包不同，核心是：

1. 先构建 TypeScript 扩展代码
2. 再构建并打包 `lsp-asun`
3. 生成 `.vsix`
4. 发布到 VS Code Marketplace

如果后续还要补发 Open VSX，可以在 Marketplace 流程稳定后再追加。

## 通用发布前清单

每个语言在发布前都建议至少做这几件事：

1. 版本号更新到本次准备发布的版本。
2. README / README_CN 示例与真实 API 保持一致。
3. 测试全部通过。
4. examples 至少跑一遍。
5. 确认最终打包产物里没有把 `node_modules`、`build/`、`obj/`、临时文件、密钥、测试缓存一起带出去。
6. 打 Git tag，例如 `v1.0.0`。
7. 准备 GitHub Release 说明，哪怕该语言还有中心仓库，也建议保留 GitHub Release 作为对外版本记录。

如果目标是“多平台版本”，再额外检查：

1. 明确你发布的是源码包、平台无关包，还是平台相关二进制。
2. 如果是原生扩展，先确定支持矩阵：Linux / macOS / Windows，以及 x86_64 / arm64。
3. 不要只在本机跑一次构建就上传；平台相关包应先聚合所有目标产物再统一发布。
4. 对支持源码回退安装的生态，尽量同时提供源码包（例如 `sdist`、源码 tarball）。

## 1. JavaScript / TypeScript：`asun-js`（npm：`@athanx/asun`）

推荐渠道：npm  
官方文档：

- https://docs.npmjs.com/creating-and-publishing-scoped-public-packages
- https://docs.npmjs.com/trusted-publishers

### 注册账号

1. 注册 npm 账号：`https://www.npmjs.com/signup`
2. 建议开启 2FA。
3. 如果以后要用组织作用域，再创建 npm organization。

### 首次发布

1. 确认 `package.json` 中：
   - `name`
   - `version`
   - `repository`
   - `homepage`
   - `bugs`
   - `license`
     都正确，其中 `name` 应为 `@athanx/asun`。
2. 本地检查：

```bash
cd asun-js
npm test
npm run build
npm run typecheck
npm pack --dry-run
```

3. 登录：

```bash
npm login
```

4. 发布：

```bash
npm publish --access public
```

### 后续更新

1. 改版本号。
2. 重新跑测试和构建。
3. 重新发布：

```bash
npm publish --access public
```

## 2. Python：`asun-py`

推荐渠道：PyPI  
官方文档：

- https://pypi.org/account/register/
- https://pypi.org/help/
- https://docs.pypi.org/trusted-publishers/using-a-publisher/

### 注册账号

1. 注册 PyPI 账号：`https://pypi.org/account/register/`
2. 验证邮箱。
3. 进入账号设置创建 API Token，或者后续改成 Trusted Publishing。

### 这个包的发布形态

`asun-py` 不是纯 Python 包，而是一个带 C++/pybind11 原生扩展的包。  
这意味着：

- `sdist` 是跨平台源码包
- `wheel` 是按 Python 版本和平台分别构建的二进制包
- 如果只在一台 Linux 机器上执行 `bdist_wheel`，得到的只会是当前平台的 wheel，不是完整多平台发布

仓库里已经补了 `cibuildwheel` 配置，位置：

- `asun-py/pyproject.toml`

### 推荐的首发方式：`sdist` + `cibuildwheel` 多平台 wheels

准备工具：

```bash
cd asun-py
python3 -m pip install -U build twine cibuildwheel
```

先构建源码包：

```bash
python3 -m build --sdist
```

在当前机器上构建本平台 wheels：

```bash
python3 -m cibuildwheel --output-dir dist
```

这条命令会按当前平台批量产出多个 wheel。  
例如在 Linux `x86_64` 上，会生成 `cp38` 到 `cp313` 的 manylinux wheels。

本地验证时，不要直接安装 `dist/*.whl`，因为其中很多 wheel 与当前 Python 版本不匹配。应只安装“当前解释器对应的那个 wheel”。

例如，当前 Python 如果是 3.13：

```bash
python3 -m pip install --force-reinstall dist/asun-1.0.0-cp313-cp313-*.whl
python3 -m pytest tests -v
```

如果想自动选择与当前 Python 版本匹配的 wheel，可以这样：

```bash
python3 -m pip install --force-reinstall \
  "$(find dist -maxdepth 1 -name "*cp$(python3 -c 'import sys; print(f"{sys.version_info.major}{sys.version_info.minor}")')-cp$(python3 -c 'import sys; print(f"{sys.version_info.major}{sys.version_info.minor}")')*.whl" | head -n 1)"
python3 -m pytest tests -v
```

上传：

```bash
python3 -m twine upload dist/*
```

### 真正的“多平台 wheel”该怎么发

最稳妥的方法是用 CI 跑构建矩阵，然后把所有产物集中上传到 PyPI。推荐至少覆盖：

- Linux `x86_64`
- Linux `aarch64`
- macOS `x86_64`
- macOS `arm64`
- Windows `x86_64`

发布后的产物形态大致会是：

- Linux：`manylinux` wheels
- macOS Intel：`macosx_x86_64` wheels
- macOS Apple Silicon：`macosx_arm64` wheels
- Windows：`win_amd64` wheels

实际执行顺序建议是：

1. 在 Linux / macOS / Windows runner 上分别跑 `python -m cibuildwheel --output-dir dist`
2. 收集所有 wheel 和一个 `sdist`
3. 最后统一执行一次 `python -m twine upload dist/*`

### 按平台如何构建

#### Linux

Linux 下 `cibuildwheel` 默认通过 Docker manylinux 容器构建，所以：

- 需要 Docker 可用
- 当前用户要有 Docker 权限

最常见的本地命令：

```bash
cd asun-py
python3 -m cibuildwheel --output-dir dist
```

如果你想显式指定架构：

```bash
CIBW_ARCHS_LINUX=x86_64 python3 -m cibuildwheel --output-dir dist
```

Linux `aarch64` 一般更适合在对应架构 runner 上构建，或者使用支持模拟/原生 ARM 的 CI。  
如果 CI 环境支持，也可以这样指定：

```bash
CIBW_ARCHS_LINUX=aarch64 python3 -m cibuildwheel --output-dir dist
```

#### macOS

macOS 不走 manylinux 容器，而是直接在宿主机 runner 上构建。推荐用两台 runner 或两个 job：

- `macos-13` / Intel：构建 `x86_64`
- `macos-14` / Apple Silicon：构建 `arm64`

命令示例：

```bash
cd asun-py
CIBW_ARCHS_MACOS=x86_64 python3 -m cibuildwheel --output-dir dist
```

```bash
cd asun-py
CIBW_ARCHS_MACOS=arm64 python3 -m cibuildwheel --output-dir dist
```

如果你以后确认扩展支持 universal2，也可以再考虑：

```bash
CIBW_ARCHS_MACOS=universal2 python3 -m cibuildwheel --output-dir dist
```

#### Windows

Windows 也直接在宿主 runner 上构建。最常见的是 `x86_64`：

```bash
cd asun-py
python -m cibuildwheel --output-dir dist
```

如果要显式指定：

```bash
set CIBW_ARCHS_WINDOWS=AMD64
python -m cibuildwheel --output-dir dist
```

PowerShell 写法：

```powershell
$env:CIBW_ARCHS_WINDOWS="AMD64"
python -m cibuildwheel --output-dir dist
```

#### 建议的最小 CI 矩阵

如果你要首发一个“多平台可安装”的版本，推荐最少准备这 5 组：

- Linux `x86_64`
- Linux `aarch64`
- macOS `x86_64`
- macOS `arm64`
- Windows `x86_64`

如果暂时做不全，优先级建议是：

1. Linux `x86_64`
2. macOS `arm64`
3. Windows `x86_64`
4. macOS `x86_64`
5. Linux `aarch64`

### 如何验证不同平台的 wheel

验证时要注意两件事：

- wheel 必须和当前 Python 主版本/次版本匹配
- wheel 还必须和当前操作系统、CPU 架构匹配

例如：

- Linux `x86_64` 机器，Python 3.13，可以安装 `cp313-manylinux-...-x86_64.whl`
- macOS 机器，不能安装 `manylinux` wheel
- Python 3.14，不能安装 `cp313` wheel

因此，多平台发布后的验证通常也是分平台完成：

1. Linux runner 验证 Linux wheels
2. macOS runner 验证 macOS wheels
3. Windows runner 验证 Windows wheels

### 如何构建不同 Python 版本的 wheel

有两种方式：

#### 方式一：推荐，用 `cibuildwheel` 批量构建

```bash
cd asun-py
python3 -m cibuildwheel --output-dir dist
```

在 Linux 上，这会一次性构建：

- `cp38`
- `cp39`
- `cp310`
- `cp311`
- `cp312`
- `cp313`
- `cp314`

优点是：

- 不需要手工切换 Python 解释器
- 输出的 wheel 命名、tag、repair 流程都更标准
- 适合正式发布
- 在不同平台 runner 上重复执行同一命令，就能得到对应平台的 wheel

#### 方式二：用 `mise` 手工切不同 Python 版本逐个构建

如果你只是本地验证某个特定 Python 版本，也可以手工切版本：

```bash
cd asun-py
mise exec python@3.10 -- python -m pip install -U pip build setuptools wheel cmake pytest
mise exec python@3.10 -- python -m build --wheel --no-isolation
mise exec python@3.10 -- python -m pip install --force-reinstall dist/asun-1.0.0-cp310-*.whl
mise exec python@3.10 -- python -m pytest tests -v
```

其它版本同理，把 `3.10` 换成：

- `3.8`
- `3.9`
- `3.11`
- `3.12`
- `3.13`
- `3.14`

注意：

- 这种手工方式更适合本地调试，不适合替代正式的多平台发布
- 如果 `mise` 里还没装对应 Python 版本，需要先安装，例如：

```bash
mise install python@3.10
```

### 推荐的上传方式

推荐不要在每个平台构建完就立刻上传，而是：

1. 先构建一个 `sdist`
2. 在各平台 runner 上分别构建 wheels
3. 把所有 `dist/*` 产物汇总到一个地方
4. 最后统一执行一次：

```bash
python3 -m twine upload dist/*
```

这样 PyPI 上会一次性出现完整的平台矩阵，用户安装体验最好。

### 版本号位置

1. `asun-py/pyproject.toml`
2. `asun-py/setup.py`

这两个地方目前都声明了版本号，发布前必须保持一致。

### 后续更新

1. 递增版本号。
2. 重新生成 `sdist`。
3. 重新跑所有目标平台的 `cibuildwheel`。
4. 聚合产物后重新上传。

### 备注

- PyPI 上同一版本号不能覆盖重传，所以版本号一旦发出就要递增。
- 如果暂时来不及补齐所有平台，至少也要上传 `sdist`，这样用户仍有机会本地编译安装。

## 3. C# / .NET：`asun-cs`

推荐渠道：NuGet.org  
官方文档：

- https://learn.microsoft.com/en-us/nuget/nuget-org/overview-nuget-org
- https://learn.microsoft.com/en-us/nuget/nuget-org/publish-a-package

### 注册账号

1. 登录/注册 NuGet：`https://www.nuget.org/account`
2. 生成 API Key。

### 首次发布

当前仓库已经是单包多目标：

- `net8.0`
- `net10.0`

这类 NuGet 包是 IL + 元数据包，不需要按 Linux / macOS / Windows 分开上传。

本地检查：

```bash
cd asun-cs
dotnet pack src/Asun/Asun.csproj -c Release
dotnet test tests/Asun.Tests/Asun.Tests.csproj -f net10.0
```

发布：

```bash
dotnet nuget push src/Asun/bin/Release/*.nupkg \
  --api-key <NUGET_API_KEY> \
  --source https://api.nuget.org/v3/index.json
```

### 后续更新

1. 改 `<Version>`。
2. 重新 `dotnet pack`。
3. 重新 `dotnet nuget push`。

版本号位置：

- `asun-cs/src/Asun/Asun.csproj`

## 4. Dart：`asun-dart`

推荐渠道：pub.dev  
官方文档：

- https://dart.dev/tools/pub/publishing
- https://dart.dev/tools/pub/verified-publishers
- https://dart.dev/tools/pub/cmd/pub-lish

### 注册账号

1. 用 Google 账号登录 `https://pub.dev/`
2. 如果你有自己的域名，建议创建 verified publisher。
3. 注意：官方文档当前说明，新包不能直接第一次就发到 verified publisher；通常先用个人账号发第一次，再转移到 publisher。

### 首次发布

```bash
cd asun-dart
dart test
dart pub publish --dry-run
dart pub publish
```

`asun-dart` 当前是纯 Dart 包，不需要按平台分别发布。

### 后续更新

```bash
cd asun-dart
dart test
dart pub publish --dry-run
dart pub publish
```

版本号位置：

- `asun-dart/pubspec.yaml`

## 5. Rust：`asun-rs`

推荐渠道：crates.io  
官方文档：

- https://doc.rust-lang.org/book/ch14-02-publishing-to-crates-io.html

### 注册账号

1. 打开 `https://crates.io/`
2. 用 GitHub 账号登录。
3. 在 crates.io 个人设置页创建 API token。

### 首次发布

```bash
cd asun-rs
cargo test
cargo package
cargo login <CRATES_IO_TOKEN>
cargo publish
```

`asun-rs` 当前发布的是 crate 源码包，不需要额外做多平台上传。

### 后续更新

1. 改 `Cargo.toml` 版本号。
2. 重新测试。
3. 再执行：

```bash
cargo publish
```

版本号位置：

- `asun-rs/Cargo.toml`

## 6. Java：`asun-java`

推荐渠道：Maven Central  
官方文档：

- https://central.sonatype.org/register/central-portal/
- https://central.sonatype.org/register/namespace/
- https://central.sonatype.org/publish/generate-portal-token/
- https://central.sonatype.org/publish/publish-portal-gradle/
- https://central.sonatype.org/publish/publish-portal-ossrh-staging-api/

### 注册账号

1. 登录 `https://central.sonatype.com/`
2. 创建/认领 namespace。
3. 验证 namespace（通常是域名或仓库来源证明）。
4. 生成 Portal User Token。

### 首次发布

`asun-java` 现在已经接入：

1. `maven-publish`
2. `signing`
3. `sourcesJar`
4. `javadocJar`
5. Sonatype Central Portal 兼容发布端点

因此首发时建议直接走当前仓库内置的 Gradle 发布方案。

发布前你需要准备：

1. Central Portal User Token
2. GPG/PGP 签名私钥
3. `groupId / artifactId / version`
4. 确认 README / API / 测试都已经对齐

先设置环境变量：

```bash
export SONATYPE_USERNAME=...
export SONATYPE_PASSWORD=...
export SIGNING_KEY='-----BEGIN PGP PRIVATE KEY BLOCK----- ...'
export SIGNING_PASSWORD=...
```

本地先验证：

```bash
cd asun-java
./gradlew test
./gradlew publishToMavenLocal
```

正式发布：

```bash
cd asun-java
./gradlew publish
```

### Kotlin 怎么发布

Kotlin **不需要单独发布一个新包**。

原因是当前仓库里：

- Java 和 Kotlin helper 在同一个 `asun-java` artifact 里
- Kotlin API 只是对同一套 JVM runtime 的 helper 封装
- 发布到 Maven Central 时，使用者拿到的是同一个坐标

也就是说，当前发布对象只有这一份：

- `groupId = io.asun`
- `artifactId = asun-java`
- `version = 例如 1.0.0`

Kotlin 用户发布后仍然通过同一个 Maven 坐标消费，不存在单独的 `asun-kotlin` 包。

### 后续更新

1. 改版本号。
2. 重新签名并发布。
3. 在 Central Portal 检查 deployment 状态是否已经 `PUBLISHED`。

版本号位置：

- `asun-java/build.gradle`

补充说明：

- Maven Central 上这类 JAR 是跨平台的，不需要额外做多操作系统矩阵发布。
- 当前构建要求 Java 21 toolchain，发布前最好用 CI 或本地 `./gradlew test publishToMavenLocal` 再验一次。

## 7. Go：`asun-go`

推荐渠道：GitHub 仓库 + 语义化 tag  
官方文档：

- https://go.dev/blog/publishing-go-modules
- https://go.dev/ref/mod

### 是否需要单独注册

不需要额外 Go 包仓库账号。  
只需要 GitHub 仓库是公开的，并且 module path 正确。

### 首次发布

1. 确认 `go.mod` 的 module path 与最终仓库地址一致。
2. 跑测试：

```bash
cd asun-go
go test ./...
```

3. 打 tag：

```bash
git tag v1.0.0
git push origin v1.0.0
```

4. 验证：

```bash
go list -m github.com/asunLab/asun-go@v1.0.0
```

`asun-go` 当前更适合按源码模块发布，不需要上传多平台二进制；真正的版本号由 Git tag 决定。

### 后续更新

每次新版本只需要：

1. 改代码
2. 跑测试
3. 打新 tag，例如 `v0.1.1`

## 8. Swift：`asun-swift`

推荐渠道：GitHub 仓库 + 语义化 tag  
官方参考：

- GitHub Release/Tag：https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository

### 是否需要单独注册

通常不需要。  
Swift Package Manager 最常见的分发方式就是直接通过 Git 仓库和语义化 tag。

### 首次发布

1. 确认 `Package.swift`、README、examples 都正确。
2. 跑：

```bash
swift test
```

3. 打 tag：

```bash
git tag 1.0.0
git push origin 1.0.0
```

4. 建一个 GitHub Release。

`asun-swift` 当前是源码形式的 Swift Package；消费侧按源码构建，不需要单独上传多平台二进制。

### 后续更新

```bash
git tag 0.1.1
git push origin 0.1.1
```

版本号位置：

- Git tag

补充说明：

- `Package.swift` 定义了支持平台范围，但发布版本号仍然以 tag 为准。

## 9. Zig：`asun-zig`

推荐渠道：GitHub 仓库 + tag/release  
官方参考：

- Zig 0.11 package management 说明：`build.zig.zon` + 无官方中心仓库  
  https://ziglang.org/download/0.11.0/release-notes.html

### 是否需要单独注册

不需要。  
当前 Zig 官方没有像 npm / PyPI / crates.io 这样的官方中心仓库。

### 首次发布

1. 确认 `build.zig.zon` 正确。
2. 跑：

```bash
cd asun-zig
zig build test
```

3. 打 tag：

```bash
git tag 1.0.0
git push origin 1.0.0
```

4. 建 GitHub Release，附带源码归档或二进制产物。

`asun-zig` 通常按源码包分发；如果想提供预编译二进制，可额外做 GitHub Actions 矩阵构建，但不是 Zig 包发布本身的必需步骤。

### 后续更新

继续用新 tag + GitHub Release。

版本号位置：

- `asun-zig/build.zig.zon`
- Git tag

## 10. C：`asun-c`

推荐渠道：GitHub Releases  
官方参考：

- GitHub Releases：https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository

### 是否需要单独注册

不需要额外注册，只需要 GitHub。

### 首次发布

```bash
cd asun-c
cmake -S . -B build
cmake --build build
ctest --test-dir build --output-on-failure
```

然后：

1. 打 tag
2. 建 GitHub Release
3. 上传源码包
4. 如果要提供多平台预编译产物，再额外上传：
   - Linux `x86_64` / `arm64`
   - macOS `x86_64` / `arm64`
   - Windows `x86_64`
     的静态库、动态库或示例二进制

### 后续更新

继续 tag + Release。

### 备注

如果以后想进一步进入系统包管理生态，可以再补：

- vcpkg
- Conan
- Homebrew

但这不是首发必需。

版本号位置：

- Git tag / GitHub Release

## 11. C++：`asun-cpp`

推荐渠道：

1. GitHub Release
2. Conan
3. vcpkg
4. Homebrew

说明：

- `asun-cpp` 当前是标准 header-only CMake package
- Conan recipe 已就绪
- vcpkg 当前是 overlay port 级别
- Homebrew 当前是 formula 模板级别，发布前需要填真实 tarball 的 `sha256`

因为当前是 header-only CMake package，最核心的发布其实是源码包和包管理器元数据，不需要像二进制 SDK 那样为每个平台单独产出 `.dll/.so/.dylib`。

### 首次发布前验证

```bash
cd asun-cpp
cmake -S . -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build
ctest --test-dir build --output-on-failure
```

建议再额外验证一次标准 CMake 包导出是否可用：

```bash
cmake --install build --prefix /tmp/asun-cpp-install
```

### 11.1 GitHub Release

这是最基础、也最先完成的一步。

流程：

1. 更新版本号（当前在 `CMakeLists.txt`、`conanfile.py`、`vcpkg/ports/asun-cpp/vcpkg.json`、`homebrew/asun-cpp.rb` 中都要一致）
2. 跑完测试和本地安装验证
3. 打 tag，例如 `v1.0.0`
4. 在 GitHub 上创建 Release
5. 上传源码归档，或直接使用 GitHub 自动生成的 source tarball / zip

### 11.2 Conan

仓库内已有：

- `asun-cpp/conanfile.py`
- `asun-cpp/test_package/`

本地验证：

```bash
cd asun-cpp
conan create . --build=missing
```

发布方式有两种：

1. 先用仓库内 recipe 做本地或私有远端发布
2. 后续再向 ConanCenter 提交 recipe

如果只是你自己的 first-party 发布，最简单的是：

```bash
cd asun-cpp
conan export . --version=1.0.0
```

然后在自己的 Conan remote 上上传对应 recipe/package。

### 11.3 vcpkg

仓库内已有 overlay port：

- `asun-cpp/vcpkg/ports/asun-cpp/portfile.cmake`
- `asun-cpp/vcpkg/ports/asun-cpp/vcpkg.json`
- `asun-cpp/vcpkg/ports/asun-cpp/vcpkg-cmake-wrapper.cmake`

当前定位：

- 这套文件已经足够用于 overlay port
- 如果要进入官方 vcpkg registry，还需要单独向 `microsoft/vcpkg` 提交 port PR

本地验证：

```bash
vcpkg install asun-cpp --overlay-ports=/path/to/asun-cpp/vcpkg/ports
```

推荐发布步骤：

1. 先完成 GitHub Release
2. 确认 release tag、源码 tarball 稳定
3. 基于当前 port 文件向 vcpkg 官方仓库提 PR

### 11.4 Homebrew

仓库内已有 formula 模板：

- `asun-cpp/homebrew/asun-cpp.rb`

注意：

- 当前 formula 里的 `sha256` 还是占位符
- 真正发布前，需要把它替换成 GitHub Release tarball 的真实 `sha256`

计算方式示例：

```bash
curl -L -o /tmp/asun-v1.0.0.tar.gz \
  https://github.com/asunLab/asun/archive/refs/tags/v1.0.0.tar.gz
shasum -a 256 /tmp/asun-v1.0.0.tar.gz
```

然后把得到的 hash 填入：

- `asun-cpp/homebrew/asun-cpp.rb`

本地验证可用：

```bash
brew install --build-from-source ./homebrew/asun-cpp.rb
brew test asun-cpp
```

如果要正式对外发布，通常有两种方式：

1. 自己维护一个 tap
2. 后续尝试提交到 Homebrew core（通常要求更严格）

### 后续更新

每次发新版本时，建议按这个顺序推进：

1. 更新 `CMakeLists.txt`
2. 更新 `conanfile.py`
3. 更新 `vcpkg/ports/asun-cpp/vcpkg.json`
4. 更新 `homebrew/asun-cpp.rb` 中的版本、下载地址与 `sha256`
5. 跑测试和打包验证
6. 打 Git tag
7. 创建 GitHub Release
8. 再分别同步 Conan / vcpkg / Homebrew

版本号位置：

- `asun-cpp/CMakeLists.txt`
- `asun-cpp/conanfile.py`
- `asun-cpp/vcpkg/ports/asun-cpp/vcpkg.json`
- `asun-cpp/homebrew/asun-cpp.rb`

### 备注

- Conan：现在已经是可用级别
- vcpkg：现在是 overlay port 可用级别
- Homebrew：现在是 formula 模板可用级别
- 所以 `asun-cpp` 目前已经不只是“GitHub Release 可发”，而是已经具备继续进入包管理器生态的基础文件

## 12. PHP 扩展：`asun-php`

推荐渠道：GitHub Releases + PIE 生态观察  
官方参考：

- PECL 首页：https://pecl.php.net/
- PECL 当前账号页提示：https://pecl.php.net/account-request.php
- GitHub Releases：https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository

### 很重要：当前不要把“新包进 PECL”当首发路线

截至 2026-03-11，PECL 官方页面已经明确提示：

- PECL 已经 deprecated
- 新包不再接受新的 PECL package 账号申请
- 建议使用 PIE 生态

因此，对 `asun-php` 这个新扩展，更实际的首发方式是：

1. GitHub Release
2. 提供源码包
3. README 里给清楚的 `phpize / ./configure / make / make install` 安装说明

### 首次发布

```bash
cd asun-php
phpize
./configure
make -j2
make test
```

然后：

1. 打 tag
2. 建 GitHub Release
3. 上传源码归档
4. 如果要提供多平台预编译扩展，再按 PHP 版本 + 操作系统分别构建对应产物
5. 在 Release 说明中写清：
   - 支持的 PHP 版本
   - 编译依赖
   - 安装命令

### 后续更新

继续用 Git tag + GitHub Release。

如果未来 PIE 生态对第三方扩展的发布路径稳定下来，再补自动化发布。

版本号位置：

- Git tag / GitHub Release

## GitHub Release 的建议做法

适用于 `asun-go`、`asun-swift`、`asun-zig`、`asun-c`、`asun-cpp`、`asun-php`，以及其它语言的辅助发布记录。

官方文档：

- https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository

建议流程：

1. 先打版本 tag，例如 `v1.0.0`
2. 在 GitHub 上创建 Release
3. 标题写成版本号
4. Release notes 至少包含：
   - 新增功能
   - breaking changes
   - 安装方式
   - 已知限制

## 推荐的首次发布实际执行顺序

如果你现在就要开始发，我建议这样排：

1. `asun-js` -> npm（包名 `@athanx/asun`）
2. `asun-cs` -> NuGet
3. `asun-dart` -> pub.dev
4. `asun-rs` -> crates.io
5. `asun-java` -> Maven Central
6. `asun-go` -> Git tag + GitHub Release
7. `asun-swift` -> Git tag + GitHub Release
8. `asun-zig` -> Git tag + GitHub Release
9. `asun-c` -> GitHub Release
10. `asun-cpp` -> GitHub Release
11. `asun-php` -> GitHub Release
12. `asun-py` -> PyPI（建议等 `sdist + cibuildwheel` 多平台产物准备好后再发）

## 13. VS Code 插件：`plugin_vscode`

推荐渠道：VS Code Marketplace  
可选补充渠道：Open VSX  
仓库位置：

- `plugin_vscode`

当前 `package.json` 关键信息：

- 扩展名：`asun-vscode`
- 显示名：`ASUN Support`
- `publisher`：`Athan`
- 版本号文件：`plugin_vscode/package.json`

### 发布前先确认

1. `plugin_vscode/package.json` 里的 `version` 已更新。
2. `publisher` 与 Marketplace 上创建的发布者一致。
3. `README.md`、截图、图标、命令说明都已经准备好。
4. `lsp-asun` 可以正常构建。
5. 插件能在本地通过 VSIX 安装并工作。

### 首次发布前要准备什么

VS Code Marketplace 的发布者不是 npm 账号，也不是 GitHub 账号本身。首次发布通常需要：

1. 注册/登录 Visual Studio Marketplace 发布者后台。
2. 创建一个发布者（publisher），名称应与 `package.json` 里的 `publisher` 保持一致。
3. 创建用于发布的 Personal Access Token。
4. 在本机安装发布工具：

```bash
mise exec -- npm install -g @vscode/vsce
```

### 本地构建与打包

仓库已经提供了跨平台打包脚本：

- `plugin_vscode/scripts/build.sh`

如果你只想先构建当前平台：

```bash
cd plugin_vscode
mise exec -- npm install
./scripts/build.sh current
```

如果你要一次生成全部平台的 VSIX：

```bash
cd plugin_vscode
mise exec -- npm install
./scripts/build.sh all
```

这个脚本会：

1. 安装插件依赖
2. 编译 TypeScript
3. 进入 `../lsp-asun` 构建语言服务器
4. 把 `lsp-asun` 打包进 `plugin_vscode/server/`
5. 生成对应平台的 `.vsix`

### 发布前建议的本地验证

至少做一次当前平台完整检查：

```bash
cd plugin_vscode
mise exec -- npm install
./scripts/build.sh current
```

生成 VSIX 后，用 VS Code 本地安装验证：

```bash
code --install-extension ./asun-vscode-<platform>-<version>.vsix
```

验证项建议至少包括：

1. `.asun` 文件语法高亮正常。
2. 格式化命令可用。
3. 压缩命令可用。
4. `ASUN -> JSON` / `JSON -> ASUN` 命令可用。
5. 诊断、hover、inlay hints 能正常工作。

### 正式发布到 VS Code Marketplace

如果已经拿到 Marketplace token，可以直接在 `plugin_vscode` 目录发布。

先登录一次：

```bash
cd plugin_vscode
mise exec -- vsce login Athan
```

然后发布：

```bash
cd plugin_vscode
mise exec -- vsce publish
```

如果本次已经改好了版本号，也可以直接指定版本发布：

```bash
cd plugin_vscode
mise exec -- vsce publish 0.1.0
```

如果你想先只生成包，不立刻发布：

```bash
cd plugin_vscode
mise exec -- npx vsce package
```

### 平台包与 Marketplace 的关系

这个插件内置了 `lsp-asun` 二进制，因此更适合按平台生成 VSIX。  
仓库脚本已经支持：

- `darwin-arm64`
- `darwin-x64`
- `linux-x64`
- `linux-arm64`
- `win32-x64`
- `win32-arm64`

你可以：

1. 只构建当前平台并做本地验证。
2. 再构建全部平台 VSIX。
3. 最后选择发布策略。

如果后续确认要把 Marketplace 也做成完整平台覆盖，建议把 `./scripts/build.sh all` 放进 CI，而不是只在本机手工打包。

### 是否还要发 Open VSX

如果目标用户包含：

- VSCodium
- OpenVSX 生态编辑器
- 一些默认不连微软 Marketplace 的环境

可以补发 Open VSX。

建议顺序是：

1. 先把 VS Code Marketplace 流程跑通。
2. 确认版本号、README、图标、打包内容稳定。
3. 再补 Open VSX 发布。

### 发布后建议补做的事情

1. 打 Git tag，例如 `plugin_vscode-v0.1.0`，或者跟随仓库统一版本策略。
2. 在 GitHub Release 里附上 `.vsix`。
3. 更新网站或文档中的安装说明。
4. 记录本次发布覆盖了哪些平台包。

## 维护建议

建议后续统一做三件事：

1. 所有语言版本号尽量保持一致。
2. 所有仓库都保留 Git tag 和 GitHub Release。
3. 平台相关包优先补 CI 矩阵，再发布；不要长期依赖“在一台机器上手工打所有二进制”。
4. 对 `asun-py`、`asun-c`、`asun-php` 这类原生包，优先提供源码包；预编译产物按支持矩阵逐步补齐。

## 参考链接

- npm publishing: https://docs.npmjs.com/creating-and-publishing-scoped-public-packages
- npm trusted publishers: https://docs.npmjs.com/trusted-publishers
- PyPI register: https://pypi.org/account/register/
- PyPI help / tokens: https://pypi.org/help/
- PyPI trusted publishing: https://docs.pypi.org/trusted-publishers/using-a-publisher/
- NuGet publish: https://learn.microsoft.com/en-us/nuget/nuget-org/publish-a-package
- NuGet overview: https://learn.microsoft.com/en-us/nuget/nuget-org/overview-nuget-org
- crates.io publishing: https://doc.rust-lang.org/book/ch14-02-publishing-to-crates-io.html
- Dart publishing: https://dart.dev/tools/pub/publishing
- Dart verified publishers: https://dart.dev/tools/pub/verified-publishers
- Go modules publishing: https://go.dev/blog/publishing-go-modules
- Go modules reference: https://go.dev/ref/mod
- Sonatype Central register: https://central.sonatype.org/register/central-portal/
- Sonatype namespace: https://central.sonatype.org/register/namespace/
- Sonatype token: https://central.sonatype.org/publish/generate-portal-token/
- Sonatype Gradle: https://central.sonatype.org/publish/publish-portal-gradle/
- Sonatype OSSRH staging compatibility: https://central.sonatype.org/publish/publish-portal-ossrh-staging-api/
- GitHub Releases: https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository
- VS Code extension publishing: https://code.visualstudio.com/api/working-with-extensions/publishing-extension
- vsce: https://github.com/microsoft/vscode-vsce
- Zig package management note: https://ziglang.org/download/0.11.0/release-notes.html
- PECL / PIE notice: https://pecl.php.net/account-request.php
