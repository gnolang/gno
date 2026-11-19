package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// TestPackageContentHashNormalizesGnomod pins the two halves of what the hash
// has to do at once: ignore the bookkeeping the keeper stamps, so the
// submitter's copy and the stored copy agree, and notice the fields the author
// declares, so an approval of a private realm cannot be cashed in on a public
// one.
func TestPackageContentHashNormalizesGnomod(t *testing.T) {
	const pkgPath = "gno.land/r/test/pkghash"

	pkgFor := func(gm gnomod.File) *std.MemPackage {
		return &std.MemPackage{Path: pkgPath, Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: gm.WriteString()},
			{Name: "pkghash.gno", Body: "package pkghash"},
		}}
	}

	approved := gnomod.File{
		Module:  pkgPath,
		Gno:     "0.9",
		Private: true,
		AddPkg: gnomod.AddPkg{
			Creator:    "g1first",
			Height:     42,
			MaxDeposit: "100ugnot",
		},
	}
	restamped := approved
	restamped.AddPkg = gnomod.AddPkg{
		Creator:    "g1second",
		Height:     99,
		MaxDeposit: "200ugnot",
	}

	approvedHash, err := PackageContentHash(pkgFor(approved))
	require.NoError(t, err)
	require.NotEmpty(t, approvedHash)

	restampedHash, err := PackageContentHash(pkgFor(restamped))
	require.NoError(t, err)
	assert.Equal(t, approvedHash, restampedHash,
		"keeper-stamped metadata must not affect the approval hash")

	// The module line is keeper-owned too: stampGnomod rewrites it to the
	// deploy path, and nothing on the submit path requires the author to have
	// written the same value.
	renamed := approved
	renamed.Module = "gno.land/r/test/somethingelse"
	renamedHash, err := PackageContentHash(pkgFor(renamed))
	require.NoError(t, err)
	assert.Equal(t, approvedHash, renamedHash,
		"a module line the keeper will rewrite must not affect the approval hash")

	changed := approved
	changed.Private = false
	changedHash, err := PackageContentHash(pkgFor(changed))
	require.NoError(t, err)
	assert.NotEqual(t, approvedHash, changedHash,
		"author-declared metadata must affect the approval hash")
}

// TestPackageContentHashSurvivesTheRealStamp is the property the whole
// normalization exists for, checked against stampGnomod itself rather than
// against an imitation of it.
//
// The approver hashes a source directory; the keeper hashes the blob it stamped
// from that directory. Any field stampGnomod writes that keeperOwnedGnomod does
// not reset makes the two disagree forever, and the package can then never be
// enabled by anybody -- with an error that says the source changed after review
// when nothing changed. Module is exactly that field, and it shipped green
// through a test that appended a hand-written [addpkg] section instead of
// calling the stamp.
//
// So this calls the stamp. A future field added to stampGnomod fails here.
func TestPackageContentHashSurvivesTheRealStamp(t *testing.T) {
	const pkgPath = "gno.land/r/test/stamped"

	// Built fresh per case: stampGnomod rewrites gnomod.toml in place.
	pkgFor := func(module string) *std.MemPackage {
		return &std.MemPackage{Name: "stamped", Path: pkgPath, Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: "module = \"" + module + "\"\ngno = \"0.9\"\n"},
			{Name: "stamped.gno", Body: "package stamped"},
		}}
	}

	for _, module := range []string{
		pkgPath,
		// A copy-pasted template whose module line was never updated. The
		// chain accepts it: stampGnomod overwrites the value and nothing
		// upstream of it compares the two.
		"gno.land/r/test/template",
	} {
		t.Run(module, func(t *testing.T) {
			mpkg := pkgFor(module)
			approver, err := PackageContentHash(mpkg)
			require.NoError(t, err)

			gm, err := parseGnomodForHash(mpkg)
			require.NoError(t, err)
			stampGnomod(gm, mpkg, pkgPath, crypto.AddressFromPreimage([]byte("c")), 42, "100ugnot")

			keeper, err := PackageContentHash(mpkg)
			require.NoError(t, err)
			assert.Equal(t, approver, keeper,
				"the approver's hash must survive the stamp, or the package can never be enabled")
		})
	}
}

// TestPackageContentHashIsPinned freezes the digest for a fully-populated
// gnomod.toml.
//
// The hash is a canonical TOML re-encode, so it moves whenever gnomod.File
// gains or reorders a field, or whenever tm2/pkg/toml's encoder changes -- none
// of which look like consensus changes at the call site that makes them. When
// it moves, every parked package's expected hash moves with it: approvals
// prepared offline with `enablepkg -broadcast=false`, and any hash an operator
// recorded for `-pkg-hash`, are refused with "it changed after review" though
// nothing changed, and a gnokey or gpao built off the validators' commit
// computes a different hash for identical bytes.
//
// That is a coordinated upgrade, not a golden update. If this fails, decide
// that deliberately and write it down; do not just paste the new value in.
func TestPackageContentHashIsPinned(t *testing.T) {
	const pinned = "47b0005e2d817770b8a4071bfd28dab1e80375dcaa7d167044ac78d8ffa01045"

	mpkg := &std.MemPackage{
		Name: "pinned",
		Path: "gno.land/r/test/pinned",
		Files: []*std.MemFile{
			{Name: "pinned.gno", Body: "package pinned\n\nfunc Who(cur realm) string { return \"x\" }\n"},
			{Name: "gnomod.toml", Body: "" +
				"module = \"gno.land/r/test/pinned\"\n" +
				"gno = \"0.9\"\n" +
				"ignore = true\n" +
				"draft = true\n" +
				"private = true\n" +
				"[addpkg]\n" +
				"creator = \"g1stamped\"\n" +
				"height = 7\n" +
				"max_deposit = \"5ugnot\"\n"},
		},
	}

	got, err := PackageContentHash(mpkg)
	require.NoError(t, err)
	assert.Equal(t, pinned, got,
		"the approval hash moved; see this test's comment before changing the constant")
}

// TestPackageContentHashRejectsMalformedGnomod pins that a package whose
// gnomod.toml cannot be read produces an ERROR rather than an empty hash.
//
// The empty string is also what a message carrying no approval decodes to, so
// returning it in band let both CLI callers build and sign an approval naming
// no source at all -- refused on chain, after the fee, with a message about the
// transaction rather than about the directory the operator got wrong.
func TestPackageContentHashRejectsMalformedGnomod(t *testing.T) {
	malformed := &std.MemPackage{Path: "gno.land/r/test/pkghash", Files: []*std.MemFile{
		{Name: "gnomod.toml", Body: "module = ["},
		{Name: "pkghash.gno", Body: "package pkghash"},
	}}
	h, err := PackageContentHash(malformed)
	require.Error(t, err)
	assert.Empty(t, h)

	// Nor may the deprecated gno.mod stand in. gnomod.ParseMemPackage accepts
	// it, which would leave the parse succeeding while the substitution below
	// matched nothing and hashed the raw bytes -- a confident hash over a file
	// checkGnomodConstraints guarantees the chain will never hold.
	legacy := &std.MemPackage{Path: "gno.land/r/test/pkghash", Files: []*std.MemFile{
		{Name: "gno.mod", Body: "module gno.land/r/test/pkghash\n"},
		{Name: "pkghash.gno", Body: "package pkghash"},
	}}
	h, err = PackageContentHash(legacy)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "gno mod tidy")
	assert.Empty(t, h)
}
