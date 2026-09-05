# Vendored fork of github.com/pelletier/go-toml. See README.md.
#
# Rebuild the tree from the upstream module + gno.patch:
#     make all
# Regenerate gno.patch after editing files here:
#     make genpatch
#
# Upstream predates the Go 1.19 doc-comment style and modern Go idiom, so imported
# files are gofmt'ed and run through the `go fix` modernizers before the patch is
# generated or applied -- the repo gates on `go fix` being clean, see
# .github/workflows/_ci-go.yml. Both sides of the diff get the same treatment, so
# gno.patch stays the semantic deviation alone -- the bounds, the vet fixes and
# the timezone-robust test -- rather than burying it in reformatting hunks.

VERSION := v1.9.5
MODPATH := $(shell go env GOMODCACHE)/github.com/pelletier/go-toml@$(VERSION)
# Files here that are ours, not upstream's.
OURS := Makefile gno.patch README.md bounds_test.go

# Pristine upstream is staged here for genpatch. It sits outside this directory
# so it is a package of its own rather than a second copy of `toml` nested inside
# it, and it is removed again at the end of the target.
UPSTREAM_STAGE := ../tomlupstream

GO_FIX_TOOLCHAIN := go1.26.1
# -omitzero is excluded to match the repo-wide setting: amino honors json
# omitempty on struct fields, so the omitzero modernizer would change output.
GO_FIX_FIXERS := any buildtag fmtappendf forvar hostport inline mapsloop minmax \
	newexpr plusbuild rangeint reflecttypefor slicescontains slicessort \
	stditerators stringsbuilder stringscut stringscutprefix stringsseq \
	testingcontext waitgroup

# Run the modernizers over the package rooted at $(1), in place.
#
# Fixers are applied one at a time on purpose. A combined `go fix ./...` over
# this package reports "4 of 416 fixes skipped (e.g. due to conflicts)", then
# writes nothing at all and exits 1 -- so the conflict discards the other 412
# fixes too (go1.26.1). Applied individually they all land.
# A fixer that finds nothing to do exits non-zero, so per-fixer failures are
# ignored -- which would also hide a package that does not typecheck, and go fix
# declines such a package silently. The gate is re-run at the end to turn that
# back into a hard error: after a successful sequential application it exits 0.
define gofix
	cd $(1) && \
	for f in $(GO_FIX_FIXERS); do \
		GOTOOLCHAIN=$(GO_FIX_TOOLCHAIN) CGO_ENABLED=0 go fix -$$f ./... >/dev/null 2>&1 || true; \
	done; \
	if ! GOTOOLCHAIN=$(GO_FIX_TOOLCHAIN) CGO_ENABLED=0 go fix -omitzero=false -diff ./... >/dev/null 2>&1; then \
		echo "gofix: $(1) still reports pending fixes; does the package typecheck?" >&2; \
		exit 1; \
	fi
endef

all: clean import patch

# Import source from the module cache, gofmt'ed and modernized. `go mod download`
# first if absent.
import:
	@test -d "$(MODPATH)" || go mod download github.com/pelletier/go-toml
	cp -a $(MODPATH)/*.go $(MODPATH)/*.toml $(MODPATH)/LICENSE .
	rm -f fuzz.go
	chmod -R u+w .
	gofmt -w .
	@# bounds_test.go is ours and survives `clean`, so at this point it
	@# references identifiers only gno.patch adds. That leaves the test package
	@# failing to typecheck, and `go fix` then declines the whole package
	@# silently. Stage it out, matching genpatch's pristine copy, which has no
	@# bounds_test.go either -- so both sides get the same modernization.
	mv bounds_test.go bounds_test.go.hold
	$(call gofix,.)
	mv bounds_test.go.hold bounds_test.go

# Diff our tree against a gofmt'ed, modernized pristine copy of upstream.
genpatch:
	rm -rf $(UPSTREAM_STAGE) && mkdir -p $(UPSTREAM_STAGE)
	cp -a $(MODPATH)/*.go $(UPSTREAM_STAGE)/
	rm -f $(UPSTREAM_STAGE)/fuzz.go
	chmod -R u+w $(UPSTREAM_STAGE)
	gofmt -w $(UPSTREAM_STAGE)
	$(call gofix,$(UPSTREAM_STAGE))
	for f in $$(find . -maxdepth 1 -name "*.go" \! -name "bounds_test.go" -printf "%f\n" | LC_ALL=C sort); do \
		diff -u --label "$$f" --label "$$f" "$(UPSTREAM_STAGE)/$$f" "$$f" || true; \
	done > gno.patch
	rm -rf $(UPSTREAM_STAGE)

# Apply patch against imported files.
patch:
	patch -p0 < gno.patch

clean:
	find . -maxdepth 1 -type f $(foreach f,$(OURS),\! -name "$(f)") -delete

.PHONY: all import genpatch patch clean
