package api

import (
	"encoding/json"
	"net/http"
)

// Error describes a stable API error payload.
type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Response is the standard JSON envelope for all HTTP APIs.
type Response struct {
	Success bool   `json:"success"`
	Data    any    `json:"data,omitempty"`
	Error   *Error `json:"error,omitempty"`
}

// WriteJSON writes a success response as JSON.
func WriteJSON(w http.ResponseWriter, status int, data any) {
	write(w, status, Response{Success: true, Data: data})
}

// WriteError writes an error response as JSON.
func WriteError(w http.ResponseWriter, status int, code, message string) {
	write(w, status, Response{
		Success: false,
		Error:   &Error{Code: code, Message: message},
	})
}

func write(w http.ResponseWriter, status int, payload Response) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
