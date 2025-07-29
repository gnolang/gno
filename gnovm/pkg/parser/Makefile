# Import source from the current Go standard library.
import:
	cp -a $$(go env GOROOT)/src/go/parser/* .
	cp -a $$(go env GOROOT)/src/go/printer/nodes.go nodes.go.src
#	mkdir -p internal && cp -a $$(go env GOROOT)/src/go/internal/typeparams internal/
#	rm -rf testdata/issue42951

# Generate a patch to be applied against imported files.
genpatch:
	for f in $$(find . -name "*.orig"); do diff -u $$f $${f%.orig} || true; done > gno.patch

# Apply patch against imported files.
patch:
	patch -b < gno.patch

clean:
	rm -rf *.go *.orig internal testdata

.PHONY: import genpatch patch clean
