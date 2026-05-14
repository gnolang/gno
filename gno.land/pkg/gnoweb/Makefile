# Configurable arguments
DEV_REMOTE ?= 127.0.0.1:26657
CHAIN_ID ?=
PUBLIC_DIR ?= public
ALIASES ?=

# tools

## go tools
reflex := go run -modfile ./tools/go.mod github.com/cespare/reflex
logname := go -C ./tools run ./cmd/logname
golangci_lint := go run -modfile ../../../misc/devdeps/go.mod github.com/golangci/golangci-lint/v2/cmd/golangci-lint

## node tools
node_modules := ./frontend/node_modules
esbuild := $(node_modules)/.bin/esbuild
postcss := $(node_modules)/.bin/postcss
biome := $(node_modules)/.bin/biome

# css config
css_config_path := frontend/css/main.css

# main css config
main_css_files := $(shell find frontend/css -name '*.css')
output_maincss := $(PUBLIC_DIR)/main.css

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
# Separate shared and controller files
shared_js := $(src_dir_js)/controller.ts
controller_js := $(filter-out $(shared_js),$(input_js))
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
	go -C ../../../contribs/gnodev run -tags=noembed . -no-web

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
generate: maincss ts static

# Main CSS
maincss: $(output_maincss)
$(output_maincss): $(main_css_files) $(templates_files) $(postcss)
	$(postcss) $(css_config_path) -o $@ --env production
	touch $@

ts: $(esbuild) $(output_js)
# Build shared chunk first (always loaded)
$(out_dir_js)/controller.js: $(shared_js) $(esbuild)
	NODE_ENV=production $(esbuild) $< --log-level=error --bundle --outfile=$@ --format=esm --minify

# Build controller files with shared chunk reference
$(out_dir_js)/%.js: $(src_dir_js)/%.ts $(out_dir_js)/controller.js
	NODE_ENV=production $(esbuild) $< --log-level=error --bundle --outdir=$(out_dir_js) --format=esm --define:process.env.NODE_ENV="\"production\"" --minify --external:./controller.js

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
dev.all: $(esbuild) $(PUBLIC_DIR)
	@echo "-- starting development tools"
	PUBLIC_DIR=$(cache_dir)/public ALIASES="$(ALIASES)" $(MAKE) -j 3 dev.gnoweb dev.ts dev.maincss

# Go server in development mode
dev.gnoweb: generate
	$(reflex) -s -r '.*\.(go|html)' -- \
		go run -tags=noembed -ldflags="-X 'github.com/gnolang/gno/gno.land/pkg/gnoweb.AssetDir=${PUBLIC_DIR}'" ../../cmd/gnoweb \
		-v -no-strict -chainid=${CHAIN_ID} -remote=${DEV_REMOTE} \
		$(if $(ALIASES),-aliases="$(ALIASES)",) \
		2>&1 | $(logname) gnoweb


# Main CSS in development mode
dev.maincss: generate | $(PUBLIC_DIR)
	$(postcss) $(css_config_path) -o $(output_maincss) --watch --env development \
		2>&1 | $(logname) maincss

# TS in development mode
dev.ts: $(esbuild) generate | $(PUBLIC_DIR)
	NODE_ENV=development $(esbuild) $(input_js) --bundle --outdir=$(out_dir_js) --sourcemap --format=esm --watch \
		2>&1 | $(logname) esbuild

# Install node modules deps located in ./tools
$(esbuild) $(postcss): $(node_modules)
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
