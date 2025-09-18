MODULE ?= server
ITERATION ?= 1

.PHONY: build test help

build:
	go build -o build/$(MODULE) cmd/$(MODULE)/main.go

test:
	./metricstest -test.v -test.run=^TestIteration$(ITERATION)$$ -binary-path=build/$(MODULE)
