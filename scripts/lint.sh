#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git rev-parse --show-toplevel)
cd "${ROOT_DIR}"

format_output=$(gofmt -s -l .)
if [[ -n "${format_output}" ]]; then
  echo "gofmt detected files that need formatting:" >&2
  echo "${format_output}" >&2
  exit 1
fi

gofumpt_output=$(gofumpt -l .)
if [[ -n "${gofumpt_output}" ]]; then
  echo "gofumpt detected files that need formatting:" >&2
  echo "${gofumpt_output}" >&2
  exit 1
fi

golangci-lint run ./...

go vet ./...

go test ./... >/dev/null

sqlc generate
if ! git diff --quiet --stat -- sqlc.yaml internal db; then
  echo "sqlc generated files are out of date with migrations or configuration." >&2
  git diff --stat -- sqlc.yaml internal db
  exit 1
fi
