###
# Find OS and Go environment
# GO contains the Go binary
# FS contains the OS file separator
###
ifeq ($(OS),Windows_NT)
  GO := $(shell where go.exe 2> NUL)
  FS := "\\"
else
  GO := $(shell command -v go 2> /dev/null)
  FS := "/"
endif

ifeq ($(GO),)
  $(error could not find go. Is it in PATH? $(GO))
endif

all: tools

tools: go-fuzz go-fuzz-build

TOOLS_DESTDIR  ?= $(GOPATH)/bin

GOFUZZ 				= $(TOOLS_DESTDIR)/go-fuzz 
GOFUZZ_BUILD 	= $(TOOLS_DESTDIR)/go-fuzz-build

# Install the runsim binary with a temporary workaround of entering an outside
# directory as the "go get" command ignores the -mod option and will pollute the
# go.{mod, sum} files.
# 
# ref: https://github.com/golang/go/issues/30515
go-fuzz: $(GOFUZZ)
$(GOFUZZ):
	@echo "Installing go-fuzz..."
	@(cd /tmp && go get -u github.com/dvyukov/go-fuzz/go-fuzz)

go-fuzz-build: $(GOFUZZ_BUILD)
$(GOFUZZ_BUILD):
	@echo "Installing go-fuzz-build..."
	@(cd /tmp && go get -u github.com/dvyukov/go-fuzz/go-fuzz-build)

tools-clean:
	rm -f $(GOFUZZ_BUILD) $(GOFUZZ)
	rm -f tools-stamp

.PHONY: all tools tools-clean