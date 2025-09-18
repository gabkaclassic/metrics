package config

import (
	"flag"
)

type (
	Config struct {
		Server Server
	}

	Server struct {
		Address string
	}
)

func ParseConfig() *Config {
	var cfg Config

	// Server
	serverAddress := flag.String("address", "0.0.0.0:8080", "HTTP server address")

	flag.Parse()

	// Server
	cfg.Server.Address = *serverAddress

	return &cfg
}
