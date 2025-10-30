SHELL := /bin/sh

.PHONY: build-local
build-local:
	GOOS=$${GOOS:-linux} GOARCH=$${GOARCH:-amd64} CGO_ENABLED=0 \
		go build -o ./bin/api ./cmd/api-gateway-full
	@echo "Built ./bin/api"


