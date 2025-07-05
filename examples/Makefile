include ../misc/makehelp/makeHelpMsgs.mk

.PHONY: help
help: # Print this help message
	@$(call RUN_MAKEFILE_HELP,..,)

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
.PHONY: transpile
transpile: # Transpile .gno files to Go code using gno's transpiler (without building)
	go run ../gnovm/cmd/gno tool transpile -v .

.PHONY: build
build: # Transpile .gno files and immediately compile the generated Go code
	go run ../gnovm/cmd/gno tool transpile -v --gobuild .

.PHONY: test
test: # Run Gno tests in the gno.land/... package path using the gno test runner
	go run ../gnovm/cmd/gno test -v ./gno.land/...

.PHONY: test.fast
test.fast: # Run all Gno tests but stop immediately on the first failure
	go run ../gnovm/cmd/gno test -failfast -v  ./... 

.PHONY: lint
lint: # Run gno's built-in linter over the current directory
	go run ../gnovm/cmd/gno lint -v .

.PHONY: fix
fix: # Run gno's cross-realm fixer over the current directory
	go run ../gnovm/cmd/gno fix -v .

.PHONY: test.sync
test.sync: # Run tests and update golden test outputs to match current results
	go run ../gnovm/cmd/gno test -v --update-golden-tests ./...

.PHONY: clean
clean: # Delete all generated Go files from transpiled Gno source
	find . \( -name "*.gno.gen.go" -or -name ".*.gno.gen_test.go" \) -delete

.PHONY: fmt
GNOFMT_FLAGS ?= -w
fmt: # Format all .gno source files recursively using gno fmt
	go run ../gnovm/cmd/gno fmt $(GNOFMT_FLAGS) ./...

.PHONY: tidy
tidy: # Recursively tidy Gno module dependencies using gno mod tidy
	go run github.com/gnolang/gno/gnovm/cmd/gno mod tidy -v --recursive

.PHONY: generate
generate: # Run Go's code generation (`go generate`) and then reformat all output with gno fmt
	go generate ./...
	$(MAKE) fmt
