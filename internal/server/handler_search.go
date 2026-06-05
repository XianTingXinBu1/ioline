package server

import (
	"encoding/json"
	"errors"
	"net/http"

	"ioline/internal/api"
	"ioline/internal/search"
)

func (s *Server) handleSearchFiles(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		api.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	rootPath, ok := s.requireWorkspace(w)
	if !ok {
		return
	}

	result, err := s.searchService.FindFiles(rootPath, r.URL.Query().Get("query"))
	if err != nil {
		if errors.Is(err, search.ErrQueryRequired) {
			api.WriteError(w, http.StatusBadRequest, "INVALID_QUERY", err.Error())
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	api.WriteJSON(w, http.StatusOK, result)
}

func (s *Server) handleSearchText(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		api.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
		return
	}

	rootPath, ok := s.requireWorkspace(w)
	if !ok {
		return
	}

	var payload search.TextSearchRequest
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		api.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON request body")
		return
	}

	result, err := s.searchService.FindText(rootPath, payload)
	if err != nil {
		if errors.Is(err, search.ErrQueryRequired) {
			api.WriteError(w, http.StatusBadRequest, "INVALID_QUERY", err.Error())
			return
		}
		api.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
		return
	}

	api.WriteJSON(w, http.StatusOK, result)
}
