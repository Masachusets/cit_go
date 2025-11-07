package config

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/knadh/koanf/parsers/dotenv"
	"github.com/knadh/koanf/parsers/toml"
	"github.com/knadh/koanf/providers/confmap"
	"github.com/knadh/koanf/providers/basicflag"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
)

type server struct {
	Port            string        `koanf:"port"`
	Host            string        `koanf:"host"`
	ReadTimeout     time.Duration `koanf:"read_timeout"`
	WriteTimeout    time.Duration `koanf:"write_timeout"`
	ShutdownTimeout time.Duration `koanf:"shutdown_timeout"`
}

type database struct {
	URL             string        `koanf:"url"`
	MaxConnections  int32         `koanf:"max_connections"`
	MinConnections  int32         `koanf:"min_connections"`
	Timeout         time.Duration `koanf:"timeout"`
	MaxConnLifetime time.Duration `koanf:"max_conn_lifetime"`
	MaxConnIdleTime time.Duration `koanf:"max_conn_idle_time"`
	RunMigrations   bool          `koanf:"run_migrations"`
	SSLMode         string        `koanf:"ssl_mode"`
}

type app struct {
	Environment string `koanf:"environment"`
	Debug       bool   `koanf:"debug"`
	LogLevel    string `koanf:"log_level"`
	LogFormat   string `koanf:"log_format"`
}

type security struct {
	TrustedProxies []string `koanf:"trusted_proxies"`
	EnableCORS    bool     `koanf:"enable_cors"`
	AllowedOrigins []string `koanf:"allowed_origins"`
}

type Config struct {
	Server   server   `koanf:"server"`
	Database database `koanf:"database"`
	App      app      `koanf:"app"`
	Security security `koanf:"security"`
}

func New() *Config {
	k := koanf.New(".")

	// Read config file
	configFiles := []string{
		"config.toml",
		"./config/config.toml",
	}
	configLoaded := false
	for _, configFile := range configFiles {
		if err := k.Load(file.Provider(configFile), toml.Parser()); err == nil {
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


	// Read .env file
	if err := k.Load(file.Provider(".env"), dotenv.Parser()); err != nil {
		slog.Error(".env file not found")
	} else {
		slog.Debug("load from .env file")
		transformEnvKeys(k)
	}

	// Read flags
	f := setFromFlags(k)

	if err := k.Load(basicflag.Provider(f, "."), nil); err != nil {
		slog.Error("flags not found")
	} else {
		slog.Debug("load from flags")
	}

	// Environment variable handling
	// err := k.Load(
	// 	env.Provider(
	// 		"",
	// 		".",
	// 		func(s string) string {
	// 			return strings.ReplaceAll(strings.ToLower(s), "_", ".")
	// 		},
	// 	),
	// 	nil,
	// )
	// if err != nil {
	// 	slog.Debug("failed to load environment variables", "error", err)
	// } else {
	// 	slog.Debug(
	// 		"loaded environment variables",
	// 	)
	// }

	// Unmarshal config
	var cfg Config
	err := k.UnmarshalWithConf("", &cfg, koanf.UnmarshalConf{Tag: "koanf"})
	if err != nil {
		slog.Error("failed to unmarshal config", "error", err)
		os.Exit(1)
	}

	// Construct database URL if not provided directly
	if cfg.Database.URL == "" {
		cfg.Database.URL = constructDatabaseURL(k)
	}

	// Production overrides
	if cfg.App.Environment == "production" {
		cfg.App.Debug = false
		cfg.App.LogFormat = "json"
		cfg.Security.AllowedOrigins = []string{}
		cfg.Database.RunMigrations = false
	}

	return &cfg
}

// Construct database URL if not provided directly
func constructDatabaseURL(k *koanf.Koanf) string {
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

	if user == "" || password == "" {
		slog.Error("DB_URL not provided and DB_USER/DB_PASSWORD not found in environment")
		os.Exit(1)
	}
	
	url := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=%s",
		user,
		password,
		host,
		port,
		name,
		sslmode,
	)
	slog.Debug("DB_URL", "DB.url", url)
	return url
}

// setFromFlags read settings from flags
func setFromFlags(k *koanf.Koanf) *flag.FlagSet {
	f := flag.NewFlagSet("config", flag.ExitOnError)

	f.String("server.port", k.String("server.port"), "Server port")
	f.String("server.host", k.String("server.host"), "Server host")

	f.String("database.url", k.String("database.url"), "Database URL")
	f.String("database.host", k.String("database.host"), "Database host")
	f.Int("database.port", k.Int("database.port"), "Database port")
	f.String("database.name", k.String("database.name"), "Database name")
	f.Bool("database.run_migrations", k.Bool("database.run_migrations"), "Run migrations")

	f.String("app.enviromment", k.String("app.environment"), "Environment")
	f.Bool("app.debug", k.Bool("app.debug"), "Debug mode")
	f.String("app.log_level", k.String("app.log_level"), "Log level")

	f.String("security.trusted_proxies", k.String("security.trusted_proxies"), "Trusted proxies")
	f.String("security.allowed_origins", k.String("security.allowed_origins"), "Allowed CORS origins")

	if err := f.Parse(os.Args[1:]); err != nil {
		slog.Error("failed to parse flags", "error", err)
		os.Exit(1)
	}

	return f
}

// transformEnvKeys transforms environment variable keys from .env file
// from format DATABASE_USER to database.user
func transformEnvKeys(k *koanf.Koanf) {
	// Get all keys from koanf
	allKeys := k.Keys()

	// Create a map to store transformed keys
	transformations := make(map[string]interface{})

	for _, key := range allKeys {
		// Check if key is uppercase (likely from .env file)
		upperKey := strings.ToUpper(key)
		if upperKey == key && strings.Contains(key, "_") {
			// Transform: DATABASE_USER -> database.user
			transformedKey := strings.Replace(strings.ToLower(key), "_", ".", 1)

			// Only transform if the transformed key doesn't already exist
			// (to avoid overwriting already correctly formatted keys)
			if !k.Exists(transformedKey) {
				value := k.Get(key)
				transformations[transformedKey] = value
			}
		}
	}

	// Apply transformations
	if len(transformations) > 0 {
		k.Load(confmap.Provider(transformations, "."), nil)
	}
}
