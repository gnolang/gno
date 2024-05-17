all: build lint

# Build the linter
build:
	cd ../misc/linter && go build -o ../misc/linter/build/linter

# Run the linter for the docs/ folder
lint:
	../misc/linter/build/linter -path .

clean:
	rm -f ./linter/linter