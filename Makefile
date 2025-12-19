MODULE ?= server
ITERATION ?= 1
COVERAGE_FILE ?= coverage.out
COVERAGE_FILTERED ?= coverage_filtered.out
URL ?= http://localhost:8080
PORT ?= 9090

.PHONY: build test help profile

build:
	go build -o build/$(MODULE) cmd/$(MODULE)/main.go

profile:
	go tool pprof -http=":${PORT}" -seconds=60 ${URL}/debug/pprof/profile

test:
	@echo "==> Running tests with coverage..."
	@go clean -testcache
	@go test ./... -coverprofile=$(COVERAGE_FILE)
	@grep -v -E '(mocks\.gen\.go)|(pkg/metric/*)|(main\.go)' $(COVERAGE_FILE) > $(COVERAGE_FILTERED)
	@go tool cover -func=$(COVERAGE_FILTERED)
	@rm $(COVERAGE_FILTERED)
