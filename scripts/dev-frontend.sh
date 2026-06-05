#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

source "$(dirname "$0")/dev-lib.sh"

cleanup_pidfile_process "$FRONTEND_PID_FILE" frontend
cleanup_legacy_frontend_processes

log "starting frontend in foreground on :${FRONTEND_PORT}"
cd "$ROOT_DIR/web"
exec npm run dev -- --host 0.0.0.0
