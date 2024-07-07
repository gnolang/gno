all: build lint

rundep=go run -modfile ../misc/devdeps/go.mod

# Build the linter
build:
	$(rundep) github.com/campoy/embedmd -w `find . -name "*.md"`
	cd ../misc/docs-linter && go build -o ./build/

# Run the linter for the docs/ folder
lint:
	../misc/docs-linter/build/linter -path .

clean:
	rm -rf ../misc/docs-linter/build
