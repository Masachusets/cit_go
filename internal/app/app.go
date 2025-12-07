package app

import (
	"context"
	"log/slog"

	"github.com/Masachusets/cit_go/config"
	"github.com/Masachusets/cit_go/internal/adapter/postgres"
	"github.com/Masachusets/cit_go/internal/server"
)

func Run(ctx context.Context, cfg *config.Config, logger *slog.Logger) error {
	// Postgres
	pgPool, err := postgres.New(ctx, cfg)
	if err != nil {
		logger.Error(
			"failed to connect to database",
			"error", err,
			"database_url", cfg.Database.URL,
		)
		return err
	}

	defer func() {
		pgPool.Close()
		logger.Info("Database connection pool closed")
	}()

	logger.Info("DB Postgres is connected")
	
	server := server.NewServer(ctx, cfg, logger, pgPool)

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