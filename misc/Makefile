BLANK :=
SPACE := $(BLANK) $(BLANK)
HASH := \#

DIR_OFFSET_OPT        = $(if $(filter $(PWD),$(CURDIR)),,-C $(patsubst $(patsubst %/,%,$(PWD))/%,%,$(CURDIR)))
MAKE_SUBDIRS          = $(patsubst %/Makefile,%,$(wildcard */Makefile))
BASH_GET_TARGET_LINES = cat Makefile | grep '^[a-z][^:]*:' | grep -v '$(HASH).*@LEGACY'
MAX_TARGET_CHARS      = $(lastword $(sort $(shell $(BASH_GET_TARGET_LINES) | sed -e 's/:.*$$//' -e 's/././g')))

.PHONY: help
help: # Print this help message
	@echo "Available make commands:"
	@$(BASH_GET_TARGET_LINES) | \
	    sort | \
	    sed \
	        -e 's/:[^$(HASH)]*$(HASH) */$(HASH) /' \
	        -e 's/:[^$(HASH)]*$$//' \
	        -e 's/$(HASH)/$(subst .,$(SPACE),$(MAX_TARGET_CHARS))   $(HASH)/' \
	        -e 's/^\($(MAX_TARGET_CHARS)...\) *$(HASH)/\1<--/' \
	        -e 's/^/  /'
	@$(if \
		$(MAKE_SUBDIRS),$\
		echo ; echo "Sub-directories with make targets:" ; \
		for d in $(sort $(MAKE_SUBDIRS)); do \
			echo '    '"$$(grep -q '^help *:' $$d/Makefile && echo '*  ' || echo '   ') $\
			    $(if $(filter $(MAKE),$(shell which make)),make,$(MAKE)) $(if $(DIR_OFFSET_OPT),$(DIR_OFFSET_OPT)/,-C$(SPACE))$$d" ; $\
	    done ; \
	    grep -q '^help *:' $(patsubst %,%/Makefile,$(MAKE_SUBDIRS)) && \
	        echo && echo '       * Is documented with a `help` target.' || echo > /dev/null,$\
		$(HASH) do nothing $\
	)

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
lint:
	$(rundep) github.com/golangci/golangci-lint/cmd/golangci-lint run --config ../.github/golangci.yml ./...

.PHONY: fmt
fmt:
	$(rundep) mvdan.cc/gofumpt $(GOFMT_FLAGS) .

########################################
# Test suite
.PHONY: test
test: _test.genstd

.PHONY: _test.genstd
_test.genstd:
	go test ./genstd/... $(GOTEST_FLAGS)

.PHONY: tidy
tidy:
	# Give execute permissions
	chmod +x ./mod_tidy.sh
	# Tidy go mods
	VERIFY_MOD_SUMS=$(VERIFY_MOD_SUMS) ./mod_tidy.sh
