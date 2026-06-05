#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

source "$(dirname "$0")/dev-lib.sh"

cleanup_pidfile_process "$BACKEND_PID_FILE" backend
cleanup_legacy_backend_processes

log "building backend binary..."
(cd "$ROOT_DIR" && go build -o "$BACKEND_BIN" ./apps/server)

log "starting backend in foreground on ${BACKEND_ADDR}"
exec env IOLINE_SERVER_ADDR="$BACKEND_ADDR" "$BACKEND_BIN"
