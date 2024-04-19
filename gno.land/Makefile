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
# flags for `make fmt`. -w will write the result to the destination files.
GOFMT_FLAGS ?= -w
# flags for `make imports`.
GOIMPORTS_FLAGS ?= $(GOFMT_FLAGS)
# test suite flags.
GOTEST_FLAGS ?= -v -p 1 -timeout=30m

########################################
# Dev tools
.PHONY: start.gnoland
start.gnoland:; go run ./cmd/gnoland start

.PHONY: start.gnoweb
start.gnoweb:; go run ./cmd/gnoweb

.PHONY: build
build: build.gnoland build.gnokey build.gnoweb

build.gnoland:;    go build -o build/gnoland   ./cmd/gnoland
build.gnoweb:;     go build -o build/gnoweb    ./cmd/gnoweb
build.gnokey:;     go build -o build/gnokey    ./cmd/gnokey

run.gnoland:;      go run ./cmd/gnoland start
run.gnoweb:;       go run ./cmd/gnoweb

.PHONY: install
install: install.gnoland install.gnoweb install.gnokey

install.gnoland:;    go install ./cmd/gnoland
install.gnoweb:;     go install ./cmd/gnoweb
install.gnokey:;     go install ./cmd/gnokey

.PHONY: fclean
fclean: clean
	rm -rf gnoland-data genesis.json

.PHONY: clean
clean:
	rm -rf build

.PHONY: lint
lint:
	$(rundep) github.com/golangci/golangci-lint/cmd/golangci-lint run --config ../.github/golangci.yml ./...

.PHONY: fmt
fmt:
	$(rundep) mvdan.cc/gofumpt $(GOFMT_FLAGS) .

.PHONY: imports
imports:
	$(rundep) golang.org/x/tools/cmd/goimports $(GOIMPORTS_FLAGS) .

########################################
# Test suite
.PHONY: test
test: _test.gnoland _test.gnoweb _test.gnokey _test.pkgs

_test.gnoland:;      go test $(GOTEST_FLAGS) ./cmd/gnoland
_test.gnoweb:;       go test $(GOTEST_FLAGS) ./cmd/gnoweb
_test.gnokey:;       go test $(GOTEST_FLAGS) ./cmd/gnokey
_test.pkgs:;         go test $(GOTEST_FLAGS) ./pkg/...
_test.pkgs.sync:;    UPDATE_SCRIPTS=true go test $(GOTEST_FLAGS) ./pkg/...
