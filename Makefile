all: gnoland goscan logos

.PHONY: logos goscan gnoland


# NB: Binaries will output in the folder github.com/gnoland
targets = windows/amd64,windows/arm64,darwin-10.14/arm64,darwin-10.14/amd64,linux/*

# The main show
gnoland:
	echo "Building gnoland"
	go build -o gnoland ./cmd/gnoland

# goscan scans go code to determine its AST
goscan:
	echo "Building goscan"
	go build -o goscan ./cmd/goscan


# Logos is the interface to Gnoland
logos:
	echo "building logos"
	go build -o logos ./logos/cmd/logos.go

clean:
	rm -rf build

test:
	echo "Running tests"
	go test
	go test tests/*.go -v -run="Test/realm.go"


# TODO stringer -type=Op
# Unsure what the above refers to. Created a basic Makefile.

