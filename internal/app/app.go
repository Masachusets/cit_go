package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
)

var (
	serverAddr = "localhost:4000"
)

func Run(ctx context.Context) error {
	router := chi.NewRouter()
	
	server := http.Server{
		Addr: serverAddr,
		Handler: router,
	}

	go func() {
		<-ctx.Done()
		slog.Info("shutting down server")
		server.Shutdown(ctx)
	}()

	slog.Info(
		"starting server", 
		slog.String("addr", serverAddr),
	)

	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}