# ASON 多语言发布手册（中文）

更新时间：2026-03-11

这份文档用于说明 ASON 各语言仓库应该发布到哪里、首次发布怎么做、后续版本怎么更新。

适用范围：

- `ason-js`
- `ason-py`
- `ason-cs`
- `ason-dart`
- `ason-rs`
- `ason-java`
- `ason-go`
- `ason-swift`
- `ason-zig`
- `ason-c`
- `ason-cpp`
- `ason-php`

不包含编辑器插件、LSP、网站。

## 先看结论

| 仓库 | 推荐发布渠道 | 是否需要单独平台注册 | 备注 |
| --- | --- | --- | --- |
| `ason-js` | npm | 是 | 标准 JS/TS 包发布渠道 |
| `ason-py` | PyPI | 是 | 标准 Python 包发布渠道 |
| `ason-cs` | NuGet.org | 是 | 标准 .NET 包发布渠道 |
| `ason-dart` | pub.dev | 是 | 标准 Dart/Flutter 包发布渠道 |
| `ason-rs` | crates.io | 是 | 标准 Rust crate 发布渠道 |
| `ason-java` | Maven Central | 是 | 标准 Java/JVM 库发布渠道 |
| `ason-go` | GitHub 仓库 + 语义化 tag | 否（只要 GitHub） | Go 没有必须单独上传的中心仓库 |
| `ason-swift` | GitHub 仓库 + 语义化 tag | 否（只要 GitHub） | SwiftPM 通常直接消费 Git tag |
| `ason-zig` | GitHub 仓库 + tag/release | 否（只要 GitHub） | Zig 当前没有官方中心仓库 |
| `ason-c` | GitHub Releases | 否（只要 GitHub） | 可后续再补 vcpkg/conan/homebrew |
| `ason-cpp` | GitHub Releases + Conan / vcpkg / Homebrew | 否（只要 GitHub） | 现在已具备包管理器接入文件，可按需要逐步发布 |
| `ason-php` | GitHub Releases + PIE 生态 | 否（建议先 GitHub） | 新 PECL 包当前不建议作为首发渠道 |

## 当前建议的发布顺序

建议先发这些“标准包管理器”语言：

1. `ason-js`
2. `ason-py`
3. `ason-cs`
4. `ason-dart`
5. `ason-rs`
6. `ason-java`

然后再发这些“基于 Git tag / Release 分发”的语言：

1. `ason-go`
2. `ason-swift`
3. `ason-zig`
4. `ason-c`
5. `ason-cpp`
6. `ason-php`

原因很简单：前一组用户通常会直接通过包管理器安装，发布收益最高；后一组更多是源码依赖或 Release 分发。

## 通用发布前清单

每个语言在发布前都建议至少做这几件事：

1. 版本号更新到本次准备发布的版本。
2. README / README_CN 示例与真实 API 保持一致。
3. 测试全部通过。
4. examples 至少跑一遍。
5. 确认最终打包产物里没有把 `node_modules`、`build/`、`obj/`、临时文件、密钥、测试缓存一起带出去。
6. 打 Git tag，例如 `v1.0.0`。
7. 准备 GitHub Release 说明，哪怕该语言还有中心仓库，也建议保留 GitHub Release 作为对外版本记录。

## 1. JavaScript / TypeScript：`ason-js`

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
   都正确。
2. 本地检查：

```bash
cd ason-js
mise exec -- npm test
mise exec -- npm run build
mise exec -- npm run typecheck
mise exec -- npm pack --dry-run
```

3. 登录：

```bash
mise exec -- npm login
```

4. 发布：

```bash
mise exec -- npm publish --access public
```

### 后续更新

1. 改版本号。
2. 重新跑测试和构建。
3. 重新发布：

```bash
mise exec -- npm publish --access public
```

## 2. Python：`ason-py`

推荐渠道：PyPI  
官方文档：

- https://pypi.org/account/register/
- https://pypi.org/help/
- https://docs.pypi.org/trusted-publishers/using-a-publisher/

### 注册账号

1. 注册 PyPI 账号：`https://pypi.org/account/register/`
2. 验证邮箱。
3. 进入账号设置创建 API Token，或者后续改成 Trusted Publishing。

### 首次发布

1. 先确保 wheel 安装后真的能 `import ason`。
2. 构建：

```bash
cd ason-py
mise exec -- python3 setup.py sdist bdist_wheel
```

3. 本地验证：

```bash
mise exec -- python3 -m pip install --force-reinstall dist/*.whl
mise exec -- python3 -m pytest tests -v
```

4. 上传（推荐 `twine`）：

```bash
mise exec -- python3 -m pip install -U twine
mise exec -- python3 -m twine upload dist/*
```

### 后续更新

1. 改版本号。
2. 重新构建 `sdist` 和 `wheel`。
3. 用新版本重新上传：

```bash
mise exec -- python3 -m twine upload dist/*
```

### 备注

PyPI 上同一版本号不能覆盖重传，所以版本号一旦发出就要递增。

## 3. C# / .NET：`ason-cs`

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

本地检查：

```bash
cd ason-cs
dotnet pack src/Ason/Ason.csproj -c Release
dotnet test tests/Ason.Tests/Ason.Tests.csproj -f net10.0
```

发布：

```bash
dotnet nuget push src/Ason/bin/Release/*.nupkg \
  --api-key <NUGET_API_KEY> \
  --source https://api.nuget.org/v3/index.json
```

### 后续更新

1. 改 `<Version>`。
2. 重新 `dotnet pack`。
3. 重新 `dotnet nuget push`。

## 4. Dart：`ason-dart`

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
cd ason-dart
mise exec -- dart test
mise exec -- dart pub publish --dry-run
mise exec -- dart pub publish
```

### 后续更新

```bash
cd ason-dart
mise exec -- dart test
mise exec -- dart pub publish --dry-run
mise exec -- dart pub publish
```

## 5. Rust：`ason-rs`

推荐渠道：crates.io  
官方文档：

- https://doc.rust-lang.org/book/ch14-02-publishing-to-crates-io.html

### 注册账号

1. 打开 `https://crates.io/`
2. 用 GitHub 账号登录。
3. 在 crates.io 个人设置页创建 API token。

### 首次发布

```bash
cd ason-rs
mise exec -- cargo test
mise exec -- cargo package
mise exec -- cargo login <CRATES_IO_TOKEN>
mise exec -- cargo publish
```

### 后续更新

1. 改 `Cargo.toml` 版本号。
2. 重新测试。
3. 再执行：

```bash
mise exec -- cargo publish
```

## 6. Java：`ason-java`

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

`ason-java` 现在已经接入：

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
cd ason-java
mise exec -- ./gradlew test
mise exec -- ./gradlew publishToMavenLocal
```

正式发布：

```bash
cd ason-java
mise exec -- ./gradlew publish
```

### Kotlin 怎么发布

Kotlin **不需要单独发布一个新包**。

原因是当前仓库里：

- Java 和 Kotlin helper 在同一个 `ason-java` artifact 里
- Kotlin API 只是对同一套 JVM runtime 的 helper 封装
- 发布到 Maven Central 时，使用者拿到的是同一个坐标

也就是说，当前发布对象只有这一份：

- `groupId = io.ason`
- `artifactId = ason-java`
- `version = 例如 1.0.0`

Kotlin 用户发布后仍然通过同一个 Maven 坐标消费，不存在单独的 `ason-kotlin` 包。

### 后续更新

1. 改版本号。
2. 重新签名并发布。
3. 在 Central Portal 检查 deployment 状态是否已经 `PUBLISHED`。

## 7. Go：`ason-go`

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
cd ason-go
mise exec -- go test ./...
```

3. 打 tag：

```bash
git tag v1.0.0
git push origin v1.0.0
```

4. 验证：

```bash
go list -m github.com/ason-lab/ason-go@v1.0.0
```

### 后续更新

每次新版本只需要：

1. 改代码
2. 跑测试
3. 打新 tag，例如 `v0.1.1`

## 8. Swift：`ason-swift`

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

### 后续更新

```bash
git tag 0.1.1
git push origin 0.1.1
```

## 9. Zig：`ason-zig`

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
cd ason-zig
mise exec -- zig build test
```

3. 打 tag：

```bash
git tag 1.0.0
git push origin 1.0.0
```

4. 建 GitHub Release，附带源码归档或二进制产物。

### 后续更新

继续用新 tag + GitHub Release。

## 10. C：`ason-c`

推荐渠道：GitHub Releases  
官方参考：

- GitHub Releases：https://docs.github.com/en/repositories/releasing-projects-on-github/managing-releases-in-a-repository

### 是否需要单独注册

不需要额外注册，只需要 GitHub。

### 首次发布

```bash
cd ason-c
cmake -S . -B build
cmake --build build
ctest --test-dir build --output-on-failure
```

然后：

1. 打 tag
2. 建 GitHub Release
3. 上传源码包，必要时上传预编译静态库/示例二进制

### 后续更新

继续 tag + Release。

### 备注

如果以后想进一步进入系统包管理生态，可以再补：

- vcpkg
- Conan
- Homebrew

但这不是首发必需。

## 11. C++：`ason-cpp`

推荐渠道：

1. GitHub Release
2. Conan
3. vcpkg
4. Homebrew

说明：

- `ason-cpp` 当前是标准 header-only CMake package
- Conan recipe 已就绪
- vcpkg 当前是 overlay port 级别
- Homebrew 当前是 formula 模板级别，发布前需要填真实 tarball 的 `sha256`

### 首次发布前验证

```bash
cd ason-cpp
cmake -S . -B build -DCMAKE_BUILD_TYPE=Release
cmake --build build
ctest --test-dir build --output-on-failure
```

建议再额外验证一次标准 CMake 包导出是否可用：

```bash
cmake --install build --prefix /tmp/ason-cpp-install
```

### 11.1 GitHub Release

这是最基础、也最先完成的一步。

流程：

1. 更新版本号（当前在 `CMakeLists.txt`、`conanfile.py`、`vcpkg/ports/ason-cpp/vcpkg.json`、`homebrew/ason-cpp.rb` 中都要一致）
2. 跑完测试和本地安装验证
3. 打 tag，例如 `v1.0.0`
4. 在 GitHub 上创建 Release
5. 上传源码归档，或直接使用 GitHub 自动生成的 source tarball / zip

### 11.2 Conan

仓库内已有：

- `ason-cpp/conanfile.py`
- `ason-cpp/test_package/`

本地验证：

```bash
cd ason-cpp
conan create . --build=missing
```

发布方式有两种：

1. 先用仓库内 recipe 做本地或私有远端发布
2. 后续再向 ConanCenter 提交 recipe

如果只是你自己的 first-party 发布，最简单的是：

```bash
cd ason-cpp
conan export . --version=1.0.0
```

然后在自己的 Conan remote 上上传对应 recipe/package。

### 11.3 vcpkg

仓库内已有 overlay port：

- `ason-cpp/vcpkg/ports/ason-cpp/portfile.cmake`
- `ason-cpp/vcpkg/ports/ason-cpp/vcpkg.json`
- `ason-cpp/vcpkg/ports/ason-cpp/vcpkg-cmake-wrapper.cmake`

当前定位：

- 这套文件已经足够用于 overlay port
- 如果要进入官方 vcpkg registry，还需要单独向 `microsoft/vcpkg` 提交 port PR

本地验证：

```bash
vcpkg install ason-cpp --overlay-ports=/path/to/ason-cpp/vcpkg/ports
```

推荐发布步骤：

1. 先完成 GitHub Release
2. 确认 release tag、源码 tarball 稳定
3. 基于当前 port 文件向 vcpkg 官方仓库提 PR

### 11.4 Homebrew

仓库内已有 formula 模板：

- `ason-cpp/homebrew/ason-cpp.rb`

注意：

- 当前 formula 里的 `sha256` 还是占位符
- 真正发布前，需要把它替换成 GitHub Release tarball 的真实 `sha256`

计算方式示例：

```bash
curl -L -o /tmp/ason-v1.0.0.tar.gz \
  https://github.com/ason-lab/ason/archive/refs/tags/v1.0.0.tar.gz
shasum -a 256 /tmp/ason-v1.0.0.tar.gz
```

然后把得到的 hash 填入：

- `ason-cpp/homebrew/ason-cpp.rb`

本地验证可用：

```bash
brew install --build-from-source ./homebrew/ason-cpp.rb
brew test ason-cpp
```

如果要正式对外发布，通常有两种方式：

1. 自己维护一个 tap
2. 后续尝试提交到 Homebrew core（通常要求更严格）

### 后续更新

每次发新版本时，建议按这个顺序推进：

1. 更新 `CMakeLists.txt`
2. 更新 `conanfile.py`
3. 更新 `vcpkg/ports/ason-cpp/vcpkg.json`
4. 更新 `homebrew/ason-cpp.rb` 中的版本、下载地址与 `sha256`
5. 跑测试和打包验证
6. 打 Git tag
7. 创建 GitHub Release
8. 再分别同步 Conan / vcpkg / Homebrew

### 备注

- Conan：现在已经是可用级别
- vcpkg：现在是 overlay port 可用级别
- Homebrew：现在是 formula 模板可用级别
- 所以 `ason-cpp` 目前已经不只是“GitHub Release 可发”，而是已经具备继续进入包管理器生态的基础文件

## 12. PHP 扩展：`ason-php`

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

因此，对 `ason-php` 这个新扩展，更实际的首发方式是：

1. GitHub Release
2. 提供源码包
3. README 里给清楚的 `phpize / ./configure / make / make install` 安装说明

### 首次发布

```bash
cd ason-php
phpize
./configure
make -j2
make test
```

然后：

1. 打 tag
2. 建 GitHub Release
3. 上传源码归档
4. 在 Release 说明中写清：
   - 支持的 PHP 版本
   - 编译依赖
   - 安装命令

### 后续更新

继续用 Git tag + GitHub Release。

如果未来 PIE 生态对第三方扩展的发布路径稳定下来，再补自动化发布。

## GitHub Release 的建议做法

适用于 `ason-go`、`ason-swift`、`ason-zig`、`ason-c`、`ason-cpp`、`ason-php`，以及其它语言的辅助发布记录。

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

1. `ason-js` -> npm
2. `ason-cs` -> NuGet
3. `ason-dart` -> pub.dev
4. `ason-rs` -> crates.io
5. `ason-java` -> Maven Central
6. `ason-go` -> Git tag + GitHub Release
7. `ason-swift` -> Git tag + GitHub Release
8. `ason-zig` -> Git tag + GitHub Release
9. `ason-c` -> GitHub Release
10. `ason-cpp` -> GitHub Release
11. `ason-php` -> GitHub Release
12. `ason-py` -> PyPI（前提是 wheel 安装问题彻底修好）

## 维护建议

建议后续统一做三件事：

1. 所有语言版本号尽量保持一致。
2. 所有仓库都保留 Git tag 和 GitHub Release。
3. 后面再补自动发布 CI，不要一开始就同时手动 + 自动两套混着来。

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
- Zig package management note: https://ziglang.org/download/0.11.0/release-notes.html
- PECL / PIE notice: https://pecl.php.net/account-request.php
