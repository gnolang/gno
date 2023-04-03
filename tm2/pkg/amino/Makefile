
all: check_tools test

########################################
###  Build

build:
	# Nothing to build!

install:
	# Nothing to install!


########################################
### Tools & dependencies

include tools.mk

########################################
### Testing

test:
	go test $(shell go list ./... | grep -v vendor | grep -v fuzz)
	go test $(shell go list ./... | grep -v vendor | grep fuzz)

gofuzz_binary:
	rm -rf tests/fuzz/binary/corpus/
	rm -rf tests/fuzz/binary/crashers/
	rm -rf tests/fuzz/binary/suppressions/
	go run tests/fuzz/binary/init-corpus/main.go --corpus-parent=tests/fuzz/binary
	go-fuzz-build github.com/gnolang/gno/tm2/pkg/amino/tests/fuzz/binary
	go-fuzz -bin=./fuzzbinary-fuzz.zip -workdir=tests/fuzz/binary
	rm -rf ./fuzzbinary-fuzz.zip

gofuzz_json:
	rm -rf tests/fuzz/json/corpus/
	rm -rf tests/fuzz/json/crashers/
	rm -rf tests/fuzz/json/suppressions/
	go-fuzz-build github.com/gnolang/gno/tm2/pkg/amino/tests/fuzz/json
	go-fuzz -bin=./fuzzjson-fuzz.zip -workdir=tests/fuzz/json
	rm -rf ./fuzzjson-fuzz.zip

########################################
### Formatting, linting, and vetting

fmt:
	@go fmt ./...

# look into .golangci.yml for enabling / disabling linters
lint:
	@echo "--> Running linter"
	@golangci-lint run
	find . -name '*.go' -type f -not -path "./vendor*" -not -path "*.git*" | xargs gofmt -d -s
	go mod verify

# To avoid unintended conflicts with file names, always add to .PHONY
# unless there is a reason not to.
# https://www.gnu.org/software/make/manual/html_node/Phony-Targets.html
.PHONY: build install check_tools get_tools fmt lint test
