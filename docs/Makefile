# Default target named all
all: build run

# Build the Go program
build:
	cd linter && go build -o linter

# Run the Go program
run:
	./linter/linter -path .

# Clean up binary
clean:
	rm -f ./linter/linter