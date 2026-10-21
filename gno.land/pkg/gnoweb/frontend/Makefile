# Per-feature TS/CSS lives outside frontend/ (under pkg/gnoweb/feature/<name>/frontend/).
# Pass those paths explicitly so biome lints them alongside the legacy tree.
feature_paths := $(shell find ../feature -path '*/frontend' -type d 2>/dev/null)

lint: node_modules
	npx @biomejs/biome check . $(feature_paths)
lint.ci: node_modules
	npx @biomejs/biome ci . $(feature_paths)

fmt: node_modules
	npx @biomejs/biome check --write . $(feature_paths)

clean fclean:
	rm -rf node_modules

node_modules: package-lock.json package.json
	npm install && touch $@

# Unit tests: esbuild bundles each tests/*.test.ts for node, then node's own
# runner executes them. No test framework dependency.
test_out := .cache/test

.PHONY: test
test: node_modules
	rm -rf $(test_out)
	./node_modules/.bin/esbuild tests/*.test.ts --bundle --platform=node --format=esm --outdir=$(test_out) --out-extension:.js=.mjs --log-level=error
	node --test '$(test_out)/*.test.mjs'
