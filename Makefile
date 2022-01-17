all: gnoland gnokey goscan logos

.PHONY: logos goscan gnoland gnokey logos reset test test1 test2 testrealm testrealm1 testrealm2 testpackages testpkgs

reset:
	rm -rf testdir
	make

tools:
	go build -o build/logjack ./pkgs/autofile/cmd

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

# genproto makes protobufs
genproto:
	echo "Building genproto"
	go build -o build/genproto ./cmd/.genproto


# gnotxport allows you to fetch valid transactions from rpc
gnotxport:
	echo "Building gnotxport"
	go build -o build/gnotxport ./cmd/gnotxport

# gnoview doesn't build yet
gnoview:
	echo "Building gnoview"
	go build -o build/gnoview ./cmd/gnoview

# Logos is the interface to Gnoland
logos:
	echo "building logos"
	go build -o build/logos ./logos/cmd/logos.go

clean:
	rm -rf build

test:
	echo "Running tests"
	go test
	go test tests/*.go -v -test.short

test1:
	echo "Running tests"
	go test
	go test tests/*.go -v -test.short -run "TestFiles1"

test2:
	echo "Running tests"
	go test
	go test tests/*.go -v -test.short -run "TestFiles2"

testrealm:
	echo "Running tests"
	go test
	go test tests/*.go -v -run "TestFiles/^zrealm"

testrealm1:
	echo "Running tests"
	go test
	go test tests/*.go -v -run "TestFiles1/^zrealm"

testrealm2:
	echo "Running tests"
	go test
	go test tests/*.go -v -run "TestFiles2/^zrealm"

testpackages:
	echo "Running tests"
	go test tests/*.go -v -run "TestPackages"

testpkgs:
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
