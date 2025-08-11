GNOROOT_DIR ?= $(abspath $(lastword $(MAKEFILE_LIST))/../../../)
GOBUILD_FLAGS ?= -ldflags "-X github.com/gnolang/gno/gnovm/pkg/gnoenv._GNOROOT=$(GNOROOT_DIR)"
GOTEST_FLAGS ?= $(GOBUILD_FLAGS) -v -p 1 -timeout=5m

rundep := go run -modfile ../../misc/devdeps/go.mod
golangci_lint := $(rundep) github.com/golangci/golangci-lint/v2/cmd/golangci-lint

help: ## Display this help message.
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: build install clean test generate

##@ Build

build: ## Build gnobro binary
	go build $(GOBUILD_FLAGS) -o build/gnobro .

install: ## Install gnobro binary
	go install $(GOBUILD_FLAGS) .

##@ Testing

test: ## Run tests
	go test $(GOTEST_FLAGS) ./...

test.integration: ## Run integration tests
	go test $(GOTEST_FLAGS) -run='^\$$TestIntegration' ./...

##@ Development

tidy: ## Run go mod tidy
	go mod tidy

fmt: ## Format code
	go fmt ./...

lint: ## Run linter
	$(golangci_lint) --config ../../.github/golangci.yml run ./...

embedmd_filters = grep -v "exit status 1" | sed 's/^\(  -[^ ]* [^ ]*\) \{2,\}/\1\t/'
generate: ## Generate documentation
	mkdir -p .tmp
	go build .
	(go run . -h 2>&1 || true) | $(embedmd_filters) > .tmp/gnobro-usage.txt
	$(rundep) github.com/campoy/embedmd -w `find . -name "*.md"`
	rm -f ./gnobro

clean: ## Clean build artifacts
	rm -rf build/

##@ Dependencies

deps: ## Install dependencies
	go mod download
