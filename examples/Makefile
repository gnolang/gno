# Official packages: more reliable and tested modules, distinct from the experimentation area.
OFFICIAL_PACKAGES = ./gno.land/p
OFFICIAL_PACKAGES += ./gno.land/r/demo
OFFICIAL_PACKAGES += ./gno.land/r/gnoland
OFFICIAL_PACKAGES += ./gno.land/r/system

.PHONY: help
help:
	@echo "Available make commands:"
	@cat Makefile | grep '^[a-z][^:]*:' | cut -d: -f1 | sort | sed 's/^/  /'

.PHONY: precompile
precompile:
	go run ../gnovm/cmd/gno precompile --verbose .

.PHONY: build
build: precompile
	go run ../gnovm/cmd/gno build --verbose .

.PHONY: test
test:
	go run ../gnovm/cmd/gno test --verbose .

.PHONY: lint
lint:
	go run ../gnovm/cmd/gno lint $(OFFICIAL_PACKAGES)

.PHONY: test.sync
test.sync:
	go run ../gnovm/cmd/gno test --verbose --update-golden-tests .

.PHONY: clean
clean:
	find . \( -name "*.gno.gen.go" -or -name ".*.gno.gen_test.go" \) -delete

.PHONY: fmt
GOFMT_FLAGS ?= -w
fmt:
	go run -modfile ../misc/devdeps/go.mod mvdan.cc/gofumpt $(GOFMT_FLAGS) `find . -name "*.gno"`

.PHONY: tidy
GNO_CMD = $(realpath ../gnovm/cmd/gno)
tidy:
	find . -name "gno.mod" -execdir go run $(GNO_CMD) mod tidy \;
