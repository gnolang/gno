GNOROOT_DIR ?= $(abspath $(lastword $(MAKEFILE_LIST))/../../../)
GOBUILD_FLAGS ?= -ldflags "-X github.com/gnolang/gno/gnovm/pkg/gnoenv._GNOROOT=$(GNOROOT_DIR)"
GOTEST_FLAGS ?= $(GOBUILD_FLAGS) -v -p 1 -timeout=5m

rundep := go run -modfile ../../misc/devdeps/go.mod
golangci_lint := $(rundep) github.com/golangci/golangci-lint/v2/cmd/golangci-lint

.PHONY: install
install:
	go install $(GOBUILD_FLAGS) .

.PHONY: build
build:
	go build $(GOBUILD_FLAGS) -o ./build/gnodev .

lint:
	$(golangci_lint) --config ../../.github/golangci.yml run ./...

test:
	go test $(GOTEST_FLAGS) -v ./...

embedmd_filters = grep -v "exit status 1" | grep -v -- " -root" | grep -v -- " -home" | sed 's/^\(  -[^ ]* [^ ]*\) \{2,\}/\1\t/'
generate:
	mkdir -p .tmp
	go build .
	(go run . -h 2>&1 || true) | $(embedmd_filters) > .tmp/gnodev-usage.txt
	(go run . local -h 2>&1 || true) | $(embedmd_filters) > .tmp/gnodev-local-usage.txt
	(go run . staging -h 2>&1 || true) | $(embedmd_filters) > .tmp/gnodev-staging-usage.txt
	$(rundep) github.com/campoy/embedmd -w `find . -name "*.md"`
	rm -f ./gnodev
