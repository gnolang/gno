########################################
# Dist suite
.PHONY: logos goscan gnoland gnokey gnofaucet logos reset gnoweb
all: gnoland gnokey goscan logos gnoweb

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

# Development tool
gnodev:
	@echo "Building gnodev"
	go build -o build/gnodev ./cmd/gnodev

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

gnoweb:
	@echo "Building website"
	go build -o build/website



# Logos is the interface to Gnoland
logos:
	@echo "building logos"
	go build -o build/logos ./logos/cmd/logos.go

clean:
	rm -rf build

examples.precompile: install_gnodev
	gnodev precompile ./examples --verbose

examples.build: install_gnodev examples.precompile
	gnodev build ./examples --verbose

########################################
# Formatting, linting.

.PHONY: fmt
fmt:
	go run -modfile ./misc/devdeps/go.mod mvdan.cc/gofumpt -w .
	go run -modfile ./misc/devdeps/go.mod mvdan.cc/gofumpt -w `find stdlibs examples -name "*.gno"`
	git checkout bak

########################################
# Test suite
.PHONY: test test.go test.go1 test.go2 test.go3 test.gno test.files1 test.files2 test.realm test.packages test.flappy test.packages0 test.packages1 test.packages2
test: test.gno test.go test.flappy
	@echo "Full test suite finished."

test.gno: test.files1 test.files2 test.packages test.examples
	go test tests/*.go -v -run "TestFileStr"
	go test tests/*.go -v -run "TestSelectors"

test.flappy:
	# flappy tests should work "sometimes" (at least once)
	TEST_STABILITY=flappy go run -modfile ./misc/devdeps/go.mod moul.io/testman test -test.v -timeout=20m -retry=10 -run ^TestFlappy \
		./pkgs/bft/consensus ./pkgs/bft/blockchain ./pkgs/bft/mempool ./pkgs/p2p ./pkgs/bft/privval

test.go: test.go1 test.go2 test.go3

test.go1:
	# test most of pkgs/* except amino and bft.
	go test . -v
	# -p 1 shows test failures as they come
	# maybe another way to do this?
	go test `go list ./pkgs/... | grep -v pkgs/amino/ | grep -v pkgs/bft/` -v -p 1 -timeout=30m

test.go2:
	# test amino.
	go test ./pkgs/amino/... -v -p 1 -timeout=30m

test.go3:
	# test bft.
	go test ./pkgs/bft/... -v -p 1 -timeout=30m

test.files1:
	go test tests/*.go -v -test.short -run "TestFiles1/" --timeout 30m

test.files2:
	go test tests/*.go -v -test.short -run "TestFiles2/" --timeout 30m

test.realm:
	go test tests/*.go -v -run "TestFiles/^zrealm" --timeout 30m

test.packages: test.packages0 test.packages1 test.packages2

test.packages0:
	go test tests/*.go -v -run "TestPackages/(bufio|crypto|encoding|errors|internal|io|math|sort|std|stdshim|strconv|strings|testing|unicode)" --timeout 30m

test.packages1:
	go test tests/*.go -v -run "TestPackages/regexp" --timeout 30m

test.packages2:
	go test tests/*.go -v -run "TestPackages/bytes" --timeout 30m

test.examples:
	go run ./cmd/gnodev test ./examples

# Code gen
stringer:
	stringer -type=Op

genproto:
	rm -rf proto/*
	find pkgs | grep -v "^pkgs\/amino" | grep "\.proto" | xargs rm
	find pkgs | grep -v "^pkgs\/amino" | grep "pbbindings" | xargs rm
	find pkgs | grep -v "^pkgs\/amino" | grep "pb.go" | xargs rm
	@rm gno.proto || true
	@rm pbbindings.go || true
	@rm gno.pb.go || true
	go run cmd/genproto/genproto.go

genproto2:
	rm -rf proto/*
	#find pkgs | grep -v "^pkgs\/amino" | grep "\.proto" | xargs rm
	find pkgs | grep -v "^pkgs\/amino" | grep "pbbindings" | xargs rm
	find pkgs | grep -v "^pkgs\/amino" | grep "pb.go" | xargs rm
	#@rm gno.proto || true
	@rm pbbindings.go || true
	@rm gno.pb.go || true
