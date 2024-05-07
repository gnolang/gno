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

########################################
# Dev tools
.PHONY: install
install: install.gnomd install.gnodev install.gnofaucet

install.gnomd:;  cd gnomd && go install .
install.gnodev:; $(MAKE) -C ./gnodev install
install.gnofaucet:; $(MAKE) -C ./gnofaucet install

.PHONY: clean
clean:
	rm -rf build

########################################
# Dev tools
rundep=go run -modfile ../misc/devdeps/go.mod

.PHONY: fmt
fmt:
	$(rundep) mvdan.cc/gofumpt $(GOFMT_FLAGS) .

.PHONY: tidy
tidy:
	@for gomod in `find . -name go.mod`; do ( \
    dir=`dirname $$gomod`; \
	  set -xe; \
	  cd $$dir; \
	  go mod tidy -v; \
	); done

########################################
# Test suite
.PHONY: test
test: test.gnodev
test.gnodev:
	$(MAKE) -C ./gnodev test

########################################
# Lint
.PHONY: test
lint: lint.gnodev
lint.gnodev:
	$(MAKE) -C ./gnodev test
