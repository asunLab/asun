#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

LANGUAGE_REPOS=(
  "asun-c"
  "asun-cpp"
  "asun-cs"
  "asun-dart"
  "asun-go"
  "asun-java"
  "asun-js"
  "asun-php"
  "asun-py"
  "asun-rs"
  "asun-swift"
  "asun-zig"
)

mode=""
explicit_tag=""
bump_kind=""
message=""
push_tags=false
dry_run=false
force=false
no_commit=false
version_only=false
targets=()

usage() {
  cat <<'EOF'
Usage:
  scripts/batch-tag-languages.sh --tag v0.1.0 [options]
  scripts/batch-tag-languages.sh --bump minor [options]
  scripts/batch-tag-languages.sh --bump patch [options]
  scripts/batch-tag-languages.sh --version-only --tag v0.1.0 [options]
  scripts/batch-tag-languages.sh --version-only --bump patch [options]

Create the same release tag across all language repositories under this mono-repo.
If a repo has explicit package metadata files, this script also syncs those
version fields to match the tag, creates a version-bump commit, and then tags it.
By default this script only tags the official language repos:
  asun-c asun-cpp asun-cs asun-dart asun-go asun-java asun-js
  asun-php asun-py asun-rs asun-swift asun-zig

Modes:
  --tag TAG           Use the exact same tag for every language repo
  --bump minor        Bump each repo's latest semver tag by 0.1.0
  --bump patch        Bump each repo's latest semver tag by 0.0.1
  --bump 0.1          Alias of minor
  --bump 0.01         Alias of patch

Options:
  --target NAME       Only tag one language repo or a comma-separated subset.
                      Examples: py, js, java, c, cpp, all
  --message MSG       Annotated tag message. Default: "Release <tag>"
  --no-commit         Sync version files in the worktree but do not auto-commit them.
                      Repos with version changes will be skipped for tagging until
                      you commit manually and rerun the script.
  --version-only      Only sync version files. Do not create or push tags.
  --push              Push created tags to origin
  --force             Replace existing local tag if it already exists
  --dry-run           Show what would happen without editing, committing, or pushing
  -h, --help          Show this help

Examples:
  scripts/batch-tag-languages.sh --tag v0.1.0
  scripts/batch-tag-languages.sh --tag v0.1.0 --target py
  scripts/batch-tag-languages.sh --tag v1.0.1 --target rs --no-commit
  scripts/batch-tag-languages.sh --version-only --tag v1.0.1 --target rs
  scripts/batch-tag-languages.sh --tag v0.1.0 --target js,py,rs
  scripts/batch-tag-languages.sh --tag v0.2.0 --push
  scripts/batch-tag-languages.sh --bump minor
  scripts/batch-tag-languages.sh --bump 0.01 --push
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

replace_version_files() {
  local repo_name="$1"
  local repo_path="$2"
  local tag="$3"
  local version="${tag#v}"

  python3 - "$repo_name" "$repo_path" "$version" <<'PY'
import json
import pathlib
import re
import sys

repo_name = sys.argv[1]
repo_path = pathlib.Path(sys.argv[2])
version = sys.argv[3]
changed = []

def write_text_if_changed(path: pathlib.Path, new_text: str):
    old_text = path.read_text()
    if old_text != new_text:
        path.write_text(new_text)
        changed.append(str(path.relative_to(repo_path)))

def replace_regex(path: pathlib.Path, pattern: str, repl: str, count: int = 0):
    old_text = path.read_text()
    new_text, n = re.subn(pattern, repl, old_text, count=count, flags=re.M)
    if n > 0 and new_text != old_text:
        path.write_text(new_text)
        changed.append(str(path.relative_to(repo_path)))

if repo_name == "asun-js":
    for rel in ("package.json", "package-lock.json"):
        path = repo_path / rel
        data = json.loads(path.read_text())
        data["version"] = version
        if rel == "package-lock.json":
            pkgs = data.get("packages", {})
            if "" in pkgs and isinstance(pkgs[""], dict):
                pkgs[""]["version"] = version
        write_text_if_changed(path, json.dumps(data, indent=2, ensure_ascii=False) + "\n")

elif repo_name == "asun-py":
    replace_regex(repo_path / "pyproject.toml", r'^version = "[^"]+"$', f'version = "{version}"', count=1)
    replace_regex(repo_path / "setup.py", r'version="[^"]+"', f'version="{version}"', count=1)

elif repo_name == "asun-cs":
    replace_regex(repo_path / "src/Asun/Asun.csproj", r'<Version>[^<]+</Version>', f'<Version>{version}</Version>', count=1)

elif repo_name == "asun-dart":
    replace_regex(repo_path / "pubspec.yaml", r'^version:\s*[^\s]+$', f'version: {version}', count=1)

elif repo_name == "asun-rs":
    replace_regex(repo_path / "Cargo.toml", r'^version = "[^"]+"$', f'version = "{version}"', count=1)
    lock_path = repo_path / "Cargo.lock"
    old_text = lock_path.read_text()
    new_text = re.sub(r'(\[\[package\]\]\nname = "asun"\nversion = ")[^"]+(")', rf'\g<1>{version}\2', old_text, count=1)
    if new_text != old_text:
        lock_path.write_text(new_text)
        changed.append("Cargo.lock")

elif repo_name == "asun-java":
    replace_regex(repo_path / "build.gradle", r"^version = '[^']+'$", f"version = '{version}'", count=1)

elif repo_name == "asun-cpp":
    replace_regex(repo_path / "CMakeLists.txt", r'project\(asun VERSION [^\s\)]+', f'project(asun VERSION {version}', count=1)
    replace_regex(repo_path / "conanfile.py", r'version = "[^"]+"', f'version = "{version}"', count=1)
    vcpkg_path = repo_path / "vcpkg/ports/asun-cpp/vcpkg.json"
    data = json.loads(vcpkg_path.read_text())
    data["version-string"] = version
    write_text_if_changed(vcpkg_path, json.dumps(data, indent=2, ensure_ascii=False) + "\n")
    replace_regex(repo_path / "homebrew/asun-cpp.rb", r'url "https://github\.com/asun-lab/asun/archive/refs/tags/v[^"]+\.tar\.gz"', f'url "https://github.com/asun-lab/asun/archive/refs/tags/v{version}.tar.gz"', count=1)

print("\n".join(changed))
PY
}

sync_repo_version() {
  local repo_name="$1"
  local repo_path="$2"
  local repo_tag="$3"
  local version="${repo_tag#v}"
  local changed_files
  local changed_array=()

  changed_files="$(replace_version_files "$repo_name" "$repo_path" "$repo_tag")"
  if [[ -n "$changed_files" ]]; then
    mapfile -t changed_array <<<"$changed_files"
  fi

  if [[ ${#changed_array[@]} -eq 0 ]]; then
    return 0
  fi

  if $dry_run; then
    log "Would sync $repo_name version files to $version:"
    printf '  %s\n' "${changed_array[@]}"
    git -C "$repo_path" checkout -- "${changed_array[@]}"
    if $no_commit; then
      return 10
    fi
    return 0
  fi

  if $no_commit; then
    log "Synced $repo_name version files to $version in the worktree (no commit):"
    printf '  %s\n' "${changed_array[@]}"
    return 10
  fi

  git -C "$repo_path" add -- "${changed_array[@]}"
  git -C "$repo_path" commit -m "chore: bump version to $repo_tag" >/dev/null
  log "Synced $repo_name version files to $version and created a version bump commit"
}

is_valid_tag() {
  [[ "$1" =~ ^v?[0-9]+\.[0-9]+\.[0-9]+$ ]]
}

normalize_tag() {
  local raw="$1"
  if [[ "$raw" == v* ]]; then
    printf '%s\n' "$raw"
  else
    printf 'v%s\n' "$raw"
  fi
}

latest_semver_tag() {
  local repo="$1"
  git -C "$repo" tag --list | grep -E '^v?[0-9]+\.[0-9]+\.[0-9]+$' | sort -V | tail -n 1 || true
}

bump_tag() {
  local tag="$1"
  local kind="$2"
  local version="${tag#v}"
  local major minor patch
  IFS='.' read -r major minor patch <<<"$version"
  case "$kind" in
    minor)
      minor=$((minor + 1))
      patch=0
      ;;
    patch)
      patch=$((patch + 1))
      ;;
    *)
      log "Unsupported bump kind: $kind"
      exit 1
      ;;
  esac
  printf 'v%s.%s.%s\n' "$major" "$minor" "$patch"
}

ensure_clean_repo() {
  local repo="$1"
  if [[ -n "$(git -C "$repo" status --short)" ]]; then
    log "Skipping $repo: worktree has uncommitted changes"
    return 1
  fi
  return 0
}

resolve_target() {
  local name="${1,,}"
  case "$name" in
    all) printf 'all\n' ;;
    c|asun-c) printf 'asun-c\n' ;;
    cpp|cxx|asun-cpp) printf 'asun-cpp\n' ;;
    cs|csharp|dotnet|asun-cs) printf 'asun-cs\n' ;;
    dart|asun-dart) printf 'asun-dart\n' ;;
    go|golang|asun-go) printf 'asun-go\n' ;;
    java|jvm|asun-java) printf 'asun-java\n' ;;
    js|ts|javascript|typescript|asun-js) printf 'asun-js\n' ;;
    php|asun-php) printf 'asun-php\n' ;;
    py|python|asun-py) printf 'asun-py\n' ;;
    rs|rust|asun-rs) printf 'asun-rs\n' ;;
    swift|asun-swift) printf 'asun-swift\n' ;;
    zig|asun-zig) printf 'asun-zig\n' ;;
    *)
      return 1
      ;;
  esac
}

while [[ $# -gt 0 ]]; do
  case "$1" in
    --tag)
      shift
      [[ $# -gt 0 ]] || { log "Missing value for --tag"; usage; exit 1; }
      mode="tag"
      explicit_tag="$(normalize_tag "$1")"
      ;;
    --bump)
      shift
      [[ $# -gt 0 ]] || { log "Missing value for --bump"; usage; exit 1; }
      mode="bump"
      bump_kind="$1"
      ;;
    --message)
      shift
      [[ $# -gt 0 ]] || { log "Missing value for --message"; usage; exit 1; }
      message="$1"
      ;;
    --no-commit)
      no_commit=true
      ;;
    --version-only)
      version_only=true
      ;;
    --target)
      shift
      [[ $# -gt 0 ]] || { log "Missing value for --target"; usage; exit 1; }
      IFS=',' read -r -a raw_targets <<<"$1"
      for raw_target in "${raw_targets[@]}"; do
        resolved="$(resolve_target "$raw_target")" || {
          log "Unknown target: $raw_target"
          usage
          exit 1
        }
        targets+=("$resolved")
      done
      ;;
    --push)
      push_tags=true
      ;;
    --dry-run)
      dry_run=true
      ;;
    --force)
      force=true
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

if [[ -z "$mode" ]]; then
  log "You must provide either --tag or --bump."
  usage
  exit 1
fi

if $version_only && $push_tags; then
  log "--version-only cannot be used with --push"
  exit 1
fi

if [[ "$mode" == "tag" ]]; then
  if ! is_valid_tag "$explicit_tag"; then
    log "Tag must look like v0.1.0 or 0.1.0"
    exit 1
  fi
fi

if [[ "$mode" == "bump" ]]; then
  case "$bump_kind" in
    minor|0.1)
      bump_kind="minor"
      ;;
    patch|0.01)
      bump_kind="patch"
      ;;
    *)
      log "--bump only supports: minor, patch, 0.1, 0.01"
      exit 1
      ;;
  esac
fi

selected_repos=()
if [[ ${#targets[@]} -eq 0 ]]; then
  selected_repos=("${LANGUAGE_REPOS[@]}")
else
  for target in "${targets[@]}"; do
    if [[ "$target" == "all" ]]; then
      selected_repos=("${LANGUAGE_REPOS[@]}")
      break
    fi
    selected_repos+=("$target")
  done
fi

deduped_repos=()
for repo_name in "${selected_repos[@]}"; do
  skip=false
  for existing in "${deduped_repos[@]}"; do
    if [[ "$existing" == "$repo_name" ]]; then
      skip=true
      break
    fi
  done
  if ! $skip; then
    deduped_repos+=("$repo_name")
  fi
done

created=()
planned=()
skipped=()
failed=()

for repo_name in "${deduped_repos[@]}"; do
  repo_path="$ROOT_DIR/$repo_name"
  if [[ ! -d "$repo_path/.git" && ! -f "$repo_path/.git" ]]; then
    skipped+=("$repo_name (missing repo)")
    continue
  fi

  if ! ensure_clean_repo "$repo_path"; then
    skipped+=("$repo_name (dirty)")
    continue
  fi

  if [[ "$mode" == "tag" ]]; then
    repo_tag="$explicit_tag"
  else
    current_tag="$(latest_semver_tag "$repo_path")"
    if [[ -z "$current_tag" ]]; then
      current_tag="v0.0.0"
    fi
    repo_tag="$(bump_tag "$current_tag" "$bump_kind")"
  fi

  repo_message="${message:-Release $repo_tag}"

  if git -C "$repo_path" rev-parse -q --verify "refs/tags/$repo_tag" >/dev/null 2>&1; then
    if ! $force; then
      skipped+=("$repo_name ($repo_tag exists)")
      continue
    fi
    run_cmd git -C "$repo_path" tag -d "$repo_tag"
  fi

  if sync_repo_version "$repo_name" "$repo_path" "$repo_tag"; then
    :
  else
    sync_status=$?
    if [[ $sync_status -eq 10 ]]; then
      skipped+=("$repo_name (version updated, commit required before tagging)")
      continue
    fi
    failed+=("$repo_name@$repo_tag (version sync failed)")
    continue
  fi

  if $version_only; then
    if $dry_run; then
      planned+=("$repo_name@$repo_tag (version only)")
    else
      created+=("$repo_name@$repo_tag (version only)")
    fi
    continue
  fi

  if run_cmd git -C "$repo_path" tag -a "$repo_tag" -m "$repo_message"; then
    if $dry_run; then
      planned+=("$repo_name@$repo_tag")
    else
      created+=("$repo_name@$repo_tag")
    fi
  else
    failed+=("$repo_name@$repo_tag")
    continue
  fi

  if $push_tags; then
    if ! run_cmd git -C "$repo_path" push origin "$repo_tag"; then
      failed+=("$repo_name@$repo_tag (push failed)")
    fi
  fi
done

log
log "Summary"
if [[ ${#planned[@]} -gt 0 ]]; then
  printf '  planned: %s\n' "${planned[@]}"
fi
if [[ ${#created[@]} -gt 0 ]]; then
  printf '  created: %s\n' "${created[@]}"
fi
if [[ ${#skipped[@]} -gt 0 ]]; then
  printf '  skipped: %s\n' "${skipped[@]}"
fi
if [[ ${#failed[@]} -gt 0 ]]; then
  printf '  failed: %s\n' "${failed[@]}"
fi

if [[ ${#failed[@]} -gt 0 ]]; then
  exit 1
fi
