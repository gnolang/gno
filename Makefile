.PHONY: help
help:
	@echo "Available make commands:"
	@cat Makefile | grep '^[a-z][^:]*:' | cut -d: -f1 | sort | sed 's/^/  /'

rundep=go run -modfile ../misc/devdeps/go.mod

# shortcuts to frequently used commands from sub-components.
install_gnokey:
	go install ./gno.land/cmd/gnokey
	@echo "gnokey installed. more info in ./gno.land/"
install_gno:
	go install ./gnovm/cmd/gno
	@echo "gno installed. more info in ./gnovm/"

.PHONY: test
test:
	go test -count=1 -v ./misc/docker-integration

.PHONY: test.components
test.components:
	$(MAKE) -C tm2      test
	$(MAKE) -C gnovm    test
	$(MAKE) -C gno.land test
	$(MAKE) -C examples test

.PHONY: fmt
fmt:
	$(MAKE) -C tm2      fmt
	$(MAKE) -C gnovm    fmt
	$(MAKE) -C gno.land fmt
	$(MAKE) -C examples fmt

.PHONY: lint
lint:
	golangci-lint run --config .github/golangci.yml
