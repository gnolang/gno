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
GOTEST_FLAGS ?= -v -p 1 -timeout=30m -tags='ledger_suite'

########################################
# Dev tools
.PHONY: build
_build.tools: _build.aminoscan _build.goscan _build.logjack _build.iaviewer

_build.aminoscan:;  go build -o build/aminoscan  ./pkg/amino/cmd/aminoscan
_build.goscan:;     go build -o build/goscan     ./pkg/amino/cmd/goscan
_build.logjack:;    go build -o build/logjack    ./pkg/autofile/cmd
_build.iaviewer:;   go build -o build/iaviewer   ./pkg/iavl/cmd/iaviewer

.PHONY: clean
clean:
	rm -rf ./build/

.PHONY: fmt
fmt:
	go fmt ./...
	$(rundep) golang.org/x/tools/cmd/goimports -w .

.PHONY: lint
lint:
	$(rundep) github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --config ../.github/golangci.yml ./...

########################################
# Test suite
.PHONY: test
test: _test.pkg.amino _test.pkg.bft _test.pkg.db _test.pkg.others _test.flappy

_test.flappy:
	# flappy tests should work "sometimes" (at least once).
	# TODO: support coverage for flappy tests.
	STABILITY_FILTER=flappy $(rundep) moul.io/testman test -test.v -timeout=20m -retry=10 -run ^TestFlappy \
		./pkg/bft/consensus ./pkg/bft/blockchain ./pkg/bft/mempool ./pkg/p2p ./pkg/bft/privval

_test.pkg.others:;  go test $(GOTEST_FLAGS) `go list ./pkg/... | grep -Ev 'pkg/(amino|bft|db|iavl/benchmarks)(/|$$)'`
_test.pkg.amino:;   go test $(GOTEST_FLAGS) ./pkg/amino/...
_test.pkg.bft:;     go test $(GOTEST_FLAGS) ./pkg/bft/...
_test.pkg.db:;      go test $(GOTEST_FLAGS) ./pkg/db/... ./pkg/iavl/benchmarks/...

.PHONY: generate
generate:
	go generate -x ./...
	$(MAKE) fmt

.PHONY: generate.backup-grpc
generate.backup-grpc:
	cd pkg/bft/backup && go run -modfile ../../../../misc/devdeps/go.mod github.com/bufbuild/buf/cmd/buf generate
	go run -modfile ../misc/devdeps/go.mod mvdan.cc/gofumpt -w pkg/bft/backup
	go run -modfile ../misc/devdeps/go.mod golang.org/x/tools/cmd/goimports -w pkg/bft/backup