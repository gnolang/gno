# Configurable arguments
DEV_REMOTE ?= 127.0.0.1:26657
CHAIN_ID ?= test3
PUBLIC_DIR ?= public

# Variable Declarations
tools_run := go run -modfile ./tools/go.mod
run_reflex := $(tools_run) github.com/cespare/reflex
run_logname := go -C ./tools run ./cmd/logname

# css config
input_css := frontend/css/input.css
output_css := $(PUBLIC_DIR)/styles.css
tw_version := 3.4.14
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
esbuild_version := 0.24.0

# cache
cache_dir := .cache

#############
# Targets
#############
.PHONY: all generate fmt css ts

# Install dependencies
all: generate

test:
	go test -v ./...

# Generate process
generate: css ts static

css: $(output_css)
$(output_css): $(input_css) $(templates_files)
	npx -y tailwindcss@$(tw_version) -c $(tw_config_path) -i $(input_css) -o $@ --minify # tailwind
	touch $@

ts: $(output_js)
$(out_dir_js)/%.js: $(src_dir_js)/%.ts
	npx -y esbuild $< --log-level=error --bundle --outdir=$(out_dir_js) --format=esm --minify

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
.PHONY: dev dev.server dev.css dev.ts deps

# Run the development dependencies in parallel
dev:
	@echo "-- starting development tools"
	@PUBLIC_DIR=$(cache_dir)/public $(MAKE) -j 3 \
		dev.gnoweb \
		dev.ts \
		dev.css

# Go server in development mode
dev.gnoweb: generate
	$(run_reflex) -s -r '.*\.(go|html)' -- \
		go run ../../cmd/gnoweb -v -no-strict -assets-dir=${PUBLIC_DIR} -chainid=${CHAIN_ID} -remote=${DEV_REMOTE} \

		2>&1 | $(run_logname) gnoweb

# Tailwind CSS in development mode
dev.css: generate | $(PUBLIC_DIR)
	npx -y tailwindcss@$(tw_version) -c $(tw_config_path) --verbose -i $(input_css) -o $(output_css) --watch \
		2>&1 | $(run_logname) tailwind

# XXX: add versioning on esbuild
# TS in development mode
dev.ts: generate | $(PUBLIC_DIR)
	npx -y esbuild@$(esbuild_version) $(input_js) --bundle --outdir=$(out_dir_js) --sourcemap --format=esm --watch \
		2>&1 | $(run_logname) esbuild

# Cleanup
clean:
	rm -rf $(cache_dir) tmp
fclean: clean
	rm -rf $(PUBLIC_DIR)

# Dirs
$(PUBLIC_DIR):; mkdir -p $@
