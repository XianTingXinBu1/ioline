package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ioline/internal/server"
)

func main() {
	logger := log.New(os.Stdout, "[ioline] ", log.LstdFlags)

	addr := os.Getenv("IOLINE_SERVER_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	app := server.New(server.Config{
		Addr:   addr,
		Logger: logger,
	})

	go func() {
		logger.Printf("server listening on %s", app.Addr())
		if err := app.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatalf("server error: %v", err)
		}
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	<-sigCh

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := app.Shutdown(ctx); err != nil {
		logger.Printf("shutdown error: %v", err)
	}
}
