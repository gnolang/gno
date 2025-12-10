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
GOTEST_FLAGS ?= -v -p 1 -timeout=30m
# used to set the default GNOROOT. This helps on local development, as it sets
# up GNOROOT to the supposed path of the Gno repository clone.
GNOROOT_DIR ?= $(abspath $(lastword $(MAKEFILE_LIST))/../../)
# We can't use '-trimpath' yet as amino use absolute path from call stack
# to find some directory: see #1236
GOBUILD_FLAGS ?= -ldflags "-X github.com/gnolang/gno/tm2/pkg/version.Version=$(VERSION) -X github.com/gnolang/gno/gnovm/pkg/gnoenv._GNOROOT=$(GNOROOT_DIR)"
# file where to place cover profile; used for coverage commands which are
# more complex than adding -coverprofile, like test.cmd.coverage.
GOTEST_COVER_PROFILE ?= cmd-profile.out

# user for gno version [branch].[N]+[hash]
VERSION ?= $(shell git describe --tags --exact-match 2>/dev/null || echo "$(shell git rev-parse --abbrev-ref HEAD).$(shell git rev-list --count HEAD)+$(shell git rev-parse --short HEAD)")

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
	$(rundep) github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --config ../.github/golangci.yml ./...

.PHONY: fmt
fmt:
	go fmt ./...
	go run ./cmd/gno fmt -w ./stdlibs/... ./tests/stdlibs/...
	$(rundep) golang.org/x/tools/cmd/goimports -w .

# Benchmarking the VM's opcode and storage access
.PHONY: opcode storage build_opcode build_storage
build.bench.opcode:
	go build -tags "benchmarkingops" -o build/gnobench ./cmd/benchops

build.bench.storage:
	go build -tags "benchmarkingstorage" -o build/gnobench ./cmd/benchops

build.bench.native:
	go build -tags "benchmarkingnative" -o build/gnobench ./cmd/benchops

# Extract the latest commit hash
COMMIT_HASH := $(shell git rev-parse --short=7 HEAD)

# Run target
run.bench.opcode: build.bench.opcode
	./build/gnobench -out opcode_results_$(COMMIT_HASH).csv

# Run target
run.bench.storage: build.bench.storage
	./build/gnobench -out store_results_$(COMMIT_HASH).csv

run.bench.native: build.bench.native
	./build/gnobench -out native_results_$(COMMIT_HASH).csv

########################################
# Test suite
.PHONY: test
test: _test.filetest _test.cmd _test.pkg _test.stdlibs

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
	go test -v ./stdlibs/...
	# 'gno test ./stdlibs/...' is different from TestStdlibs in 'go test ./pkg/..'
	# TestStdlibs does not include tests from *_test packages.
	cd stdlibs && go run ../cmd/gno test -v ./...


.PHONY: _test.filetest
_test.filetest:
	@echo "use 'go test pkg/gnolang/files_test.go -test.short --update-golden-tests' to update filetest expectations"
	go test pkg/gnolang/files_test.go -test.short -run 'TestFiles$$/' $(GOTEST_FLAGS)

########################################
# Code gen
.PHONY: generate
generate: _dev.generate _dev.docs fmt

.PHONY: _dev.docs
_dev.docs:
	mkdir -p .tmp
	(go run ./cmd/gno -h 2>&1 || true) | grep -v "exit status 1" > .tmp/gno-help.txt
	$(rundep) github.com/campoy/embedmd -w `find . -name "*.md"`

.PHONY: _dev.generate
_dev.generate:
	go generate -x ./...

# genproto:
# see top-level Makefile.
