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
GNOFMT_FLAGS ?= -w
# flags for `make imports`.
GOIMPORTS_FLAGS ?= $(GOFMT_FLAGS)
# test suite flags.
GOTEST_FLAGS ?= -v -p 1 -timeout=30m
# used to set the default GNOROOT. This helps on local development, as it sets
# up GNOROOT to the supposed path of the Gno repository clone.
GNOROOT_DIR ?= $(abspath $(lastword $(MAKEFILE_LIST))/../../)
# We can't use '-trimpath' yet as amino use absolute path from call stack
# to find some directory: see #1236
GOBUILD_FLAGS ?= -ldflags "-X github.com/gnolang/gno/gnovm/pkg/gnoenv._GNOROOT=$(GNOROOT_DIR)"
# file where to place cover profile; used for coverage commands which are
# more complex than adding -coverprofile, like test.cmd.coverage.
GOTEST_COVER_PROFILE ?= cmd-profile.out

########################################
# Dev tools
.PHONY: build
build:
	go build $(GOBUILD_FLAGS) -o build/gno ./cmd/gno

.PHONY: install
install:
	go install $(GOBUILD_FLAGS) ./cmd/gno

.PHONY: clean
clean:
	rm -rf build

########################################
# Dev tools
.PHONY: lint
lint:
	$(rundep) github.com/golangci/golangci-lint/cmd/golangci-lint run --config ../.github/golangci.yml ./...

.PHONY: fmt
fmt:
	go run ./cmd/gno fmt $(GNOFMT_FLAGS) ./stdlibs/... ./tests/stdlibs/...
	$(rundep) mvdan.cc/gofumpt $(GOFMT_FLAGS) .

.PHONY: imports
imports:
	$(rundep) golang.org/x/tools/cmd/goimports $(GOIMPORTS_FLAGS) .

########################################
# Test suite
.PHONY: test
test: _test.cmd _test.pkg _test.stdlibs

.PHONY: _test.cmd
_test.cmd:
	go test ./cmd/... $(GOTEST_FLAGS)

# Run tests on ./cmd/, saving the result of the coverage in
# GOTEST_COVER_PROFILE.
.PHONY: test.cmd.coverage
test.cmd.coverage:
	$(eval export TXTARCOVERDIR := $(shell mktemp -d --tmpdir gnovm-make.XXXXXXX))
	go test ./cmd/... -covermode atomic -test.gocoverdir='$(TXTARCOVERDIR)' $(GOTEST_FLAGS)
	@echo "coverage results:"
	go tool covdata percent -i="$(TXTARCOVERDIR)"
	go tool covdata textfmt -v 1 -i="$(TXTARCOVERDIR)" -o '$(GOTEST_COVER_PROFILE)'
	rm -rf "$(TXTARCOVERDIR)"

# Run test.cmd.coverage, then view the result in the HTML browser render
# and delete the original file.
.PHONY: test.cmd.coverage_view
test.cmd.coverage_view: test.cmd.coverage
	go tool cover -html='$(GOTEST_COVER_PROFILE)'
	rm '$(GOTEST_COVER_PROFILE)'

.PHONY: _test.pkg
_test.pkg:
	go test ./pkg/... $(GOTEST_FLAGS)

.PHONY: _test.stdlibs
_test.stdlibs:
	go run ./cmd/gno test -v ./stdlibs/...


########################################
# Code gen
# TODO: move _dev.stringer to go:generate instructions, simplify generate
#       to just go generate.
.PHONY: generate
generate: _dev.stringer _dev.generate

stringer_cmd=$(rundep) golang.org/x/tools/cmd/stringer
.PHONY: _dev.stringer
_dev.stringer:
	$(stringer_cmd) -type=Kind       ./pkg/gnolang
	$(stringer_cmd) -type=Op         ./pkg/gnolang
	$(stringer_cmd) -type=TransCtrl  ./pkg/gnolang
	$(stringer_cmd) -type=TransField ./pkg/gnolang
	$(stringer_cmd) -type=VPType     ./pkg/gnolang
	$(stringer_cmd) -type=Word       ./pkg/gnolang

.PHONY: _dev.generate
_dev.generate:
	go generate -x ./...

# genproto:
# see top-level Makefile.
