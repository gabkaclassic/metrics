// Package config provides configuration loading for server and agent applications.
//
// The package supports configuration via environment variables and command-line flags,
// with environment variables used as defaults and flags taking precedence if provided.
//
// Configuration is parsed using github.com/caarlos0/env for environment variables
// and the standard flag package for CLI arguments.
//
// Custom parsers are defined for types such as time.Duration to allow concise
// numeric configuration (values are interpreted as seconds).
//
// The package exposes separate parsing functions for server-side and agent-side
// configuration to keep concerns isolated.
package config

import (
	"flag"
	"net/url"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v10"
)

type (
	// Server represents the full configuration of the metrics server.
	Server struct {
		Address string `env:"ADDRESS" envDefault:"localhost:8080"`
		SignKey string `env:"KEY"`
		Log     Log
		Dump    Dump
		DB      DB
		Audit   Audit
	}
	// Agent represents the configuration of the metrics agent.
	Agent struct {
		PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2"`
		ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10"`
		Client         Client
		Log            Log
		BatchesEnabled bool   `env:"BATCHES" envDefault:"true"`
		SignKey        string `env:"KEY"`
		RateLimit      int    `env:"RATE_LIMIT" envDefault:"5"`
		BatchSize      int    `env:"BATCH_SIZE" envDefault:"100"`
	}
	// DB contains database-related configuration.
	DB struct {
		Driver         string        `env:"DB_DRIVER" envDefault:"postgres"`
		DSN            string        `env:"DATABASE_DSN"`
		MigrationsPath string        `env:"DB_MIGRATIONS_PATH" envDefault:"./migrations"`
		MaxConns       int           `env:"DB_MAX_CONNS" envDefault:"4"`
		MaxConnTTL     time.Duration `env:"DB_MAX_CONN_TTL" envDefault:"60"`
	}
	// Client contains HTTP client configuration used by the agent
	// to communicate with the server.
	Client struct {
		BaseURL string        `env:"ADDRESS" envDefault:"localhost:8080"`
		Timeout time.Duration `env:"TIMEOUT" envDefault:"3"`
		Retries int           `env:"RETRIES" envDefault:"3"`
	}
	// Log defines logging configuration shared between server and agent.
	Log struct {
		Level   string `env:"LOG_LEVEL" envDefault:"info"`
		File    string `env:"LOG_FILE"`
		Console bool   `env:"LOG_CONSOLE" envDefault:"false"`
		JSON    bool   `env:"LOG_JSON" envDefault:"true"`
	}
	// Dump defines configuration for periodic metrics persistence to disk.
	Dump struct {
		StoreInterval   time.Duration `env:"STORE_INTERVAL" envDefault:"300"`
		FileStoragePath string        `env:"FILE_STORAGE_PATH" envDefault:"/tmp/metrics_dumps/dump.json"`
		Restore         bool          `env:"RESTORE" envDefault:"false"`
	}
	// Audit defines configuration for audit logging destinations.
	Audit struct {
		File string `env:"AUDIT_FILE"`
		URL  string `env:"AUDIT_URL"`
	}
)

// ensureURL normalizes an address string into a valid URL.
// If the scheme is missing, "http://" is prepended.
func ensureURL(addr string) string {
	if !strings.Contains(addr, "://") {
		addr = "http://" + addr
	}
	u, err := url.Parse(addr)
	if err != nil {
		return addr
	}
	return u.String()
}

func defineEnvParsers() map[reflect.Type]env.ParserFunc {
	return map[reflect.Type]env.ParserFunc{
		reflect.TypeOf(time.Duration(0)): func(v string) (any, error) {
			secs, err := strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
			return time.Duration(secs) * time.Second, nil
		},
	}
}

// ParseServerConfig parses and returns server configuration.
//
// Configuration values are loaded from environment variables first,
// then overridden by command-line flags if provided.
//
// Supported sources:
//   - Environment variables (via caarlos0/env)
//   - Command-line flags (highest priority)
//
// The function returns a fully populated Server configuration
// or an error if parsing fails.
func ParseServerConfig() (*Server, error) {
	var cfg Server

	parsers := defineEnvParsers()

	if err := env.ParseWithOptions(&cfg, env.Options{FuncMap: parsers}); err != nil {
		return nil, err
	}

	address := flag.String("a", cfg.Address, "HTTP server address")

	logLevel := flag.String("log-level", cfg.Log.Level, "Logging level")
	logFile := flag.String("log-file", cfg.Log.File, "Log file path")
	logConsole := flag.Bool("log-console", cfg.Log.Console, "Enable console logging")
	logJSON := flag.Bool("log-json", cfg.Log.JSON, "Enable JSON output for logs")

	storeInterval := flag.Uint("i", uint(cfg.Dump.StoreInterval.Seconds()), "Store interval")
	fileStoragePath := flag.String("f", cfg.Dump.FileStoragePath, "File storage path")
	restore := flag.Bool("r", cfg.Dump.Restore, "Restore need")

	dbDSN := flag.String("d", cfg.DB.DSN, "DSN")
	dbDriver := flag.String("db-driver", cfg.DB.Driver, "Database driver")
	dbMigrationsPath := flag.String("db-migrations-path", cfg.DB.MigrationsPath, "Migrations file path")
	dbMaxConns := flag.Int("db-max-conns", int(cfg.DB.MaxConns), "Maximum DB connection amount")
	dbMaxConTTL := flag.Uint("db-max-conn-ttl", uint(cfg.DB.MaxConnTTL), "Maximum DB connection TTL")

	auditFile := flag.String("audit-file", cfg.Audit.File, "Audit dump filepath")
	auditURL := flag.String("audit-url", cfg.Audit.URL, "Audit url")

	signKey := flag.String("k", cfg.SignKey, "Key to verify requests bodies")

	flag.Parse()

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			cfg.Address = *address

		case "i":
			cfg.Dump.StoreInterval = time.Duration(*storeInterval) * time.Second
		case "f":
			cfg.Dump.FileStoragePath = *fileStoragePath
		case "r":
			cfg.Dump.Restore = *restore

		case "log-level":
			cfg.Log.Level = *logLevel
		case "log-file":
			cfg.Log.File = *logFile
		case "log-console":
			cfg.Log.Console = *logConsole
		case "log-json":
			cfg.Log.JSON = *logJSON

		case "db-driver":
			cfg.DB.Driver = *dbDriver
		case "d":
			cfg.DB.DSN = *dbDSN
		case "db-migrations-path":
			cfg.DB.MigrationsPath = *dbMigrationsPath
		case "db-max-conns":
			cfg.DB.MaxConns = *dbMaxConns
		case "db-max-conn-ttl":
			cfg.DB.MaxConnTTL = time.Duration(*dbMaxConTTL) * time.Second

		case "audit-file":
			cfg.Audit.File = *auditFile
		case "audit-url":
			cfg.Audit.URL = *auditURL

		case "k":
			cfg.SignKey = *signKey
		}
	})

	return &cfg, nil
}

// ParseAgentConfig parses and returns agent configuration.
//
// Configuration values are loaded from environment variables first,
// then overridden by command-line flags if provided.
//
// The agent client BaseURL is normalized to ensure a valid URL scheme.
//
// The function returns a fully populated Agent configuration
// or an error if parsing fails.
func ParseAgentConfig() (*Agent, error) {
	var cfg Agent

	parsers := defineEnvParsers()

	if err := env.ParseWithOptions(&cfg, env.Options{FuncMap: parsers}); err != nil {
		return nil, err
	}

	pollInterval := flag.Uint("p", uint(cfg.PollInterval.Seconds()), "Metrics polling interval (seconds)")
	reportInterval := flag.Uint("r", uint(cfg.ReportInterval.Seconds()), "Metrics reporting interval (seconds)")
	serverAddress := flag.String("a", cfg.Client.BaseURL, "Server HTTP base URL")
	retries := flag.Int("report-retries", cfg.Client.Retries, "Max update metrics retries")
	timeout := flag.Uint("report-timeout", uint(cfg.Client.Timeout.Seconds()), "Metrics update timeout (seconds)")
	batchesEnabled := flag.Bool("batches-enabled", cfg.BatchesEnabled, "Batches using enabled")
	batchSize := flag.Int("batch-size", cfg.BatchSize, "Batches sizes")

	logLevel := flag.String("log-level", cfg.Log.Level, "Logging level")
	logFile := flag.String("log-file", cfg.Log.File, "Log file path")
	logConsole := flag.Bool("log-console", cfg.Log.Console, "Enable console logging")
	logJSON := flag.Bool("log-json", cfg.Log.JSON, "Enable JSON output for logs")

	signKey := flag.String("k", cfg.SignKey, "Key to sign requests bodies")
	rateLimit := flag.Int("l", cfg.RateLimit, "Rate limits to send metric")

	flag.Parse()

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "p":
			cfg.PollInterval = time.Duration(*pollInterval) * time.Second
		case "r":
			cfg.ReportInterval = time.Duration(*reportInterval) * time.Second
		case "batches-enabled":
			cfg.BatchesEnabled = *batchesEnabled
		case "batch-size":
			cfg.BatchSize = *batchSize

		case "a":
			cfg.Client.BaseURL = *serverAddress
		case "report-retries":
			cfg.Client.Retries = *retries
		case "report-timeout":
			cfg.Client.Timeout = time.Duration(*timeout) * time.Second

		case "log-level":
			cfg.Log.Level = *logLevel
		case "log-file":
			cfg.Log.File = *logFile
		case "log-console":
			cfg.Log.Console = *logConsole
		case "log-json":
			cfg.Log.JSON = *logJSON

		case "k":
			cfg.SignKey = *signKey
		case "l":
			cfg.RateLimit = *rateLimit
		}
	})

	cfg.Client.BaseURL = ensureURL(cfg.Client.BaseURL)

	return &cfg, nil
}
