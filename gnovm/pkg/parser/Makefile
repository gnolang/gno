# Build from Go source + .patch file.
all: clean import patch

GOROOT := $(shell go env GOROOT)

# Import source from the current Go standard library.
import:
	cp -a $(GOROOT)/src/go/parser/* .
	cp -a $(GOROOT)/src/go/printer/nodes.go nodes.go.src
	rm -rf testdata/issue42951

# Generate a patch to be applied against imported files.
genpatch:
	for f in $$(find . -name "*.go" | LC_ALL=C sort); do \
		diff -u --label "$$f" --label "$$f" "$(GOROOT)/src/go/parser/$$f" "$$f" || true; \
	done > gno.patch

# Apply patch against imported files.
patch:
	patch -p1 < gno.patch

clean:
	find . -type f \! \( -name "Makefile" -o -name "gno.patch" \) -delete

.PHONY: all import genpatch patch clean
