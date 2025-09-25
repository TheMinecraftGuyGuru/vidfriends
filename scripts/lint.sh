#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
EXIT_CODE=0

mapfile -t GO_FILES < <(find "$ROOT_DIR/backend" -type f -name '*.go' -print)
if [[ ${#GO_FILES[@]} -gt 0 ]]; then
  echo "== Running gofmt checks =="
  mapfile -t UNFORMATTED < <(gofmt -l "${GO_FILES[@]}" || true)
  if [[ ${#UNFORMATTED[@]} -gt 0 ]]; then
    echo "The following Go files need formatting:"
    printf '  %s\n' "${UNFORMATTED[@]}"
    EXIT_CODE=1
  else
    echo "All Go files are properly formatted."
  fi
else
  echo "No Go files found under backend/. Skipping gofmt check."
fi

if [[ -f "$ROOT_DIR/frontend/package.json" ]]; then
  echo -e "\n== Running frontend lint =="
  pushd "$ROOT_DIR/frontend" >/dev/null
  if [[ -f pnpm-lock.yaml ]]; then
    corepack enable pnpm
    pnpm install --frozen-lockfile
    pnpm run lint
  elif [[ -f yarn.lock ]]; then
    corepack enable
    yarn install --immutable
    yarn lint
  else
    if [[ -f package-lock.json ]]; then
      npm ci
    else
      npm install
    fi
    npm run lint
  fi
  popd >/dev/null
else
  echo "No frontend/package.json found. Skipping frontend lint."
fi

exit $EXIT_CODE
