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
	go run ../gnovm/cmd/gno test --verbose ./...

.PHONY: test.sync
test.sync:
	go run ../gnovm/cmd/gno test --verbose --update-golden-tests ./...

.PHONY: clean
clean:
	find . \( -name "*.gno.gen.go" -or -name ".*.gno.gen_test.go" \) -delete

.PHONY: fmt
GOFMT_FLAGS ?= -w
fmt:
	go run -modfile ../misc/devdeps/go.mod mvdan.cc/gofumpt $(GOFMT_FLAGS) `find . -name "*.gno"`
