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
	Server struct {
		Address string `env:"ADDRESS" envDefault:"localhost:8080"`
		Log     Log
	}
	Agent struct {
		PollInterval   time.Duration `env:"POLL_INTERVAL" envDefault:"2"`
		ReportInterval time.Duration `env:"REPORT_INTERVAL" envDefault:"10"`
		Client         Client
		Log            Log
	}
	Client struct {
		BaseURL string        `env:"ADDRESS" envDefault:"localhost:8080"`
		Timeout time.Duration `env:"TIMEOUT" envDefault:"3"`
		Retries int           `env:"RETRIES" envDefault:"3"`
	}
	Log struct {
		Level   string `env:"LOG_LEVEL" envDefault:"info"`
		File    string `env:"LOG_FILE"`
		Console bool   `env:"LOG_CONSOLE" envDefault:"false"`
		JSON    bool   `env:"LOG_JSON" envDefault:"true"`
	}
)

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
		reflect.TypeOf(time.Duration(0)): func(v string) (interface{}, error) {
			secs, err := strconv.Atoi(v)
			if err != nil {
				return nil, err
			}
			return time.Duration(secs) * time.Second, nil
		},
	}
}

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

	flag.Parse()

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "a":
			cfg.Address = *address
		case "log-level":
			cfg.Log.Level = *logLevel
		case "log-file":
			cfg.Log.File = *logFile
		case "log-console":
			cfg.Log.Console = *logConsole
		case "log-json":
			cfg.Log.JSON = *logJSON
		}
	})

	return &cfg, nil
}

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

	logLevel := flag.String("log-level", cfg.Log.Level, "Logging level")
	logFile := flag.String("log-file", cfg.Log.File, "Log file path")
	logConsole := flag.Bool("log-console", cfg.Log.Console, "Enable console logging")
	logJSON := flag.Bool("log-json", cfg.Log.JSON, "Enable JSON output for logs")

	flag.Parse()

	flag.Visit(func(f *flag.Flag) {
		switch f.Name {
		case "p":
			cfg.PollInterval = time.Duration(*pollInterval) * time.Second
		case "r":
			cfg.ReportInterval = time.Duration(*reportInterval) * time.Second
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
		}
	})

	cfg.Client.BaseURL = ensureURL(cfg.Client.BaseURL)

	return &cfg, nil
}
