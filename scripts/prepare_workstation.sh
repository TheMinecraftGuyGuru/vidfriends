#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

MIN_GO="1.21"
MIN_NODE="18.0.0"

version_ge() {
  if [[ "$1" == "$2" ]]; then
    return 0
  fi
  local sorted
  sorted=$(printf '%s\n' "$1" "$2" | sort -V | head -n1)
  [[ "$sorted" == "$2" ]]
}

check_tool() {
  local cmd="$1"
  local label="$2"
  local required="$3"
  local min_version="${4:-}"
  local version_cmd="${5:-}"
  local parser="${6:-}"

  if ! command -v "$cmd" >/dev/null 2>&1; then
    if [[ "$required" == "required" ]]; then
      echo "[ERROR] Missing required tool: $label ($cmd)" >&2
      return 1
    else
      echo "[WARN] Optional tool not found: $label ($cmd)" >&2
      return 0
    fi
  fi

  if [[ -n "$min_version" ]]; then
    local version_output
    if [[ -n "$version_cmd" ]]; then
      version_output=$(eval "$version_cmd")
    else
      version_output=$("$cmd" --version 2>/dev/null)
    fi
    local version
    if [[ -n "$parser" ]]; then
      version=$(eval "$parser" <<<"$version_output")
    else
      version="$version_output"
    fi

    if ! version_ge "$version" "$min_version"; then
      echo "[ERROR] $label version $version is below required minimum $min_version" >&2
      return 1
    fi
  fi
}

extract_go_version() {
  sed -n 's/^go version go\([0-9.]*\).*/\1/p'
}

extract_node_version() {
  sed -n 's/^v\([0-9.]*\).*/\1/p'
}

copy_env_file() {
  local source="$1"
  local dest="$2"
  if [[ ! -f "$source" ]]; then
    echo "[WARN] Example file missing: $source"
    return 0
  fi
  if [[ -f "$dest" ]]; then
    echo "[SKIP] $dest already exists"
  else
    cp "$source" "$dest"
    echo "[OK] Created $dest"
  fi
}

main() {
  local errors=0

  echo "== Checking required tools =="
  if ! check_tool go "Go" required "$MIN_GO" "go version" extract_go_version; then
    errors=$((errors+1))
  fi
  if ! check_tool node "Node.js" required "$MIN_NODE" "node --version" extract_node_version; then
    errors=$((errors+1))
  fi
  if ! check_tool pnpm "pnpm" optional; then
    echo "      pnpm is recommended for frontend dependency management."
  fi
  if ! check_tool npm "npm" optional; then
    echo "      npm can be used as a fallback if pnpm is unavailable."
  fi
  if ! check_tool psql "PostgreSQL client" optional; then
    echo "      Install PostgreSQL locally to run the database without Docker."
  fi
  if ! check_tool docker "Docker" optional; then
    echo "      Docker is required for the Compose-based workflow."
  fi
  if ! check_tool yt-dlp "yt-dlp" optional; then
    echo "      yt-dlp enables richer video metadata retrieval."
  fi

  echo
  echo "== Preparing environment files =="
  copy_env_file "$ROOT_DIR/backend/.env.example" "$ROOT_DIR/backend/.env"
  copy_env_file "$ROOT_DIR/frontend/.env.example" "$ROOT_DIR/frontend/.env.local"
  copy_env_file "$ROOT_DIR/deploy/.env.example" "$ROOT_DIR/deploy/.env"

  echo
  if [[ "$errors" -gt 0 ]]; then
    echo "Completed with $errors error(s). Please install the missing required tools and re-run this script." >&2
    exit 1
  else
    echo "Workstation preparation complete. Edit the generated .env files with project-specific secrets before running the stack."
  fi
}

main "$@"
