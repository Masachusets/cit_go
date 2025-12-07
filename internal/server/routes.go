package server

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"

	// "github.com/go-chi/httplog/v3"
	"github.com/jackc/pgx/v5/pgxpool"
)

func (s *Server) RegisterRoutes(ctx context.Context, log *slog.Logger, db *pgxpool.Pool) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	// r.Use(
	// 	httplog.RequestLogger(
	// 		log, 
	// 		&httplog.Options{
	// 			Level: slog.LevelInfo,
	// 			Schema: httplog.SchemaECS.Concise(true),
	// 			RecoverPanics: true,
	// 		},
	// 	),
	// )
	r.Use(middleware.Heartbeat("/ping"))

	r.Use(
		cors.Handler(
			cors.Options{
				AllowedOrigins:   s.cfg.Security.AllowedOrigins,
				AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
				AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
				AllowCredentials: true,
				MaxAge:           300,
				Debug:            s.cfg.App.Debug,
			},
		),
	)

	// fileServer := http.Fi

	r.Get(
		"/", 
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Главная страница"))
    	},
	)
	
	return r
}