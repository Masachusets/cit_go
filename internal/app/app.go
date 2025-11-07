package app

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Masachusets/cit_go/config"
	"github.com/go-chi/chi/v5"
)

func Run(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {
	router := chi.NewRouter()
	
	server := http.Server{
		Addr: cfg.Server.Host + ":" + cfg.Server.Port,
		Handler: router,
	}

	go func() {
		<-ctx.Done()
		logger.Info("shutting down server")
		server.Shutdown(ctx)
	}()

	logger.Info(
		"starting server", 
		slog.String("addr", server.Addr),
	)

	if err := server.ListenAndServe(); err != nil {
		return err
	}

	return nil
}