.PHONY: help
help:
	@echo "Available make commands:"
	@cat Makefile | grep '^[a-z]' | grep -v '=' | cut -d: -f1 | sort | sed 's/^/  /'

rundep=go run -modfile ../misc/devdeps/go.mod

.PHONY: build
build: _build.gnoland _build.gnokey _build.gnoweb _build.gnofaucet

_build.gnoland:;    go build -o build/gnoland   ./cmd/gnoland
_build.gnoweb:;     go build -o build/gnoweb    ./cmd/gnoweb
_build.gnofaucet:;  go build -o build/gnofaucet ./cmd/gnofaucet
_build.gnokey:;     go build -o build/gnokey    ./cmd/gnokey

.PHONY: install
install:
	go install ./cmd/...

.PHONY: fclean
fclean: clean
	rm -rf testdir

.PHONY: clean
clean:
	rm -rf build

.PHONY: lint
lint:
	golangci-lint run --config ../.github/golangci.yml ./...

.PHONY: fmt
GOFMT_FLAGS ?= -w
fmt:
	$(rundep) mvdan.cc/gofumpt $(GOFMT_FLAGS) .

########################################
# Test suite
.PHONY: test
test: _test.gnoland _test.gnoweb _test.gnokey _test.pkgs
	@echo "Full test suite finished."

_test.gnoland:; go test -v ./cmd/gnoland
_test.gnoweb:;  go test -v ./cmd/gnoweb
_test.gnokey:;  go test -v ./cmd/gnokey
_test.pkgs:;    go test -v ./pkg/...
