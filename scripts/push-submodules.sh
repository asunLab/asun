#!/usr/bin/env bash

#set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"


pushGit() (
  cd "$ROOT_DIR/$1"
  if [[ "${2:-}" == "-f" ]]; then
     git push -f origin HEAD:main
     return
  fi
  git pull --rebase --autostash
  git push origin HEAD:main
)

pids=()
modules=(
  website lsp-asun js-format
  plugin_vscode plugin_zed plugin_jetbrain
  asun-c asun-cpp asun-cs asun-dart asun-go
  asun-java asun-js asun-php asun-py asun-rs asun-swift asun-zig
)

for mod in "${modules[@]}"; do
  pushGit "$mod" "${1:-}" &
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