# Import source from the current Go standard library.
import:
	cp -a $$(go env GOROOT)/src/go/parser/* .
	cp -a $$(go env GOROOT)/src/go/printer/nodes.go nodes.go.src

# Generate a patch to be applied against imported files.
genpatch:
	for f in *.orig; do diff -u $$f $${f%.orig} || true; done > gno.patch

# Apply patch against imported files.
patch:
	patch -b < gno.patch

clean:
	rm -rf *.go *.orig testdata;
	echo "package parser" > gen.go
	echo "//go:generate make import patch" >> gen.go

.PHONY: import genpatch patch clean
