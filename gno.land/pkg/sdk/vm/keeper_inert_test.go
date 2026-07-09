package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// TestVMKeeperInertPackageLifecycle exercises the full "oracle activation"
// flow introduced by the "inert" code-submission policy:
//
//  1. the chain runs with CodeSubmissionPolicy == "inert" and a single approver
//     (the oracle) in PkgApprovers;
//  2. an untrusted user submits a package via MsgAddPackage — it is stored inert
//     (no typecheck, no execution) and is NOT importable or callable;
//  3. a non-approver cannot enable it;
//  4. the approver (oracle) sends MsgEnablePackage — the chain re-typechecks,
//     executes, and the package becomes visible and callable;
//  5. the inert copy is gone once activated.
func TestVMKeeperInertPackageLifecycle(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// The oracle/approver key. Only this address may enable packages.
	approver := crypto.AddressFromPreimage([]byte("oracle"))
	// An untrusted submitter and an unrelated third party.
	submitter := crypto.AddressFromPreimage([]byte("submitter"))
	stranger := crypto.AddressFromPreimage([]byte("stranger"))

	for _, addr := range []crypto.Address{approver, submitter, stranger} {
		acc := env.acck.NewAccountWithAddress(ctx, addr)
		env.acck.SetAccount(ctx, acc)
		env.bankk.SetCoins(ctx, addr, initialBalance)
	}

	// Switch the chain into "inert" submission mode with the oracle as approver.
	params := DefaultParams()
	params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
	params.PkgApprovers = []crypto.Address{approver}
	env.vmk.SetParams(ctx, params)

	const pkgPath = "gno.land/r/test/inert"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "inert.gno", Body: `package inert

func Echo(cur realm, msg string) string {
	return "echo:" + msg
}`},
	}

	gnostore := env.vmk.getGnoTransactionStore(ctx)

	// ---- 1. Submission (permissionless) --------------------------------------
	err := env.vmk.AddPackage(ctx, NewMsgAddPackage(submitter, pkgPath, files))
	require.NoError(t, err)

	// The package is stored inert: invisible to the normal resolver...
	assert.Nil(t, gnostore.GetPackage(pkgPath, false),
		"inert package must not be resolvable before activation")
	// ...but present in the inert key space.
	assert.NotNil(t, gnostore.GetInertPackage(pkgPath),
		"submitted package must be stored inert")

	// It must not be callable while inert: the package has no executable node,
	// so the VM cannot resolve it (on-chain this surfaces as a failed message).
	assert.Panics(t, func() {
		env.vmk.Call(ctx, NewMsgCall(submitter, nil, pkgPath, "Echo", []string{"hi"}))
	}, "inert package must not be callable")

	// ---- 2. Only an approver may enable --------------------------------------
	err = env.vmk.EnablePackage(ctx, MsgEnablePackage{Approver: stranger, PkgPath: pkgPath})
	require.Error(t, err, "non-approver must not be able to enable a package")
	assert.Contains(t, err.Error(), "unauthorized")

	// Still inert after the rejected attempt.
	assert.Nil(t, gnostore.GetPackage(pkgPath, false))
	assert.NotNil(t, gnostore.GetInertPackage(pkgPath))

	// ---- 3. Enabling an unknown path fails -----------------------------------
	err = env.vmk.EnablePackage(ctx, MsgEnablePackage{Approver: approver, PkgPath: "gno.land/r/test/missing"})
	require.Error(t, err, "enabling a non-existent inert package must fail")

	// ---- 4. Oracle approves --------------------------------------------------
	err = env.vmk.EnablePackage(ctx, MsgEnablePackage{Approver: approver, PkgPath: pkgPath})
	require.NoError(t, err, "approver must be able to enable a valid inert package")

	// The package is now a normal, resolvable package...
	assert.NotNil(t, gnostore.GetPackage(pkgPath, false),
		"enabled package must be resolvable")
	// ...and the inert copy is gone.
	assert.Nil(t, gnostore.GetInertPackage(pkgPath),
		"inert copy must be removed once activated")

	// ---- 5. The contract is now callable -------------------------------------
	res, err := env.vmk.Call(ctx, NewMsgCall(submitter, nil, pkgPath, "Echo", []string{"hello world"}))
	require.NoError(t, err)
	assert.Equal(t, `("echo:hello world" string)`+"\n\n", res)
}

// TestVMKeeperEnablePackageRejectsInvalidCode verifies the design invariant
// "the oracle proposes, the chain enforces": even if an approver tries to
// activate a package, the on-chain typechecker still rejects malformed code.
func TestVMKeeperEnablePackageRejectsInvalidCode(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	submitter := crypto.AddressFromPreimage([]byte("submitter"))
	for _, addr := range []crypto.Address{approver, submitter} {
		acc := env.acck.NewAccountWithAddress(ctx, addr)
		env.acck.SetAccount(ctx, acc)
		env.bankk.SetCoins(ctx, addr, initialBalance)
	}

	params := DefaultParams()
	params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
	params.PkgApprovers = []crypto.Address{approver}
	env.vmk.SetParams(ctx, params)

	// Intentionally broken package: references an undefined symbol. Storing it
	// inert must succeed (no typecheck on submission), but enabling must fail.
	const pkgPath = "gno.land/r/test/broken"
	files := []*std.MemFile{
		// MemPackage files must be sorted by name.
		{Name: "broken.gno", Body: `package broken

func Boom(cur realm) string {
	return undefinedSymbol
}`},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
	}

	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(submitter, pkgPath, files)),
		"inert submission must not typecheck")

	err := env.vmk.EnablePackage(ctx, MsgEnablePackage{Approver: approver, PkgPath: pkgPath})
	require.Error(t, err, "chain must reject activation of ill-typed code")

	// The package never becomes callable.
	gnostore := env.vmk.getGnoTransactionStore(ctx)
	assert.Nil(t, gnostore.GetPackage(pkgPath, false))
}

// TestVMKeeperDisablePackageNotImplemented documents that MsgDisablePackage is
// approver-gated but not yet functional (tracked for a follow-up PR).
func TestVMKeeperDisablePackageNotImplemented(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	stranger := crypto.AddressFromPreimage([]byte("stranger"))

	params := DefaultParams()
	params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
	params.PkgApprovers = []crypto.Address{approver}
	env.vmk.SetParams(ctx, params)

	// Non-approver is rejected on authorization, before hitting the stub.
	err := env.vmk.DisablePackage(ctx, MsgDisablePackage{Approver: stranger, PkgPath: "gno.land/r/test/x"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unauthorized")

	// Approver reaches the not-yet-implemented stub (returned as an
	// "unknown request" abci error until the follow-up PR lands).
	err = env.vmk.DisablePackage(ctx, MsgDisablePackage{Approver: approver, PkgPath: "gno.land/r/test/x"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown request")
}
