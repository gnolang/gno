.PHONY: help
help:
	@echo "Available make commands:"
	@cat Makefile | grep '^[a-z][^:]*:' | grep -v 'install_' | cut -d: -f1 | sort | sed 's/^/  /'

# command to run dependency utilities, like goimports.
rundep=go run -modfile misc/devdeps/go.mod

########################################
# Environment variables
# You can overwrite any of the following by passing a different value on the
# command line, ie. `CGO_ENABLED=1 make test`.
# NOTE: these are not very useful in this makefile, but they serve as
# documentation for sub-makefiles.

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
.PHONY: install
install: install.gnokey install.gno install.gnodev

# shortcuts to frequently used commands from sub-components.
.PHONY: install.gnokey
install.gnokey:
	$(MAKE) --no-print-directory -C ./gno.land	install.gnokey
	# \033[0;32m ... \033[0m is ansi for green text.
	@echo "\033[0;32m[+] 'gnokey' has been installed. Read more in ./gno.land/\033[0m"
.PHONY: install.gno
install.gno:
	$(MAKE) --no-print-directory -C ./gnovm	install
	@echo "\033[0;32m[+] 'gno' has been installed. Read more in ./gnovm/\033[0m"
.PHONY: install.gnodev
install.gnodev:
	$(MAKE) --no-print-directory -C ./contribs install.gnodev
	@echo "\033[0;32m[+] 'gnodev' has been installed. Read more in ./contribs/gnodev/\033[0m"

# old aliases
.PHONY: install_gnokey
install_gnokey: install.gnokey
.PHONY: install_gno
install_gno: install.gno

.PHONY: test
test: test.components test.docker

.PHONY: test.components
test.components:
	$(MAKE) --no-print-directory -C tm2      test
	$(MAKE) --no-print-directory -C gnovm    test
	$(MAKE) --no-print-directory -C gno.land test
	$(MAKE) --no-print-directory -C examples test
	$(MAKE) --no-print-directory -C misc     test

.PHONY: test.docker
test.docker:
	@if hash docker 2>/dev/null; then \
		go test --tags=docker -count=1 -v ./misc/docker-integration; \
	else \
		echo "[-] 'docker' is missing, skipping ./misc/docker-integration tests."; \
	fi

.PHONY: fmt
fmt:
	$(MAKE) --no-print-directory -C tm2      fmt imports
	$(MAKE) --no-print-directory -C gnovm    fmt imports
	$(MAKE) --no-print-directory -C gno.land fmt imports
	$(MAKE) --no-print-directory -C examples fmt

.PHONY: lint
lint:
	$(rundep) github.com/golangci/golangci-lint/cmd/golangci-lint run --config .github/golangci.yml

.PHONY: tidy
tidy:
	$(MAKE) --no-print-directory -C misc     tidy
