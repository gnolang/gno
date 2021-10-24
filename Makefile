all: gnoland gnokey goscan logos

.PHONY: logos goscan gnoland gnokey logos reset

reset:
	rm -rf testdir
	make

# The main show (daemon)
gnoland:
	echo "Building gnoland"
	go build -o build/gnoland ./cmd/gnoland

# The main show (client)
gnokey:
	echo "Building gnokey"
	go build -o build/gnokey ./cmd/gnokey

# goscan scans go code to determine its AST
goscan:
	echo "Building goscan"
	go build -o build/goscan ./cmd/goscan


# Logos is the interface to Gnoland
logos:
	echo "Building Logos"
	go build -o logos ./logos/cmd/logos.go
	
gnokey:
	echo "Building Gnokey"
	go build -o gnokey ./cmd/gnokey/


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
