include ../misc/makehelp/makeHelpMsgs.mk

progs_with_generate := $(patsubst %/Makefile,%,$(shell grep -l '^generate:' */Makefile))

.PHONY: help
help: # Print this help message
	@$(call RUN_MAKEFILE_HELP,..,$(patsubst %/,%,$(programs)) $(patsubst %,2:%,$(progs_with_generate)))

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
install:  # Print instructions for installing one or all subpackages
	@echo 'To install a tool, go to the subdirectory, then run `make install`.'
	@echo 'To do a full installation, run `make install_all`.'

install_all: $(addprefix install.,$(programs))  # Install all subpackages
install.%:  # Install subpackage %
	@echo "[+] make -C $(subst install.,,$@) install"
	$(MAKE) --no-print-directory -C $(subst install.,,$@) install
.PHONY: install_all

########################################
# Test suite
test: $(addprefix test.,$(programs)) # Test all subpackages
test.%:  # Test subpackage %
	@echo "[+] make -C $(subst test.,,$@) install"
	$(MAKE) --no-print-directory -C $(subst test.,,$@) test
.PHONY: test

########################################
# Lint
.PHONY: lint
lint: $(addprefix lint.,$(programs)) # Lint all subpackages
lint.%:  # Lint subpackage %
	@echo "[+] make -C $(subst lint.,,$@) install"
	$(MAKE) --no-print-directory -C $(subst lint.,,$@) lint

########################################
# Dev tools
rundep=go run -modfile ../misc/devdeps/go.mod

.PHONY: fmt
fmt: # Format all Go files in this directory and subdirectories
	$(rundep) mvdan.cc/gofumpt $(GOFMT_FLAGS) .

.PHONY: tidy
tidy: # Run 'go mod tidy' in all Go modules within this directory and subdirectories
	find . -name go.mod -execdir go mod tidy -v \;


generate: $(addprefix generate.,$(progs_with_generate)) # Generate documentation (for applicable tools)
generate.%: # %2:Generate documentation for %
	@if grep '^generate:' $(subst generate.,,$@)/Makefile; then \
	  echo "[+] make -C $(subst generate.,,$@) generate"; \
	  $(MAKE) --no-print-directory -C $(subst generate.,,$@) generate; \
	fi
.PHONY: generate
