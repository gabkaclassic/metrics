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
	"encoding/json"
	"errors"
	"flag"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/caarlos0/env/v10"
)

type (
	// Server represents the full configuration of the metrics server.
	Server struct {
		Address        string `env:"ADDRESS" envDefault:"localhost:8080"`
		SignKey        string `env:"KEY"`
		Log            Log
		Dump           Dump
		DB             DB
		Audit          Audit
		PrivateKeyPath string `env:"CRYPTO_KEY"`
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
		PublicKeyPath  string `env:"CRYPTO_KEY"`
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

	// serverFileConfig represents JSON-based configuration for the metrics server.
	//
	// All fields are optional and map 1:1 to existing environment variables
	// and command-line flags. Missing or zero-value fields must NOT override
	// values provided by environment variables or flags.
	//
	// Duration values are expected to be valid Go duration strings
	// (e.g. "1s", "5m", "1h").
	serverFileConfig struct {
		Address       string `json:"address"`
		Restore       *bool  `json:"restore"`
		StoreInterval string `json:"store_interval"`
		StoreFile     string `json:"store_file"`
		DatabaseDSN   string `json:"database_dsn"`
		CryptoKeyPath string `json:"crypto_key"`
	}

	// agentFileConfig represents JSON-based configuration for the metrics agent.
	//
	// All fields are optional and map 1:1 to existing environment variables
	// and command-line flags. Missing or zero-value fields must NOT override
	// values provided by environment variables or flags.
	//
	// Duration values are expected to be valid Go duration strings
	// (e.g. "1s", "5m", "1h").
	agentFileConfig struct {
		Address        string `json:"address"`
		ReportInterval string `json:"report_interval"`
		PollInterval   string `json:"poll_interval"`
		CryptoKeyPath  string `json:"crypto_key"`
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

// getConfigPath resolves configuration file path from supported sources.
//
// Resolution order:
//  1. CONFIG environment variable
//  2. Command-line flags: -c, -config, -c=..., -config=...
//
// The function does not validate file existence.
// Empty string means configuration file was not specified.
func getConfigPath() string {
	if v := os.Getenv("CONFIG"); v != "" {
		return v
	}

	args := os.Args[1:]
	for i := 0; i < len(args); i++ {
		a := args[i]

		if a == "-c" || a == "-config" {
			if i+1 < len(args) {
				return args[i+1]
			}
			return ""
		}

		if strings.HasPrefix(a, "-c=") {
			return strings.TrimPrefix(a, "-c=")
		}

		if strings.HasPrefix(a, "-config=") {
			return strings.TrimPrefix(a, "-config=")
		}
	}

	return ""
}

// loadServerFileConfig loads and parses server JSON configuration file.
//
// If the file does not exist, os.ErrNotExist should be returned.
// If the file exists but cannot be parsed or contains invalid values,
// a non-nil error must be returned.
func loadServerFileConfig(path string) (*serverFileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg serverFileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// loadAgentFileConfig loads and parses agent JSON configuration file.
//
// If the file does not exist, os.ErrNotExist should be returned.
// If the file exists but cannot be parsed or contains invalid values,
// a non-nil error must be returned.
func loadAgentFileConfig(path string) (*agentFileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg agentFileConfig
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// applyServerFileConfig applies JSON configuration values to Server config.
//
// Only non-zero and explicitly set fields from serverFileConfig
// must be applied. Existing values in cfg must not be overwritten
// by zero-values from JSON.
//
// Duration fields must be parsed using time.ParseDuration.
func applyServerFileConfig(cfg *Server, fc *serverFileConfig) error {
	if fc == nil {
		return nil
	}

	if fc.Address != "" {
		cfg.Address = fc.Address
	}

	if fc.Restore != nil {
		cfg.Dump.Restore = *fc.Restore
	}

	if fc.StoreInterval != "" {
		d, err := time.ParseDuration(fc.StoreInterval)
		if err != nil {
			return err
		}
		cfg.Dump.StoreInterval = d
	}

	if fc.StoreFile != "" {
		cfg.Dump.FileStoragePath = fc.StoreFile
	}

	if fc.DatabaseDSN != "" {
		cfg.DB.DSN = fc.DatabaseDSN
	}

	if fc.CryptoKeyPath != "" {
		cfg.PrivateKeyPath = fc.CryptoKeyPath
	}

	return nil
}

// applyAgentFileConfig applies JSON configuration values to Agent config.
//
// Only non-zero and explicitly set fields from agentFileConfig
// must be applied. Existing values in cfg must not be overwritten
// by zero-values from JSON.
//
// Duration fields must be parsed using time.ParseDuration.
func applyAgentFileConfig(cfg *Agent, fc *agentFileConfig) error {
	if fc == nil {
		return nil
	}

	if fc.Address != "" {
		cfg.Client.BaseURL = fc.Address
	}

	if fc.ReportInterval != "" {
		d, err := time.ParseDuration(fc.ReportInterval)
		if err != nil {
			return err
		}
		cfg.ReportInterval = d
	}

	if fc.PollInterval != "" {
		d, err := time.ParseDuration(fc.PollInterval)
		if err != nil {
			return err
		}
		cfg.PollInterval = d
	}

	if fc.CryptoKeyPath != "" {
		cfg.PublicKeyPath = fc.CryptoKeyPath
	}

	return nil
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

	configPath := getConfigPath()
	if configPath != "" {
		fc, err := loadServerFileConfig(configPath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		if err == nil {
			if err := applyServerFileConfig(&cfg, fc); err != nil {
				return nil, err
			}
		}
	}

	parsers := defineEnvParsers()

	if err := env.ParseWithOptions(&cfg, env.Options{FuncMap: parsers}); err != nil {
		return nil, err
	}

	flag.String("c", "", "Path to file with config")
	flag.String("config", "", "Path to file with config")

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
	privateKeyPath := flag.String("crypto-key", cfg.PrivateKeyPath, "Path to private key to decrypt requests")

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

		case "crypto-key":
			cfg.PrivateKeyPath = *privateKeyPath
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

	configPath := getConfigPath()
	if configPath != "" {
		fc, err := loadAgentFileConfig(configPath)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		if err == nil {
			if err := applyAgentFileConfig(&cfg, fc); err != nil {
				return nil, err
			}
		}
	}

	parsers := defineEnvParsers()

	if err := env.ParseWithOptions(&cfg, env.Options{FuncMap: parsers}); err != nil {
		return nil, err
	}

	flag.String("c", "", "Path to file with config")
	flag.String("config", "", "Path to file with config")

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
	publicKeyPath := flag.String("crypto-key", cfg.PublicKeyPath, "Path to public key to encrypt requests")
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
		case "crypto-key":
			cfg.PublicKeyPath = *publicKeyPath
		case "l":
			cfg.RateLimit = *rateLimit
		}
	})

	cfg.Client.BaseURL = ensureURL(cfg.Client.BaseURL)

	return &cfg, nil
}
