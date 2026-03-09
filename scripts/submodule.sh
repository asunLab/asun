#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

remote_base="https://github.com/ason-lab"
target_branch="main"
migration_branch="chore/submodules"
dry_run=false
assume_yes=false

usage() {
  cat <<'EOF'
Usage:
  scripts/submodule.sh [options] DIR [DIR...]
  scripts/submodule.sh [options] PATH:REPO [PATH:REPO ...]

Convert one or more tracked directories in the current monorepo into git
submodules while preserving each directory's history via `git subtree split`.

Each target can be written in one of two forms:
  DIR
      Use the same name for local path and remote repo.
  PATH:REPO
      Use PATH as the directory in the current repo and REPO as the remote repo
      name under the GitHub org / remote base.

What this script does:
  1. `git subtree split --prefix=<path>` for each target
  2. push each split branch to <remote-base>/<repo>.git:<target-branch>
  3. create or switch to the migration branch
  4. `git rm -r` the selected directories
  5. `git submodule add` them back from their new repos
  6. create two parent-repo commits

Options:
  --remote-base URL        Remote base URL (default: git@github.com:ason-lab)
  --target-branch NAME     Target branch to push to (default: main)
  --migration-branch NAME  Parent repo branch used for the migration
                           (default: chore/submodules)
  --dry-run                Print what would happen without changing anything
  -y, --yes                Skip the confirmation prompt
  -h, --help               Show this help

Examples:
  scripts/submodule.sh ason-go ason-rs
  scripts/submodule.sh plugin_vscode:ason-vscode lsp-ason
  scripts/submodule.sh --dry-run ason-go ason-rs lsp-ason
  scripts/submodule.sh --remote-base git@github.com:my-org ason-go

Notes:
  - Run this from a clean parent repository.
  - The target repos should already exist on the remote.
  - Commit history is preserved for each selected path, but commit SHAs will
    change in the split repos because history is rewritten to the new root.
EOF
}

log() {
  printf '%s\n' "$*"
}

fail() {
  log "Error: $*"
  exit 1
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

require_clean_repo() {
  local status
  status="$(git status --short)"
  if [[ -n "$status" ]]; then
    log "Error: working tree is not clean."
    log "Commit or stash changes before running this script."
    log
    log "Changed files:"
    printf '%s\n' "$status"
    log
    log "Hints:"
    log "  - inspect changes: git status --short"
    log "  - commit changes:  git add -A && git commit -m \"...\""
    log "  - stash changes:   git stash push -u"
    exit 1
  fi
}

confirm() {
  if $assume_yes || $dry_run; then
    return 0
  fi

  printf 'Continue? [y/N] '
  read -r answer
  case "$answer" in
    y|Y|yes|YES)
      ;;
    *)
      fail "aborted by user"
      ;;
  esac
}

declare -a raw_targets=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --remote-base)
      shift
      [[ $# -gt 0 ]] || fail "missing value for --remote-base"
      remote_base="$1"
      ;;
    --target-branch)
      shift
      [[ $# -gt 0 ]] || fail "missing value for --target-branch"
      target_branch="$1"
      ;;
    --migration-branch)
      shift
      [[ $# -gt 0 ]] || fail "missing value for --migration-branch"
      migration_branch="$1"
      ;;
    --dry-run)
      dry_run=true
      ;;
    -y|--yes)
      assume_yes=true
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    -*)
      fail "unknown option: $1"
      ;;
    *)
      raw_targets+=("$1")
      ;;
  esac
  shift
done

[[ ${#raw_targets[@]} -gt 0 ]] || {
  usage
  exit 1
}

cd "$ROOT_DIR"

if ! git rev-parse --show-toplevel >/dev/null 2>&1; then
  fail "not inside a git repository: $ROOT_DIR"
fi

repo_root="$(git rev-parse --show-toplevel)"
cd "$repo_root"

require_clean_repo

declare -a target_paths=()
declare -a target_repos=()

for target in "${raw_targets[@]}"; do
  if [[ "$target" == *:* ]]; then
    path="${target%%:*}"
    repo="${target#*:}"
  else
    path="$target"
    repo="$target"
  fi

  [[ -n "$path" ]] || fail "empty path in target: $target"
  [[ -n "$repo" ]] || fail "empty repo name in target: $target"
  [[ -d "$path" ]] || fail "path does not exist or is not a directory: $path"

  if ! git ls-files --error-unmatch "$path" >/dev/null 2>&1; then
    fail "path is not tracked by git: $path"
  fi

  target_paths+=("$path")
  target_repos+=("$repo")
done

log "Submodule migration plan"
log "  repo root: $repo_root"
log "  remote base: $remote_base"
log "  target branch: $target_branch"
log "  migration branch: $migration_branch"
log "  dry-run: $dry_run"
log

for i in "${!target_paths[@]}"; do
  path="${target_paths[$i]}"
  repo="${target_repos[$i]}"
  log "  - $path -> ${remote_base}/${repo}.git"
done

log
log "This will:"
log "  1. split and push each target repo"
log "  2. create/switch to $migration_branch"
log "  3. remove the in-tree directories"
log "  4. add them back as submodules"
log "  5. create two parent-repo commits"
log

confirm

for i in "${!target_paths[@]}"; do
  path="${target_paths[$i]}"
  repo="${target_repos[$i]}"
  split_branch="split/${repo}"
  remote_url="${remote_base}/${repo}.git"

  log "== Splitting $path -> $remote_url =="

  if git show-ref --verify --quiet "refs/heads/$split_branch"; then
    run_cmd git branch -D "$split_branch"
  fi

  run_cmd git subtree split --prefix="$path" -b "$split_branch"
  run_cmd git push "$remote_url" "${split_branch}:${target_branch}"
done

if git show-ref --verify --quiet "refs/heads/$migration_branch"; then
  log "== Switching to existing branch $migration_branch =="
  run_cmd git checkout "$migration_branch"
else
  log "== Creating branch $migration_branch =="
  run_cmd git checkout -b "$migration_branch"
fi

log "== Removing embedded directories =="
run_cmd git rm -r "${target_paths[@]}"
run_cmd git commit -m "chore: remove embedded projects before submodule migration"

log "== Adding submodules =="
for i in "${!target_paths[@]}"; do
  path="${target_paths[$i]}"
  repo="${target_repos[$i]}"
  remote_url="${remote_base}/${repo}.git"
  run_cmd git submodule add "$remote_url" "$path"
done

run_cmd git commit -m "chore: add language repos as submodules"

log
log "Done."
log "Next steps:"
log "  - review .gitmodules and submodule SHAs"
log "  - run: git status"
log "  - push the migration branch when ready:"
log "      git push origin $migration_branch"
