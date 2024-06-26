all: build lint

# Build the linter
build:
	cd ../misc/docs-linter && go build -o ./build/

# Run the linter for the docs/ folder
lint:
	../misc/docs-linter/build/linter -path .

clean:
	rm -rf ../misc/docs-linter/build