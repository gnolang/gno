rundep := go run -modfile ../../misc/devdeps/go.mod
golangci_lint := $(rundep) github.com/golangci/golangci-lint/v2/cmd/golangci-lint

all: build

.PHONY: install
install:
	go install cmd

.PHONY: build
build:
	@echo "Building archive binary"
	go build -o build/archive ./cmd

.PHONY: lint
lint:
	$(golangci_lint) --config ../../.github/golangci.yml run ./...

.PHONY: test
test:
	go test $(GOTEST_FLAGS) -v ./...

.PHONY: gofumpt
gofumpt:
	go install mvdan.cc/gofumpt@latest
	gofumpt -l -w .

.PHONY: fixalign
fixalign:
	go install golang.org/x/tools/go/analysis/passes/fieldalignment/cmd/fieldalignment@latest
	fieldalignment -fix $(filter-out $@,$(MAKECMDGOALS)) # the full package name (not path!)
