#!/usr/bin/env bash
# ──────────────────────────────────────────────────────────────────────────────
# ASON VS Code 插件 — 跨平台构建脚本
#
# 用法:
#   ./scripts/build.sh              # 构建所有平台
#   ./scripts/build.sh current      # 仅构建当前平台
#   ./scripts/build.sh darwin-arm64  # 构建指定平台
# ──────────────────────────────────────────────────────────────────────────────
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
PLUGIN_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"
LSP_DIR="$(cd "$PLUGIN_DIR/../ason-lsp" && pwd)"
SERVER_DIR="$PLUGIN_DIR/server"

# ── 平台定义 ──────────────────────────────────────────────────────────────────
# 格式: "vscode-target:GOOS:GOARCH:binary-suffix"
ALL_TARGETS=(
    "darwin-arm64:darwin:arm64:"
    "darwin-x64:darwin:amd64:"
    "linux-x64:linux:amd64:"
    "linux-arm64:linux:arm64:"
    "win32-x64:windows:amd64:.exe"
    "win32-arm64:windows:arm64:.exe"
)

# ── 辅助函数 ──────────────────────────────────────────────────────────────────

log() { echo -e "\033[1;34m→\033[0m $*"; }
ok()  { echo -e "\033[1;32m✓\033[0m $*"; }
err() { echo -e "\033[1;31m✗\033[0m $*" >&2; }

detect_current_target() {
    local os arch
    os="$(uname -s | tr '[:upper:]' '[:lower:]')"
    arch="$(uname -m)"

    case "$os" in
        darwin) os="darwin" ;;
        linux)  os="linux"  ;;
        *)      err "Unsupported OS: $os"; exit 1 ;;
    esac

    case "$arch" in
        x86_64|amd64) arch="x64" ;;
        arm64|aarch64) arch="arm64" ;;
        *)             err "Unsupported arch: $arch"; exit 1 ;;
    esac

    echo "${os}-${arch}"
}

# ── 编译 ason-lsp ─────────────────────────────────────────────────────────────

build_lsp() {
    local goos="$1" goarch="$2" suffix="$3" output

    output="$SERVER_DIR/ason-lsp${suffix}"
    log "Compiling ason-lsp for ${goos}/${goarch} ..."

    mkdir -p "$SERVER_DIR"

    (
        cd "$LSP_DIR"
        GOOS="$goos" GOARCH="$goarch" CGO_ENABLED=0 \
            go build -trimpath -ldflags="-s -w" -o "$output" .
    )

    ok "Built: $output ($(du -h "$output" | awk '{print $1}'))"
}

# ── 打包 vsix ─────────────────────────────────────────────────────────────────

package_vsix() {
    local target="$1"

    log "Packaging VSIX for target: ${target} ..."
    (
        cd "$PLUGIN_DIR"
        npx vsce package --target "$target" --no-git-tag-version
    )
    ok "Packaged: ${target}"
}

# ── 编译 TypeScript ──────────────────────────────────────────────────────────

compile_ts() {
    log "Installing npm dependencies ..."
    (cd "$PLUGIN_DIR" && npm install --silent)
    log "Compiling TypeScript ..."
    (cd "$PLUGIN_DIR" && npm run compile)
    ok "TypeScript compiled."
}

# ── 清理 ──────────────────────────────────────────────────────────────────────

clean_server() {
    rm -rf "$SERVER_DIR"
}

# ── 主流程 ────────────────────────────────────────────────────────────────────

main() {
    local mode="${1:-all}"
    local built=0

    echo ""
    echo "╔═══════════════════════════════════════════════════╗"
    echo "║    ASON VS Code Extension — Build Script         ║"
    echo "╚═══════════════════════════════════════════════════╝"
    echo ""

    # 编译 TypeScript（只需一次）
    compile_ts

    if [[ "$mode" == "current" ]]; then
        # 仅构建当前平台
        local current
        current="$(detect_current_target)"
        log "Building for current platform: ${current}"

        for entry in "${ALL_TARGETS[@]}"; do
            IFS=':' read -r target goos goarch suffix <<< "$entry"
            if [[ "$target" == "$current" ]]; then
                clean_server
                build_lsp "$goos" "$goarch" "$suffix"
                package_vsix "$target"
                built=1
                break
            fi
        done

        if [[ $built -eq 0 ]]; then
            err "No matching target for: ${current}"
            exit 1
        fi

    elif [[ "$mode" == "all" ]]; then
        # 构建所有平台
        for entry in "${ALL_TARGETS[@]}"; do
            IFS=':' read -r target goos goarch suffix <<< "$entry"
            clean_server
            build_lsp "$goos" "$goarch" "$suffix"
            package_vsix "$target"
            built=$((built + 1))
        done

    else
        # 构建指定平台
        for entry in "${ALL_TARGETS[@]}"; do
            IFS=':' read -r target goos goarch suffix <<< "$entry"
            if [[ "$target" == "$mode" ]]; then
                clean_server
                build_lsp "$goos" "$goarch" "$suffix"
                package_vsix "$target"
                built=1
                break
            fi
        done

        if [[ $built -eq 0 ]]; then
            err "Unknown target: ${mode}"
            echo "Available targets:"
            for entry in "${ALL_TARGETS[@]}"; do
                IFS=':' read -r target _ _ _ <<< "$entry"
                echo "  - $target"
            done
            exit 1
        fi
    fi

    # 清理最后一次的 server/ 目录
    clean_server

    echo ""
    ok "Done! Generated .vsix files:"
    ls -lh "$PLUGIN_DIR"/*.vsix 2>/dev/null || echo "  (none found)"
    echo ""
}

main "$@"
