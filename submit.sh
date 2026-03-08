#!/usr/bin/env bash
set -euo pipefail

msg="${1:-chore: update}"

git submodule foreach '
  if [ -n "$(git status --porcelain)" ]; then
    echo "== committing $name =="
    git add -A
    git commit -m "'"$msg"'"
  fi
'

git submodule foreach '
  branch=$(git branch --show-current)
  if [ -n "$branch" ]; then
    echo "== pushing $name:$branch =="
    git push origin "$branch"
  else
    echo "== skip $name: detached HEAD =="
  fi
'

git add .
if [ -n "$(git status --porcelain)" ]; then
  git commit -m "$msg"
  git push
fi