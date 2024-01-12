.PHONY: help
help:
	@echo "Available make commands:"
	@cat Makefile | grep '^[a-z][^:]*:' | cut -d: -f1 | sort | sed 's/^/  /'

.PHONY: install
install: install.gnomd install.gnodev

install.gnomd:;  cd gnomd && go install .
install.gnodev:; $(MAKE) -C ./gnodev install

.PHONY: clean
clean:
	rm -rf build

########################################
# Dev tools
rundep=go run -modfile ../misc/devdeps/go.mod

.PHONY: fmt
GOFMT_FLAGS ?= -w
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
GOTEST_FLAGS ?= -v -p 1 -timeout=30m

.PHONY: test
test:
	@echo "nothing to do."
