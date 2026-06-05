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

	app := server.New(server.Config{
		Addr:   ":8080",
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
