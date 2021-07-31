all: gnoland goscan logos

.PHONY: logos goscan gnoland

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
