GNOROOT_DIR ?= $(abspath $(lastword $(MAKEFILE_LIST))/../../../)

GOBUILD_FLAGS := -ldflags "-X github.com/gnolang/gno/gnovm/pkg/gnoenv._GNOROOT=$(GNOROOT_DIR)"

install:
	go install $(GOBUILD_FLAGS) .

build:
	go build $(GOBUILD_FLAGS) -o build/gnodev ./cmd/gno
