#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

pids=()
modules=(
  website lsp-asun
  plugin_vscode plugin_zed plugin_jetbrain
  asun-c asun-cpp asun-cs asun-dart asun-go
  asun-java asun-js asun-php asun-py asun-rs asun-swift asun-zig
)

switchMain() (
    git push --set-upstream origin main
    cd "$ROOT_DIR/$1" && git switch main && git pull --autostash
)

modules=(
  website lsp-asun
  plugin_vscode plugin_zed plugin_jetbrain
  asun-c asun-cpp asun-cs asun-dart asun-go
  asun-java asun-js asun-php asun-py asun-rs asun-swift asun-zig
)

for mod in "${modules[@]}"; do
  switchMain "$mod" "${1:-}" &
  pids+=($!)
done

failed=0
for pid in "${pids[@]}"; do
  wait "$pid" || ((failed++))
done

if ((failed > 0)); then
  echo "$failed module(s) failed to push" >&2
  exit 1
fi