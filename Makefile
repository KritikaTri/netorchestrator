GO ?= go
VERIFIED := ./internal/topology/... ./internal/cfgdiff/... ./cmd/netorchctl/...

.PHONY: help test test-all vet fmt fmt-check build cli demo bench
help:            ## list targets
	@grep -E '^[a-z-]+:.*##' $(MAKEFILE_LIST) | sed 's/:.*##/ -/'

test:            ## race-tested core packages (fast, no external services)
	$(GO) test -race -count=1 -cover $(VERIFIED)

test-all:        ## every package in the module
	$(GO) test -race -count=1 ./...

vet:             ## static checks, whole module
	$(GO) vet ./...

fmt:             ## format everything
	gofmt -w $$(git ls-files '*.go')

fmt-check:       ## fail if anything is unformatted
	@test -z "$$(gofmt -l $$(git ls-files '*.go'))" || (gofmt -l $$(git ls-files '*.go'); exit 1)

build:           ## build every command into ./bin
	mkdir -p bin && $(GO) build -o bin/ ./cmd/...

cli:             ## build just the dependency-free CLI
	mkdir -p bin && $(GO) build -o bin/netorchctl ./cmd/netorchctl

demo: cli        ## run the CLI against the bundled campus example
	./bin/netorchctl spof  -topology examples/campus.json
	./bin/netorchctl blast -topology examples/campus.json -node dist2
	./bin/netorchctl path  -topology examples/campus.json -from access1 -to access3
	./bin/netorchctl cfgdiff -desired examples/desired.cfg -running examples/running.cfg

bench:           ## shortest-path benchmark on a 100x100 grid
	$(GO) test -run '^$$' -bench . -benchmem ./internal/topology
