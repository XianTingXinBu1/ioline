package server

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"
	"time"

	"ioline/internal/api"
	"ioline/internal/files"
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
}

// New creates a minimal HTTP server instance for the ioline backend.
func New(cfg Config) *Server {
	logger := cfg.Logger
	if logger == nil {
		logger = log.Default()
	}

	addr := cfg.Addr
	if addr == "" {
		addr = ":8080"
	}

	s := &Server{
		logger:           logger,
		workspaceService: workspace.NewService(),
		fileService:      files.NewService(),
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/api/healthz", s.handleHealthz)
	mux.HandleFunc("/api/system/info", s.handleSystemInfo)
	mux.HandleFunc("/api/workspace/current", s.handleWorkspaceCurrent)
	mux.HandleFunc("/api/files/list", s.handleFilesList)
	mux.HandleFunc("/api/files/stat", s.handleFilesStat)
	mux.HandleFunc("/api/files", s.handleFiles)
	mux.HandleFunc("/api/files/move", s.handleFilesMove)
	mux.HandleFunc("/api/directories", s.handleDirectories)
	mux.HandleFunc("/api/file/content", s.handleFileContent)

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
		"name":         "ioline",
		"goVersion":    runtime.Version(),
		"os":           runtime.GOOS,
		"arch":         runtime.GOARCH,
		"termux":       isTermuxEnvironment(),
		"workspaceSet": s.workspaceService.Current().IsSet,
	})
}

func (s *Server) handleWorkspaceCurrent(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		api.WriteJSON(w, http.StatusOK, s.workspaceService.Current())
	case http.MethodPut:
		var payload struct {
			RootPath string `json:"rootPath"`
		}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			api.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON request body")
			return
		}

		info, err := s.workspaceService.Set(payload.RootPath)
		if err != nil {
			api.WriteError(w, http.StatusBadRequest, "INVALID_WORKSPACE", err.Error())
			return
		}
		api.WriteJSON(w, http.StatusOK, info)
	default:
		api.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
	}
}

func (s *Server) handleFilesList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	rootPath, ok := s.requireWorkspace(w)
	if !ok {
		return
	}

	items, err := s.fileService.List(rootPath, r.URL.Query().Get("path"))
	if err != nil {
		s.writeFileError(w, err)
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]any{"items": items})
}

func (s *Server) handleFilesStat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	rootPath, ok := s.requireWorkspace(w)
	if !ok {
		return
	}

	entry, err := s.fileService.Stat(rootPath, r.URL.Query().Get("path"))
	if err != nil {
		s.writeFileError(w, err)
		return
	}

	api.WriteJSON(w, http.StatusOK, entry)
}

func (s *Server) handleFiles(w http.ResponseWriter, r *http.Request) {
	rootPath, ok := s.requireWorkspace(w)
	if !ok {
		return
	}

	switch r.Method {
	case http.MethodPost:
		var payload files.CreateFileRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			api.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON request body")
			return
		}
		result, err := s.fileService.CreateFile(rootPath, payload)
		if err != nil {
			s.writeFileError(w, err)
			return
		}
		api.WriteJSON(w, http.StatusCreated, result)
	case http.MethodDelete:
		var payload files.DeleteRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			api.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON request body")
			return
		}
		result, err := s.fileService.Delete(rootPath, payload)
		if err != nil {
			s.writeFileError(w, err)
			return
		}
		api.WriteJSON(w, http.StatusOK, result)
	default:
		api.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
	}
}

func (s *Server) handleDirectories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	rootPath, ok := s.requireWorkspace(w)
	if !ok {
		return
	}

	var payload files.CreateDirectoryRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		api.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON request body")
		return
	}
	result, err := s.fileService.CreateDirectory(rootPath, payload)
	if err != nil {
		s.writeFileError(w, err)
		return
	}
	api.WriteJSON(w, http.StatusCreated, result)
}

func (s *Server) handleFilesMove(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPatch {
		api.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	rootPath, ok := s.requireWorkspace(w)
	if !ok {
		return
	}

	var payload files.MoveRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		api.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON request body")
		return
	}
	result, err := s.fileService.Move(rootPath, payload)
	if err != nil {
		s.writeFileError(w, err)
		return
	}
	api.WriteJSON(w, http.StatusOK, result)
}

func (s *Server) handleFileContent(w http.ResponseWriter, r *http.Request) {
	rootPath, ok := s.requireWorkspace(w)
	if !ok {
		return
	}

	switch r.Method {
	case http.MethodGet:
		content, err := s.fileService.ReadText(rootPath, r.URL.Query().Get("path"))
		if err != nil {
			s.writeFileError(w, err)
			return
		}
		api.WriteJSON(w, http.StatusOK, content)
	case http.MethodPut:
		var payload files.SaveRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			api.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON request body")
			return
		}
		result, err := s.fileService.SaveText(rootPath, payload)
		if err != nil {
			s.writeFileError(w, err)
			return
		}
		api.WriteJSON(w, http.StatusOK, result)
	default:
		api.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
	}
}

func (s *Server) requireWorkspace(w http.ResponseWriter) (string, bool) {
	rootPath, err := s.workspaceService.RootPath()
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "WORKSPACE_NOT_CONFIGURED", "workspace is not configured")
		return "", false
	}
	return rootPath, true
}

func (s *Server) writeFileError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, files.ErrWorkspaceNotConfigured), errors.Is(err, workspace.ErrNotConfigured):
		api.WriteError(w, http.StatusBadRequest, "WORKSPACE_NOT_CONFIGURED", err.Error())
	case errors.Is(err, files.ErrInvalidPath), errors.Is(err, files.ErrPathRequired), errors.Is(err, files.ErrSourceDestinationEqual):
		api.WriteError(w, http.StatusBadRequest, "INVALID_PATH", err.Error())
	case errors.Is(err, files.ErrNotRegularFile):
		api.WriteError(w, http.StatusBadRequest, "NOT_REGULAR_FILE", err.Error())
	case errors.Is(err, files.ErrBinaryFile):
		api.WriteError(w, http.StatusBadRequest, "UNSUPPORTED_FILE", err.Error())
	case errors.Is(err, files.ErrFileTooLarge):
		api.WriteError(w, http.StatusBadRequest, "FILE_TOO_LARGE", err.Error())
	case errors.Is(err, files.ErrAlreadyExists):
		api.WriteError(w, http.StatusConflict, "ALREADY_EXISTS", err.Error())
	case errors.Is(err, files.ErrDirectoryNotEmpty), errors.Is(err, fs.ErrNotExist):
		api.WriteError(w, http.StatusBadRequest, "DIRECTORY_NOT_EMPTY", err.Error())
	case errors.Is(err, os.ErrNotExist):
		api.WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	default:
		api.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
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
