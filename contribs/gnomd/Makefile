rundep := go run -modfile ../../misc/devdeps/go.mod
golangci_lint := $(rundep) github.com/golangci/golangci-lint/v2/cmd/golangci-lint


.PHONY: install test lint
install:
	go install .

test:
	@echo "XXX: add tests"

lint:
	$(golangci_lint) --config ../../.github/golangci.yml run ./...
