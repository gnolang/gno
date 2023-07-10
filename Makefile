.PHONY: help
help:
	@echo "Available make commands:"
	@cat Makefile | grep '^[a-z][^:]*:' | cut -d: -f1 | sort | sed 's/^/  /'

rundep=go run -modfile ../misc/devdeps/go.mod

.PHONY: install
install: install_gnokey install_gno

# shortcuts to frequently used commands from sub-components.
install_gnokey:
	$(MAKE) --no-print-directory -C ./gno.land	install.gnokey
	@echo "[+] 'gnokey' is installed. more info in ./gno.land/."
install_gno:
	$(MAKE) --no-print-directory -C ./gnovm	install
	@echo "[+] 'gno' is installed. more info in ./gnovm/."

.PHONY: test
test: test.components test.docker

.PHONY: test.components
test.components:
	$(MAKE) --no-print-directory -C tm2      test
	$(MAKE) --no-print-directory -C gnovm    test
	$(MAKE) --no-print-directory -C gno.land test
	$(MAKE) --no-print-directory -C examples test

.PHONY: test.docker
test.docker:
	@if hash docker 2>/dev/null; then \
		go test --tags=docker -count=1 -v ./misc/docker-integration; \
	else \
		echo "[-] 'docker' is missing, skipping ./misc/docker-integration tests."; \
	fi

.PHONY: fmt
fmt:
	$(MAKE) --no-print-directory -C tm2      fmt
	$(MAKE) --no-print-directory -C gnovm    fmt
	$(MAKE) --no-print-directory -C gno.land fmt
	$(MAKE) --no-print-directory -C examples fmt

.PHONY: lint
lint:
	golangci-lint run --config .github/golangci.yml

.PHONE: reset
reset:
	cd /Users/yo/Projects/teritori/gno/gno.land && make install && rm -fr testdir && gnoland

.PHONE: faucet
faucet:
	cd /Users/yo/Projects/teritori/gno/gno.land && gnofaucet serve test1 --chain-id dev --send 500000000ugnot

.PHONE: init
init:
# cd /Users/yo/Projects/teritori/gno/gno.land && make install && rm -fr testdir && gnoland
# cd /Users/yo/Projects/teritori/gno/gno.land && gnofaucet serve test1 --chain-id dev --send 500000000ugnot

	curl --location --request POST 'http://localhost:5050' \
	--header 'Content-Type: application/x-www-form-urlencoded' \
	--data-urlencode 'toaddr=g1ckn395mpttp0vupgtratyufdaakgh8jgkmr3ym'

	gnokey maketx call -gas-fee="1ugnot" -gas-wanted="5000000" -broadcast="true" -pkgpath="gno.land/r/demo/users" -func="Register" -args="" -args="yo_account" -args="" -send="200000000ugnot" yo

	gnokey maketx call -gas-fee="1ugnot" -gas-wanted="5000000" -broadcast="true" -pkgpath="gno.land/r/demo/users" -func="Register" -args="" -args="test1_account" -args="" -send="200000000ugnot" test1

	gnokey maketx call -pkgpath "gno.land/r/demo/social_feeds" -func "CreateFeed" -gas-fee 1000000ugnot -gas-wanted 2000000 -send "" -broadcast -chainid "dev" -args "teritori" -remote "127.0.0.1:26657" test1

	gnokey maketx call -pkgpath "gno.land/r/demo/social_feeds" -func "CreatePost" -gas-fee 1000000ugnot -gas-wanted 2000000 -send "" -broadcast -chainid "dev" -args "1" -args "0" -args "0" -args '{"gifs": [], "files": [], "title": "", "message": "this is 2 reply inside t sdv ds d ds  he thread", "hashtags": [], "mentions": [], "createdAt": "2023-04-20T09:39:45.522Z", "updatedAt": "2023-04-20T09:39:45.522Z"}' -remote "127.0.0.1:26657" test1

