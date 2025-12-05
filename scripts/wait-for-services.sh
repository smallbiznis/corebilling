#!/usr/bin/env bash
set -euo pipefail

wait_for() {
  local name=$1
  local host=$2
  local port=$3
  local retries=${4:-60}
  local delay=${5:-2}

  echo "Waiting for ${name} at ${host}:${port}..."
  for ((i=1; i<=retries; i++)); do
    if nc -z "$host" "$port" >/dev/null 2>&1; then
      echo "${name} is available"
      return 0
    fi
    sleep "$delay"
  done
  echo "${name} did not become ready in time" >&2
  return 1
}

wait_for "Postgres" "${POSTGRES_HOST:-localhost}" "${POSTGRES_PORT:-5432}" "${POSTGRES_RETRIES:-60}" "${POSTGRES_WAIT:-2}"
wait_for "Redis" "${REDIS_HOST:-localhost}" "${REDIS_PORT:-6379}" "${REDIS_RETRIES:-60}" "${REDIS_WAIT:-2}"
wait_for "NATS" "${NATS_HOST:-localhost}" "${NATS_PORT:-4222}" "${NATS_RETRIES:-60}" "${NATS_WAIT:-2}"
wait_for "Kafka" "${KAFKA_HOST:-localhost}" "${KAFKA_PORT:-9092}" "${KAFKA_RETRIES:-90}" "${KAFKA_WAIT:-3}"
