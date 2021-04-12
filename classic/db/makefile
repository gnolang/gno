GOTOOLS = github.com/golangci/golangci-lint/cmd/golangci-lint
PACKAGES=$(shell go list ./...)

export GO111MODULE = on

all: lint test

### go tests
test:
	@echo "--> Running go test"
	@go test -p 1 $(PACKAGES)

lint:
	@echo "--> Running linter"
	@golangci-lint run

tools:
	go get -v $(GOTOOLS)
