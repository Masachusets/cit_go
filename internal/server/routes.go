package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (s *Server) RegisterRoutes(ctx context.Context, log *slog.Logger, db *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	
	return r
}