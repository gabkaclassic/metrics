MODULE ?= server
ITERATION ?= 1
COVERAGE_FILE ?= coverage.out
COVERAGE_FILTERED ?= coverage_filtered.out
URL ?= http://localhost:8080
PORT ?= 9090
VERSION := 1.0.0
BUILD_DATE := $(shell date +'%Y.%m.%d_%H:%M:%S')
COMMIT_HASH := $(shell git rev-parse --short HEAD)


.PHONY: build test help profile

build:
	go build -ldflags "\
	-X main.buildVersion=$(VERSION) \
	-X main.buildDate=$(BUILD_DATE) \
	-X main.buildCommit=$(COMMIT_HASH)" \
	-o build/$(MODULE) cmd/$(MODULE)/main.go

profile:
	go tool pprof -http=":${PORT}" -seconds=60 ${URL}/debug/pprof/profile

swagger:
	swag init -d ./cmd/server,./internal/handler,./internal/model,./pkg/error --output ./api

test:
	@echo "==> Running tests with coverage..."
	@go clean -testcache
	@go test ./... -coverprofile=$(COVERAGE_FILE)
	@grep -v -E '(mocks\.gen\.go)|(pkg/metric/*)|(main\.go)|(doc\.go)|(reset\.gen\.go)' $(COVERAGE_FILE) > $(COVERAGE_FILTERED)
	@go tool cover -func=$(COVERAGE_FILTERED)
	@rm $(COVERAGE_FILTERED)
