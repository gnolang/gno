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
# when running `make tidy`, use it to check that the go.mods are up-to-date.
VERIFY_MOD_SUMS ?= false

########################################
# Dev tools
.PHONY: lint
lint: # Run golangci-lint using the shared configuration on all code under ./misc
	$(rundep) github.com/golangci/golangci-lint/cmd/golangci-lint run --config ../.github/golangci.yml ./...

.PHONY: fmt
fmt: # Format all Go files in the current directory using gofumpt
	$(rundep) mvdan.cc/gofumpt $(GOFMT_FLAGS) .

########################################
# Test suite
.PHONY: test
test: _test.makefile_help _test.genstd # Run the genstd and makefile_help test suites


.PHONY: _test.makefile_help
_test.makefile_help:
	go test makehelp/makefile_help.go makehelp/makefile_help_test.go $(GOTEST_FLAGS)

.PHONY: _test.genstd
_test.genstd:
	go test ./genstd/... $(GOTEST_FLAGS)

.PHONY: tidy
tidy: # Run mod_tidy.sh to clean up and optionally verify go.mod files
	# Give execute permissions
	chmod +x ./mod_tidy.sh
	# Tidy go mods
	VERIFY_MOD_SUMS=$(VERIFY_MOD_SUMS) ./mod_tidy.sh
