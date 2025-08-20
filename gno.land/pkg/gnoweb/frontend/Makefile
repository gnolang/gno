lint: node_modules
	npx @biomejs/biome check
lint.ci: node_modules
	npx @biomejs/biome ci

fmt: node_modules
	npx @biomejs/biome check --write

clean fclean:
	rm -rf node_modules

node_modules: package-lock.json package.json
	npm install && touch $@
