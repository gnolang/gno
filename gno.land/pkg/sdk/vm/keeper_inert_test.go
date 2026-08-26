package vm

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
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
		_, _ = env.vmk.Call(ctx, NewMsgCall(submitter, nil, pkgPath, "Echo", []string{"hi"}))
	}, "inert package must not be callable")

	// ---- 2. Only an approver may enable --------------------------------------
	err = env.vmk.EnablePackage(ctx, approvalFor(t, env, ctx, stranger, pkgPath))
	require.Error(t, err, "non-approver must not be able to enable a package")
	assert.Contains(t, err.Error(), "unauthorized")

	// Still inert after the rejected attempt.
	assert.Nil(t, gnostore.GetPackage(pkgPath, false))
	assert.NotNil(t, gnostore.GetInertPackage(pkgPath))

	// ---- 3. Enabling an unknown path fails -----------------------------------
	err = env.vmk.EnablePackage(ctx, MsgEnablePackage{Approver: approver, PkgPath: "gno.land/r/test/missing"})
	require.Error(t, err, "enabling a non-existent inert package must fail")

	// ---- 4. Oracle approves --------------------------------------------------
	err = env.vmk.EnablePackage(ctx, approvalFor(t, env, ctx, approver, pkgPath))
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

	err := env.vmk.EnablePackage(ctx, approvalFor(t, env, ctx, approver, pkgPath))
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

// TestVMKeeperEnableTakesStorageDepositFromCreator pins who pays for the realm
// state that MsgEnablePackage brings into existence.
//
// Neither path charged it before. processStorageDeposit is driven entirely by
// RealmStorageDiffs(), which is empty at submit because nothing has executed
// yet, and EnablePackage never called processStorageDeposit at all — so under
// "inert" every byte of realm state created at activation was free.
//
// It is charged to the creator, not the approver. An approver is typically an
// automated oracle holding a hot key; billing it for other people's storage
// would let an attacker bleed its balance with large submissions and stall
// approvals for everyone. Only the creator's own submission can lock the
// creator's own funds, so the charge is consented by the act of submitting.
//
// The approver is funded in both cases below, so the failing case can only be
// about the creator's balance.
func TestVMKeeperEnableTakesStorageDepositFromCreator(t *testing.T) {
	const pkgPath = "gno.land/r/test/deposit"
	files := []*std.MemFile{
		{Name: "deposit.gno", Body: `package deposit

var Greeting = "hello"

func Set(cur realm, s string) { Greeting = s }`},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
	}

	t.Run("funded creator pays", func(t *testing.T) {
		env := setupTestEnv()
		ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

		approver := crypto.AddressFromPreimage([]byte("oracle"))
		creator := crypto.AddressFromPreimage([]byte("fundedcreator"))
		for _, addr := range []crypto.Address{approver, creator} {
			acc := env.acck.NewAccountWithAddress(ctx, addr)
			env.acck.SetAccount(ctx, acc)
			require.NoError(t, env.bankk.SetCoins(ctx, addr, initialBalance))
		}

		params := DefaultParams()
		params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
		params.PkgApprovers = []crypto.Address{approver}
		env.vmk.SetParams(ctx, params)

		require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, files)))

		before := env.bankk.GetCoins(ctx, creator).AmountOf(ugnot.Denom)
		approverBefore := env.bankk.GetCoins(ctx, approver).AmountOf(ugnot.Denom)

		require.NoError(t, env.vmk.EnablePackage(ctx,
			approvalFor(t, env, ctx, approver, pkgPath)))

		after := env.bankk.GetCoins(ctx, creator).AmountOf(ugnot.Denom)
		assert.Less(t, after, before,
			"the creator must have paid a storage deposit for the realm state enable created")
		assert.Equal(t, approverBefore, env.bankk.GetCoins(ctx, approver).AmountOf(ugnot.Denom),
			"the approver must not pay for someone else's storage")
	})

	t.Run("unfunded creator blocks activation", func(t *testing.T) {
		env := setupTestEnv()
		ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

		approver := crypto.AddressFromPreimage([]byte("oracle"))
		creator := crypto.AddressFromPreimage([]byte("brokecreator"))
		for _, addr := range []crypto.Address{approver, creator} {
			acc := env.acck.NewAccountWithAddress(ctx, addr)
			env.acck.SetAccount(ctx, acc)
		}
		// Approver funded, creator not.
		require.NoError(t, env.bankk.SetCoins(ctx, approver, initialBalance))

		params := DefaultParams()
		params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
		params.PkgApprovers = []crypto.Address{approver}
		env.vmk.SetParams(ctx, params)

		require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, files)))

		err := env.vmk.EnablePackage(ctx,
			approvalFor(t, env, ctx, approver, pkgPath))
		require.Error(t, err,
			"activation must fail when the creator cannot cover the storage deposit")
		assert.Contains(t, err.Error(), "deposit",
			"the failure must be about the deposit, not something incidental")

		// Deliberately not asserting that the package is absent from the store
		// here. EnablePackage writes into the gno transaction store as it goes,
		// and discarding those writes on error is the tx boundary's job
		// (SetEndTxHook commits only when the result is OK) — a direct keeper
		// call like this one bypasses it, so the package IS present in the
		// store at this point. The returned error is the part this change
		// controls; the rollback is pre-existing machinery tested elsewhere.
	})
}
