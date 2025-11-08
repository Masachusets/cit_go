package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Masachusets/cit_go/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Server struct {
	adrress string
	logger  *slog.Logger
	db      *pgxpool.Pool
}

func NewServer(ctx context.Context, cfg *config.Config, log *slog.Logger, pgpool *pgxpool.Pool) *http.Server {
	NewServ := &Server{
		adrress: cfg.Server.Host + ":" + cfg.Server.Port,
		logger:  log,
		db:      pgpool,
	}

	return &http.Server{
		Addr:         NewServ.adrress,
		Handler:      NewServ.RegisterRoutes(ctx, log, pgpool),
		IdleTimeout:  cfg.Server.IdleTimeout,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}
}
