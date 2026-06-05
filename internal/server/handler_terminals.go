package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/websocket"

	"ioline/internal/api"
	"ioline/internal/terminal"
)

func (s *Server) handleTerminals(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		api.WriteJSON(w, http.StatusOK, map[string]any{"items": s.terminalService.List()})
	case http.MethodPost:
		rootPath, ok := s.requireWorkspace(w)
		if !ok {
			return
		}
		var payload terminal.CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			api.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON request body")
			return
		}
		result, err := s.terminalService.Create(rootPath, payload)
		if err != nil {
			s.writeTerminalError(w, err)
			return
		}
		api.WriteJSON(w, http.StatusCreated, result)
	default:
		api.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
	}
}

func (s *Server) handleTerminalSessionRoutes(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/api/terminals/")
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) == 0 || parts[0] == "" {
		api.WriteError(w, http.StatusNotFound, "NOT_FOUND", "terminal route not found")
		return
	}

	id := parts[0]
	if len(parts) == 1 {
		if r.Method != http.MethodDelete {
			api.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
			return
		}
		if err := s.terminalService.Close(id); err != nil {
			s.writeTerminalError(w, err)
			return
		}
		api.WriteJSON(w, http.StatusOK, map[string]string{"id": id, "status": "closed"})
		return
	}

	switch parts[1] {
	case "resize":
		if r.Method != http.MethodPost {
			api.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
			return
		}
		var payload terminal.ResizeRequest
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			api.WriteError(w, http.StatusBadRequest, "INVALID_JSON", "invalid JSON request body")
			return
		}
		if err := s.terminalService.Resize(id, payload); err != nil {
			s.writeTerminalError(w, err)
			return
		}
		api.WriteJSON(w, http.StatusOK, map[string]any{"id": id, "cols": payload.Cols, "rows": payload.Rows})
	case "stream":
		if r.Method != http.MethodGet {
			api.WriteError(w, http.StatusMethodNotAllowed, "METHOD_NOT_ALLOWED", "method not allowed")
			return
		}
		s.handleTerminalStream(w, r, id)
	default:
		api.WriteError(w, http.StatusNotFound, "NOT_FOUND", "terminal route not found")
	}
}

func (s *Server) handleTerminalStream(w http.ResponseWriter, r *http.Request, id string) {
	session, err := s.terminalService.Get(id)
	if err != nil {
		s.writeTerminalError(w, err)
		return
	}

	conn, err := s.upgrader.Upgrade(w, r, nil)
	if err != nil {
		s.logger.Printf("websocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	errCh := make(chan error, 2)

	go func() {
		buffer := make([]byte, 4096)
		for {
			n, readErr := session.PTY().Read(buffer)
			if n > 0 {
				if writeErr := conn.WriteMessage(websocket.TextMessage, buffer[:n]); writeErr != nil {
					errCh <- writeErr
					return
				}
			}
			if readErr != nil {
				errCh <- readErr
				return
			}
		}
	}()

	go func() {
		for {
			messageType, payload, readErr := conn.ReadMessage()
			if readErr != nil {
				errCh <- readErr
				return
			}
			if messageType != websocket.TextMessage && messageType != websocket.BinaryMessage {
				continue
			}
			if _, writeErr := session.PTY().Write(payload); writeErr != nil {
				errCh <- writeErr
				return
			}
		}
	}()

	<-errCh
}

func (s *Server) writeTerminalError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, terminal.ErrSessionLimitReached):
		api.WriteError(w, http.StatusConflict, "TERMINAL_LIMIT_REACHED", err.Error())
	case errors.Is(err, terminal.ErrSessionNotFound):
		api.WriteError(w, http.StatusNotFound, "TERMINAL_NOT_FOUND", err.Error())
	case errors.Is(err, terminal.ErrInvalidSize):
		api.WriteError(w, http.StatusBadRequest, "INVALID_TERMINAL_SIZE", err.Error())
	default:
		api.WriteError(w, http.StatusInternalServerError, "INTERNAL_ERROR", fmt.Sprintf("terminal error: %v", err))
	}
}
