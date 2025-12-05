#!/usr/bin/env bash
set -euo pipefail

# Configure Testcontainers to run reliably in CI and locally.
export TESTCONTAINERS_RYUK_DISABLED=${TESTCONTAINERS_RYUK_DISABLED:-false}
export TESTCONTAINERS_CHECKS_DISABLE=${TESTCONTAINERS_CHECKS_DISABLE:-false}
export TESTCONTAINERS_HOST_OVERRIDE=${TESTCONTAINERS_HOST_OVERRIDE:-"localhost"}
export TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE=${TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE:-"unix:///var/run/docker.sock"}

if command -v sysctl >/dev/null 2>&1; then
  sudo sysctl -w kernel.unprivileged_userns_clone=1 || true
fi

echo "Testcontainers configured:"
echo "  TESTCONTAINERS_HOST_OVERRIDE=${TESTCONTAINERS_HOST_OVERRIDE}"
echo "  TESTCONTAINERS_RYUK_DISABLED=${TESTCONTAINERS_RYUK_DISABLED}"
echo "  Docker socket: ${TESTCONTAINERS_DOCKER_SOCKET_OVERRIDE}"
