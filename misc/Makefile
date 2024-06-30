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
# when running `make tidy`, use it to check that the go.mods are up-to-date.
VERIFY_MOD_SUMS ?= false

########################################
# Dev tools
.PHONY: lint
lint:
	$(rundep) github.com/golangci/golangci-lint/cmd/golangci-lint run --config ../.github/golangci.yml ./...

.PHONY: fmt
fmt:
	$(rundep) mvdan.cc/gofumpt $(GOFMT_FLAGS) .

########################################
# Test suite
.PHONY: test
test: _test.genstd

.PHONY: _test.genstd
_test.genstd:
	go test ./genstd/... $(GOTEST_FLAGS)

.PHONY: tidy
tidy:
	# Give execute permissions
	chmod +x ./mod_tidy.sh
	# Tidy go mods
	VERIFY_MOD_SUMS=$(VERIFY_MOD_SUMS) ./mod_tidy.sh
