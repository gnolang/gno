# Configurable arguments
DEV_REMOTE ?= 127.0.0.1:26657
CHAIN_ID ?=
PUBLIC_DIR ?= public

# tools

## go tools
reflex := go run -modfile ./tools/go.mod github.com/cespare/reflex
logname := go -C ./tools run ./cmd/logname
golangci_lint := go run -modfile ../../../misc/devdeps/go.mod github.com/golangci/golangci-lint/v2/cmd/golangci-lint

## node tools
node_modules := ./frontend/node_modules
esbuild := $(node_modules)/.bin/esbuild
tailwind := $(node_modules)/.bin/tailwindcss
biome := $(node_modules)/.bin/biome

# css config
input_css := frontend/css/input.css
output_css := $(PUBLIC_DIR)/styles.css
tw_config_path := frontend/css/tx.config.js
templates_files := $(shell find . -iname '*.html')

# static config
src_dir_static := frontend/static
out_dir_static := $(PUBLIC_DIR)
input_static := $(shell find $(src_dir_static) -type f)
output_static := $(patsubst $(src_dir_static)/%, $(out_dir_static)/%, $(input_static))

# esbuild config
src_dir_js := frontend/js
out_dir_js := $(PUBLIC_DIR)/js
input_js := $(shell find $(src_dir_js) -name '*.ts')
output_js := $(patsubst $(src_dir_js)/%.ts,$(out_dir_js)/%.js,$(input_js))

# cache
cache_dir := .cache

#####################
# General Targets
#####################
.PHONY: all dev start-node generate fmt css ts

# Generate public dir
all: generate

# Run gnoweb in development mode
dev: dev.all

# Run gnodev instance
node-start:
	go -C ../../../contribs/gnodev run . -no-web

# Test process
test:
	go test -v ./...

# Format process
fmt: fmt.go fmt.front
fmt.go:; go fmt ./...
fmt.front:; $(MAKE) -C frontend fmt

# Lint process
lint: lint.go lint.front
lint.go:; $(golangci_lint) --config ../../../.github/golangci.yml run ./...
lint.front:; $(MAKE) -C frontend lint

# Generate process
generate: css ts static

css: $(tailwind) $(output_css)
$(output_css): $(input_css) $(templates_files)
	$(tailwind) -c $(tw_config_path) -i $(input_css) -o $@ --minify # tailwind
	touch $@

ts: $(esbuild) $(output_js)
$(out_dir_js)/%.js: $(src_dir_js)/%.ts
	$(esbuild) $< --log-level=error --bundle --outdir=$(out_dir_js) --format=esm --minify

# Rule to copy static files while preserving directory structure
static: $(output_static)
$(out_dir_static)/%: $(src_dir_static)/%
	@mkdir -p $(dir $@)
	@cp -v $< $@

 ###############################
 # Developments
 ###############################
.PHONY: dev.all dev.server dev.css dev.ts

# Run the development dependencies in parallel
dev.all: $(esbuild) $(tailwind) $(PUBLIC_DIR)
	@echo "-- starting development tools"
	PUBLIC_DIR=$(cache_dir)/public $(MAKE) -j 3 dev.gnoweb dev.ts dev.css

# Go server in development mode
dev.gnoweb: generate
	$(reflex) -s -r '.*\.(go|html)' -- \
		go run -tags=noembed -ldflags="-X 'github.com/gnolang/gno/gno.land/pkg/gnoweb.AssetDir=${PUBLIC_DIR}'" ../../cmd/gnoweb \
		-v -no-strict -chainid=${CHAIN_ID} -remote=${DEV_REMOTE} \
		2>&1 | $(logname) gnoweb

# Tailwind CSS in development mode
dev.css: $(tailwind) generate | $(PUBLIC_DIR)
	$(tailwind) -c $(tw_config_path) --verbose -i $(input_css) -o $(output_css) --watch \
		2>&1 | $(logname) tailwind

# TS in development mode
dev.ts: $(esbuild) generate | $(PUBLIC_DIR)
	$(esbuild) $(input_js) --bundle --outdir=$(out_dir_js) --sourcemap --format=esm --watch \
		2>&1 | $(logname) esbuild

# Install node modules deps located in ./tools
$(esbuild) $(tailwind): $(node_modules)
$(node_modules):
	$(MAKE) -C frontend node_modules

# Cleanup
clean:
	rm -rf $(cache_dir) tmp
fclean: clean
	rm -rf $(PUBLIC_DIR)
	rm -rf $(node_modules)

# Dirs
$(PUBLIC_DIR):; mkdir -p $@
