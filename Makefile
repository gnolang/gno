all: gnoland goscan logos

.PHONY: logos


# Uses xgo to compile gnoland, gnoscan and logos
# NB: Binaries will output in the folder github.com/gnoland
targets = windows/amd64,windows/arm64,darwin-10.14/arm64,darwin-10.14/amd64,linux/*

omni: xgo
	echo "Building all gnoland components for arm64 and amd64"
	xgo -v -x --targets=$(targets) ./cmd/gnoland
	xgo -v -x --targets=$(targets) ./cmd/goscan
	xgo -v -x --targets=$(targets) ./logos/cmd

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
	go test tests/*.go -v -run="Test/realm.go"

xgo:
	echo "installing xgo"
	go get src.techknowlogick.com/xgo


# TODO stringer -type=Op
# Unsure what the above refers to. Created a basic Makefile.

