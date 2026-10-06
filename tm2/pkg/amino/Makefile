
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
# See README.md for documentation on test and fuzz targets.

test:
	go test $(shell go list ./... | grep -v vendor | grep -v fuzz)
	go test $(shell go list ./... | grep -v vendor | grep fuzz)

FUZZTIME ?= 1h

# Run both fuzzers concurrently for FUZZTIME total:
#   - Property fuzz: TestCodec{Struct,Def,AminoTags} with gofuzz-generated values,
#     asserting amino-reflect / pbbindings / genproto2 byte parity and roundtrip.
#   - Coverage fuzz: FuzzUnmarshalBinary2 with random bytes, asserting no panics.
# Usage: make fuzz [FUZZTIME=30m]
#        make fuzz FUZZTIME=0    # run forever (Ctrl-C to stop)
# Exits non-zero if either fuzzer reports a failure.
fuzz:
	@echo "=== Running property and coverage fuzzers in parallel for $(FUZZTIME) ==="
	@( while true; do \
	  AMINO_FUZZ_BUDGET=$(FUZZTIME) go test -v -count=1 ./... -run 'TestCodec(Struct|Def|AminoTags)' -timeout=0 2>&1; \
	  E=$$?; \
	  if [ "$$E" != "0" ]; then echo "FAIL: property fuzz exit=$$E" >&2; exit $$E; fi; \
	  if [ "$(FUZZTIME)" != "0" ]; then break; fi; \
	  echo "=== property fuzz: starting next rotation ===" >&2; \
	done > /tmp/amino-property-fuzz.log 2>&1 ) & P1=$$!; \
	go test ./tests/fuzz/binary/ -fuzz=FuzzUnmarshalBinary2 -fuzztime=$(FUZZTIME) -timeout=0 2>&1 \
	  | tee /tmp/amino-coverage-fuzz.log & P2=$$!; \
	( while kill -0 $$P1 2>/dev/null; do \
	  sleep 30; \
	  echo "[property fuzz] passed=$$(cat /tmp/amino-property-fuzz.log 2>/dev/null | grep -c -- '--- PASS') failed=$$(cat /tmp/amino-property-fuzz.log 2>/dev/null | grep -c -- '--- FAIL')"; \
	done ) & P3=$$!; \
	wait $$P1; E1=$$?; \
	wait $$P2; E2=$$?; \
	kill $$P3 2>/dev/null; wait $$P3 2>/dev/null; \
	echo "=== property fuzz summary ==="; grep -E '^\s*--- (PASS|FAIL)' /tmp/amino-property-fuzz.log | tail -20; \
	echo "=== coverage fuzz summary ==="; tail -5 /tmp/amino-coverage-fuzz.log; \
	P_FAIL=$$(grep -cE '^\s*--- FAIL' /tmp/amino-property-fuzz.log || true); \
	if [ "$$P_FAIL" -gt 0 ]; then \
	  echo "FAIL: property fuzz had $$P_FAIL failing subtests (see /tmp/amino-property-fuzz.log)"; exit 1; \
	fi; \
	if [ "$$E2" != "0" ]; then \
	  echo "FAIL: coverage fuzz exit=$$E2 (see /tmp/amino-coverage-fuzz.log)"; exit 1; \
	fi; \
	echo "PASS (ran for $(FUZZTIME) each, in parallel)"

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
.PHONY: build install check_tools get_tools fmt lint test fuzz
