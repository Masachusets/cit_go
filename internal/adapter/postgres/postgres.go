package postgres

import (
	"context"

	"github.com/Masachusets/cit_go/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Pool struct {}

func New(ctx context.Context, cfg *config.Config) (*Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.Database.URL)
	if err != nil {}
	return &Pool{}, nil
}