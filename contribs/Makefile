BLANK :=
SPACE := $(BLANK) $(BLANK)
HASH := \#


MAX_programs_CHARS    = $(lastword $(sort $(foreach d,$(programs),$(shell echo $(d) | sed -e 's/././g'))))
BASH_GET_TARGET_LINES = cat Makefile | grep '^[a-z][^:]*:' | grep -v '$(HASH).*@LEGACY'
MAX_TARGET_CHARS      = $(lastword $(sort $(shell $(BASH_GET_TARGET_LINES) | sed -e 's/:.*$$//' -e 's/%/$(MAX_programs_CHARS)/' -e 's/././g')))

.PHONY: help
help: # Print this help message
	@echo "Available make commands:"
	@( \
	    $(BASH_GET_TARGET_LINES) | \
	        grep -v '%' | \
	        sed \
	            -e 's/:[^$(HASH)]*$(HASH) */$(HASH) /' \
	            -e 's/:[^$(HASH)]*$$//' \
	            -e 's/$(HASH)/$(subst .,$(SPACE),$(MAX_TARGET_CHARS))   $(HASH)/' \
	            -e 's/^\($(MAX_TARGET_CHARS)...\) *$(HASH)/\1<--/' \
	            -e 's/^/  /' ; \
	    for d in $(patsubst %/,%,$(programs)) ; do \
	        $(BASH_GET_TARGET_LINES) | \
	            grep '\%' | \
	            sed \
	                -e 's,\%,'"$$d,g" \
	                -e 's/:[^$(HASH)]*$(HASH) */$(HASH) /' \
	                -e 's/:[^$(HASH)]*$$//' \
	                -e 's/$(HASH)/$(subst .,$(SPACE),$(MAX_TARGET_CHARS))   $(HASH)/' \
	                -e 's/^\($(MAX_TARGET_CHARS)...\) *$(HASH)/\1<--/' \
	                -e 's/^/  /' ; \
	    done \
	) | \
	sort

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
fmt: # Format contribs (Please clarify if this operates recursively)
	$(rundep) mvdan.cc/gofumpt $(GOFMT_FLAGS) .

.PHONY: tidy
tidy: # Tidy contribs (Please clarify if this operates recursively)
	find . -name go.mod -execdir go mod tidy -v \;
