package server

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestHealthzAndSystemInfo(t *testing.T) {
	t.Parallel()

	srv := New(Config{Logger: log.New(io.Discard, "", 0)})

	healthz := doRequest(t, srv, http.MethodGet, "/api/healthz", "")
	assertStatus(t, healthz, http.StatusOK)
	assertSuccess(t, healthz)
	assertJSONContains(t, healthz.Body.String(), `"status":"ok"`)

	systemInfo := doRequest(t, srv, http.MethodGet, "/api/system/info", "")
	assertStatus(t, systemInfo, http.StatusOK)
	assertSuccess(t, systemInfo)
	assertJSONContains(t, systemInfo.Body.String(), `"name":"ioline"`)
	assertJSONContains(t, systemInfo.Body.String(), `"terminalMaxSessions":4`)
}

func TestMethodNotAllowed(t *testing.T) {
	t.Parallel()

	srv := New(Config{Logger: log.New(io.Discard, "", 0)})

	cases := []struct {
		name   string
		method string
		path   string
	}{
		{name: "healthz post", method: http.MethodPost, path: "/api/healthz"},
		{name: "system info post", method: http.MethodPost, path: "/api/system/info"},
		{name: "workspace directories post", method: http.MethodPost, path: "/api/workspace/directories"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := doRequest(t, srv, tc.method, tc.path, "")
			assertStatus(t, resp, http.StatusMethodNotAllowed)
			assertErrorCode(t, resp, "METHOD_NOT_ALLOWED")
		})
	}
}

func TestWorkspaceRequiredEndpoints(t *testing.T) {
	t.Parallel()

	srv := New(Config{Logger: log.New(io.Discard, "", 0)})

	cases := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{name: "files list", method: http.MethodGet, path: "/api/files/list"},
		{name: "search files", method: http.MethodGet, path: "/api/search/files?query=server"},
		{name: "search text", method: http.MethodPost, path: "/api/search/text", body: `{"query":"workspace"}`},
		{name: "terminal create", method: http.MethodPost, path: "/api/terminals", body: `{"cols":80,"rows":24}`},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := doRequest(t, srv, tc.method, tc.path, tc.body)
			assertStatus(t, resp, http.StatusBadRequest)
			assertErrorCode(t, resp, "WORKSPACE_NOT_CONFIGURED")
		})
	}
}

func TestInvalidJSONRequests(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	srv := newWorkspaceServer(t, root)

	cases := []struct {
		name string
		path string
	}{
		{name: "workspace put", path: "/api/workspace/current"},
		{name: "search text", path: "/api/search/text"},
		{name: "files create", path: "/api/files"},
		{name: "directories create", path: "/api/directories"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			resp := doRequest(t, srv, http.MethodPost, tc.path, "{")
			if tc.path == "/api/workspace/current" {
				resp = doRequest(t, srv, http.MethodPut, tc.path, "{")
			}
			assertStatus(t, resp, http.StatusBadRequest)
			assertErrorCode(t, resp, "INVALID_JSON")
		})
	}
}

func TestSearchInvalidQuery(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	srv := newWorkspaceServer(t, root)

	fileResp := doRequest(t, srv, http.MethodGet, "/api/search/files?query=%20%20%20", "")
	assertStatus(t, fileResp, http.StatusBadRequest)
	assertErrorCode(t, fileResp, "INVALID_QUERY")

	textResp := doRequest(t, srv, http.MethodPost, "/api/search/text", `{"query":"   "}`)
	assertStatus(t, textResp, http.StatusBadRequest)
	assertErrorCode(t, textResp, "INVALID_QUERY")
}

func TestFilesInvalidPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	srv := newWorkspaceServer(t, root)

	listResp := doRequest(t, srv, http.MethodGet, "/api/files/list?path=../outside", "")
	assertStatus(t, listResp, http.StatusBadRequest)
	assertErrorCode(t, listResp, "INVALID_PATH")

	contentResp := doRequest(t, srv, http.MethodGet, "/api/file/content?path=/tmp/demo.txt", "")
	assertStatus(t, contentResp, http.StatusBadRequest)
	assertErrorCode(t, contentResp, "INVALID_PATH")
}

func TestWorkspaceDirectoriesErrors(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	filePath := filepath.Join(root, "file.txt")
	if err := os.WriteFile(filePath, []byte("demo"), 0o644); err != nil {
		t.Fatalf("WriteFile() error = %v", err)
	}

	srv := New(Config{Logger: log.New(io.Discard, "", 0)})

	notFoundPath := strings.ReplaceAll(filepath.ToSlash(filepath.Join(root, "missing")), " ", "%20")
	notFoundResp := doRequest(t, srv, http.MethodGet, "/api/workspace/directories?path="+notFoundPath, "")
	assertStatus(t, notFoundResp, http.StatusNotFound)
	assertErrorCode(t, notFoundResp, "NOT_FOUND")

	invalidResp := doRequest(t, srv, http.MethodGet, "/api/workspace/directories?path="+strings.ReplaceAll(filepath.ToSlash(filePath), " ", "%20"), "")
	assertStatus(t, invalidResp, http.StatusBadRequest)
	assertErrorCode(t, invalidResp, "INVALID_PATH")
}

func TestTerminalRouteErrors(t *testing.T) {
	t.Parallel()

	srv := New(Config{Logger: log.New(io.Discard, "", 0)})

	notFoundResp := doRequest(t, srv, http.MethodGet, "/api/terminals/unknown/unknown", "")
	assertStatus(t, notFoundResp, http.StatusNotFound)
	assertErrorCode(t, notFoundResp, "NOT_FOUND")

	methodResp := doRequest(t, srv, http.MethodGet, "/api/terminals/unknown", "")
	assertStatus(t, methodResp, http.StatusMethodNotAllowed)
	assertErrorCode(t, methodResp, "METHOD_NOT_ALLOWED")

	resizeResp := doRequest(t, srv, http.MethodPost, "/api/terminals/unknown/resize", `{"cols":80,"rows":24}`)
	assertStatus(t, resizeResp, http.StatusNotFound)
	assertErrorCode(t, resizeResp, "TERMINAL_NOT_FOUND")
}

func newWorkspaceServer(t *testing.T, root string) *Server {
	t.Helper()
	srv := New(Config{Logger: log.New(io.Discard, "", 0)})
	if _, err := srv.workspaceService.Set(root); err != nil {
		t.Fatalf("workspaceService.Set() error = %v", err)
	}
	return srv
}

func doRequest(t *testing.T, srv *Server, method, path, body string) *httptest.ResponseRecorder {
	t.Helper()
	var reader io.Reader
	if body != "" {
		reader = bytes.NewBufferString(body)
	}
	request := httptest.NewRequest(method, path, reader)
	if body != "" {
		request.Header.Set("Content-Type", "application/json")
	}
	response := httptest.NewRecorder()
	srv.httpServer.Handler.ServeHTTP(response, request)
	return response
}

func assertStatus(t *testing.T, resp *httptest.ResponseRecorder, want int) {
	t.Helper()
	if resp.Code != want {
		t.Fatalf("status = %d, want %d, body=%s", resp.Code, want, resp.Body.String())
	}
}

func assertSuccess(t *testing.T, resp *httptest.ResponseRecorder) {
	t.Helper()
	var payload struct {
		Success bool `json:"success"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; body=%s", err, resp.Body.String())
	}
	if !payload.Success {
		t.Fatalf("expected success response, body=%s", resp.Body.String())
	}
}

func assertErrorCode(t *testing.T, resp *httptest.ResponseRecorder, want string) {
	t.Helper()
	var payload struct {
		Success bool `json:"success"`
		Error   struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(resp.Body.Bytes(), &payload); err != nil {
		t.Fatalf("json.Unmarshal() error = %v; body=%s", err, resp.Body.String())
	}
	if payload.Success {
		t.Fatalf("expected error response, body=%s", resp.Body.String())
	}
	if payload.Error.Code != want {
		t.Fatalf("error code = %q, want %q, body=%s", payload.Error.Code, want, resp.Body.String())
	}
}

func assertJSONContains(t *testing.T, body, needle string) {
	t.Helper()
	if !strings.Contains(body, needle) {
		t.Fatalf("expected body to contain %q, got %s", needle, body)
	}
}
