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

CGO_ENABLED ?= 0
export CGO_ENABLED
# test suite flags.
GOTEST_FLAGS ?= -v -p 1 -timeout=30m

########################################
# Dev tools
.PHONY: lint
lint:
	$(rundep) github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --config ../.github/golangci.yml ./...

.PHONY: fmt
fmt:
	go fmt ./...
	$(rundep) golang.org/x/tools/cmd/goimports -w .

########################################
# Test suite
#
# The Go driver in quarantine_test.go runs gno tests against the quarantine
# set and a genesis-load smoke test. See README.md and quarantine_test.go for
# details.
.PHONY: test
test:
	go test $(GOTEST_FLAGS) ./...

.PHONY: test.load
test.load:
	go test $(GOTEST_FLAGS) -run TestQuarantineRealmsLoad ./...
