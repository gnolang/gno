help: ## Display this help message.
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: build install clean test

##@ Build

build: ## Build gnobro binary
	go build -o build/gnobro .

install: ## Install gnobro binary
	go install .

##@ Testing

test: ## Run tests
	go test -v ./...

test.integration: ## Run integration tests
	go test -v -run='^\$$TestIntegration' ./...

##@ Development

tidy: ## Run go mod tidy
	go mod tidy

fmt: ## Format code
	go fmt ./...

lint: ## Run linter
	golangci-lint run ./...

clean: ## Clean build artifacts
	rm -rf build/

##@ Dependencies

deps: ## Install dependencies
	go mod download