.PHONY: help
help:
	@echo "Available make commands:"
	@cat Makefile | grep '^[a-z][^:]*:' | cut -d: -f1 | sort | sed 's/^/  /'

# command to run dependency utilities, like goimports.
rundep=go run -modfile ../misc/devdeps/go.mod

########################################
# Environment variables
# You can overwrite any of the following by passing a different value on the
# command line, ie. `CGO_ENABLED=1 make test`.

# disable cgo by default. cgo requires some additional dependencies in some
# cases, and is not strictly required by any tm2 code.
CGO_ENABLED ?= 0
export CGO_ENABLED
# test suite flags.
GOTEST_FLAGS ?= -v -p 1 -timeout=30m
# used for gnokey version [branch].[N]+[hash]
VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || echo "$(shell git rev-parse --abbrev-ref HEAD).$(shell git rev-list --count HEAD)+$(shell git rev-parse --short HEAD)")

GOBUILD_FLAGS ?= -ldflags "-X github.com/gnolang/gno/tm2/pkg/version.Version=$(VERSION)"

########################################
# Dev tools
.PHONY: start.gnoland
start.gnoland:; go run ./cmd/gnoland start -lazy -skip-genesis-sig-verification

.PHONY: start.gnoweb
start.gnoweb:; go run ./cmd/gnoweb

.PHONY: build
build: build.gnoland build.gnokey build.gnoweb

build.gnoland:;    go build -o build/gnoland   ./cmd/gnoland
build.gnoweb:;     go build -o build/gnoweb    ./cmd/gnoweb
build.gnokey:;     go build $(GOBUILD_FLAGS) -o build/gnokey    ./cmd/gnokey

run.gnoland:;      go run ./cmd/gnoland start
run.gnoweb:;       go run ./cmd/gnoweb

.PHONY: install
install: install.gnoland install.gnoweb install.gnokey

install.gnoland:;    go install ./cmd/gnoland
install.gnoweb:;     go install ./cmd/gnoweb
install.gnokey:;     go install ./cmd/gnokey

.PHONY: dev.gnoweb generate.gnoweb
dev.gnoweb:
	make -C ./pkg/gnoweb dev

.PHONY: generate
generate:
	go generate -x ./...
	make -C ./pkg/gnoweb generate


.PHONY: fclean
fclean: clean
	rm -rf gnoland-data genesis.json

.PHONY: clean
clean:
	rm -rf build

.PHONY: lint
lint:
	$(rundep) github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --config ../.github/golangci.yml ./...

.PHONY: fmt
fmt:
	go fmt ./...
	$(rundep) golang.org/x/tools/cmd/goimports -w .

########################################
# Test suite
.PHONY: test
test: _test.help _test.gnoland _test.gnoweb _test.gnokey _test.pkgs

_test.help:
	@echo "run \`INMEMORY_TS=true make test\` if you encounter 'context deadline exceeded' for non-parallel testing"
_test.gnoland:;      go test $(GOTEST_FLAGS) ./cmd/gnoland
_test.gnoweb:;       go test $(GOTEST_FLAGS) ./cmd/gnoweb
_test.gnokey:;       go test $(GOTEST_FLAGS) ./cmd/gnokey
_test.pkgs:;         go test $(GOTEST_FLAGS) ./pkg/...
_test.pkgs.sync:;    UPDATE_SCRIPTS=true go test $(GOTEST_FLAGS) ./pkg/...
