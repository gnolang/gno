all: build lint

# Build the linter
build:
	cd ../misc/docs-linter && go build -o ../misc/linter/build/linter

# Run the linter for the docs/ folder
lint:
	../misc/docs-linter/build/linter -path .

clean:
	rm -f ../misc/docs-linter/linter