package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type server struct {
	Port            string        `toml:"port" env:"PORT" env-default:"8080"`
	Host            string        `toml:"host" env:"HOST" env-default:"localhost"`
	ReadTimeout     time.Duration `toml:"read_timeout" env-default:"5s"`
	WriteTimeout    time.Duration `toml:"write_timeout" env-default:"5s"`
	IdleTimeout     time.Duration `toml:"idle_timeout" env-default:"60s"`
	ShutdownTimeout time.Duration `toml:"shutdown_timeout" env-default:"30s"`
}

type database struct {
	URL             string        `toml:"url"`
	MaxConnections  int32         `toml:"max_connections" env-default:"25"`
	MinConnections  int32         `toml:"min_connections" env-default:"5"`
	Timeout         time.Duration `toml:"timeout"  env-default:"30s"`
	MaxConnLifetime time.Duration `toml:"max_conn_lifetime" env-default:"1h"`
	MaxConnIdleTime time.Duration `toml:"max_conn_idle_time" env-default:"30m"`
	RunMigrations   bool          `toml:"run_migrations" env:"DATABASE_RUN_MIGRATIONS" env-default:"true"`
	SSLMode         string        `toml:"ssl_mode" env-default:"disable"`
}

type app struct {
	Environment string `toml:"environment"`
	Debug       bool   `toml:"debug"`
	LogLevel    string `toml:"log_level"`
	LogFormat   string `toml:"log_format"`
}

type security struct {
	TrustedProxies []string `toml:"trusted_proxies"`
	EnableCORS     bool     `toml:"enable_cors"`
	AllowedOrigins []string `toml:"allowed_origins"`
}

type Config struct {
	Server   server   `toml:"server"`
	Database database `toml:"database"`
	App      app      `toml:"app"`
	Security security `toml:"security"`
}

func MustLoad() *Config {
	configPath := os.Getenv("CONFIG_FILE")
	if configPath == "" {
		log.Fatal("CONFIG_FILE is not set")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config file: %s", configPath)
	}

	return &cfg
}