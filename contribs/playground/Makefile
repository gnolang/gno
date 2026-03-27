.PHONY: install build-wasm build-rootzip dev build build-release clean

GNOVM_DIR ?= ../../gnovm
EXAMPLES_DIR ?= ../../examples
TOOLS_DATA = tools/data
WASM_OUT = $(TOOLS_DATA)/wasm/gno.wasm
ROOTZIP_OUT = $(TOOLS_DATA)/gno/root.zip

# Release tag for downloading pre-built WASM from GitHub releases
TAG ?=
GH_RELEASE_BASE = https://github.com/gnolang/gno/releases/download

install:
	pnpm install

run:
	pnpm dev

build-wasm:
	GOOS=js GOARCH=wasm go build -o $(WASM_OUT) $(GNOVM_DIR)/cmd/gno

build-rootzip:
	@mkdir -p $(TOOLS_DATA)/gno
	@cd $(GNOVM_DIR)/.. && go run ./gnovm/cmd/gno env GNOROOT | xargs -I{} sh -c \
		'cd {} && zip -r -i "*.toml" -i "*.gno" -x "*_test.gno" - gnovm/stdlibs gnovm/tests/stdlibs examples' \
		> $(CURDIR)/$(ROOTZIP_OUT)

# Download WASM and root.zip from a GitHub release tag
download-release:
ifndef TAG
	$(error TAG is required, e.g. make download-release TAG=chain/gnoland1.0)
endif
	@mkdir -p $(TOOLS_DATA)/wasm $(TOOLS_DATA)/gno
	curl -sSfL "$(GH_RELEASE_BASE)/$(TAG)/gno.wasm" -o $(WASM_OUT)
	curl -sSfL "$(GH_RELEASE_BASE)/$(TAG)/root.zip" -o $(ROOTZIP_OUT)

dev: build-wasm build-rootzip install
	pnpm dev

build: build-wasm build-rootzip install
	pnpm build

build-release: download-release install
	pnpm build

clean:
	rm -f $(WASM_OUT) $(ROOTZIP_OUT)
	rm -rf app/dist
