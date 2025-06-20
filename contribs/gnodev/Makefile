GNOROOT_DIR ?= $(abspath $(lastword $(MAKEFILE_LIST))/../../../)
GOBUILD_FLAGS ?= -ldflags "-X github.com/gnolang/gno/gnovm/pkg/gnoenv._GNOROOT=$(GNOROOT_DIR)"
GOTEST_FLAGS ?= $(GOBUILD_FLAGS) -v -p 1 -timeout=5m

rundep := go run -modfile ../../misc/devdeps/go.mod
golangci_lint := $(rundep) github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: install
install: install.gnodev
install.gnodev:
	go install $(GOBUILD_FLAGS) ./cmd/gnodev

# keep gnobro out the default install for now
install.gnobro:
	go install $(GOBUILD_FLAGS) ./cmd/gnobro

.PHONY: build
build: build.gnodev build.gnobro
build.gnodev:
	go build $(GOBUILD_FLAGS) -o ./build/gnodev ./cmd/gnodev

build.gnobro:
	go build $(GOBUILD_FLAGS) -o ./build/gnobro ./cmd/gnobro

lint:
	$(golangci_lint) --config ../../.github/golangci.yml run ./...

test:
	go test $(GOTEST_FLAGS) -v ./...

embedmd_filters = grep -v "exit status 1" | grep -v -- " -root" | grep -v -- " -home"
generate:
	mkdir -p .tmp
	go build ./cmd/gnodev ./cmd/gnobro
	(go run ./cmd/gnodev         -h 2>&1 || true) | $(embedmd_filters) > .tmp/gnodev-usage.txt
	(go run ./cmd/gnodev local   -h 2>&1 || true) | $(embedmd_filters) > .tmp/gnodev-local-usage.txt
	(go run ./cmd/gnodev staging -h 2>&1 || true) | $(embedmd_filters) > .tmp/gnodev-staging-usage.txt
	(go run ./cmd/gnobro         -h 2>&1 || true) | $(embedmd_filters) > .tmp/gnobro-usage.txt
	$(rundep) github.com/campoy/embedmd -w `find . -name "*.md"`
