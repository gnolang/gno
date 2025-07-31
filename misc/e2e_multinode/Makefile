.PHONY: build install test clean

BIN=build/e2e_multinode

# Build the binary
build/e2e_multinode:
	@mkdir -p $(dir $@)
	go build -o $@

# Install the binary to GOPATH/bin
install:
	go install .

# Run go tests
test:
	go test -v ./...

# Clean build artifacts
clean:
	rm -f build
