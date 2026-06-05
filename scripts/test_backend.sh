#!/data/data/com.termux/files/usr/bin/bash
set -euo pipefail

BASE_URL="http://127.0.0.1:8080"
WORKSPACE="/data/data/com.termux/files/home/project/ioline"
TMP_WS_REL="tmp/api-test"
TMP_FILE_REL="$TMP_WS_REL/hello.txt"
TMP_MOVED_REL="$TMP_WS_REL/hello-renamed.txt"

cleanup() {
  curl -sS -X DELETE "$BASE_URL/api/files" \
    -H 'Content-Type: application/json' \
    -d '{"path":"tmp/api-test","recursive":true}' >/dev/null 2>&1 || true
  curl -sS -X DELETE "$BASE_URL/api/workspace/current" >/dev/null 2>&1 || true
}
trap cleanup EXIT

assert_contains() {
  local haystack="$1"
  local needle="$2"
  if [[ "$haystack" != *"$needle"* ]]; then
    echo "ASSERT FAILED: expected to find '$needle'"
    echo "$haystack"
    exit 1
  fi
}

echo "[1] healthz"
resp=$(curl -sS "$BASE_URL/api/healthz")
assert_contains "$resp" '"success":true'
assert_contains "$resp" '"status":"ok"'

echo "[2] system info"
resp=$(curl -sS "$BASE_URL/api/system/info")
assert_contains "$resp" '"name":"ioline"'

echo "[3] workspace candidates"
resp=$(curl -sS "$BASE_URL/api/workspaces/candidates")
assert_contains "$resp" '"items"'

echo "[4] set workspace"
resp=$(curl -sS -X PUT "$BASE_URL/api/workspace/current" \
  -H 'Content-Type: application/json' \
  -d "{\"rootPath\":\"$WORKSPACE\"}")
assert_contains "$resp" '"isSet":true'

echo "[5] get workspace"
resp=$(curl -sS "$BASE_URL/api/workspace/current")
assert_contains "$resp" "$WORKSPACE"

echo "[6] list files"
resp=$(curl -sS "$BASE_URL/api/files/list?path=.")
assert_contains "$resp" '"items"'

echo "[7] stat file"
resp=$(curl -sS "$BASE_URL/api/files/stat?path=go.mod")
assert_contains "$resp" '"name":"go.mod"'

echo "[8] read file content"
resp=$(curl -sS "$BASE_URL/api/file/content?path=go.mod")
assert_contains "$resp" '"content":"module ioline'

echo "[9] create directory"
resp=$(curl -sS -X POST "$BASE_URL/api/directories" \
  -H 'Content-Type: application/json' \
  -d '{"path":"tmp/api-test"}')
assert_contains "$resp" '"type":"directory"'

echo "[10] create file"
resp=$(curl -sS -X POST "$BASE_URL/api/files" \
  -H 'Content-Type: application/json' \
  -d '{"path":"tmp/api-test/hello.txt","content":"hello terminal api"}')
assert_contains "$resp" '"type":"file"'

echo "[11] save file"
resp=$(curl -sS -X PUT "$BASE_URL/api/file/content" \
  -H 'Content-Type: application/json' \
  -d '{"path":"tmp/api-test/hello.txt","content":"hello updated"}')
assert_contains "$resp" '"path":"tmp/api-test/hello.txt"'

echo "[12] move file"
resp=$(curl -sS -X PATCH "$BASE_URL/api/files/move" \
  -H 'Content-Type: application/json' \
  -d '{"fromPath":"tmp/api-test/hello.txt","toPath":"tmp/api-test/hello-renamed.txt"}')
assert_contains "$resp" '"toPath":"tmp/api-test/hello-renamed.txt"'

echo "[13] list terminals"
resp=$(curl -sS "$BASE_URL/api/terminals")
assert_contains "$resp" '"items"'

echo "[14] create terminal"
resp=$(curl -sS -X POST "$BASE_URL/api/terminals" \
  -H 'Content-Type: application/json' \
  -d '{"cols":80,"rows":24}')
assert_contains "$resp" '"status":"running"'
term_id=$(printf '%s' "$resp" | sed -n 's/.*"id":"\([^"]*\)".*/\1/p')
if [[ -z "$term_id" ]]; then
  echo "ASSERT FAILED: terminal id empty"
  exit 1
fi

echo "[15] resize terminal"
resp=$(curl -sS -X POST "$BASE_URL/api/terminals/$term_id/resize" \
  -H 'Content-Type: application/json' \
  -d '{"cols":100,"rows":30}')
assert_contains "$resp" '"cols":100'

echo "[16] websocket terminal smoke test"
cat <<'EOF' >.tmp/ioline_ws_check.go
package main

import (
  "fmt"
  "strings"
  "time"
  "github.com/gorilla/websocket"
)

func main() {
  conn, _, err := websocket.DefaultDialer.Dial("ws://127.0.0.1:8080/api/terminals/TERM_ID/stream", nil)
  if err != nil { panic(err) }
  defer conn.Close()

  _ = conn.SetReadDeadline(time.Now().Add(5 * time.Second))
  if err := conn.WriteMessage(websocket.TextMessage, []byte("pwd\necho IOLINE_WS_OK\n")); err != nil { panic(err) }

  deadline := time.Now().Add(5 * time.Second)
  found := false
  for time.Now().Before(deadline) {
    _, msg, err := conn.ReadMessage()
    if err != nil { panic(err) }
    text := string(msg)
    if len(text) > 0 {
      fmt.Print(text)
    }
    if strings.Contains(text, "IOLINE_WS_OK") {
      found = true
      break
    }
  }
  if !found { panic("missing terminal websocket output marker") }
}
EOF
sed -i "s/TERM_ID/$term_id/g" .tmp/ioline_ws_check.go
go run .tmp/ioline_ws_check.go >.tmp/ioline_ws_test.out
assert_contains "$(cat .tmp/ioline_ws_test.out)" 'IOLINE_WS_OK'

echo "[17] close terminal"
resp=$(curl -sS -X DELETE "$BASE_URL/api/terminals/$term_id")
assert_contains "$resp" '"status":"closed"'

echo "[18] delete test directory recursively"
resp=$(curl -sS -X DELETE "$BASE_URL/api/files" \
  -H 'Content-Type: application/json' \
  -d '{"path":"tmp/api-test","recursive":true}')
assert_contains "$resp" '"type":"directory"'

echo "[19] clear workspace"
resp=$(curl -sS -X DELETE "$BASE_URL/api/workspace/current")
assert_contains "$resp" '"isSet":false'

echo "ALL BACKEND API TESTS PASSED"
