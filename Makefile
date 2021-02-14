all: xgo test gnoland goscan logos

# gnoland is the main executable of gnolang
# bring this back when ready
gnoland:
	echo "Building Gnoland for a wide range of systems"
	mkdir build || true
	xgo ./cmd/gnoland

# goscan scans go code to determine its AST
goscan:
	echo "Building goscan for a wide range of systems"
	mkdir build || true
	xgo ./cmd/goscan


# Is logos test code or will it be a part of the system?
logos:
	echo "Building logos for a wide range of systems"
	mkdir build || true
	xgo ./logos/cmd


test:
	go test tests/*.go -v -run="Test/realm.go"

xgo:
	go get src.techknowlogick.com/xgo





# TODO stringer -type=Op
# Unsure what the above refers to. Created a basic Makefile.

# TODO can probobaly make this more specific, but that may not be desirable.  Sometimes it's nice to check if it builds
# everywhere.

