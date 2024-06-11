.PHONY: help
help:
	@echo "Available make commands:"
	@cat Makefile | grep '^[a-z][^:]*:' | cut -d: -f1 | sort | sed 's/^/  /'

programs=$(wildcard */)

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
install:
	@echo 'To install a tool, go to the subdirectory, then run `make install`.'
	@echo 'To do a full installation, run `make install_all`.'

.PHONY: install_all
install_all:
	@set -e; for program in $(programs); do ( set -e; \
	  echo "[+] make -C $$program install"; \
	  $(MAKE) --no-print-directory -C $$program install; \
	); done

########################################
# Test suite
.PHONY: test
test:
	@set -e; for program in $(programs); do ( set -e; \
	  echo "[+] make -C $$program test"; \
	  $(MAKE) --no-print-directory -C $$program test; \
	); done

########################################
# Lint
.PHONY: lint
lint:
	@set -e; for program in $(programs); do ( set -e; \
	  echo "[+] make -C $$program lint"; \
	  $(MAKE) --no-print-directory -C $$program lint; \
	); done

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
