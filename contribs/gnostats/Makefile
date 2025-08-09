rundep := go run -modfile ../../misc/devdeps/go.mod
golangci_lint := $(rundep) github.com/golangci/golangci-lint/cmd/golangci-lint

.PHONY: protoc
protoc:
	# Make sure the following prerequisites are installed before running these commands:
	# https://grpc.io/docs/languages/go/quickstart/#prerequisites
	protoc --proto_path=./ --go_out=./ --go-grpc_out=./ ./proto/*.proto

lint:
	$(golangci_lint) --config ../../.github/golangci.yml run ./...
