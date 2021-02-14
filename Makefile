all: xgo gnoland goscan logos

# gnoland is the main executable of gnolang
# bring this back when ready
gnoland:
	echo "Building gnoland"
	xgo ./cmd/gnoland

# goscan scans go code to determine its AST
goscan:
	echo "Building goscan"
	xgo ./cmd/goscan


# Logos is the interface to Gnoland
logos:
	echo "building logos"
	xgo ./logos/cmd


test:
	echo "Running tests"
	go test tests/*.go -v -run="Test/realm.go"

xgo:
	echo "installing xgo"
	go get src.techknowlogick.com/xgo

# NB: Binaries will output in the folder github.com/gnoland

# TODO stringer -type=Op
# Unsure what the above refers to. Created a basic Makefile.

