#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:9650}"
WORKSPACE_PATH="${WORKSPACE_PATH:-}"
QUERY="${QUERY:-search-keyword}"
REQUESTS="${REQUESTS:-12}"
CONCURRENCY="${CONCURRENCY:-3}"

if [ -z "$WORKSPACE_PATH" ]; then
  echo "[stress] WORKSPACE_PATH is required" >&2
  exit 1
fi

WORKSPACE_PATH="$(cd "$WORKSPACE_PATH" && pwd)"

curl -fsS -X PUT "$BASE_URL/api/workspace/current" \
  -H 'Content-Type: application/json' \
  -d "{\"rootPath\":\"$WORKSPACE_PATH\"}" >/dev/null

run_request() {
  local output http_code duration_ms body
  body="{\"query\":\"$QUERY\"}"
  output="$(curl -sS -o /dev/null -w '%{http_code} %{time_total}' -X POST "$BASE_URL/api/search/text" -H 'Content-Type: application/json' -d "$body")"
  http_code="${output%% *}"
  duration_ms="$(awk -v sec="${output##* }" 'BEGIN { printf "%d", sec * 1000 }')"
  printf '%s %s\n' "$http_code" "$duration_ms"
}

export BASE_URL QUERY
export -f run_request

started_at="$(date +%s)"
results="$(seq 1 "$REQUESTS" | xargs -I{} -P "$CONCURRENCY" bash -lc 'run_request')"
ended_at="$(date +%s)"

success="$(printf '%s\n' "$results" | awk '$1 == 200 { count++ } END { print count + 0 }')"
failed="$((REQUESTS - success))"
max_ms="$(printf '%s\n' "$results" | awk 'BEGIN { max = 0 } { if ($2 > max) max = $2 } END { print max + 0 }')"
avg_ms="$(printf '%s\n' "$results" | awk 'BEGIN { sum = 0; count = 0 } { sum += $2; count++ } END { if (count == 0) print 0; else printf "%d", sum / count }')"
total_seconds="$((ended_at - started_at))"

echo "[stress] search/text"
echo "  workspace:   $WORKSPACE_PATH"
echo "  query:       $QUERY"
echo "  requests:    $REQUESTS"
echo "  concurrency: $CONCURRENCY"
echo "  success:     $success"
echo "  failed:      $failed"
echo "  totalTime:   ${total_seconds}s"
echo "  avgTime:     ${avg_ms}ms"
echo "  maxTime:     ${max_ms}ms"
