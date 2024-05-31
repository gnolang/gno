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

# Official packages (non-overridable): more reliable and tested modules, distinct from the experimentation area.
OFFICIAL_PACKAGES = ./gno.land/p
OFFICIAL_PACKAGES += ./gno.land/r/demo
OFFICIAL_PACKAGES += ./gno.land/r/gnoland
OFFICIAL_PACKAGES += ./gno.land/r/sys

########################################
# Dev tools
.PHONY: transpile
transpile:
	go run ../gnovm/cmd/gno transpile -v .

.PHONY: build
build:
	go run ../gnovm/cmd/gno transpile -v --gobuild .

.PHONY: test
test:
	go run ../gnovm/cmd/gno test -v ./...

.PHONY: lint
lint:
	go run ../gnovm/cmd/gno lint $(OFFICIAL_PACKAGES)

.PHONY: test.sync
test.sync:
	go run ../gnovm/cmd/gno test -v --update-golden-tests ./...

.PHONY: clean
clean:
	find . \( -name "*.gno.gen.go" -or -name ".*.gno.gen_test.go" \) -delete

.PHONY: fmt
GOFMT_FLAGS ?= -w
fmt:
	go run -modfile ../misc/devdeps/go.mod mvdan.cc/gofumpt $(GOFMT_FLAGS) `find . -name "*.gno"`

.PHONY: imports
GOIMPORTS_FLAGS ?= -w
imports:
	$(rundep) golang.org/x/tools/cmd/goimports $(GOIMPORTS_FLAGS) .

.PHONY: tidy
tidy:
	find . -name "gno.mod" -execdir go run github.com/gnolang/gno/gnovm/cmd/gno mod tidy \;
