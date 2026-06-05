package server

import (
	"context"
	"log"
	"net/http"
)

// Config defines the minimum configuration for the HTTP server.
type Config struct {
	Addr   string
	Logger *log.Logger
}

// Server wraps the standard library HTTP server to keep startup wiring isolated.
type Server struct {
	httpServer *http.Server
	logger     *log.Logger
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

	mux := http.NewServeMux()
	mux.HandleFunc("/healthz", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	return &Server{
		httpServer: &http.Server{
			Addr:    addr,
			Handler: mux,
		},
		logger: logger,
	}
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
