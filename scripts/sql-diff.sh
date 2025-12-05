#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR=$(git rev-parse --show-toplevel)
cd "${ROOT_DIR}"

tmpdir=$(mktemp -d)
trap 'rm -rf "${tmpdir}"' EXIT

sqlc generate

if ! git diff --quiet -- db internal sqlc.yaml; then
  echo "Detected drift between migrations and generated sqlc code." >&2
  git diff --stat -- db internal sqlc.yaml >&2
  exit 1
fi

echo "No SQL drift detected."
