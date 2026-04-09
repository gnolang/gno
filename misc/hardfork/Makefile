GOPATH ?= $(shell go env GOPATH)
GOBIN  ?= $(GOPATH)/bin

.PHONY: help install build test clean

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  %-16s %s\n", $$1, $$2}'

install: ## Install hardfork binary to GOBIN
	go install .

build: ## Build hardfork binary in current directory
	go build -o hardfork .

test: ## Run unit tests
	go test ./...

clean: ## Remove build artifacts
	rm -f hardfork
