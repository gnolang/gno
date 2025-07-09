lint: node_modules
	npx @biomejs/biome check

fmt: node_modules
	npx @biomejs/biome check --write

fclean:
	rm -rf node_modules

node_modules: package-lock.json package.json
	npm install && touch $@
