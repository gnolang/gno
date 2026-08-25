package vm

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	bft "github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// inertEnv builds a chain running the "inert" policy with one approver.
func inertEnv(t *testing.T, approver crypto.Address, funded ...crypto.Address) (testEnv, sdk.Context) {
	t.Helper()
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	for _, addr := range append([]crypto.Address{approver}, funded...) {
		acc := env.acck.NewAccountWithAddress(ctx, addr)
		env.acck.SetAccount(ctx, acc)
		require.NoError(t, env.bankk.SetCoins(ctx, addr, initialBalance))
	}
	params := DefaultParams()
	params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
	params.PkgApprovers = []crypto.Address{approver}
	env.vmk.SetParams(ctx, params)
	return env, ctx
}

// TestVMKeeperInertHonorsDeclaredMaxDeposit pins that the ceiling a creator
// declares on MsgAddPackage still binds the charge that MsgEnablePackage makes
// later.
//
// Inert is the only path where the message that DECLARES the deposit ceiling is
// not the message that SPENDS against it. If the declaration is not carried
// across, enable falls back to params.DefaultDeposit — so a creator who asked
// to risk 1000ugnot could be charged up to the chain default (100 GNOT) on a
// transaction they do not sign and cannot refuse.
func TestVMKeeperInertHonorsDeclaredMaxDeposit(t *testing.T) {
	const pkgPath = "gno.land/r/test/declared"

	// Built fresh per submission. stampGnomod rewrites gnomod.toml in place via
	// SetFile, so a shared slice leaks one subtest's stamped max_deposit into the
	// next -- which made the "no declaration" subtest below look covered when it
	// was really reading the previous subtest's ceiling.
	newFiles := func() []*std.MemFile {
		return []*std.MemFile{
			{Name: "declared.gno", Body: `package declared

var Greeting = "hello"

func Set(cur realm, s string) { Greeting = s }`},
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		}
	}

	submit := func(t *testing.T, env testEnv, ctx sdk.Context, creator crypto.Address, ceiling std.Coins) error {
		t.Helper()
		msg := NewMsgAddPackage(creator, pkgPath, newFiles())
		msg.MaxDeposit = ceiling
		return env.vmk.AddPackage(ctx, msg)
	}

	t.Run("a ceiling below the real cost blocks the charge", func(t *testing.T) {
		approver := crypto.AddressFromPreimage([]byte("oracle"))
		creator := crypto.AddressFromPreimage([]byte("frugalcreator"))
		env, ctx := inertEnv(t, approver, creator)

		// Far below the ~210_200ugnot this package actually needs, but far
		// below params.DefaultDeposit too — so if the declaration is dropped,
		// the fallback silently covers it and the enable succeeds.
		require.NoError(t, submit(t, env, ctx, creator, std.NewCoins(std.NewCoin(ugnot.Denom, 1000))))

		before := env.bankk.GetCoins(ctx, creator).AmountOf(ugnot.Denom)
		err := env.vmk.EnablePackage(ctx, MsgEnablePackage{Approver: approver, PkgPath: pkgPath})
		require.Error(t, err,
			"enable must respect the ceiling the creator declared at submit")
		assert.Contains(t, err.Error(), "deposit",
			"the refusal must be about the deposit, not something incidental")
		assert.Equal(t, before, env.bankk.GetCoins(ctx, creator).AmountOf(ugnot.Denom),
			"a refused enable must not move the creator's funds")
	})

	t.Run("a sufficient declared ceiling still activates", func(t *testing.T) {
		approver := crypto.AddressFromPreimage([]byte("oracle"))
		creator := crypto.AddressFromPreimage([]byte("generouscreator"))
		env, ctx := inertEnv(t, approver, creator)

		require.NoError(t, submit(t, env, ctx, creator, std.NewCoins(std.NewCoin(ugnot.Denom, 500_000_000))))

		before := env.bankk.GetCoins(ctx, creator).AmountOf(ugnot.Denom)
		require.NoError(t, env.vmk.EnablePackage(ctx,
			MsgEnablePackage{Approver: approver, PkgPath: pkgPath}),
			"a ceiling above the real cost must not get in the way")

		// Pin the MAGNITUDE against the realm's own recorded deposit, not just
		// "something was taken". A ceiling must cap the charge, never set it:
		// asserting only that the balance fell would pass just as happily if
		// the charge were doubled, or if an extra amount were siphoned on top.
		charged := before - env.bankk.GetCoins(ctx, creator).AmountOf(ugnot.Denom)
		rlm := env.vmk.getGnoTransactionStore(ctx).GetPackageRealm(pkgPath)
		require.NotNil(t, rlm)
		// Derive the expectation from bytes and price, NOT from rlm.Deposit:
		// lockStorageDeposit adds the charge to rlm.Deposit, so comparing the
		// two is self-referential and a doubled charge would satisfy it.
		price := std.MustParseCoin(DefaultParams().StoragePrice).Amount
		assert.Equal(t, int64(rlm.Storage)*price, charged,
			"the creator must be charged exactly bytes x price, no more")
		assert.Less(t, charged, int64(500_000_000),
			"and strictly less than the declared ceiling, which caps rather than sets")
	})

	t.Run("no declaration falls back to the chain default", func(t *testing.T) {
		approver := crypto.AddressFromPreimage([]byte("oracle"))
		creator := crypto.AddressFromPreimage([]byte("silentcreator"))
		env, ctx := inertEnv(t, approver, creator)

		// Asserting only that an undeclared submission activates would prove
		// nothing: it holds for ANY ceiling at or above the real cost,
		// including no ceiling at all. So squeeze the chain default until it
		// bites, and require the refusal to follow it. That pins WHICH value
		// the fallback reads.
		params := DefaultParams()
		params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
		params.PkgApprovers = []crypto.Address{approver}
		params.DefaultDeposit = "1000ugnot"
		env.vmk.SetParams(ctx, params)

		require.NoError(t, submit(t, env, ctx, creator, nil))
		err := env.vmk.EnablePackage(ctx,
			MsgEnablePackage{Approver: approver, PkgPath: pkgPath})
		require.Error(t, err,
			"with nothing declared the charge must be capped by params.DefaultDeposit")
		assert.Contains(t, err.Error(), "deposit")
	})
}

// TestVMKeeperInertResubmissionIsCreatorBound pins that only the original
// submitter may replace a package already parked at a path.
//
// AddPackage's ErrPkgAlreadyExists guard reads GetPackage, which sees only the
// ACTIVE store; parked packages live in a separate key space and are invisible
// to it, and AddInertPackage overwrites unconditionally. Without a guard, an
// attacker can wait for an approver to review source at a path and then
// front-run the enable, having their own bytes type-checked, init()ed and
// stamped as creator under the reviewed path.
func TestVMKeeperInertResubmissionIsCreatorBound(t *testing.T) {
	const pkgPath = "gno.land/r/test/contested"
	filesFor := func(body string) []*std.MemFile {
		return []*std.MemFile{
			{Name: "contested.gno", Body: body},
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		}
	}
	original := `package contested

func Who(cur realm) string { return "original" }`
	swapped := `package contested

func Who(cur realm) string { return "swapped" }`

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("firstsubmitter"))
	attacker := crypto.AddressFromPreimage([]byte("frontrunner"))
	env, ctx := inertEnv(t, approver, creator, attacker)

	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, filesFor(original))))

	t.Run("a stranger cannot overwrite a parked submission", func(t *testing.T) {
		err := env.vmk.AddPackage(ctx, NewMsgAddPackage(attacker, pkgPath, filesFor(swapped)))
		require.Error(t, err, "a third party must not be able to replace parked bytes")
		// The detail lives in the wrapped trace, which is what reaches the ABCI
		// log; Error() alone is the generic "package already exists" that the
		// live-package guard also returns, so asserting on it would not tell
		// the two apart.
		assert.Contains(t, fmt.Sprintf("%+v", err), "already awaiting approval")

		parked := env.vmk.getGnoTransactionStore(ctx).GetInertPackage(pkgPath)
		require.NotNil(t, parked)
		assert.Contains(t, parked.GetFile("contested.gno").Body, "original",
			"the parked bytes must be the ones the approver would review")
	})

	t.Run("the original submitter may still retry", func(t *testing.T) {
		require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, filesFor(swapped))),
			"resubmission by the same creator is the retry path after a failed enable")

		parked := env.vmk.getGnoTransactionStore(ctx).GetInertPackage(pkgPath)
		require.NotNil(t, parked)
		assert.Contains(t, parked.GetFile("contested.gno").Body, "swapped")
	})
}

// TestVMKeeperInertSubmitKeepsLivePrivateSource pins that parking a package over
// a live PRIVATE package does not destroy the live package's source.
//
// AddPackage deletes a private package's stored mempackage blobs before
// redeploying, because the prod / #allbutprod pair is not fully replaced by a
// re-add. That delete must not run on the inert path, which returns without
// ever calling AddMemPackage: the source would be gone while the realm, its
// objects and the package index survive. At boot,
// PreprocessAllFilesAndSaveBlockNodes skips a nil mempackage silently, so a
// restarted node rebuilds no PackageNode and panics on call, while a node that
// has not restarted keeps answering from its in-memory node cache. That is a
// consensus split keyed on restart history, which is exactly the hazard the
// surrounding code was written to avoid.
func TestVMKeeperInertSubmitKeepsLivePrivateSource(t *testing.T) {
	const pkgPath = "gno.land/r/test/livepriv"
	privateFiles := []*std.MemFile{
		{Name: "gnomod.toml", Body: `module = "gno.land/r/test/livepriv"
gno = "0.9"
private = true`},
		{Name: "livepriv.gno", Body: `package livepriv

func Echo(cur realm) string { return "live" }`},
	}

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("privowner"))

	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	for _, addr := range []crypto.Address{approver, creator} {
		acc := env.acck.NewAccountWithAddress(ctx, addr)
		env.acck.SetAccount(ctx, acc)
		require.NoError(t, env.bankk.SetCoins(ctx, addr, initialBalance))
	}

	// Deploy it for real first, under the default permissionless policy.
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, privateFiles)))
	require.NotNil(t, env.vmk.getGnoTransactionStore(ctx).GetMemPackage(pkgPath),
		"precondition: the package is live and its source is stored")

	// Now the chain switches to inert and the same path is submitted again.
	params := DefaultParams()
	params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
	params.PkgApprovers = []crypto.Address{approver}
	env.vmk.SetParams(ctx, params)

	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, privateFiles)))

	assert.NotNil(t, env.vmk.getGnoTransactionStore(ctx).GetMemPackage(pkgPath),
		"parking a submission must not delete the live package's source: the realm "+
			"and package index survive, so a restarted node would diverge from one that has not restarted")
}

// TestVMKeeperEnableCannotTakeOverALivePackage pins that activating a parked
// package cannot replace one that is already live.
//
// EnablePackage is the deferred second half of a deploy, but it enforced none
// of the deploy's preconditions: its entire precondition set was "the sender is
// an approver" and "something is parked here". A path can be parked and live at
// the same time, because the two live in different key spaces and nothing
// clears a parked blob when governance moves the policy off "inert". So an
// attacker parks at a path, waits for the policy to open up, lets someone else
// deploy there for real, and then any approver's routine enable silently
// replaces the live package with the attacker's bytes -- running init() with
// the attacker as OriginCaller, which is what p/nt/ownable records as owner.
func TestVMKeeperEnableCannotTakeOverALivePackage(t *testing.T) {
	const pkgPath = "gno.land/r/test/takeover"
	filesFor := func(body string) []*std.MemFile {
		return []*std.MemFile{
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
			{Name: "takeover.gno", Body: body},
		}
	}

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	attacker := crypto.AddressFromPreimage([]byte("squatter"))
	owner := crypto.AddressFromPreimage([]byte("realowner"))
	env, ctx := inertEnv(t, approver, attacker, owner)

	// 1. Under inert, the attacker parks a package at the path.
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(attacker, pkgPath, filesFor(
		`package takeover

func Who(cur realm) string { return "attacker" }`))))

	// 2. Governance opens the chain. The parked blob is not cleaned up, and
	//    the approver list is not cleared.
	params := DefaultParams()
	params.CodeSubmissionPolicy = CodeSubmissionPolicyPermissionless
	params.PkgApprovers = []crypto.Address{approver}
	env.vmk.SetParams(ctx, params)

	// 3. Somebody deploys at that path for real.
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(owner, pkgPath, filesFor(
		`package takeover

func Who(cur realm) string { return "owner" }`))))
	require.NotNil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false),
		"precondition: the path is live")

	// 3b. Governance returns to inert. Enable refuses outright under any other
	//     policy, so without this the refusal below would be the policy check
	//     rather than the takeover guard -- and the guard would go untested.
	//     A round trip is also the harder case: the parked blob survives it.
	params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
	env.vmk.SetParams(ctx, params)

	// 4. A routine enable must not hand the path back to the attacker.
	err := env.vmk.EnablePackage(ctx, MsgEnablePackage{Approver: approver, PkgPath: pkgPath})
	require.Error(t, err, "enabling over a live package is a takeover and must be refused")
	// Discriminate. "already exists in cache" is a DIFFERENT failure --
	// RunMemPackage's SetCachePackage panic, which an earlier version of this
	// guard triggered on every private redeploy -- and matching the bare phrase
	// would let that stand in for the guard actually firing. The guard's own
	// message names the path, and lives in the wrapped trace rather than in
	// Error().
	full := fmt.Sprintf("%+v", err)
	assert.Contains(t, full, "package already exists: "+pkgPath)
	assert.NotContains(t, full, "in cache",
		"the refusal must be the guard, not a cache-collision panic")

	stored := env.vmk.getGnoTransactionStore(ctx).GetMemPackage(pkgPath)
	require.NotNil(t, stored, "the live package must survive the refused enable")
	assert.Contains(t, stored.GetFile("takeover.gno").Body, "owner",
		"the live package's source must still be the one its deployer published")
}

// TestVMKeeperInertIgnoresHandWrittenMaxDeposit pins that the deposit ceiling
// read at enable comes from the MESSAGE, never from the file.
//
// The [addpkg] section is keeper bookkeeping, but it lives in a file the
// submitter authors, so any field the keeper does not overwrite is
// attacker-supplied. A hand-written max_deposit that survived the stamp would
// be read back at enable as though the message had declared it.
func TestVMKeeperInertIgnoresHandWrittenMaxDeposit(t *testing.T) {
	const pkgPath = "gno.land/r/test/handwritten"
	// Declares a ceiling far below what this package needs, in the file only.
	// If it were honoured, the enable would be refused.
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: `module = "gno.land/r/test/handwritten"
gno = "0.9"

[addpkg]
max_deposit = "1ugnot"`},
		{Name: "handwritten.gno", Body: `package handwritten

var Greeting = "hello"

func Set(cur realm, s string) { Greeting = s }`},
	}

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("handwriter"))
	env, ctx := inertEnv(t, approver, creator)

	// Message declares nothing, so the chain default applies.
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, files)))

	parked := env.vmk.getGnoTransactionStore(ctx).GetInertPackage(pkgPath)
	require.NotNil(t, parked)
	assert.NotContains(t, parked.GetFile("gnomod.toml").Body, "1ugnot",
		"the stamp must overwrite a hand-written ceiling, not preserve it")

	require.NoError(t, env.vmk.EnablePackage(ctx,
		MsgEnablePackage{Approver: approver, PkgPath: pkgPath}),
		"a ceiling the message never declared must not be able to block activation")
}

// TestVMKeeperEnableCanRedeployALivePrivatePackage pins the other half of the
// live-package rule: a PRIVATE package may be replaced, exactly as the ordinary
// deploy path allows.
//
// This is the branch that is easy to add and never exercise. It also sits on a
// trap: reading the live PackageValue to decide whether the path is private
// populates the object cache, and RunMemPackage's SetCachePackage panics when
// the package is already cached. AddPackage escapes that only by accident --
// checkNamespacePermission re-enters getGnoTransactionStore, which clears the
// object cache between the read and the run -- and EnablePackage calls neither
// of those. So the liveness probe here must not load the package value.
func TestVMKeeperEnableCanRedeployALivePrivatePackage(t *testing.T) {
	const pkgPath = "gno.land/r/test/privcycle"
	filesFor := func(body string) []*std.MemFile {
		return []*std.MemFile{
			{Name: "gnomod.toml", Body: `module = "gno.land/r/test/privcycle"
gno = "0.9"
private = true`},
			{Name: "privcycle.gno", Body: body},
		}
	}

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("privcycler"))
	env, ctx := inertEnv(t, approver, creator)

	// v1: park, then activate.
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, filesFor(
		`package privcycle

func Which(cur realm) string { return "v1" }`))))
	require.NoError(t, env.vmk.EnablePackage(ctx,
		MsgEnablePackage{Approver: approver, PkgPath: pkgPath}))

	// v2: park over the now-live private package, then activate again.
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, filesFor(
		`package privcycle

func Which(cur realm) string { return "v2" }`))),
		"the same creator may park a replacement for their own private package")

	require.NoError(t, env.vmk.EnablePackage(ctx,
		MsgEnablePackage{Approver: approver, PkgPath: pkgPath}),
		"a private package may be replaced, so activating the replacement must work")

	stored := env.vmk.getGnoTransactionStore(ctx).GetMemPackage(pkgPath)
	require.NotNil(t, stored)
	assert.Contains(t, stored.GetFile("privcycle.gno").Body, "v2",
		"the activated replacement must be the source that is stored")
}

// TestVMKeeperInertIgnoresHandWrittenCreator pins the security-relevant half of
// the unconditional-stamp contract: the creator recorded on a parked package is
// the message signer, never a value the submitter wrote into their own file.
//
// This matters more than the ceiling. EnablePackage reads AddPkg.Creator back to
// decide who init() runs as and who pays the deposit, and the re-submission
// guard compares against it to decide who may replace the parked bytes. A
// hand-written creator that survived the stamp would let a submitter park a
// package attributed to someone else -- charging them, initializing under their
// identity, and locking them into the retry path.
func TestVMKeeperInertIgnoresHandWrittenCreator(t *testing.T) {
	const pkgPath = "gno.land/r/test/forgedcreator"
	victim := crypto.AddressFromPreimage([]byte("victim"))
	files := []*std.MemFile{
		{Name: "forgedcreator.gno", Body: `package forgedcreator

func Who(cur realm) string { return "x" }`},
		{Name: "gnomod.toml", Body: `module = "gno.land/r/test/forgedcreator"
gno = "0.9"

[addpkg]
creator = "` + victim.String() + `"
height = 999`},
	}

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	forger := crypto.AddressFromPreimage([]byte("forger"))
	env, ctx := inertEnv(t, approver, forger, victim)

	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(forger, pkgPath, files)))

	parked := env.vmk.getGnoTransactionStore(ctx).GetInertPackage(pkgPath)
	require.NotNil(t, parked)
	stamped := parked.GetFile("gnomod.toml").Body
	assert.Contains(t, stamped, forger.String(),
		"the stamped creator must be the message signer")
	assert.NotContains(t, stamped, victim.String(),
		"a hand-written creator must not survive the stamp")

	// And the consequence that makes it matter: the deposit follows the signer.
	victimBefore := env.bankk.GetCoins(ctx, victim).AmountOf(ugnot.Denom)
	forgerBefore := env.bankk.GetCoins(ctx, forger).AmountOf(ugnot.Denom)
	require.NoError(t, env.vmk.EnablePackage(ctx,
		MsgEnablePackage{Approver: approver, PkgPath: pkgPath}))
	assert.Equal(t, victimBefore, env.bankk.GetCoins(ctx, victim).AmountOf(ugnot.Denom),
		"the named victim must not be charged for someone else's submission")

	// Pinned to bytes x price rather than "the balance went down". Asserting
	// only a decrease would pass just as happily if the forger were charged
	// twice over, or if something unrelated were taken alongside the deposit.
	charged := forgerBefore - env.bankk.GetCoins(ctx, forger).AmountOf(ugnot.Denom)
	rlm := env.vmk.getGnoTransactionStore(ctx).GetPackageRealm(pkgPath)
	require.NotNil(t, rlm)
	price := std.MustParseCoin(DefaultParams().StoragePrice).Amount
	assert.Equal(t, int64(rlm.Storage)*price, charged,
		"the actual submitter pays, and pays exactly the storage deposit")
}

// TestVMKeeperEnableCannotPublicizeALivePrivateRealm pins that activating a
// parked PUBLIC package over a live PRIVATE realm is refused.
//
// checkGnomodConstraints enforces "a private package cannot be overridden by a
// public package" at deploy, but on the inert path it runs at SUBMIT, against
// whatever was live then — and for a package parked before anything existed at
// the path, that was nothing. So the rule has to be re-applied at enable.
//
// The consequence of missing it is not cosmetic. `private` means other realms
// cannot import the package and cannot store references to its objects.
// Flipping a live private realm public retroactively exposes state that was
// persisted under that invariant.
func TestVMKeeperEnableCannotPublicizeALivePrivateRealm(t *testing.T) {
	const pkgPath = "gno.land/r/test/privflip"
	mod := func(private bool) string {
		m := "module = \"" + pkgPath + "\"\ngno = \"0.9\""
		if private {
			m += "\nprivate = true"
		}
		return m
	}
	filesFor := func(private bool, body string) []*std.MemFile {
		return []*std.MemFile{
			{Name: "gnomod.toml", Body: mod(private)},
			{Name: "privflip.gno", Body: body},
		}
	}

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	attacker := crypto.AddressFromPreimage([]byte("earlyparker"))
	owner := crypto.AddressFromPreimage([]byte("privowner2"))
	env, ctx := inertEnv(t, approver, attacker, owner)

	// 1. Nothing is live yet, so the submit-time constraint check sees no
	//    private package to protect. The attacker parks a PUBLIC package.
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(attacker, pkgPath, filesFor(false,
		`package privflip

func Who(cur realm) string { return "attacker" }`))))

	// 2. Governance opens the chain and the owner deploys a PRIVATE realm there.
	params := DefaultParams()
	params.CodeSubmissionPolicy = CodeSubmissionPolicyPermissionless
	params.PkgApprovers = []crypto.Address{approver}
	env.vmk.SetParams(ctx, params)
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(owner, pkgPath, filesFor(true,
		`package privflip

func Who(cur realm) string { return "owner" }`))))

	// 2b. Back to inert, so the refusal below is the private-override rule and
	//     not EnablePackage's policy check.
	params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
	env.vmk.SetParams(ctx, params)

	// 3. The stale parked public package must not be activatable over it.
	err := env.vmk.EnablePackage(ctx, MsgEnablePackage{Approver: approver, PkgPath: pkgPath})
	require.Error(t, err,
		"a public package must not be activated over a live private realm")
	assert.Contains(t, fmt.Sprintf("%+v", err), "cannot be overridden by a public package")

	stored := env.vmk.getGnoTransactionStore(ctx).GetMemPackage(pkgPath)
	require.NotNil(t, stored)
	assert.Contains(t, stored.GetFile("gnomod.toml").Body, "private = true",
		"the live realm must still be private")
}

// TestVMKeeperInertChargesPreprocessGasAtSubmit pins the economic argument the
// inert policy rests on: the submitter pays for the compile work their bytes
// will cause, at submit, priced by source length.
//
// Deleting chargePreprocessGas from the inert branch previously passed every
// test in the repository, which left the entire anti-DoS justification for
// "inert" unguarded -- a submitter could park an arbitrarily large package for
// the price of one amino write and leave the compile bill to whoever enables it.
func TestVMKeeperInertChargesPreprocessGasAtSubmit(t *testing.T) {
	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("gascreator"))

	// Gas charged for a submission, measured on the env's own meter. The package
	// clause has to match the last path element, so the body is generated.
	submitGas := func(t *testing.T, name string, extraLines int) (int64, int) {
		t.Helper()
		pkgPath := "gno.land/r/test/" + name
		var body strings.Builder
		body.WriteString("package " + name + "\n\nfunc F(cur realm) int { return 1 }\n")
		for range extraLines {
			body.WriteString("// padding to make this package substantially longer\n")
		}

		env, ctx := inertEnv(t, approver, creator)
		// Sorted by name: "gas.gno" < "gnomod.toml".
		files := []*std.MemFile{
			{Name: "gas.gno", Body: body.String()},
			{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		}
		before := ctx.GasMeter().GasConsumed()
		require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, files)))
		return ctx.GasMeter().GasConsumed() - before, len(body.String())
	}

	smallGas, smallLen := submitGas(t, "gassmall", 0)
	largeGas, largeLen := submitGas(t, "gaslarge", 400)

	perByte := DefaultParams().PreprocessGasPerByte
	require.Positive(t, perByte)

	delta := int64(largeLen - smallLen)
	require.Positive(t, delta)
	assert.GreaterOrEqual(t, largeGas-smallGas, perByte*delta,
		"the submitter must be charged at least PreprocessGasPerByte for every "+
			"additional source byte parked; without that, parking a large "+
			"package is nearly free and the compile bill falls on whoever enables it")
}

// TestInertPolicyDoesNotParkGenesisPackages pins that a chain launching under
// the "inert" policy still deploys its own genesis packages.
//
// Genesis content is the chain's own, already reviewed by whoever wrote the
// genesis file -- there is nobody for an approver to protect it from. Parking it
// would mean a chain that boots with nothing deployed: no r/sys/params, no
// govdao, so no way to propose the change that would let anything be enabled,
// and no approver able to act because the realms an approver needs do not exist
// yet. The policy is about what STRANGERS may submit after the chain is running.
//
// Every other height-sensitive rule in AddPackage already knows this: the type
// checker drops to genesis mode at height 0, and the draft-package rule is
// waived there too.
func TestInertPolicyDoesNotParkGenesisPackages(t *testing.T) {
	const pkgPath = "gno.land/r/test/atgenesis"
	files := []*std.MemFile{
		{Name: "atgenesis.gno", Body: `package atgenesis

func Hello(cur realm) string { return "hi" }`},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
	}

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("genesiscreator"))
	env, ctx := inertEnv(t, approver, creator)

	// Height 0 is genesis. The shared env runs at 42, which is why the inert
	// branch was reached unconditionally before.
	genesisCtx := ctx.WithBlockHeader(&bft.Header{ChainID: ctx.ChainID(), Height: 0})

	require.NoError(t, env.vmk.AddPackage(genesisCtx, NewMsgAddPackage(creator, pkgPath, files)))

	store := env.vmk.getGnoTransactionStore(genesisCtx)
	assert.Nil(t, store.GetInertPackage(pkgPath),
		"a genesis package must not be parked: there is no functioning chain yet "+
			"for an approver to enable it from")
	assert.NotNil(t, store.GetMemPackage(pkgPath),
		"it must be deployed, like every other genesis package")
}

// TestOrdinaryDeployStoresNoMaxDeposit pins that adding the max_deposit field
// did not change what an ordinary deploy stores.
//
// A package's gnomod.toml is stored in chain state, so its bytes are covered by
// the app hash. If the ordinary path started writing a max_deposit line, every
// deploy on every chain would produce different bytes -- a consensus break for a
// field only the inert path needs. The empty value plus `omitempty` is what
// keeps that from happening, and nothing else was checking it.
func TestOrdinaryDeployStoresNoMaxDeposit(t *testing.T) {
	const pkgPath = "gno.land/r/test/plaindeploy"
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "plaindeploy.gno", Body: `package plaindeploy

func Hello(cur realm) string { return "hi" }`},
	}

	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	creator := crypto.AddressFromPreimage([]byte("plaincreator"))
	acc := env.acck.NewAccountWithAddress(ctx, creator)
	env.acck.SetAccount(ctx, acc)
	require.NoError(t, env.bankk.SetCoins(ctx, creator, initialBalance))

	// Default policy: permissionless, so this is the ordinary deploy path.
	msg := NewMsgAddPackage(creator, pkgPath, files)
	msg.MaxDeposit = std.NewCoins(std.NewCoin(ugnot.Denom, 500_000_000))
	require.NoError(t, env.vmk.AddPackage(ctx, msg))

	stored := env.vmk.getGnoTransactionStore(ctx).GetMemPackage(pkgPath)
	require.NotNil(t, stored)
	body := stored.GetFile("gnomod.toml").Body

	// Declared on the message and still absent from what is stored: the
	// ordinary path consumes the ceiling in the same transaction, so there is
	// nothing to carry.
	assert.NotContains(t, body, "max_deposit",
		"an ordinary deploy must store the same bytes it always did")
	// The two fields that DO round-trip are still there, so this is not passing
	// because the stamp stopped working altogether.
	assert.Contains(t, body, "creator")
	assert.Contains(t, body, creator.String())
}

// TestVMKeeperInertRefusesPayableSubmission pins that a submission under
// "inert" may not carry coins.
//
// The ordinary deploy path credits msg.Send to the package address AND presents
// it to init() as the origin-send envelope. Inert can only do the first half:
// the coins would move at submit while init() runs at enable, in a message that
// sends nothing. EnablePackage therefore builds its ExecContext with an empty
// OriginSend and no OriginSendRecipient, and a payable init() does not merely
// see an empty envelope -- it panics on the recipient mismatch.
//
// Left unrefused, the same source deploys under "permissionless" and fails
// under "inert", which makes a governance parameter change program semantics.
// Deterministic, so never a fork, but a submitter has no way to see it coming.
func TestVMKeeperInertRefusesPayableSubmission(t *testing.T) {
	const pkgPath = "gno.land/r/test/payable"

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("payer"))
	env, ctx := inertEnv(t, approver, creator)

	msg := NewMsgAddPackage(creator, pkgPath, []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "payable.gno", Body: `package payable

func Noop(cur realm) {}`},
	})
	msg.Send = std.MustParseCoins(ugnot.ValueString(1_000_000))

	err := env.vmk.AddPackage(ctx, msg)
	require.Error(t, err, "a payment that init() will never see must be refused")
	assert.Contains(t, fmt.Sprintf("%+v", err), "cannot carry a payment")

	// Nothing moved and nothing was parked: the refusal is total, not partial.
	assert.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetInertPackage(pkgPath),
		"a refused submission must not be parked")
	assert.True(t, env.bankk.GetCoins(ctx, gnolang.DerivePkgCryptoAddr(pkgPath)).IsZero(),
		"no coins may reach the package address")

	// The same package without a payment goes through, so the refusal is about
	// the coins and not about the package.
	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "payable.gno", Body: `package payable

func Noop(cur realm) {}`},
	})))
}

// TestVMKeeperEnableRequiresInertPolicy pins that enable is valid only while
// "inert" is the policy in force.
//
// Enable exists to finish a submission that "inert" split in two. Without this
// check a parked package stays activatable forever: governance moves to
// "permissioned" precisely to stop strangers getting code onto the chain, and
// every package parked during the "inert" era would remain a stranger's pending
// deploy that a single approver could still land -- under a policy that would
// have refused the submission outright.
//
// PkgApprovers is not a substitute. It is not cleared when the policy changes,
// and an approver's mandate was to activate what the policy of the day
// accepted, not to carry it across a governance decision.
func TestVMKeeperEnableRequiresInertPolicy(t *testing.T) {
	const pkgPath = "gno.land/r/test/policygone"

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("submitter"))
	env, ctx := inertEnv(t, approver, creator)

	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "policygone.gno", Body: `package policygone

func Who(cur realm) string { return "parked" }`},
	})))

	for _, policy := range []CodeSubmissionPolicy{
		CodeSubmissionPolicyPermissionless,
		CodeSubmissionPolicyPermissioned,
	} {
		params := DefaultParams()
		params.CodeSubmissionPolicy = policy
		params.PkgApprovers = []crypto.Address{approver}
		env.vmk.SetParams(ctx, params)

		err := env.vmk.EnablePackage(ctx, MsgEnablePackage{Approver: approver, PkgPath: pkgPath})
		require.Error(t, err, "enable must be refused under policy %q", policy)
		assert.Contains(t, fmt.Sprintf("%+v", err), "packages cannot be enabled")
		assert.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false),
			"the package must not have been activated under policy %q", policy)
	}

	// Returning to inert makes it activatable again: the check is about the
	// policy in force, not a permanent disqualification.
	params := DefaultParams()
	params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
	params.PkgApprovers = []crypto.Address{approver}
	env.vmk.SetParams(ctx, params)
	require.NoError(t, env.vmk.EnablePackage(ctx,
		MsgEnablePackage{Approver: approver, PkgPath: pkgPath}))
	assert.NotNil(t, env.vmk.getGnoTransactionStore(ctx).GetPackage(pkgPath, false))
}

// sortMemFiles puts files in the order ValidateMemPackage requires.
//
// Worth a helper rather than a hand-written order: whether "<name>.gno" sorts
// before or after "gnomod.toml" depends on the name, so a fixed order is
// correct for some packages and rejected for others.
func sortMemFiles(files []*std.MemFile) []*std.MemFile {
	slices.SortFunc(files, func(a, b *std.MemFile) int {
		return strings.Compare(a.Name, b.Name)
	})
	return files
}

// replayFiles builds a trivial package whose directory name is also its package
// name, for the genesis-replay and submission-charge tests below.
func replayFiles(name string) []*std.MemFile {
	path := "gno.land/r/test/" + name
	return sortMemFiles([]*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(path)},
		{Name: name + ".gno", Body: "package " + name + `

func Who(cur realm) string { return "live" }`},
	})
}

// TestVMKeeperGenesisReplayFollowsTheReplayedPolicy pins that replayed history
// is executed under the policy that governed it, and that the policy is simply
// read rather than skipped.
//
// The params already carry the answer: `gnogenesis fork generate` copies the
// source chain's vm params into the fork's genesis untouched, and every
// historical governance tx that moved the policy re-applies as it replays.
//
// Taking the ordinary path regardless would bring up a package the source chain
// left awaiting approval -- running its init() and charging its creator a
// storage deposit that the source chain never charged.
func TestVMKeeperGenesisReplayFollowsTheReplayedPolicy(t *testing.T) {
	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("historical"))
	env, ctx := inertEnv(t, approver, creator)

	replayCtx := ctx.WithValue(auth.GenesisReplayKey{}, true)

	// Replayed while the policy reads "inert": parks, as it did on the source.
	const parkedPath = "gno.land/r/test/replayparked"
	require.NoError(t, env.vmk.AddPackage(replayCtx,
		NewMsgAddPackage(creator, parkedPath, replayFiles("replayparked"))))
	gs := env.vmk.getGnoTransactionStore(ctx)
	assert.Nil(t, gs.GetPackage(parkedPath, false),
		"a package the source chain parked must not be deployed live by replay")
	assert.NotNil(t, gs.GetInertPackage(parkedPath),
		"it must park, exactly as it did on the source chain")

	// Replayed while the policy reads something else: deploys live. This is the
	// same history a governance tx earlier in the replay would produce.
	open := DefaultParams()
	open.CodeSubmissionPolicy = CodeSubmissionPolicyPermissionless
	open.PkgApprovers = []crypto.Address{approver}
	require.NoError(t, env.vmk.SetParams(ctx, open))

	const livePath = "gno.land/r/test/replaylive"
	require.NoError(t, env.vmk.AddPackage(replayCtx,
		NewMsgAddPackage(creator, livePath, replayFiles("replaylive"))))
	assert.NotNil(t, gs.GetPackage(livePath, false),
		"a package the source chain deployed live must not be parked by replay")
	assert.Nil(t, gs.GetInertPackage(livePath))
}

// TestVMKeeperGenesisReplayEnableActivatesAParkedPackage continues the story
// TestVMKeeperGenesisReplayFollowsTheReplayedPolicy starts.
//
// The replayed MsgAddPackage parks, so the matching MsgEnablePackage -- also in
// the replayed history -- finds its package and activates it. That reproduces
// the source chain's outcome rather than approximating it.
//
// The two authorization gates are exempt: replay runs after InitGenesis has
// installed the fork's params, so a fork that moved off "inert" or rotated
// pkg_approvers must not refuse the enables in its own history.
func TestVMKeeperGenesisReplayEnableActivatesAParkedPackage(t *testing.T) {
	const pkgPath = "gno.land/r/test/replayenable"

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("historical"))
	stranger := crypto.AddressFromPreimage([]byte("rotated-out-approver"))
	env, ctx := inertEnv(t, approver, creator, stranger)

	replayCtx := ctx.WithValue(auth.GenesisReplayKey{}, true)

	require.NoError(t, env.vmk.AddPackage(replayCtx,
		NewMsgAddPackage(creator, pkgPath, replayFiles("replayenable"))))
	gs := env.vmk.getGnoTransactionStore(ctx)
	require.NotNil(t, gs.GetInertPackage(pkgPath), "the replayed add must park")

	// Sent by an approver the fork has since rotated out, so this also covers
	// the approver gate's exemption.
	require.NoError(t, env.vmk.EnablePackage(replayCtx,
		MsgEnablePackage{Approver: stranger, PkgPath: pkgPath}),
		"a fork that rotated pkg_approvers must not refuse the enables in its own history")
	assert.NotNil(t, gs.GetPackage(pkgPath, false),
		"the replayed enable must activate the package it parked")
	assert.Nil(t, gs.GetInertPackage(pkgPath),
		"and clear the parked blob, as an ordinary enable does")
}

// TestVMKeeperGenesisReplayEnableWithNothingParkedIsANoOp covers the other half:
// a replayed enable whose package went live on the ordinary path.
//
// It has to succeed. Every genesis exported before this branch existed looks
// like this: the replayed add deploys live and leaves nothing parked, so the
// matching enable has nothing to do and must not turn a working fork into a
// replay failure.
//
// The no-op is conditional on the package actually being live -- see
// TestVMKeeperGenesisReplayEnableRefusesWhenNothingIsLive for why.
func TestVMKeeperGenesisReplayEnableWithNothingParkedIsANoOp(t *testing.T) {
	const pkgPath = "gno.land/r/test/replaynothing"

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("historical"))
	env, ctx := inertEnv(t, approver, creator)

	// A fork that has moved off "inert": the replayed add deploys live, so the
	// replayed enable finds nothing.
	offInert := DefaultParams()
	offInert.CodeSubmissionPolicy = CodeSubmissionPolicyPermissionless
	offInert.PkgApprovers = []crypto.Address{approver}
	require.NoError(t, env.vmk.SetParams(ctx, offInert))

	replayCtx := ctx.WithValue(auth.GenesisReplayKey{}, true)
	require.NoError(t, env.vmk.AddPackage(replayCtx,
		NewMsgAddPackage(creator, pkgPath, replayFiles("replaynothing"))))
	gs := env.vmk.getGnoTransactionStore(ctx)
	require.Nil(t, gs.GetInertPackage(pkgPath))

	require.NoError(t, env.vmk.EnablePackage(replayCtx,
		MsgEnablePackage{Approver: approver, PkgPath: pkgPath}),
		"a replayed enable with nothing parked must be a no-op, not an error")
	assert.NotNil(t, gs.GetPackage(pkgPath, false), "the package must still be live")

	// Live traffic is unaffected: the same call outside replay still refuses,
	// and for the policy reason -- not because the setup left nothing parked.
	err := env.vmk.EnablePackage(ctx, MsgEnablePackage{Approver: approver, PkgPath: pkgPath})
	require.Error(t, err, "the exemption must be scoped to replay")
	// %+v, not Error(): tm2's abci errors keep the detail on the wrapped trace.
	assert.Contains(t, fmt.Sprintf("%+v", err), "code_submission_policy",
		"it must refuse on the policy, not because nothing was parked")
}

// TestVMKeeperGenesisReplayEnableRefusesWhenNothingIsLive pins the condition on
// the no-op above.
//
// "Nothing parked" has two causes and they need opposite answers. If the package
// is live, the replayed add took the ordinary path and the enable is genuinely
// spare work. If nothing is live either, the replayed add FAILED -- and a
// replayed MsgAddPackage has plenty of ways to fail that its own history did
// not: a creator whose account an earlier diverging tx never created, a
// namespace that changed hands, a prior park by someone else.
//
// Returning nil there records success for a package that is not on the chain.
// The fork boots with the realm missing, and the replay report's last word on
// that path is "enabled OK" -- so the one failure it does show reads like an
// isolated hiccup rather than a missing package.
func TestVMKeeperGenesisReplayEnableRefusesWhenNothingIsLive(t *testing.T) {
	const pkgPath = "gno.land/r/test/replaymissing"

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("historical"))
	env, ctx := inertEnv(t, approver, creator)

	// Stand in for the replayed add having failed: nothing parked, nothing live.
	gs := env.vmk.getGnoTransactionStore(ctx)
	require.Nil(t, gs.GetInertPackage(pkgPath))
	require.Nil(t, gs.GetMemPackage(pkgPath))

	replayCtx := ctx.WithValue(auth.GenesisReplayKey{}, true)
	err := env.vmk.EnablePackage(replayCtx,
		MsgEnablePackage{Approver: approver, PkgPath: pkgPath})
	require.Error(t, err,
		"a replayed enable must not report success for a package that is not on the chain")
	// %+v, not Error(): tm2's abci errors keep the detail on the wrapped trace.
	assert.Contains(t, fmt.Sprintf("%+v", err), "no inert package at path",
		"and it must say what is actually wrong")
}

// TestVMKeeperEnableChecksChainDomain covers the last of AddPackage's path rules
// at enable.
//
// chain_domain is a governance param, so a change between submit and enable
// would otherwise let a parked package go live under a domain AddPackage would
// refuse. Exercised with sys_names_pkgpath empty, which is where the gap is
// reachable: checkNamespacePermission applies the same prefix rule but returns
// early when that param is unset, while AddPackage applies it unconditionally.
func TestVMKeeperEnableChecksChainDomain(t *testing.T) {
	const pkgPath = "gno.land/r/test/domainshift"

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("domaincreator"))
	env, ctx := inertEnv(t, approver, creator)

	params := DefaultParams()
	params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
	params.PkgApprovers = []crypto.Address{approver}
	params.SysNamesPkgPath = ""
	env.vmk.SetParams(ctx, params)

	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, []*std.MemFile{
		{Name: "domainshift.gno", Body: `package domainshift

func Who(cur realm) string { return "parked" }`},
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
	})))

	// Governance moves the chain domain after the package was parked.
	params.ChainDomain = "example.com"
	env.vmk.SetParams(ctx, params)

	err := env.vmk.EnablePackage(ctx, MsgEnablePackage{Approver: approver, PkgPath: pkgPath})
	require.Error(t, err,
		"a parked package must not go live under a domain AddPackage would refuse")
	assert.Contains(t, fmt.Sprintf("%+v", err), "invalid domain")
}

// TestInertSubmissionChargeIsTakenFromTheCreator covers the charge that prices
// the init() an inert submission defers onto the approver.
//
// The approver runs that init() on its own transaction and its own gas meter,
// and fees are flat — so its exposure is its fee times the number of approvals
// it can be induced to make, and it stops approving for everyone once its spend
// limit is reached. Submitting is otherwise nearly free, so provoking those
// approvals is cheap. The charge puts the price on the party that chose the
// cost.
func TestInertSubmissionChargeIsTakenFromTheCreator(t *testing.T) {
	const pkgPath = "gno.land/r/test/charged"
	const charge = int64(3_000_000)

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("submitter"))
	env, ctx := inertEnv(t, approver, creator)

	params := DefaultParams()
	params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
	params.PkgApprovers = []crypto.Address{approver}
	params.InertSubmissionCharge = ugnot.ValueString(charge)
	require.NoError(t, env.vmk.SetParams(ctx, params))
	collector := params.InertChargeCollector

	creatorBefore := env.bankk.GetCoins(ctx, creator).AmountOf(ugnot.Denom)
	collectorBefore := env.bankk.GetCoins(ctx, collector).AmountOf(ugnot.Denom)

	require.NoError(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(creator, pkgPath, replayFiles("charged"))))

	assert.Equal(t, creatorBefore-charge,
		env.bankk.GetCoins(ctx, creator).AmountOf(ugnot.Denom),
		"the creator must pay the charge")
	assert.Equal(t, collectorBefore+charge,
		env.bankk.GetCoins(ctx, collector).AmountOf(ugnot.Denom),
		"and the collector must receive exactly it")
	assert.NotNil(t, env.vmk.getGnoTransactionStore(ctx).GetInertPackage(pkgPath),
		"the package must still park")
}

// TestInertSubmissionChargeIsOffByDefault pins the default.
//
// Empty means off, and it is the only spelling of off: ParseCoins("") yields no
// coins and no error, while "0ugnot" fails Coins.validate. This matters beyond
// tidiness -- empty-by-default is what makes genesis replay correct without a
// carve-out. A default charge would be levied on replayed history that never
// paid one, and a fork's balances would drift from its source chain.
func TestInertSubmissionChargeIsOffByDefault(t *testing.T) {
	const pkgPath = "gno.land/r/test/uncharged"

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("submitter"))
	env, ctx := inertEnv(t, approver, creator)

	require.Empty(t, DefaultParams().InertSubmissionCharge,
		"the shipped default must be off")

	before := env.bankk.GetCoins(ctx, creator).AmountOf(ugnot.Denom)
	require.NoError(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(creator, pkgPath, replayFiles("uncharged"))))
	assert.Equal(t, before, env.bankk.GetCoins(ctx, creator).AmountOf(ugnot.Denom),
		"with the charge off, an inert submit must move no coins at all")
}

// TestInertSubmissionChargeIsNotTakenFromARefusedSubmission pins the placement.
//
// The charge sits last in the branch, after every refusal, so a submission that
// is turned away pays nothing and reports its real reason rather than an
// insufficient-funds error. Everything in the message phase reverts together
// regardless, but the ordering is what makes that legible at the call site
// instead of requiring a reader to know baseapp's revert semantics.
func TestInertSubmissionChargeIsNotTakenFromARefusedSubmission(t *testing.T) {
	const pkgPath = "gno.land/r/test/refused"
	const charge = int64(3_000_000)

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	first := crypto.AddressFromPreimage([]byte("firstsubmitter"))
	second := crypto.AddressFromPreimage([]byte("frontrunner"))
	env, ctx := inertEnv(t, approver, first, second)

	params := DefaultParams()
	params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
	params.PkgApprovers = []crypto.Address{approver}
	params.InertSubmissionCharge = ugnot.ValueString(charge)
	require.NoError(t, env.vmk.SetParams(ctx, params))

	require.NoError(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(first, pkgPath, replayFiles("refused"))))

	// A different creator cannot replace a parked package, and must not be
	// charged for being told so.
	before := env.bankk.GetCoins(ctx, second).AmountOf(ugnot.Denom)
	err := env.vmk.AddPackage(ctx,
		NewMsgAddPackage(second, pkgPath, replayFiles("refused")))
	require.Error(t, err, "a stranger must not overwrite a parked package")
	assert.Equal(t, before, env.bankk.GetCoins(ctx, second).AmountOf(ugnot.Denom),
		"and must not be charged for the refusal")
}

// TestInertSubmissionChargeRespectsATokenLock pins the use of SendCoins rather
// than SendCoinsUnrestricted at the charge.
//
// The charge is a one-way transfer, so a locked account has to be refused. The
// storage deposit next to it uses the unrestricted call because it refunds, and
// copying that here would let the charge spend locked coins.
func TestInertSubmissionChargeRespectsATokenLock(t *testing.T) {
	const pkgPath = "gno.land/r/test/locked"
	const charge = int64(3_000_000)

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("lockedsubmitter"))
	env, ctx := inertEnv(t, approver, creator)

	params := DefaultParams()
	params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
	params.PkgApprovers = []crypto.Address{approver}
	params.InertSubmissionCharge = ugnot.ValueString(charge)
	require.NoError(t, env.vmk.SetParams(ctx, params))

	// Lock ugnot. The creator is a plain BaseAccount, which does not implement
	// AccountUnrestricter, so it is not on the token-lock whitelist.
	env.bankk.SetRestrictedDenoms(ctx, []string{ugnot.Denom})

	before := env.bankk.GetCoins(ctx, creator).AmountOf(ugnot.Denom)
	err := env.vmk.AddPackage(ctx,
		NewMsgAddPackage(creator, pkgPath, replayFiles("locked")))
	require.Error(t, err, "a locked account must not be able to pay the charge")
	// The charge's own wording reaches the ABCI log, not Error(), which returns
	// a bare "insufficient coins error". Checking the log stops this passing on
	// some unrelated refusal, and pins the surface a creator actually reads.
	require.Contains(t, sdk.ABCIResultFromError(err).Log, "submission charge")
	assert.Equal(t, before, env.bankk.GetCoins(ctx, creator).AmountOf(ugnot.Denom),
		"and must keep its coins")
	assert.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetInertPackage(pkgPath),
		"and the package must not park, because the charge gates it")
}

// TestInertSubmissionChargeValidation covers the bounds on the param.
//
// The ceiling is the point: a charge governance can raise without limit is a
// deploy freeze, which is the outcome the charge exists to prevent.
func TestInertSubmissionChargeValidation(t *testing.T) {
	for _, tc := range []struct {
		name   string
		charge string
		ok     bool
	}{
		{"empty is off", "", true},
		{"a plain ugnot amount", "3000000ugnot", true},
		{"at the ceiling", ugnot.ValueString(maxInertSubmissionCharge), true},
		{"above the ceiling", ugnot.ValueString(maxInertSubmissionCharge + 1), false},
		{"zero is not a second spelling of off", "0ugnot", false},
		{"a foreign denom", "5foo", false},
		{"more than one coin", "1000ugnot,5foo", false},
		{"not a coin at all", "banana", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			params := DefaultParams()
			params.InertSubmissionCharge = tc.charge
			err := params.Validate()
			if tc.ok {
				assert.NoError(t, err)
				return
			}
			assert.Error(t, err)
		})
	}

	t.Run("a zero collector validates, and is defaulted before it is read", func(t *testing.T) {
		// Validate deliberately says nothing about the collector. An
		// unconditional non-zero rule breaks `gnogenesis fork generate`, which
		// builds Params without applyLegacyDefaults and would then refuse to
		// produce a fork genesis. A cross-field rule ("charge set implies
		// collector set") would abort a governance proposal mid-execution,
		// because WillSetParam re-validates the whole struct and panics while
		// r/sys/params sets one key per proposal.
		//
		// applyLegacyDefaults is what makes that safe: the keeper's one read path
		// never sees a zero collector. The guard at the charge itself is
		// unreachable defence, kept so a misconfiguration would skip the charge
		// rather than burn it at the zero address.
		params := DefaultParams()
		params.InertChargeCollector = crypto.Address{}
		params.InertSubmissionCharge = "3000000ugnot"
		require.NoError(t, params.Validate(),
			"an unconditional collector rule would break the fork tool")
		assert.False(t, params.ApplyLegacyDefaults().InertChargeCollector.IsZero(),
			"and the read path must supply one regardless")
	})

	t.Run("a legacy params blob still validates", func(t *testing.T) {
		// A blob written before these fields existed reads the collector as
		// zero. Validate rejects that, and WillSetParam re-validates the whole
		// struct and PANICS -- so without the legacy default every unrelated
		// governance param update would abort on such a chain.
		legacy := DefaultParams()
		legacy.InertChargeCollector = crypto.Address{}
		assert.NoError(t, legacy.ApplyLegacyDefaults().Validate(),
			"legacy state must keep validating")
		assert.Empty(t, legacy.ApplyLegacyDefaults().InertSubmissionCharge,
			"but the charge itself must stay off, or replay levies it on history that never paid")
	})
}

// TestChainDomainAccessorsAgree pins that the two ways of reading the chain
// domain never disagree.
//
// AddPackage reads it through getChainDomainParam and EnablePackage through the
// same accessor; they are the two halves of one deploy applying one rule, so a
// divergence means a package AddPackage accepted becomes one EnablePackage
// refuses. The accessors default differently on purpose-free grounds:
// getChainDomainParam seeds "gno.land" and lets GetString overwrite it, while
// GetStruct leaves the field at "" for an absent key and applyLegacyDefaults --
// which does fill PreprocessGasPerByte and CodeSubmissionPolicy -- does not fill
// ChainDomain. Reading params.ChainDomain here would therefore test
// HasPrefix(path, "/") and refuse everything.
//
// Asserted across the states a chain can actually be in rather than by
// constructing an absent key directly, since SetParams always writes every
// field. If a future field is added to Params without a legacy default, this
// fails as soon as any of these states stops round-tripping.
func TestChainDomainAccessorsAgree(t *testing.T) {
	for _, tt := range []struct {
		name   string
		domain string
	}{
		{"the shipped default", chainDomainDefault},
		{"a custom domain", "example.com"},
		{"explicitly empty", ""},
	} {
		t.Run(tt.name, func(t *testing.T) {
			env := setupTestEnv()
			ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

			p := DefaultParams()
			p.ChainDomain = tt.domain
			require.NoError(t, p.Validate())
			require.NoError(t, env.vmk.SetParams(ctx, p))

			assert.Equal(t, env.vmk.GetParams(ctx).ChainDomain, env.vmk.getChainDomainParam(ctx),
				"GetParams().ChainDomain and getChainDomainParam() must agree; "+
					"EnablePackage and AddPackage apply the same domain rule")
		})
	}
}

// TestVMKeeperEnableTypeChecksProductionFilesOnly pins ProdOnly on the enable
// type check.
//
// A parked blob is stored exactly as submitted, test files included, and enable
// type-checks it. Without ProdOnly it would check the _test.gno files too, and
// resolving their test-stdlib imports needs a getter the consensus path does not
// have -- so the answer would depend on what the node has on disk rather than on
// what the chain agreed. That is a consensus split, and it fails silently: nodes
// that agree still agree, so nothing shows up until one does not.
//
// AddPackage takes the same option for the same reason. This test exists because
// mutating ProdOnly to false broke nothing else in the suite.
func TestVMKeeperEnableTypeChecksProductionFilesOnly(t *testing.T) {
	const pkgPath = "gno.land/r/test/prodonly"

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("submitter"))
	env, ctx := inertEnv(t, approver, creator)

	// The production file compiles. The test file does not: it calls something
	// that does not exist, so type-checking it is an error.
	files := sortMemFiles([]*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(pkgPath)},
		{Name: "prodonly.gno", Body: `package prodonly

func Who(cur realm) string { return "live" }`},
		{Name: "prodonly_test.gno", Body: `package prodonly

func BrokenIfTypeChecked() { thisSymbolDoesNotExist() }`},
	})

	require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, pkgPath, files)))
	gs := env.vmk.getGnoTransactionStore(ctx)
	parked := gs.GetInertPackage(pkgPath)
	require.NotNil(t, parked)
	require.Len(t, parked.Files, 3,
		"the parked blob keeps the test file, which is what makes ProdOnly matter")

	require.NoError(t, env.vmk.EnablePackage(ctx,
		MsgEnablePackage{Approver: approver, PkgPath: pkgPath}),
		"enable must type-check production files only; a broken _test.gno is not its business")
	assert.NotNil(t, gs.GetPackage(pkgPath, false), "and the package must go live")
}

// TestInertSubmissionChargeNamesItselfWhenUnaffordable covers the refusal a
// creator who cannot pay actually sees.
//
// The bank's own error talks about funds and not about why the amount is being
// asked for, so on its own it reads as a wrong gas fee. That matters more than
// usual here: the charge defaults to off, so the first person to meet it is
// someone whose chain just turned it on by governance vote, with nothing in the
// message to connect the two.
func TestInertSubmissionChargeNamesItselfWhenUnaffordable(t *testing.T) {
	const pkgPath = "gno.land/r/test/toopoor"

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("brokecreator"))
	env, ctx := inertEnv(t, approver, creator)

	// Above the funded balance, so the transfer cannot succeed.
	broke := env.bankk.GetCoins(ctx, creator).AmountOf(ugnot.Denom)
	params := DefaultParams()
	params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
	params.PkgApprovers = []crypto.Address{approver}
	params.InertSubmissionCharge = ugnot.ValueString(broke + 1)
	require.NoError(t, env.vmk.SetParams(ctx, params))

	err := env.vmk.AddPackage(ctx,
		NewMsgAddPackage(creator, pkgPath, replayFiles("toopoor")))
	require.Error(t, err)
	// %+v, not Error(): tm2's abci errors keep the detail on the wrapped trace.
	detail := fmt.Sprintf("%+v", err)
	assert.Contains(t, detail, "submission charge",
		"the refusal must name the charge, not just report missing funds")
	assert.Contains(t, detail, string(CodeSubmissionPolicyInert),
		"and name the policy that requires it, so the cause is findable")

	assert.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetInertPackage(pkgPath),
		"and nothing may be parked when the charge was not paid")
}

// TestQueryPackageMetaTellsTheThreeStatesApart covers vm/qpkgmeta_json.
//
// The point of the query is the distinction, so all three states are asserted
// against the same keeper: a parked package that reads as "absent" would leave
// its creator exactly where they were before the query existed.
func TestQueryPackageMetaTellsTheThreeStatesApart(t *testing.T) {
	const parkedPath = "gno.land/r/test/parked"
	const livePath = "gno.land/r/test/wentlive"

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("submitter"))
	env, ctx := inertEnv(t, approver, creator)

	decode := func(t *testing.T, path string) PackageMeta {
		t.Helper()
		raw, err := env.vmk.QueryPackageMeta(ctx, path)
		require.NoError(t, err, "an unknown path must not be an error")
		var got PackageMeta
		require.NoError(t, json.Unmarshal([]byte(raw), &got))
		assert.Equal(t, path, got.Path)
		return got
	}

	t.Run("absent", func(t *testing.T) {
		got := decode(t, "gno.land/r/test/neversubmitted")
		assert.Equal(t, PackageStatusAbsent, got.Status)
		assert.Empty(t, got.Creator, "there is nothing to report about a path nobody used")
	})

	t.Run("inert", func(t *testing.T) {
		require.NoError(t, env.vmk.AddPackage(ctx,
			NewMsgAddPackage(creator, parkedPath, replayFiles("parked"))))

		got := decode(t, parkedPath)
		assert.Equal(t, PackageStatusInert, got.Status,
			"a submitted package must not read as absent; that is the blind spot this closes")
		assert.Equal(t, creator.String(), got.Creator,
			"and it must name the creator, who pays the deposit at enable")
	})

	t.Run("a redeploy parked over a live private realm is reported pending", func(t *testing.T) {
		// AddPackage refuses to park over a live PUBLIC package, so this is the
		// only way the two key spaces hold the same path at once. Without the
		// second lookup the status reads plain "live" and the creator cannot
		// see that their redeploy is waiting -- the blind spot this query
		// exists to close, reappearing one case further in.
		const privPath = "gno.land/r/test/privpending"
		priv := func() []*std.MemFile {
			return sortMemFiles([]*std.MemFile{
				{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(privPath) + "\nprivate = true\n"},
				{Name: "privpending.gno", Body: "package privpending\n\nfunc Who(cur realm) string { return \"live\" }"},
			})
		}

		require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, privPath, priv())))
		require.NoError(t, env.vmk.EnablePackage(ctx,
			MsgEnablePackage{Approver: approver, PkgPath: privPath}))
		live := decode(t, privPath)
		require.Equal(t, PackageStatusLive, live.Status, "premise: live before the redeploy")
		require.False(t, live.Pending, "premise: nothing pending yet")

		require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, privPath, priv())),
			"a private realm may be parked over while live; that is the redeploy path")

		got := decode(t, privPath)
		assert.Equal(t, PackageStatusLive, got.Status, "the old code is still what runs")
		assert.True(t, got.Pending, "but the parked redeploy must be visible")
	})

	t.Run("live after enable", func(t *testing.T) {
		require.NoError(t, env.vmk.AddPackage(ctx,
			NewMsgAddPackage(creator, livePath, replayFiles("wentlive"))))
		require.Equal(t, PackageStatusInert, decode(t, livePath).Status, "premise: parked first")

		require.NoError(t, env.vmk.EnablePackage(ctx,
			MsgEnablePackage{Approver: approver, PkgPath: livePath}))

		got := decode(t, livePath)
		assert.Equal(t, PackageStatusLive, got.Status,
			"enabling must flip the status, or the query cannot show progress")
		assert.Equal(t, creator.String(), got.Creator,
			"the creator survives activation; the approver never becomes the owner")
	})
}
