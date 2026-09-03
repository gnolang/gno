rundep := go run -modfile ../devdeps/go.mod
golangci_lint := $(rundep) github.com/golangci/golangci-lint/v2/cmd/golangci-lint

GOTEST_FLAGS ?= -v -timeout 30m

# Every directory of scenarios. A scenario declares its own cluster, so a new
# shape costs a directory, not a target.
SCENARIO_DIRS := $(wildcard testdata/*/)

.PHONY: install
install:
	go install .

.PHONY: build
build:
	go build -o build/gnoe2e .

.PHONY: lint
lint:
	$(golangci_lint) --config ../../.github/golangci.yml run ./...

.PHONY: test
test:
	go test $(GOTEST_FLAGS) ./...

# test-master runs the same suite against gnoland and gpao built from master,
# so a scenario written for a fix can be shown red before the fix lands. The
# ref is unpacked with git archive into a temp dir the trap removes, so a
# Ctrl-C leaves nothing registered in .git.
.PHONY: test-master
test-master:
	tmp="$$(mktemp -d)" && trap 'rm -rf "$$tmp"' EXIT && \
	git -C ../.. archive master | tar -x -C "$$tmp" && \
	GNOROOT="$$tmp" go test $(GOTEST_FLAGS) ./...

.PHONY: test-short
test-short:
	go test -short ./...

# The CLI route: coloured output, and every scenario runs even after one
# fails. Use "test" for the go test route.
.PHONY: scenarios
scenarios: build
	build/gnoe2e run -verbose $(SCENARIO_DIRS)
