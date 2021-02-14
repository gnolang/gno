all: xgo gnoland goscan logos

targets = windows/amd64,windows/arm64,darwin-10.14/arm64,darwin-10.14/amd64,linux/*

# gnoland is the main executable of gnolang
# bring this back when ready
gnoland:
	echo "Building gnoland"
	xgo -v -x --targets=$(targets) ./cmd/gnoland

# goscan scans go code to determine its AST
goscan:
	echo "Building goscan"
	xgo -v -x --targets=$(targets) ./cmd/goscan


# Logos is the interface to Gnoland
logos:
	echo "building logos"
	xgo -v -x --targets=4(targets) ./logos/cmd


test:
	echo "Running tests"
	go test tests/*.go -v -run="Test/realm.go"

xgo:
	echo "installing xgo"
	go get src.techknowlogick.com/xgo

# NB: Binaries will output in the folder github.com/gnoland

# TODO stringer -type=Op
# Unsure what the above refers to. Created a basic Makefile.

