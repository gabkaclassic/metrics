package main

import (
	"github.com/gabkaclassic/metrics/config"
	"github.com/gabkaclassic/metrics/pkg/httpserver"
)

func main() {

	cfg := config.ParseConfig()

	server := httpserver.New(
		httpserver.Address(cfg.Server.Address),
	)
	server.Run()
}
