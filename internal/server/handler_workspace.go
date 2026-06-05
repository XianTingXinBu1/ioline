package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"ioline/internal/api"
)

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
	case http.MethodDelete:
		api.WriteJSON(w, http.StatusOK, s.workspaceService.Clear())
	default:
		api.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
	}
}

func (s *Server) handleWorkspaceCandidates(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	api.WriteJSON(w, http.StatusOK, map[string]any{"items": s.workspaceService.Candidates()})
}

func (s *Server) handleWorkspaceDirectories(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	result, err := s.workspaceService.BrowseDirectories(r.URL.Query().Get("path"))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			api.WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		api.WriteError(w, http.StatusBadRequest, "INVALID_PATH", err.Error())
		return
	}

	api.WriteJSON(w, http.StatusOK, result)
}

func (s *Server) requireWorkspace(w http.ResponseWriter) (string, bool) {
	rootPath, err := s.workspaceService.RootPath()
	if err != nil {
		api.WriteError(w, http.StatusBadRequest, "WORKSPACE_NOT_CONFIGURED", "workspace is not configured")
		return "", false
	}
	return rootPath, true
}
