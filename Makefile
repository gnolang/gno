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
	@# \033[0;32m ... \033[0m is ansi for green text.
	@printf "\033[0;32m[+] 'gnokey' has been installed. Read more in ./gno.land/\033[0m\n"
.PHONY: install.gno
install.gno:
	$(MAKE) --no-print-directory -C ./gnovm	install
	@printf "\033[0;32m[+] 'gno' has been installed. Read more in ./gnovm/\033[0m\n"
.PHONY: install.gnodev
install.gnodev:
	$(MAKE) --no-print-directory -C ./contribs/gnodev install
	@printf "\033[0;32m[+] 'gnodev' has been installed. Read more in ./contribs/gnodev/\033[0m\n"
.PHONY: install.gnobro
install.gnobro:
	$(MAKE) --no-print-directory -C ./contribs/gnobro install
	@printf "\033[0;32m[+] 'gnobro' has been installed. Read more in ./contribs/gnobro/\033[0m\n"


# old aliases
.PHONY: install_gnokey
install_gnokey: install.gnokey
.PHONY: install_gno
install_gno: install.gno

.PHONY: test
test: test.components

.PHONY: test.components
test.components:
	$(MAKE) --no-print-directory -C tm2      test
	$(MAKE) --no-print-directory -C gnovm    test
	$(MAKE) --no-print-directory -C gno.land test
	$(MAKE) --no-print-directory -C examples test
	$(MAKE) --no-print-directory -C misc     test

.PHONY: fmt
fmt:
	$(MAKE) --no-print-directory -C tm2      fmt
	$(MAKE) --no-print-directory -C gnovm    fmt
	$(MAKE) --no-print-directory -C gno.land fmt
	$(MAKE) --no-print-directory -C examples fmt
	$(MAKE) --no-print-directory -C contribs fmt

.PHONY: lint
lint:
	$(rundep) github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --config .github/golangci.yml

.PHONY: tidy
tidy:
	$(MAKE) --no-print-directory -C misc     tidy

.PHONY: mocks
mocks:
	$(rundep) github.com/golang/mock/mockgen -source=tm2/pkg/db/types.go -package mockdb -destination tm2/pkg/db/mockdb/mockdb.go
