lint:
	go run -C ../misc/docs/tools/linter . -path "$(PWD)" -treat-urls-as-err=true

generate:
	go run -modfile ../misc/devdeps/go.mod github.com/campoy/embedmd -w `find . -name "*.md"`
	go run -C ../misc/docs/tools/indexparser . -path "$(PWD)/README.md" > ../misc/docs/sidebar.json

test:
	@echo "Nothing to do."
