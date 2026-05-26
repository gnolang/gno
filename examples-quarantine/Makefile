.PHONY: help
help:
	@echo "Available make commands:"
	@cat Makefile | grep '^[a-z][^:]*:' | cut -d: -f1 | sort | sed 's/^/  /'

# command to run dependency utilities, like goimports.
rundep=go run -modfile ../misc/devdeps/go.mod

########################################
# Environment variables

CGO_ENABLED ?= 0
export CGO_ENABLED
GOFMT_FLAGS ?= -w
GOIMPORTS_FLAGS ?= $(GOFMT_FLAGS)
GOTEST_FLAGS ?= -v -p 1 -timeout=30m

########################################
# Quarantine test driver
#
# `make test` runs the Go driver at the root of this directory. The driver
# builds a unified package list spanning examples/ and examples-quarantine/
# so cross-tree imports resolve, then runs gno tests against each quarantined
# package. See ./quarantine_test.go.
.PHONY: test
test:
	cd .. && go test $(GOTEST_FLAGS) ./examples-quarantine/...

.PHONY: test.load
test.load:
	cd .. && go test $(GOTEST_FLAGS) -run TestQuarantineRealmsLoad ./examples-quarantine/...

.PHONY: transpile
transpile:
	go run ../gnovm/cmd/gno tool transpile -v .

.PHONY: build
build:
	go run ../gnovm/cmd/gno tool transpile -v --gobuild .

.PHONY: lint
lint:
	go run ../gnovm/cmd/gno lint -v ./...

.PHONY: fmt
GNOFMT_FLAGS ?= -w
fmt:
	go run ../gnovm/cmd/gno fmt $(GNOFMT_FLAGS) ./...

.PHONY: tidy
tidy:
	go run github.com/gnolang/gno/gnovm/cmd/gno mod tidy -v --recursive

.PHONY: clean
clean:
	find . \( -name "*.gno.gen.go" -or -name ".*.gno.gen_test.go" \) -delete
