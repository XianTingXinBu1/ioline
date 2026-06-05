#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

source "$(dirname "$0")/dev-lib.sh"

cleanup_pidfile_process "$BACKEND_PID_FILE" backend
cleanup_pidfile_process "$FRONTEND_PID_FILE" frontend
cleanup_legacy_backend_processes
cleanup_legacy_frontend_processes
rm -f "$BACKEND_PID_FILE" "$FRONTEND_PID_FILE"
log "dev services stopped"
