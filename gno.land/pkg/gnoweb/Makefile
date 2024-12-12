# XXX: refine and cleanup this Makefile. Most of these were designed to be used
# in a single monorepo

# configurable arguments
DEV_REMOTE ?= 127.0.0.1:26657
CHAIN_ID ?= test3

# Variable Declarations
COMPONENTS_FOLDER := ./components

tools_run := go run -modfile ./tools/go.mod

run_air := $(tools_run) github.com/air-verse/air
run_reflex := $(tools_run) github.com/cespare/reflex

run_logname := go -C ./tools run ./cmd/logname

# css config
input_css := ./frontend/css/input.css
output_css := ./public/styles.css
tw_version := 3.4.14
tw_config_path := ./frontend/css/tx.config.js

# static config
static_folder := ./frontend/static
output_statics := ./public

# Esbuild config
input_js := ./frontend/js/**/*.ts
output_js := ./public/js
esbuild_version := 0.24.0

#############
# Targets
#############
.PHONY: all generate fmt css ts

# Install dependencies
all: generate

# Generate process
generate: clean css ts fonts
css:
	npx -y tailwindcss@$(tw_version) -c $(tw_config_path) -i $(input_css) -o $(output_css) --minify # tailwind
ts:
	npx -y esbuild $(input_js) --bundle --outdir=$(output_js) --format=esm --minify
fonts:
	mkdir -p $(output_statics)
	cp -v -r $(static_folder)/* $(output_statics)

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
	@$(MAKE) -j 3 \
		dev.gnoweb \
		dev.ts \
		dev.css

# Go server in development mode
dev.gnoweb: | .cache
	$(run_reflex) -s -r '.*\.go(html)?' -- \
		go run ../../cmd/gnoweb -chainid=${CHAIN_ID} -remote=${DEV_REMOTE} \
		2>&1 | $(run_logname) gnoweb

# Tailwind CSS in development mode
dev.css: | public
	npx -y tailwindcss@$(tw_version) -c $(tw_config_path) --verbose -i $(input_css) -o $(output_css) --watch \
		2>&1 | $(run_logname) tailwind

# XXX: add versioning on esbuild
# TS in development mode
dev.ts: | public
	npx -y esbuild@$(esbuild_version) $(input_js) --bundle --outdir=$(output_js) --format=esm --minify --watch \
		2>&1 | $(run_logname) esbuild

# Cleanup
clean:
	rm -rf public tmp
fclean: clean
	rm -rf .cache

# Dirs
.cache public:; mkdir -p $@
