#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

dry_run=false
force_push=false
push_all_tags=false
targets=()
requested_tags=()

usage() {
  cat <<'EOF'
Usage:
  scripts/push-tags.sh [options]

Push git tags for selected submodule repositories.

Options:
  --target NAME          Only process one repo or a comma-separated subset.
                         Examples: rs, rust, java, go, py, python, lsp,
                         vscode, zed, website, all
  --tag TAG              Push a specific tag. Repeatable and also accepts comma-separated tags.
                         Examples: --tag v1.0.0 --tag v1.0.1 or --tag v1.0.0,v1.0.1
  --all-tags             Push all local tags in each selected repo
  --force                Force-push tags when the same tag already exists on origin
  --dry-run              Print what would happen without pushing
  -h, --help             Show this help

Examples:
  scripts/push-tags.sh --target rs --tag v1.0.0
  scripts/push-tags.sh --target rs,java --tag v1.0.0,v1.0.1
  scripts/push-tags.sh --target website --all-tags
  scripts/push-tags.sh --target all --tag v1.0.0 --dry-run
EOF
}

log() {
  printf '%s\n' "$*"
}

run_cmd() {
  if $dry_run; then
    printf '[dry-run] '
    printf '%q ' "$@"
    printf '\n'
  else
    "$@"
  fi
}

normalize_tag() {
  local raw="$1"
  if [[ "$raw" == v* ]]; then
    printf '%s\n' "$raw"
  else
    printf 'v%s\n' "$raw"
  fi
}

resolve_target() {
  local raw="${1,,}"
  local submodule="$2"
  local base short

  base="$(basename "$submodule")"
  short="${base#asun-}"

  case "$raw" in
    all) return 0 ;;
    "$submodule"|"$base"|"$short") return 0 ;;
    rust|rs) [[ "$base" == "asun-rs" ]] && return 0 ;;
    python|py) [[ "$base" == "asun-py" ]] && return 0 ;;
    javascript|typescript|js|ts) [[ "$base" == "asun-js" ]] && return 0 ;;
    cpp|cxx) [[ "$base" == "asun-cpp" ]] && return 0 ;;
    csharp|cs|dotnet) [[ "$base" == "asun-cs" ]] && return 0 ;;
    java|jvm) [[ "$base" == "asun-java" ]] && return 0 ;;
    golang|go) [[ "$base" == "asun-go" ]] && return 0 ;;
    dart) [[ "$base" == "asun-dart" ]] && return 0 ;;
    php) [[ "$base" == "asun-php" ]] && return 0 ;;
    swift) [[ "$base" == "asun-swift" ]] && return 0 ;;
    zig) [[ "$base" == "asun-zig" ]] && return 0 ;;
    c) [[ "$base" == "asun-c" ]] && return 0 ;;
    lsp|lsp-asun) [[ "$base" == "lsp-asun" ]] && return 0 ;;
    vscode|plugin_vscode) [[ "$base" == "plugin_vscode" ]] && return 0 ;;
    zed|plugin_zed) [[ "$base" == "plugin_zed" ]] && return 0 ;;
    jetbrains|jetbrain|plugin_jetbrain) [[ "$base" == "plugin_jetbrain" ]] && return 0 ;;
    js-format|formatter|format) [[ "$base" == "js-format" ]] && return 0 ;;
    website) [[ "$base" == "website" ]] && return 0 ;;
  esac

  return 1
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --target)
      shift
      [[ $# -gt 0 ]] || { log "Missing value for --target"; exit 1; }
      IFS=',' read -r -a raw_targets <<<"$1"
      for raw_target in "${raw_targets[@]}"; do
        targets+=("$raw_target")
      done
      ;;
    --tag)
      shift
      [[ $# -gt 0 ]] || { log "Missing value for --tag"; exit 1; }
      IFS=',' read -r -a raw_tag_values <<<"$1"
      for raw_tag in "${raw_tag_values[@]}"; do
        requested_tags+=("$(normalize_tag "$raw_tag")")
      done
      ;;
    --all-tags)
      push_all_tags=true
      ;;
    --force)
      force_push=true
      ;;
    --dry-run)
      dry_run=true
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      log "Unknown option: $1"
      usage
      exit 1
      ;;
  esac
  shift
done

if ! $push_all_tags && [[ ${#requested_tags[@]} -eq 0 ]]; then
  log "You must provide either --tag or --all-tags."
  usage
  exit 1
fi

cd "$ROOT_DIR"

submodules=()
while IFS= read -r line; do
  submodules+=("$line")
done < <(git config --file .gitmodules --get-regexp '^submodule\..*\.path$' | awk '{print $2}')

if [[ ${#submodules[@]} -eq 0 ]]; then
  log "No submodules found."
  exit 0
fi

selected_submodules=()
if [[ ${#targets[@]} -eq 0 ]]; then
  selected_submodules=("${submodules[@]}")
else
  for submodule in "${submodules[@]}"; do
    for target in "${targets[@]}"; do
      if resolve_target "$target" "$submodule"; then
        selected_submodules+=("$submodule")
        break
      fi
    done
  done
fi

if [[ ${#selected_submodules[@]} -eq 0 ]]; then
  log "No matching submodules for the requested target(s)."
  exit 1
fi

pushed=()
unchanged=()
missing=()
warnings=()

for submodule in "${selected_submodules[@]}"; do
  if [[ ! -d "$submodule" ]]; then
    missing+=("$submodule")
    continue
  fi

  log "== $submodule =="

  repo_tags=()
  if $push_all_tags; then
    while IFS= read -r tag; do
      [[ -n "$tag" ]] && repo_tags+=("$tag")
    done < <(git -C "$submodule" tag --list)
  else
    repo_tags=("${requested_tags[@]}")
  fi

  if [[ ${#repo_tags[@]} -eq 0 ]]; then
    warnings+=("$submodule: no local tags to push")
    log "No local tags to push"
    continue
  fi

  for tag in "${repo_tags[@]}"; do
    if ! git -C "$submodule" rev-parse -q --verify "refs/tags/$tag" >/dev/null 2>&1; then
      warnings+=("$submodule: missing local tag $tag")
      log "Skipping missing local tag: $tag"
      continue
    fi

    if git -C "$submodule" ls-remote --exit-code --tags origin "refs/tags/$tag" >/dev/null 2>&1; then
      if $force_push; then
        log "Force-pushing tag $tag"
        run_cmd git -C "$submodule" push --force origin "refs/tags/$tag"
        pushed+=("$submodule@$tag (forced)")
      else
        log "No push needed: origin already has tag $tag"
        unchanged+=("$submodule@$tag")
      fi
    else
      log "Pushing tag $tag"
      run_cmd git -C "$submodule" push origin "refs/tags/$tag"
      pushed+=("$submodule@$tag")
    fi
  done
done

log
log "Summary"
if [[ ${#pushed[@]} -gt 0 ]]; then
  printf '  pushed: %s\n' "${pushed[@]}"
fi
if [[ ${#unchanged[@]} -gt 0 ]]; then
  printf '  unchanged: %s\n' "${unchanged[@]}"
fi
if [[ ${#missing[@]} -gt 0 ]]; then
  printf '  missing: %s\n' "${missing[@]}"
fi
if [[ ${#warnings[@]} -gt 0 ]]; then
  printf '  warning: %s\n' "${warnings[@]}"
fi
