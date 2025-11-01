package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/env"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type server struct {
	Port            string        `mapstruct:"port"`
	Host            string        `mapstruct:"host"`
	ReadTimeout     time.Duration `mapstruct:"read_timeout"`
	WriteTimeout    time.Duration `mapstruct:"write_timeout"`
	ShutdownTimeout time.Duration `mapstruct:"shutdown_timeout"`
}

type database struct {
	URL             string        `mapstruct:"url"`
	MaxConnections  int32         `mapstruct:"max_connections"`
	MinConnections  int32         `mapstruct:"min_connections"`
	Timeout         time.Duration `mapstruct:"timeout"`
	MaxConnLifetime time.Duration `mapstruct:"max_conn_lifetime"`
	MaxConnIdleTime time.Duration `mapstruct:"max_conn_idle_time"`
	RunMigrations   bool          `mapstruct:"run_migrations"`
	SSLMode         string        `mapstruct:"ssl_mode"`
}

type app struct {
	Environment string `mapstruct:"environment"`
	Debug       bool   `mapstruct:"debug"`
	LogLevel    string `mapstruct:"log_level"`
	LogFormat   string `mapstruct:"log_format"`
}

type security struct {
	TrustedProxies []string `mapstruct:"trusted_proxies"`
	EnambleCORS    bool     `mapstruct:"enable_cors"`
	AllowedOrigins []string `mapstruct:"allowed_origins"`
}

type Config struct {
	Server   server   `mapstruct:"server"`
	DB       database `mapstruct:"database"`
	App      app      `mapstruct:"app"`
	Security security `mapstruct:"security"`
}

func New() *Config {
	k := koanf.New(".")

	// Set defaults settings
	setDefaults(k)

	// Read .env file first
	if err := k.Load(file.Provider(".env"), dotenv.Parser()); err != nil {
		slog.Debug(".env file not found")
	} else {
		slog.Debug("load from .env file")
	}

	// Read config file (optional)
	configFiles := []string{
		"config.yaml",
		"config.yml",
		"./config/config.yaml",
		"./config/config.yml",
	}
	configLoaded := false
	for _, configFile := range configFiles {
		if err := k.Load(file.Provider(configFile), yaml.Parser()); err == nil {
			slog.Info("loaded configuration from file", "file", configFile)
			configLoaded = true
			break
		}
	}
	if !configLoaded {
		slog.Debug(
			"no config file found, using defaults and environment variables",
		)
	}

	// Environment variable handling
	k.Load(
		env.Provider(
			"",
			".",
			func(s string) string {
				return strings.ReplaceAll(strings.ToLower(s), "_", ".")
			},
		),
		nil,
	)

	// Unmarshal config
	var cfg Config
	err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "mapstruct"})
	if err != nil {
		slog.Error("failed to unmarshal config", "error", err)
		os.Exit(1)
	}

	// Construct database URL if not provided directly
	if cfg.DB.URL == "" {
		user := k.String("database.user")
		password := k.String("database.password")
		host := k.String("database.host")
		port := k.String("database.port")
		name := k.String("database.name")
		sslmode := k.String("database.sslmode")

		if host == "" {
			host = "localhost"
		}

		if port == "" {
			port = "5432"
		}

		if name == "" {
			name = "gowebserver"
		}

		if sslmode == "" {
			sslmode = "disable"
		}

		if user != "" && password != "" {
			cfg.DB.URL = fmt.Sprintf(
				"postgres://%s:%s@%s:%s/%s&sslmode=%s",
				user,
				password,
				host,
				port,
				name,
				sslmode,
			)
		} else {
			slog.Error("DB_URL not provided and DB_USER/DB_PASSWORD not found in environment")
			os.Exit(1)
		}
	}

	// Production overrides
	if cfg.App.Environment == "production" {
		cfg.App.Debug = false
		cfg.App.LogFormat = "json"
		cfg.Security.AllowedOrigins = []string{}
		cfg.DB.RunMigrations = false
	}

	return &cfg
}

func setDefaults(k *koanf.Koanf) {
	defaults := map[string]interface{}{
		// Server defaults
		"server.port":             "8080",
		"server.host":             "",
		"server.read_timeout":     10 * time.Second,
		"server.write_timeout":    10 * time.Second,
		"server.shutdown_timeout": 30 * time.Second,

		// Database defaults - will be overridden by environment variables
		"database.url":                "", // Will be constructed from individual vars if not set
		"database.max_connections":    25,
		"database.min_connections":    5,
		"database.timeout":            30 * time.Second,
		"database.max_conn_lifetime":  time.Hour,
		"database.max_conn_idle_time": 30 * time.Minute,
		"database.run_migrations":     true,
		"database.ssl_mode":           "disable",

		// Application defaults
		"app.environment": "development",
		"app.debug":       false,
		"app.log_level":   "info",
		"app.log_format":  "text",

		// Security defaults
		"security.trusted_proxies": []string{"127.0.0.1"},
		"security.enable_cors":     true,
		"security.allowed_origins": []string{"*"},
	}

	k.Load(confmap.Provider(defaults, "."), nil)
}

// GetLogLevel converts the string log level to slog.Level.
func (c *Config) GetLogLevel() slog.Level {
	switch strings.ToLower(c.App.LogLevel) {
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
