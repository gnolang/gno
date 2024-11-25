# configurable arguments
DEV_REMOTE ?= 127.0.0.1:26657
CHAINID ?= test3

# Variable Declarations
COMPONENTS_FOLDER := ./components

tools_run := go run -modfile ./tools/go.mod
tools_install := go install -modfile ./tools/go.mod

run_air := $(tools_run) github.com/air-verse/air

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

#############
# Targets
#############
.PHONY: all generate fmt css ts templ

# Install dependencies
all: generate

install: generate
	go install -v .

# Generate process
generate: clean css ts fonts
css:
	npx tailwindcss@$(tw_version) -c $(tw_config_path) -i $(input_css) -o $(output_css) --minify # tailwind
ts:
	npx -y esbuild $(input_js) --bundle --outdir=$(output_js) --format=esm --minify
fonts:
	mkdir -p $(output_statics)
	cp -r $(static_folder)/* $(output_statics)

# Format process
fmt:
	go fmt ./...
	$(run_templ) fmt $(COMPONENTS_FOLDER)

 ###############################
 # Developments
 ###############################
.PHONY: dev dev.templ dev.server dev.css dev.ts deps

# Run the development dependencies in parallel
dev:
	@echo "-- starting development tools"
	@$(MAKE) -j 4 \
		dev.gnoweb \
		dev.fonts \
		dev.ts \
		dev.css

# Go server in development mode
dev.gnoweb: | .cache
	@killall gnowebserve || true # make sure that previous instances as been kill
	$(run_air) \
		--build.cmd "go build -o .cache/gnowebserve ../../cmd/gonoweb" \
		--build.bin ".cache/gnowebserve -chain-id=${CHAIN_ID} -remote=${DEV_REMOTE}" \
		--build.exclude_unchanged "true" \
		--build.exclude_dir "gotools,node_modules" \
		--build.exclude_regex '^\.#' \
		--build.include_ext "go,gohtml" \
		--build.stop_on_error "false" \
		--build.rerun_delay "1000" \
		2>&1 | $(run_logname) gnoweb

# Tailwind CSS in development mode
dev.css: | public
	npx tailwindcss@$(tw_version) -c $(tw_config_path) --verbose -i $(input_css) -o $(output_css) --watch \
		2>&1 | $(run_logname) tailwind

# TS in development mode
dev.ts: | public
	npx -y esbuild $(input_js) \
		--bundle \
		--outdir=$(output_js) \
		--format=esm \
		--watch \
		--minify

dev.fonts:
	mkdir -p $(output_statics)
	cp -r $(static_folder)/* $(output_statics)

# Install deps for development purpose
deps:
	$(tools_install) github.com/a-h/templ/cmd/templ

# Cleanup
clean:
	rm -rf public tmp
fclean: clean
	rm -rf node_modules .cache

# Dirs
.cache:; mkdir -p $@
public:; mkdir -p $@
