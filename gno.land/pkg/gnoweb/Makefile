# Configurable arguments
DEV_REMOTE ?= 127.0.0.1:26657
CHAIN_ID ?= dev
PUBLIC_DIR ?= public

# tools
tools_run := go run -modfile ./tools/go.mod
run_reflex := $(tools_run) github.com/cespare/reflex
run_logname := go -C ./tools run ./cmd/logname

# css config
input_css := frontend/css/input.css
output_css := $(PUBLIC_DIR)/styles.css
tw_config_path := frontend/css/tx.config.js
templates_files := $(shell find . -iname '*.html')

# static config
tailwind := ./tools/node_modules/.bin/tailwindcss
src_dir_static := frontend/static
out_dir_static := $(PUBLIC_DIR)
input_static := $(shell find $(src_dir_static) -type f)
output_static := $(patsubst $(src_dir_static)/%, $(out_dir_static)/%, $(input_static))

# esbuild config
esbuild := ./tools/node_modules/.bin/esbuild
src_dir_js := frontend/js
out_dir_js := $(PUBLIC_DIR)/js
input_js := $(shell find $(src_dir_js) -name '*.ts')
output_js := $(patsubst $(src_dir_js)/%.ts,$(out_dir_js)/%.js,$(input_js))

# cache
cache_dir := .cache

#############
# Targets
#############
.PHONY: all generate fmt css ts

# Generate public dir
all: generate

# Run gnoweb in development mode
dev: dev.all

test:
	go test -v ./...

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

# Format process
fmt:
	go fmt ./...

 ###############################
 # Developments
 ###############################
.PHONY: dev dev.server dev.css dev.ts

# Run the development dependencies in parallel
dev.all: $(esbuild) $(tailwind) $(PUBLIC_DIR)
	@echo "-- starting development tools"
	PUBLIC_DIR=$(cache_dir)/public $(MAKE) -j 3 dev.gnoweb dev.ts dev.css

# Go server in development mode
dev.gnoweb: generate
	$(run_reflex) -s -r '.*\.(go|html)' -- \
		go run -tags=noembed -ldflags="-X 'github.com/gnolang/gno/gno.land/pkg/gnoweb.AssetDir=${PUBLIC_DIR}'" ../../cmd/gnoweb \
		-v -no-strict -chainid=${CHAIN_ID} -remote=${DEV_REMOTE} \
		2>&1 | $(run_logname) gnoweb

# Tailwind CSS in development mode
dev.css: $(tailwind) generate | $(PUBLIC_DIR)
	$(tailwind) -c $(tw_config_path) --verbose -i $(input_css) -o $(output_css) --watch \
		2>&1 | $(run_logname) tailwind

# TS in development mode
dev.ts: $(esbuild) generate | $(PUBLIC_DIR)
	$(esbuild) $(input_js) --bundle --outdir=$(out_dir_js) --sourcemap --format=esm --watch \
		2>&1 | $(run_logname) esbuild

# Install node modules deps located in ./tools
$(esbuild) $(tailwind): ./tools/node_modules
./tools/node_modules:
	cd tools && npm install
	touch $@

# Cleanup
clean:
	rm -rf $(cache_dir) tmp
fclean: clean
	rm -rf $(PUBLIC_DIR)
	rm -rf ./tools/node_modules


# Dirs
$(PUBLIC_DIR):; mkdir -p $@
