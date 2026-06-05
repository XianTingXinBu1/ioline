#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

source "$(dirname "$0")/dev-lib.sh"

print_status() {
  local name="$1"
  local pid_file="$2"
  local port="$3"
  local log_file="$4"
  local pid
  pid=$(read_pid "$pid_file" || true)

  if [[ -n "$pid" ]] && is_pid_alive "$pid"; then
    echo "$name: running (pid=$pid, port=$port)"
    echo "  log: $log_file"
  else
    echo "$name: stopped"
  fi
}

print_status backend "$BACKEND_PID_FILE" "$BACKEND_PORT" "$BACKEND_LOG_FILE"
print_status frontend "$FRONTEND_PID_FILE" "$FRONTEND_PORT" "$FRONTEND_LOG_FILE"
