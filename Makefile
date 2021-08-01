all: gnoland goscan logos gnokey

.PHONY: logos goscan gnoland gnokey

# The main show
gnoland: folder
	echo "Building gnoland"
	go build -o build/gnoland ./cmd/gnoland

# goscan scans go code to determine its AST
goscan: folder
	echo "Building goscan"
	go build -o build/goscan ./cmd/goscan


# Logos is the interface to Gnoland
logos: folder
	echo "building logos"
	go build -o build/logos ./logos/cmd/logos.go

gnokey: folder
	echo "building gnokey"
	go build -o build/gnokey ./cmd/gnokey/

clean:
	rm -rf build

test:
	echo "Running tests"
	go test
	go test tests/*.go -v -test.short

test2:
	# -p 1 shows test failures as they come
	# maybe another way to do this?
	go test ./pkgs/... -p 1 -count 1

stringer:
	stringer -type=Op

genproto:
	rm -rf proto/*
	find pkgs/crypto/ | grep "\.proto" | xargs rm
	find pkgs/crypto/ | grep "pbbindings" | xargs rm
	find pkgs/crypto/ | grep "pb.go" | xargs rm
	find pkgs/bft/ | grep "\.proto" | xargs rm
	find pkgs/bft/ | grep "pbbindings" | xargs rm
	find pkgs/bft/ | grep "pb.go" | xargs rm
	go run cmd/genproto/genproto.go

folder:
	mkdir build