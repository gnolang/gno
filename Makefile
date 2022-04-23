########################################
# Dist suite
.PHONY: logos goscan gnoland gnokey gnofaucet logos reset
all: gnoland gnokey goscan logos

reset:
	rm -rf testdir
	make

tools:
	go build -o build/logjack ./pkgs/autofile/cmd

# The main show (daemon)
gnoland:
	@echo "Building gnoland"
	go build -o build/gnoland ./cmd/gnoland

# The main show (client)
gnokey:
	@echo "Building gnokey"
	go build -o build/gnokey ./cmd/gnokey

install_gnokey:
	@echo "Installing gnokey"
	go install ./cmd/gnokey

install_gnodev:
	@echo "Installing gnodev"
	go install ./cmd/gnodev

# The faucet (daemon)
gnofaucet:
	@echo "Building gnofaucet"
	go build -o build/gnofaucet ./cmd/gnofaucet

# goscan scans go code to determine its AST
goscan:
	@echo "Building goscan"
	go build -o build/goscan ./cmd/goscan


# Logos is the interface to Gnoland
logos:
	@echo "building logos"
	go build -o build/logos ./logos/cmd/logos.go

clean:
	rm -rf build

precompile_examples: install_gnodev
	cd examples; gnodev precompile .

########################################
# Test suite
.PHONY: test test.go test.gno test.files1 test.files2 test.realm test.packages
test: test.gno test.go
	@echo "Full test suite finished."

test.gno: test.files1 test.files2 test.realm test.packages
	go test tests/*.go -v -run "TestFileStr"
	go test tests/*.go -v -run "TestSelectors"

test.go:
	go test . -v
	# -p 1 shows test failures as they come
	# maybe another way to do this?
	go test ./pkgs/... -v -p 1

test.files1:
	go test tests/*.go -v -test.short -run "TestFiles1/" --timeout 20m

test.files2:
	go test tests/*.go -v -test.short -run "TestFiles2/" --timeout 20m

test.realm:
	go test tests/*.go -v -run "TestFiles/^zrealm" --timeout 20m

test.packages:
	go test tests/*.go -v -run "TestPackages" --timeout 20m

# Code gen
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
