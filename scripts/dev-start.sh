#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

source "$(dirname "$0")/dev-lib.sh"

start_backend_managed() {
  cleanup_pidfile_process "$BACKEND_PID_FILE" backend
  cleanup_legacy_backend_processes

  log "building backend binary..."
  (cd "$ROOT_DIR" && go build -o "$BACKEND_BIN" ./apps/server)

  log "starting backend on ${BACKEND_ADDR}"
  : > "$BACKEND_LOG_FILE"
  env IOLINE_SERVER_ADDR="$BACKEND_ADDR" "$BACKEND_BIN" >> "$BACKEND_LOG_FILE" 2>&1 &
  local pid=$!
  echo "$pid" > "$BACKEND_PID_FILE"

  if ! wait_for_http "http://127.0.0.1:${BACKEND_PORT}/api/healthz" "$pid" 20; then
    log "backend failed to become ready"
    cat "$BACKEND_LOG_FILE" >&2 || true
    stop_pid "$pid"
    rm -f "$BACKEND_PID_FILE"
    exit 1
  fi

  log "backend is ready (pid=$pid, port=$BACKEND_PORT)"
}

start_frontend_managed() {
  cleanup_pidfile_process "$FRONTEND_PID_FILE" frontend
  cleanup_legacy_frontend_processes

  log "starting frontend on :${FRONTEND_PORT}"
  : > "$FRONTEND_LOG_FILE"
  (
    cd "$ROOT_DIR/web"
    exec npm run dev -- --host 0.0.0.0
  ) >> "$FRONTEND_LOG_FILE" 2>&1 &
  local pid=$!
  echo "$pid" > "$FRONTEND_PID_FILE"

  if ! wait_for_http "http://127.0.0.1:${FRONTEND_PORT}" "$pid" 20; then
    log "frontend failed to become ready"
    cat "$FRONTEND_LOG_FILE" >&2 || true
    stop_pid "$pid"
    rm -f "$FRONTEND_PID_FILE"
    exit 1
  fi

  log "frontend is ready (pid=$pid, port=$FRONTEND_PORT)"
}

"$ROOT_DIR/scripts/dev-stop.sh" >/dev/null 2>&1 || true
log "starting backend and frontend dev services..."
start_backend_managed
start_frontend_managed
"$ROOT_DIR/scripts/dev-status.sh"
