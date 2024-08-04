.PHONY: opcode storage build_opcode build_storage

# Build target
build_opcode:
	go build -tags "benchmarkingops" -o build/gnobench ./cmd

# Build target
build_storage:
	go build -tags "benchmarkingstorage" -o build/gnobench ./cmd


# Run target
opcode: build_opcode
	./build/gnobench

# Run target
storage: build_storage
	./build/gnobench

# command to run dependency utilities, like goimports.
rundep=go run -modfile ../misc/devdeps/go.mod

# flags for `make fmt`. -w will write the result to the destination files.
GOFMT_FLAGS ?= -w

.PHONY: fmt
fmt:
	$(rundep) mvdan.cc/gofumpt $(GOFMT_FLAGS) .
