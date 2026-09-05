package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	tmerrors "github.com/gnolang/gno/tm2/pkg/errors"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// An approval names a digest, and the digest has to name every byte that
// decides what the chain does with the submission. gnomod.toml carries two
// halves: the fields AddPackage stamps at submit (module, addpkg.creator,
// addpkg.height, addpkg.max_deposit) and the five the submitter authors (gno,
// ignore, draft, private, replace). PackageContentHash excludes the whole
// file, so the authored half sits outside what an approver signs.
//
// One of the five reaches enable on a live chain. `private` is read back by
// checkGnomodConstraints, which governs whether a public submission may go live
// over a live private realm; the value is then carried onto the PackageValue,
// where it decides whether any other realm may import the package. The rest
// are stopped earlier: replaces and post-genesis drafts are refused at submit,
// a gno version other than the current one panics the version gate before the
// package runs, and a submission with no gnomod.toml has no module path to
// park under. Their cases below pin the digest against a relaxation of those
// checks rather than against a hole open today, and say so.
//
// These tests assert the property, not today's behaviour: they are red on
// master. Canonicalizing the stamped fields before hashing, or hashing a
// projection of the authored fields, would turn them green as they stand.
// Refusing a re-park that moves an authored field closes the same hole at
// submit instead, and would want the first test re-pointed at the second
// AddPackage. Choosing between them is a design decision, so no fix is made
// here.

// storedGnomod renders a gnomod.toml in the shape the chain stores one: the
// stamped section first, held byte-identical across every case below, then
// whatever the submitter authored. Holding the stamp fixed is what makes a
// moved digest attributable to the authored half alone.
func storedGnomod(pkgPath, gnoVersion, authored string) string {
	mod := "module = \"" + pkgPath + "\"\ngno = \"" + gnoVersion + "\"\n"
	if authored != "" {
		mod += authored + "\n"
	}
	return mod + "[addpkg]\ncreator = \"" +
		crypto.AddressFromPreimage([]byte("stamped-creator")).String() +
		"\"\nheight = 42\nmax_deposit = \"1000000ugnot\"\n"
}

// TestEnableRefusesBytesTheApproverDidNotReview drives the whole two-phase
// deploy: a creator parks a public package, an approver digests what the chain
// serves at that path, the creator re-parks the same .gno source with
// `private = true`, and the approval of the public submission is presented.
//
// Activating it turns the reviewed public API into a realm no other package
// may import and whose slot the creator may silently overwrite from then on --
// none of which the approver saw or signed.
func TestEnableRefusesBytesTheApproverDidNotReview(t *testing.T) {
	const (
		pkgPath = "gno.land/r/test/privacyflip"
		srcName = "privacyflip.gno"
		source  = "package privacyflip\n\nfunc Who(cur realm) string { return \"reviewed\" }"
	)
	files := func(private bool) []*std.MemFile {
		authored := ""
		if private {
			authored = "private = true"
		}
		return []*std.MemFile{
			{Name: "gnomod.toml", Body: storedGnomod(pkgPath, "0.9", authored)},
			{Name: srcName, Body: source},
		}
	}

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("privacyflip-creator"))
	env, ctx := inertEnv(t, approver, creator)
	store := env.vmk.getGnoTransactionStore(ctx)

	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, files(false))))

	// The approval an approver would actually send: the digest is taken from
	// the bytes the chain serves for this path, not from a package built here.
	approval := approvalFor(t, env, ctx, approver, pkgPath)
	reviewed := store.GetInertPackage(pkgPath)
	require.NotNil(t, reviewed, "premise: the public submission is parked")
	require.NotContains(t, reviewed.GetFile("gnomod.toml").Body, "private",
		"premise: what the approver reviews is a public package")

	// Same creator, same path, identical source, private = true. Replacement by
	// the original submitter is the retry path after a failed enable, so nothing
	// ahead of the digest stands in the way.
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, files(true))))
	swapped := store.GetInertPackage(pkgPath)
	require.NotNil(t, swapped)
	require.Contains(t, swapped.GetFile("gnomod.toml").Body, "private = true",
		"premise: the bytes now parked are private")
	require.Equal(t, reviewed.GetFile(srcName).Body, swapped.GetFile(srcName).Body,
		"premise: the re-park moved gnomod.toml and nothing else")

	err := env.vmk.EnablePackage(ctx, approval)
	require.Error(t, err,
		"an approval of the public submission activated the private one: the "+
			"approver's signature named source it never saw")
	// The class, not the message: a fix is free to reword the detail.
	assert.Equal(t, InvalidPackageError{}, tmerrors.Cause(err),
		"the refusal has to be about the submission's contents, so the approver "+
			"is told their approval no longer names what is parked")
	assert.Nil(t, store.GetMemPackage(pkgPath),
		"a refused enable must leave the path empty, not half-deployed")
}

// TestApprovalDigestSeparatesSubmitterAuthoredGnomodFields walks the fields a
// submitter writes into gnomod.toml and pins that each one moves the digest.
//
// Every case differs from the base in exactly one authored field and in
// nothing else -- same source file, same stamped section -- so a digest that
// does not move is an approver signing for a package the chain reads as a
// different module from the one they reviewed.
//
// The digest is blind to all four. Only the first two reach a live chain; the
// other two are stopped by a submit-time check that the digest knows nothing
// about, so they pin the digest against that check being relaxed rather than
// against a hole open today. `reaches` records which is which, because the two
// carry very different weight and a reader deserves to be told.
func TestApprovalDigestSeparatesSubmitterAuthoredGnomodFields(t *testing.T) {
	t.Parallel()

	const (
		pkgPath = "gno.land/r/test/authored"
		source  = "package authored\n\nfunc Who(cur realm) string { return \"reviewed\" }"
	)
	pkgWith := func(mod string) *std.MemPackage {
		return &std.MemPackage{Name: "authored", Path: pkgPath, Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: mod},
			{Name: "authored.gno", Body: source},
		}}
	}

	reviewed := pkgWith(storedGnomod(pkgPath, "0.9", ""))
	reviewedMod, err := gnomod.ParseMemPackage(reviewed)
	require.NoError(t, err, "premise: the reviewed gnomod.toml is one the chain accepts")

	cases := map[string]struct {
		mod string
		// reaches is what the field does on a live chain today: what an
		// unmoved digest actually buys a submitter.
		reaches string
	}{
		"private": {
			mod:     storedGnomod(pkgPath, "0.9", "private = true"),
			reaches: "a realm no other package may import goes live over an approval of a public one",
		},
		"gno": {
			mod: storedGnomod(pkgPath, "0.8", ""),
			reaches: "nothing; ParseCheckGnoMod panics on any version but the current one before " +
				"the package runs, so this pins the digest against that gate being relaxed",
		},
		"draft": {
			mod: storedGnomod(pkgPath, "0.9", "draft = true"),
			reaches: "nothing; checkGnomodConstraints refuses a post-genesis draft at submit, " +
				"so this pins the digest against that rule being relaxed",
		},
		"replace": {
			mod: storedGnomod(pkgPath, "0.9",
				"[[replace]]\nold = \"gno.land/p/demo/avl\"\nnew = \"gno.land/p/demo/avl/v2\""),
			reaches: "nothing; checkGnomodConstraints refuses any replace at submit, " +
				"so this pins the digest against that rule being relaxed",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			submitted := pkgWith(tc.mod)
			submittedMod, err := gnomod.ParseMemPackage(submitted)
			require.NoError(t, err, "premise: this gnomod.toml is one the chain accepts")
			require.NotEqual(t, *reviewedMod, *submittedMod,
				"premise: the chain reads the two submissions as different modules")

			assert.NotEqual(t, PackageContentHash(reviewed), PackageContentHash(submitted),
				"%s is the submitter's to write, yet an approver who reviewed one value "+
					"signs a digest that names the other just as well.\n"+
					"What that buys on a live chain today: %s",
				name, tc.reaches)
		})
	}
}

// TestApprovalDigestSeparatesASubmissionMissingItsGnomod pins the degenerate
// end of the same property: a submission with no gnomod.toml at all must not
// share a digest with one that has it.
//
// Nothing reaches a live chain through this one. A package with no module
// declaration has no path to park under, which the premise below is the proof
// of. It pins the digest at the far end of the range the exclusion covers: the
// whole file is outside the hash, so an approver's signature covers a package
// whose module path, privacy, gno version and replaces were all dropped, and
// only a check elsewhere stops that mattering.
func TestApprovalDigestSeparatesASubmissionMissingItsGnomod(t *testing.T) {
	t.Parallel()

	const (
		pkgPath = "gno.land/r/test/nomodfile"
		source  = "package nomodfile\n\nfunc Who(cur realm) string { return \"reviewed\" }"
	)
	src := &std.MemFile{Name: "nomodfile.gno", Body: source}
	mod := &std.MemFile{Name: "gnomod.toml", Body: storedGnomod(pkgPath, "0.9", "")}

	reviewed := &std.MemPackage{Name: "nomodfile", Path: pkgPath, Files: []*std.MemFile{mod, src}}
	stripped := &std.MemPackage{Name: "nomodfile", Path: pkgPath, Files: []*std.MemFile{src}}

	_, err := gnomod.ParseMemPackage(reviewed)
	require.NoError(t, err, "premise: the reviewed submission declares a module")
	_, err = gnomod.ParseMemPackage(stripped)
	require.Error(t, err, "premise: the stripped submission declares nothing at all")

	assert.NotEqual(t, PackageContentHash(reviewed), PackageContentHash(stripped),
		"deleting gnomod.toml outright leaves the digest untouched, so the file the "+
			"approver reviewed contributes nothing to what their approval names")
}
