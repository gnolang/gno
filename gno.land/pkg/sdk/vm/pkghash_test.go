package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// TestAddPackageRecordsTheSubmittedDigest pins that the digest an approver
// computes from a submission is the one AddPackage records for it. Approvers
// hash the submitted bytes (gpao the MsgAddPackage it decodes out of the block,
// gnokey -pkgdir a local copy) while the parked gnomod.toml is stamped, so the
// record has to be taken from the bytes as received.
//
// The submission hand-writes the section the stamp overwrites, digest included,
// as a submitter is free to: a recorded digest it chose would approve itself.
func TestAddPackageRecordsTheSubmittedDigest(t *testing.T) {
	const pkgPath = "gno.land/r/test/stamped"
	// Fresh per call: stampGnomod rewrites msg.Package in place, so the copy
	// hashed as "submitted" has to be a second instance of the same bytes.
	files := func() []*std.MemFile {
		return []*std.MemFile{
			{Name: "gnomod.toml", Body: "module = \"" + pkgPath + "\"\n" +
				"gno = \"0.9\"\n\n" +
				"[addpkg]\n" +
				"creator = \"g1forged\"\n" +
				"pkg_hash = \"forged\"\n"},
			{Name: "stamped.gno", Body: "package stamped\n\nfunc Who(cur realm) string { return \"reviewed\" }"},
		}
	}

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("stamped-creator"))
	env, ctx := inertEnv(t, approver, creator)

	submitted := NewMsgAddPackage(creator, pkgPath, files()).Package
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, files())))

	parked := env.vmk.getGnoTransactionStore(ctx).GetInertPackage(pkgPath)
	require.NotNil(t, parked)
	gm, err := gnomod.ParseMemPackage(parked)
	require.NoError(t, err)
	require.Equal(t, creator.String(), gm.AddPkg.Creator, "premise: the stamp overwrote [addpkg]")
	require.NotEqual(t, PackageContentHash(submitted), PackageContentHash(parked),
		"premise: the stamp changed the stored bytes")

	assert.Equal(t, PackageContentHash(submitted), gm.AddPkg.PkgHash,
		"AddPackage recorded a digest other than the submission's, so no approval computed from it can land")
	require.NoError(t, env.vmk.EnablePackage(ctx, MsgEnablePackage{
		Approver: approver, PkgPath: pkgPath, PkgHash: PackageContentHash(submitted),
	}), "an approval computed from the submitted bytes has to activate what was parked")
}

// TestPackageContentHashNamesTheBytes pins that the digest names the bytes a
// submitter sent, not what a parser reads out of them: an approver reviews the
// file as written, comments included.
func TestPackageContentHashNamesTheBytes(t *testing.T) {
	const pkgPath = "gno.land/r/test/rawbytes"
	pkgWith := func(mod string) *std.MemPackage {
		return &std.MemPackage{Name: "rawbytes", Path: pkgPath, Files: []*std.MemFile{
			{Name: "gnomod.toml", Body: mod},
			{Name: "rawbytes.gno", Body: "package rawbytes\n\nfunc Who(cur realm) string { return \"reviewed\" }"},
		}}
	}
	plain := pkgWith("module = \"" + pkgPath + "\"\ngno = \"0.9\"\n")
	commented := pkgWith("# the approver reads this line too\nmodule = \"" + pkgPath + "\"\ngno = \"0.9\"\n")

	plainMod, err := gnomod.ParseMemPackage(plain)
	require.NoError(t, err)
	commentedMod, err := gnomod.ParseMemPackage(commented)
	require.NoError(t, err)
	require.Equal(t, plainMod, commentedMod, "premise: the parser reads both bodies as one module")

	assert.NotEqual(t, PackageContentHash(plain), PackageContentHash(commented),
		"two gnomod.toml bodies that read differently share a digest")
}
