include ../misc/makeHelpMsgs.mk

.PHONY: help
help: # Print this help message
	@echo "Available make targets:"
	@$(call BASH_DISPLAY_TARGETS_AND_COMMENTS)
	@$(BASH_DISPLAY_SUB_MAKES)

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
clean: # Remove the ./build directory and its contents
	rm -rf ./build/

.PHONY: fmt
fmt: # Format all Go files in this directory using gofumpt
	$(rundep) mvdan.cc/gofumpt $(GOFMT_FLAGS) .

.PHONY: imports
imports: # Organize imports using goimports for all Go files
	$(rundep) golang.org/x/tools/cmd/goimports $(GOIMPORTS_FLAGS) .

.PHONY: lint
lint: # Run golangci-lint with shared config on all packages
	$(rundep) github.com/golangci/golangci-lint/cmd/golangci-lint run --config ../.github/golangci.yml ./...

########################################
# Test suite
.PHONY: test
test: _test.pkg.amino _test.pkg.bft _test.pkg.db _test.pkg.others _test.flappy # Run all test subtargets (amino, bft, db, others, flappy)

_test.flappy:
	# flappy tests should work "sometimes" (at least once).
	# TODO: support coverage for flappy tests.
	TEST_STABILITY=flappy $(rundep) moul.io/testman test -test.v -timeout=20m -retry=10 -run ^TestFlappy \
		./pkg/bft/consensus ./pkg/bft/blockchain ./pkg/bft/mempool ./pkg/p2p ./pkg/bft/privval

_test.pkg.others:;  go test $(GOTEST_FLAGS) `go list ./pkg/... | grep -Ev 'pkg/(amino|bft|db|iavl/benchmarks)(/|$$)'`
_test.pkg.amino:;   go test $(GOTEST_FLAGS) ./pkg/amino/...
_test.pkg.bft:;     go test $(GOTEST_FLAGS) ./pkg/bft/...
_test.pkg.db:;      go test $(GOTEST_FLAGS) ./pkg/db/... ./pkg/iavl/benchmarks/...

.PHONY: generate
generate: # Run Go generate with verbose output, then format generated code
	go generate -x ./...
	$(MAKE) fmt
