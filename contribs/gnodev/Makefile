GNOROOT_DIR ?= $(abspath $(lastword $(MAKEFILE_LIST))/../../../)
GOBUILD_FLAGS ?= -ldflags "-X github.com/gnolang/gno/gnovm/pkg/gnoenv._GNOROOT=$(GNOROOT_DIR)"

rundep := go run -modfile ../../misc/devdeps/go.mod
golangci_lint := $(rundep) github.com/golangci/golangci-lint/cmd/golangci-lint

install:
	go install $(GOBUILD_FLAGS) ./cmd/gnodev

build:
	go build $(GOBUILD_FLAGS) -o build/gnodev ./cmd/gnodev

lint:
	$(golangci_lint) --config ../../.github/golangci.yml run ./...
