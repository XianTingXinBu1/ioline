package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"os"

	"ioline/internal/api"
	"ioline/internal/files"
	"ioline/internal/workspace"
)

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
	case errors.Is(err, files.ErrDirectoryNotEmpty):
		api.WriteError(w, http.StatusBadRequest, "DIRECTORY_NOT_EMPTY", err.Error())
	case errors.Is(err, os.ErrNotExist):
		api.WriteError(w, http.StatusNotFound, "NOT_FOUND", err.Error())
	default:
		api.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "internal server error")
	}
}
