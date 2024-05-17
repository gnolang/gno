all: build lint

# Build the linter
build:
	cd linter && go build -o ./build/linter

# Run the linter for the docs/ folder
lint:
	./linter/build/linter -path .

clean:
	rm -f ./linter/linter