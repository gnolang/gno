lint:
	go run -C ../misc/docs/tools/linter . -path "$(PWD)" -treat-urls-as-err=true

generate:
	go run -modfile ../misc/devdeps/go.mod github.com/campoy/embedmd -w `find . -name "*.md"`
	go run -C ../misc/docs/tools/indexparser . -path "$(PWD)/README.md" > ../misc/docs/sidebar.json

# Preview these docs in the docs.gno.land frontend locally. Hot-reloads on edits
# by default (NO_WATCH=1 for a static snapshot); UPDATE=1 refreshes the frontend
# clone. See ../misc/docs/preview.sh.
preview:
	../misc/docs/preview.sh

test:
	@echo "Nothing to do."
