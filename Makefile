.PHONY: help
help:
	@echo "Available make commands:"
	@cat Makefile | grep '^[a-z][^:]*:' | grep -v 'install_' | cut -d: -f1 | sort | sed 's/^/  /'

rundep=go run -modfile misc/devdeps/go.mod

.PHONY: install
install: install.gnokey install.gno
	@if ! command -v gnodev > /dev/null; then \
		echo ------------------------------; \
		echo "For local realm development, gnodev is recommended: https://docs.gno.land/gno-tooling/cli/gno-tooling-gnodev"; \
		echo "You can install it by calling 'make install.gnodev'"; \
	fi

# shortcuts to frequently used commands from sub-components.
.PHONY: install.gnokey
install.gnokey:
	$(MAKE) --no-print-directory -C ./gno.land	install.gnokey
	@echo "[+] 'gnokey' is installed. more info in ./gno.land/."
.PHONY: install.gno
install.gno:
	$(MAKE) --no-print-directory -C ./gnovm	install
	@echo "[+] 'gno' is installed. more info in ./gnovm/."
.PHONY: install.gnodev
install.gnodev:
	$(MAKE) --no-print-directory -C ./contribs install.gnodev
	@echo "[+] 'gnodev' is installed."

# old aliases
.PHONY: install_gnokey
install_gnokey: install.gnokey
.PHONY: install_gno
install_gno: install.gno

.PHONY: test
test: test.components test.docker

.PHONY: test.components
test.components:
	$(MAKE) --no-print-directory -C tm2      test
	$(MAKE) --no-print-directory -C gnovm    test
	$(MAKE) --no-print-directory -C gno.land test
	$(MAKE) --no-print-directory -C examples test
	$(MAKE) --no-print-directory -C misc     test

.PHONY: test.docker
test.docker:
	@if hash docker 2>/dev/null; then \
		go test --tags=docker -count=1 -v ./misc/docker-integration; \
	else \
		echo "[-] 'docker' is missing, skipping ./misc/docker-integration tests."; \
	fi

.PHONY: fmt
fmt:
	$(MAKE) --no-print-directory -C tm2      fmt imports
	$(MAKE) --no-print-directory -C gnovm    fmt imports
	$(MAKE) --no-print-directory -C gno.land fmt imports
	$(MAKE) --no-print-directory -C examples fmt

.PHONY: lint
lint:
	$(rundep) github.com/golangci/golangci-lint/cmd/golangci-lint run --config .github/golangci.yml

.PHONY: tidy
tidy:
	$(MAKE) --no-print-directory -C misc     tidy
