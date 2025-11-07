package logger

import (
	"log/slog"
	"os"
	"strings"
)

// SetupLogger configures slog based on provided flags.
// It respects debug, logLevel and logFormat.
func SetupLogger(debug bool, logLevel string, logFormat string) *slog.Logger {
	var level slog.Level

	// If debug is enabled, cap level to debug (most verbose)
	if debug {
		level = slog.LevelDebug
	} else {
		level = getLogLevel(logLevel)
	}

	var handler slog.Handler
	if strings.ToLower(logFormat) == "json"{
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else { // "text" or any unknown -> text
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	log := slog.New(handler)

	// Make it the default logger so slog.Info/... use our config
	slog.SetDefault(log)

	return log
}

func getLogLevel(level string) slog.Level {
	switch strings.ToLower(level) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
