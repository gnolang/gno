.PHONY: help
help:
	@echo "Available make commands:"
	@cat Makefile | grep '^[a-z][^:]*:' | cut -d: -f1 | sort | sed 's/^/  /'

rundep=go run -modfile ../misc/devdeps/go.mod

.PHONY: build
build:
	go build -o build/gno ./cmd/gno

.PHONY: install
install:
	go install ./cmd/gno

.PHONY: clean
clean:
	rm -rf build

########################################
# Dev tools
.PHONY: lint
lint:
	$(rundep) github.com/golangci/golangci-lint/cmd/golangci-lint run --config ../.github/golangci.yml ./...

.PHONY: fmt
GOFMT_FLAGS ?= -w
fmt:
	$(rundep) mvdan.cc/gofumpt $(GOFMT_FLAGS) .
	$(rundep) mvdan.cc/gofumpt $(GOFMT_FLAGS) `find stdlibs -name "*.gno"`

########################################
# Test suite
.PHONY: test
test: _test.cmd _test.pkg _test.gnolang

GOTEST_FLAGS ?= -v -p 1 -timeout=30m

.PHONY: _test.cmd
_test.cmd:
	go test ./cmd/... $(GOTEST_FLAGS)

.PHONY: _test.pkg
_test.pkg:
	go test ./pkg/... $(GOTEST_FLAGS)

.PHONY: _test.gnolang
_test.gnolang: _test.gnolang.native _test.gnolang.stdlibs _test.gnolang.realm _test.gnolang.pkg0 _test.gnolang.pkg1 _test.gnolang.pkg2 _test.gnolang.other
_test.gnolang.other:;        go test $(GOTEST_FLAGS) tests/*.go -run "(TestFileStr|TestSelectors)"
_test.gnolang.realm:;        go test $(GOTEST_FLAGS) tests/*.go -run "TestFiles/^zrealm"
_test.gnolang.pkg0:;         go test $(GOTEST_FLAGS) tests/*.go -run "TestPackages/(bufio|crypto|encoding|errors|internal|io|math|sort|std|stdshim|strconv|strings|testing|unicode)"
_test.gnolang.pkg1:;         go test $(GOTEST_FLAGS) tests/*.go -run "TestPackages/regexp"
_test.gnolang.pkg2:;         go test $(GOTEST_FLAGS) tests/*.go -run "TestPackages/bytes"
_test.gnolang.native:;       go test $(GOTEST_FLAGS) tests/*.go -test.short -run "TestFilesNative/"
_test.gnolang.stdlibs:;      go test $(GOTEST_FLAGS) tests/*.go -test.short -run 'TestFiles$$/'
_test.gnolang.native.sync:;  go test $(GOTEST_FLAGS) tests/*.go -test.short -run "TestFilesNative/" --update-golden-tests
_test.gnolang.stdlibs.sync:; go test $(GOTEST_FLAGS) tests/*.go -test.short -run 'TestFiles$$/' --update-golden-tests

########################################
# Code gen
stringer_cmd=$(rundep) golang.org/x/tools/cmd/stringer
.PHONY: _dev.stringer
_dev.stringer:
	$(stringer_cmd) -type=Kind       ./pkg/gnolang
	$(stringer_cmd) -type=Op         ./pkg/gnolang
	$(stringer_cmd) -type=TransCtrl  ./pkg/gnolang
	$(stringer_cmd) -type=TransField ./pkg/gnolang
	$(stringer_cmd) -type=VPType     ./pkg/gnolang
	$(stringer_cmd) -type=Word       ./pkg/gnolang

# genproto:
# see top-level Makefile.
