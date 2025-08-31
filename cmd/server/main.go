package main

import (
	"github.com/gabkaclassic/metrics/pkg/httpserver"
)

func main() {
	server := httpserver.New()
	server.Run()
}
