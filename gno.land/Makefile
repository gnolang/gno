include ../misc/makehelp/makeHelpMsgs.mk

.PHONY: help
help: # Print this help message
	@$(call RUN_MAKEFILE_HELP,..,)

# command to run dependency utilities, like goimports.
rundep=go run -modfile ../misc/devdeps/go.mod

########################################
# Environment variables
# You can overwrite any of the following by passing a different value on the
# command line, ie. `CGO_ENABLED=1 make test`.

# disable cgo by default. cgo requires some additional dependencies in some
# cases, and is not strictly required by any tm2 code.
CGO_ENABLED ?= 0
export CGO_ENABLED
# test suite flags.
GOTEST_FLAGS ?= -v -p 1 -timeout=30m

########################################
# Dev tools
.PHONY: start.gnoland
start.gnoland:  # Start the gno.land node with lazy loading enabled
	go run ./cmd/gnoland start -lazy

.PHONY: start.gnoweb
start.gnoweb:  # Start the gnoweb frontend service
	go run ./cmd/gnoweb

.PHONY: build
build: build.gnoland build.gnokey build.gnoweb # Build gnoland, gnokey, and gnoweb binaries into ./build

build.gnoland:  # Build the gnoland binary into ./build/gnoland
	go build -o build/gnoland   ./cmd/gnoland
build.gnoweb:  # Build the gnoweb binary into ./build/gnoweb
	go build -o build/gnoweb    ./cmd/gnoweb
build.gnokey:  # Build the gnokey binary into ./build/gnokey
	go build -o build/gnokey    ./cmd/gnokey

run.gnoland:  # Run the gnoland node directly from source
	go run ./cmd/gnoland start
run.gnoweb:  # Run the gnoweb frontend directly from source
	go run ./cmd/gnoweb

.PHONY: install
install: install.gnoland install.gnoweb install.gnokey  # Install gnoland, gnoweb, and gnokey into $GOBIN

install.gnoland:  # Install the gnoland CLI binary
	go install ./cmd/gnoland
install.gnoweb:  # Install the gnoweb frontend CLI binary
	go install ./cmd/gnoweb
install.gnokey:  # Install the gnokey CLI binary
	go install ./cmd/gnokey

.PHONY: dev.gnoweb generate.gnoweb
dev.gnoweb:  # Run gnoweb in development mode (delegates to pkg/gnoweb)
	make -C ./pkg/gnoweb dev

.PHONY: generate
generate:  # Run go generate for the entire repo and then regenerate gnoweb
	go generate -x ./...
	make -C ./pkg/gnoweb generate


.PHONY: fclean
fclean: clean  # Clean build artifacts and remove gnoland-data and genesis.json
	rm -rf gnoland-data genesis.json

.PHONY: clean
clean:  # Remove build artifacts from ./build
	rm -rf build

.PHONY: lint
lint:  # Run golangci-lint on all Go code in gno.land using the shared config
	$(rundep) github.com/golangci/golangci-lint/cmd/golangci-lint run --config ../.github/golangci.yml ./...

.PHONY: fmt
fmt:  # Format Go code in this directory recursively using gofumpt and goimports
	go fmt ./...
	$(rundep) golang.org/x/tools/cmd/goimports -w .

########################################
# Test suite
.PHONY: test
test: _test.help _test.gnoland _test.gnoweb _test.gnokey _test.pkgs # Run all test targets (gnoland, gnoweb, gnokey, and pkg/*)

_test.help:
	@echo "run \`INMEMORY_TS=true make test\` if you encounter 'context deadline exceeded' for non-parallel testing" 
_test.gnoland:
	go test $(GOTEST_FLAGS) ./cmd/gnoland
_test.gnoweb:
	go test $(GOTEST_FLAGS) ./cmd/gnoweb
_test.gnokey:
	go test $(GOTEST_FLAGS) ./cmd/gnokey
_test.pkgs:
	go test $(GOTEST_FLAGS) ./pkg/...
_test.pkgs.sync:
	UPDATE_SCRIPTS=true go test $(GOTEST_FLAGS) ./pkg/...
