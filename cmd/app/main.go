package main

import (
	"context"
	"log"
	"os"
	"os/signal"

	"github.com/Masachusets/cit_go/config"
	"github.com/Masachusets/cit_go/internal/app"
	"github.com/Masachusets/cit_go/pkg/logger"
)
func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg := config.New()

	logger := logger.SetupLogger(cfg.App.Debug, cfg.App.LogLevel, cfg.App.LogFormat)

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()

	if err := app.Run(ctx, cfg, logger); err != nil {
		return err
	}

	return nil
}