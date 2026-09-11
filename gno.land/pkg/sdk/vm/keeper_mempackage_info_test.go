package vm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

func infoPayload() *std.MemFile {
	return &std.MemFile{Name: "payload.gno", Body: strings.Repeat("A", 4096)}
}

// A submitter cannot attach a payload to std.MemPackage.Info.
//
// Info is amino field 5, typed any, so any registered concrete type decodes
// into it off the wire. Nothing on chain writes the field and nothing reads
// it, so an accepted payload is dead weight in the merkleized package blob,
// and it is cheap dead weight: chargePreprocessGas sums .gno, gnomod.toml and
// gno.mod bodies only, so Info pays amino-encode and store-write rates rather
// than the per-source-byte preprocess rate the same bytes would pay in a file.
func TestVMKeeperAddPackageRejectsSubmitterInfo(t *testing.T) {
	const pkgPath = "gno.land/r/test/carrier"
	newFiles := func() []*std.MemFile {
		return []*std.MemFile{
			{Name: "carrier.gno", Body: `package carrier

func Hello(cur realm) string { return "hi" }`},
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		}
	}

	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	creator := crypto.AddressFromPreimage([]byte("carrier"))
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, creator))
	require.NoError(t, env.bankk.SetCoins(ctx, creator, initialBalance))

	msg := NewMsgAddPackage(creator, pkgPath, newFiles())
	msg.Package.Info = infoPayload()

	// The tx codec carries the field through untouched, and the payload is
	// most of what the package encodes to, so nothing but a keeper refusal
	// keeps those bytes out of the store.
	var decoded MsgAddPackage
	require.NoError(t, amino.Unmarshal(amino.MustMarshal(msg), &decoded))
	require.NotNil(t, decoded.Package.Info, "Info survives amino binary decode")
	require.Greater(t, len(amino.MustMarshal(decoded.Package)), len(infoPayload().Body),
		"the decoded package still carries the payload")

	err := env.vmk.AddPackage(ctx, msg)
	require.Error(t, err)
	require.ErrorIs(t, err, InvalidPackageError{})
	assert.Contains(t, fmt.Sprintf("%+v", err), "info field is not accepted")
	assert.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false),
		"a refused submission stores nothing")

	// Causality control: the same submission without the payload is accepted,
	// so the refusal above is the Info field and not the rest of the package.
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, newFiles())))
	assert.NotNil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false))
}

// Enable is where a payload parked before the submit-time refusal existed is
// met again, and it has to answer for it: AddMemPackage revalidates on the way
// in and panics rather than returning, so the refusal has to happen before
// RunMemPackage reaches it. The check covers every stored-blob failure
// ValidateMemPackageAny returns as an error, not only Info; the ones it raises
// as a panic of its own, mptype.Validate on a type/path mismatch above all, are
// caught by validateParkedBlob and returned too.
func TestVMKeeperEnableRefusesAParkedInfoPayload(t *testing.T) {
	const pkgPath = "gno.land/r/test/legacy"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "legacy.gno", Body: `package legacy

func Hello(cur realm) string { return "hi" }`},
	}

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("legacy"))
	env, ctx := inertEnv(t, approver, creator)

	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, files)))

	// Attach the payload to the stamped blob through the store, not through
	// AddPackage, which now refuses one. This is what a blob parked before the
	// refusal looks like, and PackageContentHash does not cover Info, so the
	// approval an approver computes still matches.
	gnostore := env.vmk.getGnoTransactionStore(ctx)
	parked := gnostore.GetInertPackage(pkgPath)
	require.NotNil(t, parked)
	parked.Info = infoPayload()
	gnostore.AddInertPackage(parked)

	err := env.vmk.EnablePackage(ctx, approvalFor(t, env, ctx, approver, pkgPath))
	require.Error(t, err)
	require.ErrorIs(t, err, InvalidPackageError{})
	detail := fmt.Sprintf("%+v", err)
	assert.Contains(t, detail, "info field is not accepted")
	assert.NotContains(t, detail, "VM panic",
		"the refusal is a validation error, not a recovered panic")
	assert.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetMemPackage(pkgPath),
		"the live blob is not written")
}

// MsgRun is the other message that carries a submitter-authored mempackage,
// and it reaches the same validator through ValidateMemPackage, so a payload
// has no second way in. Its package is never saved, but the bytes are still
// in the block and chargePreprocessGas still prices file bodies only.
func TestVMKeeperRunRejectsSubmitterInfo(t *testing.T) {
	newFiles := func() []*std.MemFile {
		return []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest("gno.land/r/test")},
			{Name: "script.gno", Body: `
package main

func main() {
	println("hello world!")
}
`},
		}
	}

	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	caller := crypto.AddressFromPreimage([]byte("runner"))
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, caller))

	coins := std.MustParseCoins("")
	msg := NewMsgRun(caller, coins, newFiles())
	msg.Package.Info = infoPayload()

	_, err := env.vmk.Run(ctx, msg)
	require.Error(t, err)
	require.ErrorIs(t, err, InvalidPackageError{})
	assert.Contains(t, fmt.Sprintf("%+v", err), "info field is not accepted")

	// Causality control.
	res, err := env.vmk.Run(ctx, NewMsgRun(caller, coins, newFiles()))
	require.NoError(t, err)
	assert.Equal(t, "hello world!\n", res)
}

// The re-validation runs before the machine exists, so it is outside
// doRecover. ValidateMemPackageAny raises part of its answer as a panic rather
// than an error -- mptype.Validate panics on every failure path -- and an
// escaping panic reaches consensus as std.ErrInternal with a Go stack instead
// of the vm error every other refusal on this path returns.
func TestVMKeeperEnableReturnsAPanickingValidation(t *testing.T) {
	const pkgPath = "gno.land/r/test/mistyped"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "mistyped.gno", Body: `package mistyped

func Hello(cur realm) string { return "hi" }`},
	}

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("legacy"))
	env, ctx := inertEnv(t, approver, creator)

	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, files)))

	// A stdlib type on a /r/ path, which is what a blob parked by a binary
	// with different type rules looks like. mptype.Validate panics on it.
	gnostore := env.vmk.getGnoTransactionStore(ctx)
	parked := gnostore.GetInertPackage(pkgPath)
	require.NotNil(t, parked)
	parked.Type = gnolang.MPStdlibProd
	gnostore.AddInertPackage(parked)

	var err error
	require.NotPanics(t, func() {
		err = env.vmk.EnablePackage(ctx, approvalFor(t, env, ctx, approver, pkgPath))
	}, "the refusal must be returned, not raised past the keeper")
	require.Error(t, err)
	assert.Contains(t, fmt.Sprintf("%+v", err), "expected stdlib package path")
}

// Every refusal about a stored blob is an InvalidPackageError, which is what
// the rest of EnablePackage returns. InvalidPkgPathError is reserved here for
// the two failures that really are about the path: no inert package at it, and
// a foreign domain. The type is amino-encoded into ABCIResult.Error and hashed,
// so this one is settled at the fork, not later.
func TestVMKeeperEnableTypesEveryBlobRefusalAsInvalidPackage(t *testing.T) {
	const pkgPath = "gno.land/r/test/badextension"
	files := []*std.MemFile{
		{Name: "badextension.gno", Body: `package badextension

func Hello(cur realm) string { return "hi" }`},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
	}

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("legacy"))
	env, ctx := inertEnv(t, approver, creator)

	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, files)))

	// A file extension the validator refuses, as a blob parked under looser
	// rules would carry. This is a returned error, not one of the panics.
	gnostore := env.vmk.getGnoTransactionStore(ctx)
	parked := gnostore.GetInertPackage(pkgPath)
	require.NotNil(t, parked)
	parked.Files = append(parked.Files, &std.MemFile{Name: "notes.txt", Body: "x"})
	gnostore.AddInertPackage(parked)

	err := env.vmk.EnablePackage(ctx, approvalFor(t, env, ctx, approver, pkgPath))
	require.Error(t, err)
	assert.ErrorIs(t, err, InvalidPackageError{},
		"a refusal about the blob's files must not be typed as a path error")
}
