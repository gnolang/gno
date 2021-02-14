all: xgo test gnoland goscan logos

build-dir:
	mkdir build || echo "Build folder is already present."

# gnoland is the main executable of gnolang
# bring this back when ready
gnoland:
	echo "Building gnoland"
	xgo ./cmd/gnoland

# goscan scans go code to determine its AST
goscan:
	echo "Building goscan"
	xgo ./cmd/goscan


# Is logos test code or will it be a part of the system?
logos:
	echo "building logos"
	xgo ./logos/cmd


test:
	echo "Running tests"
	go test tests/*.go -v -run="Test/realm.go"

xgo:
	echo "installing xgo"
	go get src.techknowlogick.com/xgo



# TODO stringer -type=Op
# Unsure what the above refers to. Created a basic Makefile.

