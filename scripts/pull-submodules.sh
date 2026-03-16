#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

stage_parent=false
commit_parent=false
push_parent=false
dry_run=false
commit_message="chore: update submodule pointers"

usage() {
  cat <<'EOF'
Usage:
  scripts/pull-submodules.sh [options]

Sync and initialize submodules, then rebase each submodule onto its
origin default branch when possible. Optionally stage/commit/push the parent
repository's updated submodule pointers.

Options:
  --stage-parent            Run `git add` for changed submodule pointers in the parent repo
  --commit-parent           Create a parent commit after staging submodule pointers
  --commit-message MESSAGE  Commit message for `--commit-parent`
  --push-parent             Push the parent repo after committing
  --dry-run                 Print what would happen without changing git state
  -h, --help                Show this help

Examples:
  scripts/pull-submodules.sh
  scripts/pull-submodules.sh --stage-parent --commit-parent
  scripts/pull-submodules.sh --stage-parent --commit-parent --push-parent
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

while [[ $# -gt 0 ]]; do
  case "$1" in
    --stage-parent)
      stage_parent=true
      ;;
    --commit-parent)
      stage_parent=true
      commit_parent=true
      ;;
    --commit-message)
      shift
      if [[ $# -eq 0 ]]; then
        log "Missing value for --commit-message"
        exit 1
      fi
      commit_message="$1"
      ;;
    --push-parent)
      stage_parent=true
      commit_parent=true
      push_parent=true
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

cd "$ROOT_DIR"

if ! git rev-parse --show-toplevel >/dev/null 2>&1; then
  log "Not inside a git repository: $ROOT_DIR"
  exit 1
fi

if [[ ! -f .gitmodules ]]; then
  log "No .gitmodules file found."
  exit 0
fi

submodules=()
while IFS= read -r line; do
  submodules+=("$line")
done < <(git config --file .gitmodules --get-regexp '^submodule\..*\.path$' | awk '{print $2}')

if [[ ${#submodules[@]} -eq 0 ]]; then
  log "No submodules found."
  exit 0
fi

log "Syncing submodule metadata"
run_cmd git submodule sync --recursive
log "Initializing missing submodules"
run_cmd git submodule update --init --recursive

updated=()
unchanged=()
warnings=()

for submodule in "${submodules[@]}"; do
  if [[ ! -d "$submodule" ]]; then
    warnings+=("$submodule: directory missing after init")
    log "== $submodule =="
    log "Warning: directory missing after init"
    continue
  fi

  log "== $submodule =="

  if [[ -n "$(git -C "$submodule" status --short)" ]]; then
    warnings+=("$submodule: has uncommitted changes; skipped")
    log "Warning: worktree has uncommitted changes; skipping"
    continue
  fi

  remote_ref="$(git -C "$submodule" symbolic-ref --short -q refs/remotes/origin/HEAD || true)"
  if [[ -z "$remote_ref" ]]; then
    remote_branch="main"
    log "origin/HEAD not set, defaulting to origin/$remote_branch"
  else
    remote_branch="${remote_ref#origin/}"
  fi

  current_branch="$(git -C "$submodule" branch --show-current)"
  before_sha="$(git -C "$submodule" rev-parse --short HEAD)"

  run_cmd git -C "$submodule" fetch origin "$remote_branch" --quiet

  if ! git -C "$submodule" show-ref --verify --quiet "refs/remotes/origin/$remote_branch"; then
    warnings+=("$submodule: missing origin/$remote_branch")
    log "Warning: origin/$remote_branch not found; skipping"
    continue
  fi

  behind_count="$(git -C "$submodule" rev-list --count "HEAD..origin/$remote_branch")"
  ahead_count="$(git -C "$submodule" rev-list --count "origin/$remote_branch..HEAD")"

  if [[ "$behind_count" == "0" && "$ahead_count" == "0" ]]; then
    unchanged+=("$submodule@$before_sha")
    log "Already up to date at $before_sha"
    continue
  fi

  if [[ "$ahead_count" != "0" ]]; then
    warnings+=("$submodule: local commits ahead of origin/$remote_branch; skipped")
    log "Warning: local HEAD is ahead of origin/$remote_branch; skipping"
    continue
  fi

  if [[ -n "$current_branch" ]]; then
    log "Pulling $current_branch from origin/$remote_branch with rebase+autostash"
    run_cmd git -C "$submodule" pull --rebase --autostash origin "$remote_branch"
  else
    log "Updating detached HEAD to origin/$remote_branch"
    run_cmd git -C "$submodule" checkout --detach "origin/$remote_branch"
  fi

  after_sha="$(git -C "$submodule" rev-parse --short HEAD)"
  updated+=("$submodule@$before_sha->$after_sha")
done

if $stage_parent; then
  changed_submodules=()
  for submodule in "${submodules[@]}"; do
    if [[ -n "$(git status --short -- "$submodule")" ]]; then
      changed_submodules+=("$submodule")
    fi
  done

  if [[ ${#changed_submodules[@]} -gt 0 ]]; then
    log "Staging updated submodule pointers in parent repo"
    run_cmd git add "${changed_submodules[@]}"

    if $commit_parent; then
      if $dry_run; then
        run_cmd git commit -m "$commit_message"
      elif git diff --cached --quiet; then
        log "No staged parent changes to commit"
      else
        git commit -m "$commit_message"
      fi

      if $push_parent; then
        run_cmd git push origin HEAD
      fi
    fi
  else
    log "No parent submodule pointer changes to stage"
  fi
fi

log
log "Summary"
if [[ ${#updated[@]} -gt 0 ]]; then
  printf '  updated: %s\n' "${updated[@]}"
fi
if [[ ${#unchanged[@]} -gt 0 ]]; then
  printf '  unchanged: %s\n' "${unchanged[@]}"
fi
if [[ ${#warnings[@]} -gt 0 ]]; then
  printf '  warning: %s\n' "${warnings[@]}"
fi
