package vm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/types"
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

// TestVMKeeperEnablePackage_TypeExpansionGasCharged pins that the type-expansion
// walk is priced on the ENABLE path, not just on AddPackage.
//
// This is the one that matters under the inert policy: AddPackage deliberately does
// not type-check there ("the work is deferred, not avoided"), so MsgEnablePackage is
// the message that actually walks the submitted bytes. An unmetered walk here is the
// same consensus DoS the charge exists to stop, reached by a different message.
//
// The wiring has now slipped twice — once on AddPackage during development, and once
// here, where EnablePackage assembled TypeCheckOptions by hand and omitted the meter
// because it predated txTypeCheckOptions. Both times every other suite still passed.
// This test and its AddPackage twin are the only things that notice.
//
// Same discriminator as the AddPackage twin: two chains with equal declaration
// counts and equal source LENGTH, one through value containment ([0]tN, which
// validType expands) and one through a pointer (*tN, which it does not). Holding
// bytes equal keeps PreprocessGasPerByte from carrying the assertion.
func TestVMKeeperEnablePackage_TypeExpansionGasCharged(t *testing.T) {
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

	src := func(pkgName, elem string) string {
		var b strings.Builder
		fmt.Fprintf(&b, "package %s\n\ntype t0 struct{ v int }\n", pkgName)
		for i := 1; i <= 10; i++ {
			fmt.Fprintf(&b, "type t%d struct{ a, b %st%d }\n", i, elem, i-1)
		}
		return b.String()
	}
	// Pad the pointer variant so both weigh the same in source bytes.
	pad := func(body string, to int) string {
		if d := to - len(body); d > 0 {
			return body + "\n//" + strings.Repeat("x", d-3)
		}
		return body
	}

	// enable submits a package inert, then measures ONLY the gas the enable
	// message consumes — which is where the walk happens under this policy.
	enable := func(name, elem string) int64 {
		t.Helper()
		pkgPath := "gno.land/r/test/" + name
		width := len(src(name, "[0]"))
		files := []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
			{Name: name + ".gno", Body: pad(src(name, elem), width)},
		}
		require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(submitter, pkgPath, files)))

		parked := env.vmk.getGnoTransactionStore(ctx).GetInertPackage(pkgPath)
		require.NotNil(t, parked, "package must be parked before enabling")

		gm := types.NewInfiniteGasMeter()
		ectx := env.vmk.MakeGnoTransactionStore(env.ctx.WithGasMeter(gm))
		require.NoError(t, env.vmk.EnablePackage(ectx, MsgEnablePackage{
			Approver: approver, PkgPath: pkgPath,
			PkgHash: PackageContentHash(parked),
		}))
		return gm.GasConsumed()
	}

	ptrGas := enable("ptrchain", "*")
	valGas := enable("valchain", "[0]")

	t.Logf("enable: pointer %d gas, value %d gas, delta %d", ptrGas, valGas, valGas-ptrGas)
	require.Greater(t, valGas, ptrGas+ptrGas/4,
		"value containment (%d gas) must cost materially more to ENABLE than the "+
			"same-length source through pointers (%d gas); bytes are equal, so if "+
			"these match the expansion charge is not reaching EnablePackage",
		valGas, ptrGas)
}

// TestVMKeeperEnablePackage_TypeExpansionGasAccumulatesPerTx pins that the walk
// accumulates across MsgEnablePackage messages in one transaction.
//
// This is the cheapest form of the vector, and the one the approval oracle
// structurally cannot see. gpao verifies one PACKAGE at a time against a budget
// (-verify-budget, 10s by default); it has no notion of a transaction, so N
// packages that each pass honestly can be enabled together and cost N times as
// much. A per-package bound — the oracle's budget, or the per-package ceiling this
// change removed — is blind to composition by construction.
//
// Enable is cheaper than AddPackage for this because it charges NO byte gas:
// chargePreprocessGas runs at inert submit, at normal AddPackage and at Run, but
// never here ("EnablePackage deliberately does not charge it a second time"). The
// source bytes were paid for by an earlier transaction, so without the expansion
// charge the only gas an enable pays is store traffic — a ~100-byte message
// triggering an unbounded walk.
//
// The two shapes are byte-EQUAL (the pointer variant is padded with a comment) and
// carry the same declaration count, differing only in [0]tN versus *tN. That is
// what makes the expansion charge the discriminator: a first version of this test
// sized the budget from a measured single enable, which scales with whatever that
// enable costs, so it passed with the meter removed — proving only that SOMETHING
// accumulates. Verified against a reverted wiring: this version fails, that one did
// not.
func TestVMKeeperEnablePackage_TypeExpansionGasAccumulatesPerTx(t *testing.T) {
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

	const msgs = 4

	src := func(pkgName, elem string) string {
		var b strings.Builder
		fmt.Fprintf(&b, "package %s\n\ntype t0 struct{ v int }\n", pkgName)
		for i := 1; i <= 10; i++ {
			fmt.Fprintf(&b, "type t%d struct{ a, b %st%d }\n", i, elem, i-1)
		}
		return b.String()
	}
	pad := func(body string, to int) string {
		if d := to - len(body); d > 0 {
			return body + "\n//" + strings.Repeat("x", d-3)
		}
		return body
	}

	type parked struct{ path, hash string }
	// park submits a package inert. Names start with "pkg" so the .gno file sorts
	// after gnomod.toml; a MemPackage with unsorted files is rejected outright.
	park := func(ns, elem string, n int) parked {
		t.Helper()
		name := fmt.Sprintf("pkg%s%d", ns, n)
		path := "gno.land/r/test/" + name
		width := len(src(name, "[0]"))
		require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(submitter, path, []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(path)},
			{Name: name + ".gno", Body: pad(src(name, elem), width)},
		})))
		mp := env.vmk.getGnoTransactionStore(ctx).GetInertPackage(path)
		require.NotNil(t, mp)
		return parked{path: path, hash: PackageContentHash(mp)}
	}
	parkSet := func(ns, elem string) []parked {
		out := make([]parked, 0, msgs)
		for n := 1; n <= msgs; n++ {
			out = append(out, park(ns, elem, n))
		}
		return out
	}

	// enableAll runs every enable through ONE ctx and ONE finite meter, which is
	// what baseapp does for a multi-message tx.
	enableAll := func(set []parked, limit int64) (delivered int, err error) {
		gm := types.NewGasMeter(limit)
		ectx := env.vmk.MakeGnoTransactionStore(env.ctx.WithGasMeter(gm))
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("%v", r)
			}
		}()
		for _, p := range set {
			if e := env.vmk.EnablePackage(ectx, MsgEnablePackage{
				Approver: approver, PkgPath: p.path, PkgHash: p.hash,
			}); e != nil {
				return delivered, e
			}
			delivered++
		}
		return delivered, nil
	}
	costOne := func(p parked) int64 {
		gm := types.NewInfiniteGasMeter()
		ectx := env.vmk.MakeGnoTransactionStore(env.ctx.WithGasMeter(gm))
		require.NoError(t, env.vmk.EnablePackage(ectx, MsgEnablePackage{
			Approver: approver, PkgPath: p.path, PkgHash: p.hash,
		}))
		return gm.GasConsumed()
	}

	ptrOne := costOne(park("probeptr", "*", 0))
	valOne := costOne(park("probeval", "[0]", 0))
	require.Greater(t, valOne, ptrOne+ptrOne/4,
		"value containment (%d gas) must cost materially more to enable than the "+
			"same-length source through pointers (%d gas); bytes are equal, so if "+
			"these match the expansion charge is not reaching EnablePackage",
		valOne, ptrOne)

	// A budget covering every pointer enable, and a single value enable with room
	// to spare, but not all of them. The gap is expansion gas.
	budget := int64(msgs)*ptrOne + (int64(msgs)*(valOne-ptrOne))/2
	t.Logf("enable: pointer %d gas, value %d gas each; budget %d for %d messages",
		ptrOne, valOne, budget, msgs)
	require.Greater(t, budget, valOne, "each single enable must fit the budget")

	got, err := enableAll(parkSet("txptr", "*"), budget)
	require.NoError(t, err)
	assert.Equal(t, msgs, got, "pointer chains must all fit a %d budget", budget)

	got, err = enableAll(parkSet("txval", "[0]"), budget)
	require.Error(t, err,
		"%d value-containment enables shared a %d budget and all succeeded; the "+
			"expansion charge is not accumulating across messages", msgs, budget)
	assert.Contains(t, err.Error(), "out of gas")
	assert.Less(t, got, msgs, "the tx must abort partway, not after every message")
	t.Logf("value chains: %d/%d enables delivered before out of gas", got, msgs)
}
