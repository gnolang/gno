GOTEST_FLAGS ?= $(GOBUILD_FLAGS) -v -p 1 -timeout=5m

.PHONY: install
install:
	go install .

.PHONY: build
build:
	go build -o build/gnodepgraph .

test:
	go test -v -p 1 -timeout=5m ./...

lint:
	@echo "XXX: add lint"
