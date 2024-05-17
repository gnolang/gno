all: build run

# Build the linter
build:
	cd linter && go build -o linter

# Run the linter for the docs/ folder
run:
	./linter/linter -path .

clean:
	rm -f ./linter/linter