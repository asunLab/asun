#!/usr/bin/env bash

set -euo pipefail

REMOTE_BASE="https://github.com/ason-lab"
TARGET_BRANCH="main"
DRY_RUN=false
ASSUME_YES=false

usage() {
  cat <<'EOF'
Usage:
  scripts/submodule.sh [options] DIR
  scripts/submodule.sh [options] DIR REPO

Split one tracked directory out of the current repository, push its history to
its own repo, then replace the directory with a git submodule.

Arguments:
  DIR        Directory in the current repo, for example: ason-swift
  REPO       Remote repo name. Defaults to DIR.

Options:
  --remote-base URL   Remote base URL (default: https://github.com/ason-lab)
  --branch NAME       Branch to push to in the child repo (default: main)
  --dry-run           Print commands without executing them
  -y, --yes           Skip confirmation prompt
  -h, --help          Show this help

Examples:
  scripts/submodule.sh ason-swift
  scripts/submodule.sh plugin_vscode ason-vscode
  scripts/submodule.sh --dry-run ason-swift

What this script does:
  1. git subtree split --prefix=<DIR>
  2. git push <remote-base>/<REPO>.git split/<REPO>:<branch>
  3. git rm -r <DIR>
  4. git commit
  5. git submodule add <remote-base>/<REPO>.git <DIR>
  6. git commit

Important:
  - Run this from a clean git worktree.
  - Create the target GitHub repo before running the script.
  - Run this on a migration branch if you do not want to modify main directly.
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
  if $DRY_RUN; then
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
    log
    printf '%s\n' "$status"
    log
    log "Commit or stash changes before running this script."
    exit 1
  fi
}

confirm() {
  if $ASSUME_YES || $DRY_RUN; then
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

POSITIONAL=()

while [[ $# -gt 0 ]]; do
  case "$1" in
    --remote-base)
      shift
      [[ $# -gt 0 ]] || fail "missing value for --remote-base"
      REMOTE_BASE="$1"
      ;;
    --branch)
      shift
      [[ $# -gt 0 ]] || fail "missing value for --branch"
      TARGET_BRANCH="$1"
      ;;
    --dry-run)
      DRY_RUN=true
      ;;
    -y|--yes)
      ASSUME_YES=true
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    -*)
      fail "unknown option: $1"
      ;;
    *)
      POSITIONAL+=("$1")
      ;;
  esac
  shift
done

if [[ ${#POSITIONAL[@]} -lt 1 || ${#POSITIONAL[@]} -gt 2 ]]; then
  usage
  exit 1
fi

DIR="${POSITIONAL[0]}"
REPO="${POSITIONAL[1]:-$DIR}"
SPLIT_BRANCH="split/$REPO"
REMOTE_URL="${REMOTE_BASE}/${REPO}.git"

if ! git rev-parse --show-toplevel >/dev/null 2>&1; then
  fail "not inside a git repository"
fi

REPO_ROOT="$(git rev-parse --show-toplevel)"
cd "$REPO_ROOT"

require_clean_repo

[[ -d "$DIR" ]] || fail "directory does not exist: $DIR"
git ls-files --error-unmatch "$DIR" >/dev/null 2>&1 || fail "directory is not tracked by git: $DIR"

log "Submodule migration"
log "  repo root: $REPO_ROOT"
log "  directory: $DIR"
log "  child repo: $REPO"
log "  child remote: $REMOTE_URL"
log "  child branch: $TARGET_BRANCH"
log "  dry-run: $DRY_RUN"
log
log "This will push the split history and then replace $DIR with a submodule."
log "Run this on a migration branch if you do not want to change the current branch."
log

confirm

if git show-ref --verify --quiet "refs/heads/$SPLIT_BRANCH"; then
  run_cmd git branch -D "$SPLIT_BRANCH"
fi

log "== Splitting $DIR =="
run_cmd git subtree split --prefix="$DIR" -b "$SPLIT_BRANCH"

log "== Pushing $SPLIT_BRANCH to $REMOTE_URL =="
run_cmd git push "$REMOTE_URL" "${SPLIT_BRANCH}:${TARGET_BRANCH}"

log "== Removing in-tree directory $DIR =="
run_cmd git rm -r "$DIR"
run_cmd git commit -m "chore: remove $DIR before submodule migration"

log "== Adding submodule $DIR =="
run_cmd git submodule add "$REMOTE_URL" "$DIR"
run_cmd git commit -m "chore: add $DIR as submodule"

log
log "Done."
log "Next steps:"
log "  - review: git status"
log "  - inspect .gitmodules and submodule SHA"
log "  - push the parent repo branch when ready"
