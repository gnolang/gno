.PHONY: help
help:
	@echo "Available make commands:"
	@cat Makefile | grep '^[a-z][^:]*:' | cut -d: -f1 | sort | sed 's/^/  /'

rundep=go run -modfile ../misc/devdeps/go.mod

.PHONY: gnoland.start
gnoland.start:; go run ./cmd/gnoland start

.PHONY: build
build: build.gnoland build.gnokey build.gnoweb build.gnofaucet build.gnotxsync

build.gnoland:;    go build -o build/gnoland   ./cmd/gnoland
build.gnoweb:;     go build -o build/gnoweb    ./cmd/gnoweb
build.gnofaucet:;  go build -o build/gnofaucet ./cmd/gnofaucet
build.gnokey:;     go build -o build/gnokey    ./cmd/gnokey
build.gnotxsync:;  go build -o build/gnotxsync  ./cmd/gnotxsync

.PHONY: install
install: install.gnoland install.gnoweb install.gnofaucet install.gnokey install.gnotxsync

install.gnoland:;    go install ./cmd/gnoland
install.gnoweb:;     go install ./cmd/gnoweb
install.gnofaucet:;  go install ./cmd/gnofaucet
install.gnokey:;     go install ./cmd/gnokey
install.gnotxsync:;  go install ./cmd/gnotxsync

.PHONY: fclean
fclean: clean
	rm -rf testdir

.PHONY: clean
clean:
	rm -rf build

.PHONY: lint
lint:
	$(rundep) github.com/golangci/golangci-lint/cmd/golangci-lint run --config ../.github/golangci.yml ./...

.PHONY: fmt
GOFMT_FLAGS ?= -w
fmt:
	$(rundep) mvdan.cc/gofumpt $(GOFMT_FLAGS) .

########################################
# Test suite
.PHONY: test
test: _test.gnoland _test.gnoweb _test.gnokey _test.pkgs

GOTEST_FLAGS ?= -v -p 1 -timeout=30m

_test.gnoland:; go test $(GOTEST_FLAGS) ./cmd/gnoland
_test.gnoweb:;  go test $(GOTEST_FLAGS) ./cmd/gnoweb
_test.gnokey:;  go test $(GOTEST_FLAGS) ./cmd/gnokey
_test.pkgs:;    go test $(GOTEST_FLAGS) ./pkg/...
