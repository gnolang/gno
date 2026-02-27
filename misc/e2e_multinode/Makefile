.PHONY: build install test clean

# Build the binary
build: build/e2e_multinode
build/e2e_multinode: *.go
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
