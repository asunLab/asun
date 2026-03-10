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
  scripts/push-submodules.sh [options]

Fetch and sync each submodule to its origin default branch first, then push the
submodule's current HEAD back to that branch if it is ahead. Optionally
stage/commit/push the parent repository's updated submodule pointers.

Options:
  --stage-parent            Run `git add` for changed submodule pointers in the parent repo
  --commit-parent           Create a parent commit after staging submodule pointers
  --commit-message MESSAGE  Commit message for `--commit-parent`
  --push-parent             Push the parent repo after committing
  --dry-run                 Print what would happen without pushing or committing
  -h, --help                Show this help

Examples:
  scripts/push-submodules.sh
  scripts/push-submodules.sh --stage-parent --commit-parent
  scripts/push-submodules.sh --stage-parent --commit-parent --push-parent
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

mapfile -t submodules < <(git config --file .gitmodules --get-regexp '^submodule\..*\.path$' | awk '{print $2}')

if [[ ${#submodules[@]} -eq 0 ]]; then
  log "No submodules found."
  exit 0
fi

pushed=()
updated=()
unchanged=()
missing=()
warnings=()

for submodule in "${submodules[@]}"; do
  if [[ ! -d "$submodule" ]]; then
    missing+=("$submodule")
    continue
  fi

  log "== $submodule =="

  if ! git -C "$submodule" rev-parse --verify HEAD >/dev/null 2>&1; then
    warnings+=("$submodule: missing HEAD")
    log "Skipping: no HEAD"
    continue
  fi

  head_sha="$(git -C "$submodule" rev-parse --short HEAD)"
  remote_ref="$(git -C "$submodule" symbolic-ref --short -q refs/remotes/origin/HEAD || true)"
  if [[ -z "$remote_ref" ]]; then
    remote_branch="main"
    log "origin/HEAD not set, defaulting to origin/$remote_branch"
  else
    remote_branch="${remote_ref#origin/}"
  fi

  if [[ -n "$(git -C "$submodule" status --short)" ]]; then
    warnings+=("$submodule: has uncommitted changes; skipped sync/push")
    log "Warning: worktree has uncommitted changes; skipping"
    continue
  fi

  run_cmd git -C "$submodule" fetch origin "$remote_branch" --quiet

  remote_exists=false
  if git -C "$submodule" show-ref --verify --quiet "refs/remotes/origin/$remote_branch"; then
    remote_exists=true
  fi

  ahead_count=0
  behind_count=0
  if $remote_exists; then
    ahead_count="$(git -C "$submodule" rev-list --count "origin/$remote_branch..HEAD")"
    behind_count="$(git -C "$submodule" rev-list --count "HEAD..origin/$remote_branch")"
  else
    ahead_count=1
  fi

  if $remote_exists && [[ "$behind_count" != "0" ]]; then
    current_branch="$(git -C "$submodule" branch --show-current)"
    if [[ -n "$current_branch" ]]; then
      log "Rebasing $current_branch onto origin/$remote_branch"
      run_cmd git -C "$submodule" pull --rebase origin "$remote_branch"
    elif [[ "$ahead_count" == "0" ]]; then
      log "Updating detached HEAD to origin/$remote_branch"
      run_cmd git -C "$submodule" checkout --detach "origin/$remote_branch"
    else
      log "Rebasing detached HEAD onto origin/$remote_branch"
      run_cmd git -C "$submodule" rebase "origin/$remote_branch"
    fi

    head_sha="$(git -C "$submodule" rev-parse --short HEAD)"
    updated+=("$submodule@$head_sha")

    ahead_count="$(git -C "$submodule" rev-list --count "origin/$remote_branch..HEAD")"
    behind_count="$(git -C "$submodule" rev-list --count "HEAD..origin/$remote_branch")"
  fi

  if [[ "$behind_count" != "0" ]]; then
    warnings+=("$submodule: still behind origin/$remote_branch after sync")
    log "Warning: still behind origin/$remote_branch after sync; skipping push"
    continue
  fi

  if [[ "$ahead_count" == "0" ]]; then
    unchanged+=("$submodule@$head_sha")
    log "No push needed: HEAD already reachable from origin/$remote_branch"
    continue
  fi

  log "Pushing $head_sha -> origin/$remote_branch"
  run_cmd git -C "$submodule" push origin "HEAD:refs/heads/$remote_branch"
  pushed+=("$submodule@$head_sha -> origin/$remote_branch")
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
        if $dry_run; then
          run_cmd git push origin HEAD
        else
          git push origin HEAD
        fi
      fi
    fi
  else
    log "No parent submodule pointer changes to stage"
  fi
fi

log
log "Summary"
if [[ ${#pushed[@]} -gt 0 ]]; then
  printf '  pushed: %s\n' "${pushed[@]}"
fi
if [[ ${#updated[@]} -gt 0 ]]; then
  printf '  updated: %s\n' "${updated[@]}"
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
