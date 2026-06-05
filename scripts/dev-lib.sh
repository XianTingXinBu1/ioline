#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

ROOT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
RUNTIME_DIR="$ROOT_DIR/.runtime"
BACKEND_PORT="${IOLINE_SERVER_PORT:-9650}"
BACKEND_ADDR=":${BACKEND_PORT}"
BACKEND_BIN="$ROOT_DIR/bin/ioline-server"
BACKEND_PID_FILE="$RUNTIME_DIR/backend.pid"
BACKEND_LOG_FILE="$RUNTIME_DIR/backend.log"
FRONTEND_PID_FILE="$RUNTIME_DIR/frontend.pid"
FRONTEND_LOG_FILE="$RUNTIME_DIR/frontend.log"
FRONTEND_PORT="${IOLINE_FRONTEND_PORT:-5173}"
mkdir -p "$RUNTIME_DIR" "$ROOT_DIR/bin"

log() {
  echo "[ioline] $*"
}

read_pid() {
  local file="$1"
  [[ -f "$file" ]] || return 1
  tr -d '[:space:]' < "$file"
}

is_pid_alive() {
  local pid="$1"
  [[ -n "$pid" ]] || return 1
  kill -0 "$pid" 2>/dev/null
}

stop_pid() {
  local pid="$1"
  [[ -n "$pid" ]] || return 0
  kill "$pid" 2>/dev/null || true
  sleep 1
  if is_pid_alive "$pid"; then
    kill -9 "$pid" 2>/dev/null || true
    sleep 1
  fi
}

cleanup_pidfile_process() {
  local pid_file="$1"
  local name="$2"
  local pid
  pid=$(read_pid "$pid_file" || true)
  [[ -n "$pid" ]] || return 0
  log "stopping managed $name process (pid=$pid)..."
  stop_pid "$pid"
  rm -f "$pid_file"
}

wait_for_http() {
  local url="$1"
  local pid="$2"
  local attempts="${3:-30}"

  for _ in $(seq 1 "$attempts"); do
    if ! is_pid_alive "$pid"; then
      return 1
    fi
    if curl -fsS "$url" >/dev/null 2>&1; then
      return 0
    fi
    sleep 1
  done

  return 1
}

cleanup_legacy_backend_processes() {
  local patterns=(
    'go run ./apps/server'
    '/data/data/.*/project/ioline/server'
    '/data/data/.*/project/ioline/bin/ioline-server'
    "${BACKEND_BIN//\//\\/}"
    "sh -c \(trap 'kill 0' INT TERM EXIT;.*make dev-backend.*make dev-frontend.*wait\)"
  )

  local pattern
  for pattern in "${patterns[@]}"; do
    pkill -f "$pattern" 2>/dev/null || true
  done

  fuser -k -n tcp "$BACKEND_PORT" >/dev/null 2>&1 || true
  sleep 2
}

cleanup_legacy_frontend_processes() {
  pkill -f 'node .*/vite' 2>/dev/null || true
  pkill -f 'npm run dev -- --host 0.0.0.0' 2>/dev/null || true
  fuser -k -n tcp "$FRONTEND_PORT" >/dev/null 2>&1 || true
  sleep 2
}
