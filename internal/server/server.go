package server

import (
	"context"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/gorilla/websocket"

	"ioline/internal/api"
	"ioline/internal/files"
	"ioline/internal/search"
	"ioline/internal/terminal"
	"ioline/internal/workspace"
)

// Config defines the minimum configuration for the HTTP server.
type Config struct {
	Addr   string
	Logger *log.Logger
}

// Server wraps the standard library HTTP server to keep startup wiring isolated.
type Server struct {
	httpServer       *http.Server
	logger           *log.Logger
	workspaceService *workspace.Service
	fileService      *files.Service
	searchService    *search.Service
	terminalService  *terminal.Service
	upgrader         websocket.Upgrader
}

// New creates a minimal HTTP server instance for the ioline backend.
func New(cfg Config) *Server {
	logger := cfg.Logger
	if logger == nil {
		logger = log.Default()
	}

	addr := cfg.Addr
	if addr == "" {
		addr = ":9650"
	}

	s := &Server{
		logger:           logger,
		workspaceService: workspace.NewService(),
		fileService:      files.NewService(),
		searchService:    search.NewService(),
		terminalService:  terminal.NewService(),
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
		},
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/healthz", s.handleHealthz)
	mux.HandleFunc("/api/system/info", s.handleSystemInfo)
	mux.HandleFunc("/api/workspace/current", s.handleWorkspaceCurrent)
	mux.HandleFunc("/api/workspace/directories", s.handleWorkspaceDirectories)
	mux.HandleFunc("/api/workspaces/candidates", s.handleWorkspaceCandidates)
	mux.HandleFunc("/api/files/list", s.handleFilesList)
	mux.HandleFunc("/api/search/files", s.handleSearchFiles)
	mux.HandleFunc("/api/search/text", s.handleSearchText)
	mux.HandleFunc("/api/files/stat", s.handleFilesStat)
	mux.HandleFunc("/api/files", s.handleFiles)
	mux.HandleFunc("/api/files/move", s.handleFilesMove)
	mux.HandleFunc("/api/directories", s.handleDirectories)
	mux.HandleFunc("/api/file/content", s.handleFileContent)
	mux.HandleFunc("/api/terminals", s.handleTerminals)
	mux.HandleFunc("/api/terminals/", s.handleTerminalSessionRoutes)

	s.httpServer = &http.Server{
		Addr:              addr,
		Handler:           s.logRequests(mux),
		ReadHeaderTimeout: 5 * time.Second,
	}

	return s
}

// Addr returns the configured listen address.
func (s *Server) Addr() string {
	return s.httpServer.Addr
}

// ListenAndServe starts the HTTP server.
func (s *Server) ListenAndServe() error {
	s.logger.Printf("http server starting")
	return s.httpServer.ListenAndServe()
}

// Shutdown gracefully stops the HTTP server.
func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Printf("http server shutting down")
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func (s *Server) handleSystemInfo(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]any{
		"name":                "ioline",
		"goVersion":           runtime.Version(),
		"os":                  runtime.GOOS,
		"arch":                runtime.GOARCH,
		"termux":              isTermuxEnvironment(),
		"workspaceSet":        s.workspaceService.Current().IsSet,
		"terminalMaxSessions": 4,
	})
}

func (s *Server) logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		startedAt := time.Now()
		next.ServeHTTP(w, r)
		s.logger.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(startedAt))
	})
}

func isTermuxEnvironment() bool {
	prefix := os.Getenv("PREFIX")
	home := os.Getenv("HOME")
	return strings.Contains(prefix, "/com.termux/") || strings.Contains(home, "/com.termux/")
}

type ptyAccessor interface {
	Read([]byte) (int, error)
	Write([]byte) (int, error)
}

func (s *Server) copyPTYToWriter(src ptyAccessor, dst io.Writer) error {
	_, err := io.Copy(dst, src)
	return err
}
