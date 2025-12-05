rundep := go run -modfile ../../misc/devdeps/go.mod
golangci_lint := $(rundep) github.com/golangci/golangci-lint/cmd/golangci-lint


.PHONY: install
install:
	go install ./cmd/tessera

.PHONY: build
build:
	go build -o build/tessera ./cmd/tessera/

lint:
	$(golangci_lint) --config ../../.github/golangci.yml run ./...

test:
	go test $(GOTEST_FLAGS) -v ./...

