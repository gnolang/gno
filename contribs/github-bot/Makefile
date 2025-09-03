GNOROOT_DIR ?= $(abspath $(lastword $(MAKEFILE_LIST))/../../../)

rundep := go run -modfile ../../misc/devdeps/go.mod
golangci_lint := $(rundep) github.com/golangci/golangci-lint/v2/cmd/golangci-lint

.PHONY: install
install:
	go install .

.PHONY: lint
lint:
	$(golangci_lint) --config ../../.github/golangci.yml run ./...

.PHONY: test
test:
	go test $(GOTEST_FLAGS) -v ./...
