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

# main css config — includes per-feature CSS pulled into the Cube CSS Block
# layer via @import from frontend/css/main.css (postcss-import resolves at
# compile time, so the final bundle and layer order are unchanged).
main_css_files := $(shell find frontend/css feature -name '*.css' 2>/dev/null)
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
# Pick up both legacy controllers under frontend/js/ and per-feature
# controllers under feature/<name>/frontend/. All compile to
# $(out_dir_js)/<basename>.js (flat output, no source-path mirroring).
# code-editor.ts is a library the feature controllers import (it ships in
# their shared chunk-<hash>.js), not a page entry point; so is any feature
# module not named controller-*.
feature_js := $(sort $(shell find feature -path 'feature/*/frontend/controller-*.ts' 2>/dev/null))
input_js := $(shell find $(src_dir_js) -name '*.ts' -not -name '*.d.ts' -not -name 'code-editor.ts') \
            $(feature_js)
# Guard against silent overwrites: two features cannot ship the same
# controller-X.ts basename (they would compile to the same target).
_dup_check := $(shell echo "$(notdir $(input_js))" | tr ' ' '\n' | sort | uniq -d)
ifneq ($(_dup_check),)
$(error duplicate JS controller basenames: $(_dup_check))
endif
# Separate shared and controller files
shared_js := $(src_dir_js)/controller.ts
# Flatten: any .ts (legacy or feature) → $(out_dir_js)/<basename>.js.
output_js := $(addprefix $(out_dir_js)/,$(patsubst %.ts,%.js,$(notdir $(input_js))))
feature_out_js := $(addprefix $(out_dir_js)/,$(patsubst %.ts,%.js,$(notdir $(feature_js))))
# Tell make where to look for .ts sources so the %.js pattern rule below
# resolves the legacy controllers.
vpath %.ts $(src_dir_js)

# Per-feature controllers import shared modules as @gnoweb/js/..., through the
# `paths` of tsconfig.json, which esbuild reads. The one exception is the
# shared chunk every page loads: it stays an import of ./controller.js, so a
# feature bundle uses that chunk instead of carrying its own copy.
esbuild_controller_alias := --alias:@gnoweb/js/controller.js=./controller.js

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
test.front:; $(MAKE) -C frontend test

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

# Community share cards, from their wording in frontend/og-cards/generate.sh
# (macOS only). Run `make static` after to copy them to public/.
.PHONY: og-cards
og-cards:
	./frontend/og-cards/generate.sh

# Icons callable as <gno-icon />: regenerate the vendored set from
# markdown/icons/icons.txt, then the icon table (markdown/icons_gen.go) and
# the catalog (ICONS.md, ICONS.svg). Two test runs: the catalog reads the
# table compiled by the first.
.PHONY: icons
icons:
	go -C ./tools run ./cmd/iconset $(CURDIR)
	go test ./markdown -run 'TestIconTable$$' -update-golden-tests
	go test ./markdown -run 'TestIconCatalog$$' -update-golden-tests

# Main CSS
maincss: $(output_maincss)
$(output_maincss): $(main_css_files) $(templates_files) $(postcss)
	$(postcss) $(css_config_path) -o $@ --env production
	touch $@

# Per-feature controllers live under feature/<name>/frontend/, outside
# frontend/js/. NODE_PATH points esbuild at the single frontend/node_modules
# so npm imports (e.g. "htmx.org") resolve from any source location.
node_path := $(node_modules)

ts: $(esbuild) $(output_js)
# Build shared chunk first (always loaded). It bundles the modules it
# imports (utils.ts), and so do the controllers built after it, so any
# frontend module rebuilds it.
$(out_dir_js)/controller.js: $(shared_js) $(shell find $(src_dir_js) -name '*.ts' -not -name '*.d.ts') $(esbuild)
	NODE_ENV=production NODE_PATH=$(node_path) $(esbuild) $< --log-level=error --bundle --outfile=$@ --format=esm --minify

# sa-bootstrap is loaded as a synchronous classic <script src=...> so it blocks
# parsing until executed and guarantees window.sa_metadata is set before SA's
# async latest.js loads. It must be an IIFE, not an ES module.
$(out_dir_js)/sa-bootstrap.js: $(src_dir_js)/sa-bootstrap.ts $(esbuild)
	NODE_ENV=production $(esbuild) $< --log-level=error --outfile=$@ --format=iife --minify

# Build controller files with shared chunk reference. The bare %.ts
# prerequisite lets `vpath` resolve sources from frontend/js/ (set above).
# $< is the path Make found.
$(out_dir_js)/%.js: %.ts $(out_dir_js)/controller.js
	NODE_ENV=production NODE_PATH=$(node_path) $(esbuild) $< --log-level=error --bundle --outdir=$(out_dir_js) --format=esm --define:process.env.NODE_ENV="\"production\"" --minify --external:./controller.js

# Feature controllers build in one esbuild run with --splitting, so code that
# several of them import (CodeMirror) ships once, as a chunk-<hash>.js they
# all import, instead of once per controller. The run writes every feature
# controller: the first one is the target, the others follow from it. Chunks
# of a previous build are removed first, as their names change with content.
# Any frontend module may be imported, so every one is a prerequisite. If a
# follower or the chunk is missing, the run is forced, as make would otherwise
# find nothing to do.
feature_js_first := $(firstword $(feature_out_js))
feature_js_missing := $(filter-out $(wildcard $(feature_out_js)),$(feature_out_js)) \
                      $(if $(wildcard $(out_dir_js)/chunk-*.js),,chunk)
$(feature_js_first): $(shell find feature -path 'feature/*/frontend/*.ts' 2>/dev/null) \
                     $(shell find $(src_dir_js) -name '*.ts' -not -name '*.d.ts') \
                     $(out_dir_js)/controller.js $(if $(strip $(feature_js_missing)),FORCE)
	rm -f $(out_dir_js)/chunk-*.js
	NODE_ENV=production NODE_PATH=$(node_path) $(esbuild) $(feature_js) --log-level=error --bundle --splitting --outdir=$(out_dir_js) --entry-names=[name] --format=esm --define:process.env.NODE_ENV="\"production\"" --minify --external:./controller.js $(esbuild_controller_alias)
$(filter-out $(feature_js_first),$(feature_out_js)): $(feature_js_first) ;
.PHONY: FORCE
FORCE:

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
	NODE_ENV=development NODE_PATH=$(node_path) $(esbuild) $(input_js) --bundle --outdir=$(out_dir_js) --sourcemap --format=esm --watch \
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
