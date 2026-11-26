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

# `go fix` applies the Go 1.26+ modernizers. Pin the toolchain so `make fix`
# matches CI regardless of the version in go.mod; override with
# GO_FIX_TOOLCHAIN=local to use your installed Go. Bump when the repo's Go
# version advances (keep in sync with .github/workflows/_ci-go.yml).
GO_FIX_TOOLCHAIN ?= go1.26.1

# -omitzero=false disables the omitzero modernizer: it strips `json:",omitempty"`
# from struct-typed fields (a no-op for encoding/json), but Amino honors omitempty
# on struct fields, so removing it changes JSON output. Keep in sync with the CI
# check in .github/workflows/_ci-go.yml.
GO_FIX_FLAGS ?= -omitzero=false

.PHONY: fix fix.contribs fix.misc print.misc-modules
# Apply `go fix` across every Go module that CI checks (see the lint job).
# CGO_ENABLED=1 so `//go:build cgo` files (e.g. tm2/pkg/db/lmdbdb, mdbxdb) are
# covered, matching CI; otherwise those files are silently skipped here but
# flagged by the CI check. Requires a C compiler.
fix: fix.contribs fix.misc
	CGO_ENABLED=1 GOTOOLCHAIN=$(GO_FIX_TOOLCHAIN) go fix $(GO_FIX_FLAGS) ./...
	@# gnoe2e is its own module, which ci-gnoe2e.yml lints rather than
	@# ci-dir-misc.yml.
	@echo "==> go fix misc/gnoe2e/"
	@cd misc/gnoe2e && CGO_ENABLED=1 GOTOOLCHAIN=$(GO_FIX_TOOLCHAIN) go fix $(GO_FIX_FLAGS) ./...

# The contribs/ part of `fix`, which ci-dir-contribs-lint.yml runs on its own.
fix.contribs:
	@for d in $(wildcard contribs/*/); do \
		echo "==> go fix $$d"; \
		( cd "$$d" && CGO_ENABLED=1 GOTOOLCHAIN=$(GO_FIX_TOOLCHAIN) go fix $(GO_FIX_FLAGS) ./... ); \
	done

# The Go programs under misc/ that ci-dir-misc.yml lints and tests; it reads
# this list with `make print.misc-modules`. It is fixed because other programs
# there are covered elsewhere (genproto2 by ci-codegen-verify.yml,
# multiarch-determinism by its own workflow, gnoe2e by ci-gnoe2e.yml,
# stdlib_diff by deploy-pages.yml upstream) or not at all: a new one that is in
# neither place is never built, with nothing to say so.
MISC_MODULES := misc/audit-pattern-harness misc/autocounterd misc/genstd misc/gnopreview misc/goscan misc/loop

# The misc/ part of `fix`, which ci-dir-misc.yml runs on its own.
fix.misc:
	@for d in $(MISC_MODULES); do \
		echo "==> go fix $$d"; \
		( cd "$$d" && CGO_ENABLED=1 GOTOOLCHAIN=$(GO_FIX_TOOLCHAIN) go fix $(GO_FIX_FLAGS) ./... ); \
	done

print.misc-modules:
	@echo $(MISC_MODULES)

.PHONY: lint
lint:
	$(rundep) github.com/golangci/golangci-lint/v2/cmd/golangci-lint run --config .github/golangci.yml

.PHONY: tidy
tidy:
	$(MAKE) --no-print-directory -C misc     tidy

.PHONY: mocks
mocks:
	$(rundep) github.com/golang/mock/mockgen -source=tm2/pkg/db/types.go -package mockdb -destination tm2/pkg/db/mockdb/mockdb.go
