rundep := go run -modfile ../../misc/devdeps/go.mod
golangci_lint := $(rundep) github.com/golangci/golangci-lint/v2/cmd/golangci-lint

.PHONY: install
install:
	go install .

.PHONY: build
build:
	go build -o build/gnodrand .

.PHONY: lint
lint:
	$(golangci_lint) --config ../../.github/golangci.yml run --build-tags e2e ./...

.PHONY: test
test:
	go test ./...

# Needs gnodev on PATH and network access to drand.
.PHONY: test.e2e
test.e2e:
	go test -tags e2e -count=1 -v ./e2e
