all: build lint

# Build the linter
build:
	go run -modfile ../misc/devdeps/go.mod github.com/campoy/embedmd -w `find . -name "*.md"`

# Run the linter for the docs/ folder
lint:
	go build -C ../misc/docs-linter -o ./build
	../misc/docs-linter/build -path .
