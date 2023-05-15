.PHONY: help
help:
	@echo "Available make commands:"
	@cat Makefile | grep '^[a-z][^:]*:' | cut -d: -f1 | sort | sed 's/^/  /'

rundep=go run -modfile ../misc/devdeps/go.mod

.PHONY: build
build: _build.tm2txsync
_build.tools: _build.aminoscan _build.goscan _build.logjack _build.iaviewer

_build.tm2txsync:;  go build -o build/tm2txsync  ./cmd/tm2txsync
_build.aminoscan:;  go build -o build/aminoscan  ./pkg/amino/cmd/aminoscan
_build.goscan:;     go build -o build/goscan     ./pkg/amino/cmd/goscan
_build.logjack:;    go build -o build/logjack    ./pkg/autofile/cmd
_build.iaviewer:;   go build -o build/iaviewer   ./pkg/iavl/cmd/iaviewer

.PHONY: install
install:
	go install ./cmd/...

.PHONY: clean
clean:
	rm -rf ./build/

.PHONY: fmt
GOFMT_FLAGS ?= -w
fmt:
	$(rundep) mvdan.cc/gofumpt $(GOFMT_FLAGS) .

.PHONY: lint
lint:
	golangci-lint run --config ../.github/golangci.yml ./...

########################################
# Test suite
.PHONY: test
test: _test.pkg.amino _test.pkg.bft _test.pkg.others _test.flappy _test.cmds

_test.flappy:
	# flappy tests should work "sometimes" (at least once)
	TEST_STABILITY=flappy $(rundep) moul.io/testman test -test.v -timeout=20m -retry=10 -run ^TestFlappy \
		./pkg/bft/consensus ./pkg/bft/blockchain ./pkg/bft/mempool ./pkg/p2p ./pkg/bft/privval

_test.pkg.others:;	go test `go list ./pkg/... | grep -v pkg/amino/ | grep -v pkg/bft/` -v -p 1 -timeout=30m
_test.pkg.amino:;		go test ./pkg/amino/... -v -p 1 -timeout=30m
_test.pkg.bft:;			go test ./pkg/bft/... -v -p 1 -timeout=30m
_test.cmds:;        go test ./cmd/... -v -p 1 -timeout=30m
