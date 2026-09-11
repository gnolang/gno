package vm

import (
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gnovm/pkg/gnomod"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk/auth"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// privateFilesFor builds a private realm at pkgPath whose init() records the
// address that deployed it, plus a ledger value the caller picks so one version
// can be told from another through a real MsgCall.
//
// Built fresh per submission: stampGnomod rewrites gnomod.toml in place through
// SetFile, so a shared slice would carry one submitter's stamped creator into
// the next.
func privateFilesFor(pkgPath, name, ledger string) []*std.MemFile {
	return []*std.MemFile{
		// Name-sorted: "gnomod.toml" before any name starting past "g".
		{Name: "gnomod.toml", Body: `module = "` + pkgPath + `"
gno = "0.9"
private = true`},
		{Name: name + ".gno", Body: `package ` + name + `

import "chain/runtime/unsafe"

var (
	owner  address
	ledger = "` + ledger + `"
)

func init() { owner = unsafe.OriginCaller() }

func Owner(cur realm) string         { return owner.String() }
func Ledger(cur realm) string        { return ledger }
func Record(cur realm, entry string) { ledger = entry }
`},
	}
}

// storedCreator reads back the creator the keeper stamped into the package
// stored at pkgPath, which is the field that decides who init() runs as and who
// pays the storage deposit.
func storedCreator(t *testing.T, mpkg *std.MemPackage) string {
	t.Helper()
	require.NotNil(t, mpkg)
	gm, err := gnomod.ParseMemPackage(mpkg)
	require.NoError(t, err)
	return gm.AddPkg.Creator
}

// TestVMKeeperPrivateRedeployIsCreatorBound pins that only the address that
// deployed a live PRIVATE realm may replace it, on the ordinary path: one
// MsgAddPackage under the default permissionless policy, with no approver and
// nothing parked.
//
// The already-exists refusal waives itself for a private package
// (`pv != nil && !pv.Private`), which asks what the live package is and never
// who it belongs to, and the creator-bound guard beside it reads
// GetInertPackage, so it covers bytes that are still PARKED and never a live
// package. A stranger's submission therefore strips the live source, re-runs
// init() with OriginCaller set to them -- which is what p/nt/ownable records as
// owner -- and has stampGnomod record them as creator, while the realm keeps
// its address, its coins and its callable surface.
//
// checkNamespacePermission is not that binding: it returns nil while
// r/sys/names is undeployed or not Enable()d, and disabled is that realm's
// compiled-in initial value, so it is the state this repo's genesis produces.
func TestVMKeeperPrivateRedeployIsCreatorBound(t *testing.T) {
	const (
		pkgPath = "gno.land/r/test/privowned"
		name    = "privowned"
	)

	owner := crypto.AddressFromPreimage([]byte("privownedowner"))
	stranger := crypto.AddressFromPreimage([]byte("privownedstranger"))
	// The reads below go through a real MsgCall signed by an unrelated funded
	// address, so the owner's and the stranger's balances stay readable.
	reader := crypto.AddressFromPreimage([]byte("privownedreader"))

	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	for _, addr := range []crypto.Address{owner, stranger, reader} {
		acc := env.acck.NewAccountWithAddress(ctx, addr)
		env.acck.SetAccount(ctx, acc)
		require.NoError(t, env.bankk.SetCoins(ctx, addr, initialBalance))
	}

	// Pinned so the test cannot be read as being about the "inert" policy: if
	// the shipped default ever moves, this says so instead of the test silently
	// changing subject.
	params := env.vmk.GetParams(ctx)
	require.Equal(t, CodeSubmissionPolicyPermissionless, params.CodeSubmissionPolicy)
	require.Empty(t, params.PkgApprovers, "no approver takes part in this redeploy")

	read := func(fn string) string {
		t.Helper()
		res, err := env.vmk.Call(ctx, NewMsgCall(reader, nil, pkgPath, fn, nil))
		require.NoError(t, err)
		return res
	}
	live := func() *std.MemPackage {
		return env.vmk.getGnoTransactionStore(ctx).GetMemPackage(pkgPath)
	}
	quoted := func(s string) string { return fmt.Sprintf("(%q string)\n\n", s) }

	// The owner deploys their private realm and puts state in it.
	require.NoError(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(owner, pkgPath, privateFilesFor(pkgPath, name, "v1"))))
	_, err := env.vmk.Call(ctx, NewMsgCall(owner, nil, pkgPath, "Record", []string{"owner entry"}))
	require.NoError(t, err)

	require.Equal(t, owner.String(), storedCreator(t, live()))
	require.Equal(t, quoted(owner.String()), read("Owner"),
		"precondition: init() recorded the owner")
	require.Equal(t, quoted("owner entry"), read("Ledger"),
		"precondition: the realm holds state written after the deploy")

	t.Run("a stranger cannot redeploy over a live private realm", func(t *testing.T) {
		err := env.vmk.AddPackage(ctx,
			NewMsgAddPackage(stranger, pkgPath, privateFilesFor(pkgPath, name, "seized")))
		require.Error(t, err, "a third party must not be able to replace a live private realm")
		assert.True(t, errors.Is(err, PkgExistError{}))
		// The refusal has to name the live creator, as the parked-blob one
		// does: Error() alone is the generic "package already exists" that a
		// public path returns too, and the detail rides the wrapped trace into
		// the ABCI log.
		assert.Contains(t, fmt.Sprintf("%+v", err), owner.String(),
			"the refusal must name the address the realm belongs to")

		assert.Equal(t, owner.String(), storedCreator(t, live()),
			"the live package must still record the owner as creator")
		assert.Equal(t, quoted(owner.String()), read("Owner"),
			"init() must not re-run with the stranger as OriginCaller")
		assert.Equal(t, quoted("owner entry"), read("Ledger"),
			"the state the owner wrote must stay reachable")
		assert.Contains(t, live().GetFile(name+".gno").Body, `ledger = "v1"`,
			"qfile must keep serving the owner's source at the owner's path")
	})

	t.Run("the owner may still redeploy", func(t *testing.T) {
		require.NoError(t, env.vmk.AddPackage(ctx,
			NewMsgAddPackage(owner, pkgPath, privateFilesFor(pkgPath, name, "v2"))),
			"replacing one's own private realm is what the exemption is for")

		assert.Contains(t, live().GetFile(name+".gno").Body, `ledger = "v2"`,
			"the redeployed source must be the one that is stored")
		assert.Equal(t, owner.String(), storedCreator(t, live()))
		assert.Equal(t, quoted(owner.String()), read("Owner"),
			"init() ran again, as the owner")
	})
}

// TestVMKeeperInertPrivateRedeployIsCreatorBound pins the same binding on the
// inert path, where the takeover needs no policy change of its own: parking is
// permissionless, and PackageContentHash resets the [addpkg] section the keeper
// stamps -- creator included -- so byte-identical source parked by a DIFFERENT
// address yields the digest the approver already signed off on for the owner.
// (MsgEnablePackage.PkgHeight can pin the submission, but it is optional, and an
// approval that does not carry it is exactly this case.)
//
// EnablePackage ends with DelInertPackage, so after every successful enable the
// parked slot is empty and the creator-bound guard over PARKED bytes has
// nothing to compare against.
func TestVMKeeperInertPrivateRedeployIsCreatorBound(t *testing.T) {
	const (
		pkgPath = "gno.land/r/test/privinert"
		name    = "privinert"
	)

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	alice := crypto.AddressFromPreimage([]byte("privinertalice"))
	bob := crypto.AddressFromPreimage([]byte("privinertbob"))
	env, ctx := inertEnv(t, approver, alice, bob)

	// Alice parks her private realm and the approver activates it.
	require.NoError(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(alice, pkgPath, privateFilesFor(pkgPath, name, "v1"))))
	require.NoError(t, env.vmk.EnablePackage(ctx,
		approvalFor(t, env, ctx, approver, pkgPath)))
	require.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetInertPackage(pkgPath),
		"precondition: the enable emptied the parked slot")
	require.Equal(t, alice.String(),
		storedCreator(t, env.vmk.getGnoTransactionStore(ctx).GetMemPackage(pkgPath)))

	t.Run("a stranger cannot park over a live private realm", func(t *testing.T) {
		err := env.vmk.AddPackage(ctx,
			NewMsgAddPackage(bob, pkgPath, privateFilesFor(pkgPath, name, "seized")))
		require.Error(t, err, "a third party must not be able to queue a replacement")
		assert.True(t, errors.Is(err, PkgExistError{}))
		assert.Contains(t, fmt.Sprintf("%+v", err), alice.String(),
			"the refusal must name the address the realm belongs to")

		assert.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetInertPackage(pkgPath),
			"nothing may be left parked for an approver to activate")
	})

	t.Run("the owner may still park and activate a replacement", func(t *testing.T) {
		require.NoError(t, env.vmk.AddPackage(ctx,
			NewMsgAddPackage(alice, pkgPath, privateFilesFor(pkgPath, name, "v2"))))
		require.NoError(t, env.vmk.EnablePackage(ctx,
			approvalFor(t, env, ctx, approver, pkgPath)))

		stored := env.vmk.getGnoTransactionStore(ctx).GetMemPackage(pkgPath)
		require.NotNil(t, stored)
		assert.Contains(t, stored.GetFile(name+".gno").Body, `ledger = "v2"`)
		assert.Equal(t, alice.String(), storedCreator(t, stored))
	})
}

// TestVMKeeperEnableCannotRedeployAnotherAddressesPrivateRealm pins the binding
// on the branch AddPackage cannot cover: a package parked BEFORE anything was
// live at the path, activated after somebody else deployed there.
//
// The parked and the live blob are both private, so the private-override rule
// passes and the creator binding is the only thing that refuses. EnablePackage
// parses the live gnomod to decide whether the path may be replaced at all, and
// reads the parked one for the identity init() runs as.
func TestVMKeeperEnableCannotRedeployAnotherAddressesPrivateRealm(t *testing.T) {
	const (
		pkgPath = "gno.land/r/test/privstale"
		name    = "privstale"
	)

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	stranger := crypto.AddressFromPreimage([]byte("privstaleparker"))
	owner := crypto.AddressFromPreimage([]byte("privstaleowner"))
	env, ctx := inertEnv(t, approver, stranger, owner)

	// 1. Nothing is live, so nothing binds the park: the stranger parks a
	//    private package and it is never approved.
	require.NoError(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(stranger, pkgPath, privateFilesFor(pkgPath, name, "seized"))))

	// 2. Governance opens the chain and the owner deploys a private realm at
	//    that path through the now-open ordinary path.
	params := DefaultParams()
	params.CodeSubmissionPolicy = CodeSubmissionPolicyPermissionless
	params.PkgApprovers = []crypto.Address{approver}
	env.vmk.SetParams(ctx, params)
	require.NoError(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(owner, pkgPath, privateFilesFor(pkgPath, name, "v1"))))

	// 3. Back to inert, so the refusal below is the creator binding and not
	//    EnablePackage's policy check.
	params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
	env.vmk.SetParams(ctx, params)

	err := env.vmk.EnablePackage(ctx, approvalFor(t, env, ctx, approver, pkgPath))
	require.Error(t, err, "an approver's routine enable must not hand over a live private realm")
	assert.True(t, errors.Is(err, PkgExistError{}))
	assert.Contains(t, fmt.Sprintf("%+v", err), owner.String(),
		"the refusal must name the address the realm belongs to")

	stored := env.vmk.getGnoTransactionStore(ctx).GetMemPackage(pkgPath)
	require.NotNil(t, stored)
	assert.Contains(t, stored.GetFile(name+".gno").Body, `ledger = "v1"`,
		"the owner's live source must survive the enable")
	assert.Equal(t, owner.String(), storedCreator(t, stored))
}

// TestVMKeeperPrivateRedeployReplaysAtGenesis pins that the binding is waived
// for every transaction InitChain delivers, on both call sites.
//
// EnablePackage already waives its policy, approver and pkg_hash gates on
// replay, on the rule that a fork reproduces a record rather than granting it
// again. The creator binding is a fourth gate of the same kind, and enforcing it
// during replay would mean a chain that took a cross-address private redeploy
// under the old binary can no longer replay its own history: the boot aborts
// under the default PanicOnFailingTxResultHandler, or comes up silently diverged
// under -skip-failing-genesis-txs. It would also refuse the hardfork migration
// path, whose transaction is stamped with `gnogenesis fork addpkg --deployer`
// rather than the realm's original creator.
func TestVMKeeperPrivateRedeployReplaysAtGenesis(t *testing.T) {
	approver := crypto.AddressFromPreimage([]byte("oracle"))
	owner := crypto.AddressFromPreimage([]byte("replayowner"))
	stranger := crypto.AddressFromPreimage([]byte("replaystranger"))

	t.Run("AddPackage", func(t *testing.T) {
		const (
			pkgPath = "gno.land/r/test/zreplayadd"
			name    = "zreplayadd"
		)
		env, ctx := inertEnv(t, approver, owner, stranger)
		params := DefaultParams()
		params.CodeSubmissionPolicy = CodeSubmissionPolicyPermissionless
		env.vmk.SetParams(ctx, params)
		// InitChain marks every transaction it delivers, including a fresh
		// chain's own genesis txs.
		ctx = ctx.WithValue(auth.GenesisReplayKey{}, true)
		require.True(t, auth.IsGenesisReplay(ctx))

		require.NoError(t, env.vmk.AddPackage(ctx,
			NewMsgAddPackage(owner, pkgPath, privateFilesFor(pkgPath, name, "v1"))))
		require.NoError(t, env.vmk.AddPackage(ctx,
			NewMsgAddPackage(stranger, pkgPath, privateFilesFor(pkgPath, name, "v2"))),
			"replay must reproduce a redeploy the source chain accepted")

		stored := env.vmk.getGnoTransactionStore(ctx).GetMemPackage(pkgPath)
		assert.Contains(t, stored.GetFile(name+".gno").Body, `ledger = "v2"`,
			"the replayed source is the one that ends up stored")
		assert.Equal(t, stranger.String(), storedCreator(t, stored))
	})

	t.Run("EnablePackage", func(t *testing.T) {
		const (
			pkgPath = "gno.land/r/test/zreplayenable"
			name    = "zreplayenable"
		)
		env, ctx := inertEnv(t, approver, owner, stranger)

		// The stranger parks while nothing is live, then the owner deploys
		// there through the ordinary path -- the state AddPackage's binding
		// cannot see.
		require.NoError(t, env.vmk.AddPackage(ctx,
			NewMsgAddPackage(stranger, pkgPath, privateFilesFor(pkgPath, name, "v2"))))
		params := DefaultParams()
		params.CodeSubmissionPolicy = CodeSubmissionPolicyPermissionless
		params.PkgApprovers = []crypto.Address{approver}
		env.vmk.SetParams(ctx, params)
		require.NoError(t, env.vmk.AddPackage(ctx,
			NewMsgAddPackage(owner, pkgPath, privateFilesFor(pkgPath, name, "v1"))))

		approval := approvalFor(t, env, ctx, approver, pkgPath)
		ctx = ctx.WithValue(auth.GenesisReplayKey{}, true)
		require.NoError(t, env.vmk.EnablePackage(ctx, approval),
			"replay must reproduce an enable the source chain accepted")

		stored := env.vmk.getGnoTransactionStore(ctx).GetMemPackage(pkgPath)
		assert.Equal(t, stranger.String(), storedCreator(t, stored))
	})
}

// TestVMKeeperRejectUnjamsASquattedPrivatePath pins that the owner of a live
// private realm can clear a parked blob squatting their path.
//
// Reachable state: a stranger parks while nothing is live, the policy moves to
// permissionless, the owner deploys a private realm there, and the policy
// returns to inert. That blob can never be enabled -- which is what
// TestVMKeeperEnableCannotRedeployAnotherAddressesPrivateRealm pins -- but
// AddPackage's parked-blob guard still names its submitter, so it blocks the
// OWNER's redeploys for as long as it sits there. Before the binding an
// approver's enable consumed it, which was the takeover; the owner needs a move
// of their own or the realm is frozen.
func TestVMKeeperRejectUnjamsASquattedPrivatePath(t *testing.T) {
	const (
		pkgPath = "gno.land/r/test/zsquatted"
		name    = "zsquatted"
	)

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	stranger := crypto.AddressFromPreimage([]byte("zsquattedstranger"))
	owner := crypto.AddressFromPreimage([]byte("zsquattedowner"))
	bystander := crypto.AddressFromPreimage([]byte("zsquattedbystander"))
	env, ctx := inertEnv(t, approver, stranger, owner, bystander)

	require.NoError(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(stranger, pkgPath, privateFilesFor(pkgPath, name, "squat"))))

	params := DefaultParams()
	params.CodeSubmissionPolicy = CodeSubmissionPolicyPermissionless
	params.PkgApprovers = []crypto.Address{approver}
	env.vmk.SetParams(ctx, params)
	require.NoError(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(owner, pkgPath, privateFilesFor(pkgPath, name, "v1"))))
	params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
	env.vmk.SetParams(ctx, params)

	// Precondition: the squat blocks the owner's own redeploy.
	require.Error(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(owner, pkgPath, privateFilesFor(pkgPath, name, "v2"))),
		"precondition: the parked blob blocks the owner")

	// A bystander still has no standing.
	require.Error(t, env.vmk.RejectPackage(ctx,
		MsgRejectPackage{Sender: bystander, PkgPath: pkgPath}),
		"standing is not open to anyone who asks")
	require.NotNil(t, env.vmk.getGnoTransactionStore(ctx).GetInertPackage(pkgPath))

	// The live owner may clear it, and can then ship their own upgrade.
	require.NoError(t, env.vmk.RejectPackage(ctx,
		MsgRejectPackage{Sender: owner, PkgPath: pkgPath}),
		"the owner of the live package may clear a blob squatting their path")
	assert.Nil(t, env.vmk.getGnoTransactionStore(ctx).GetInertPackage(pkgPath))
	require.NoError(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(owner, pkgPath, privateFilesFor(pkgPath, name, "v2"))),
		"with the squat cleared the owner is no longer frozen")
}

// TestQueryPackageMetaFlagsAnUnenablableRedeploy pins that vm/qpkgmeta_json
// tells "an approver has yet to get to it" from "no approver ever can".
//
// enableBlockedReason answers from params alone and so cannot see this gate; an
// oracle polling a pending submission would otherwise pay a flat fee per attempt
// on a refusal the chain can predict.
func TestQueryPackageMetaFlagsAnUnenablableRedeploy(t *testing.T) {
	const (
		pkgPath = "gno.land/r/test/zmetajam"
		name    = "zmetajam"
	)

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	stranger := crypto.AddressFromPreimage([]byte("zmetajamstranger"))
	owner := crypto.AddressFromPreimage([]byte("zmetajamowner"))
	env, ctx := inertEnv(t, approver, stranger, owner)

	meta := func(t *testing.T) PackageMeta {
		t.Helper()
		raw, err := env.vmk.QueryPackageMeta(ctx, pkgPath)
		require.NoError(t, err)
		var got PackageMeta
		require.NoError(t, json.Unmarshal([]byte(raw), &got))
		return got
	}

	require.NoError(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(stranger, pkgPath, privateFilesFor(pkgPath, name, "squat"))))
	require.Equal(t, ReasonAwaitingApprover, meta(t).Reason,
		"precondition: nothing is live yet, so this one is genuinely enablable")

	params := DefaultParams()
	params.CodeSubmissionPolicy = CodeSubmissionPolicyPermissionless
	params.PkgApprovers = []crypto.Address{approver}
	env.vmk.SetParams(ctx, params)
	require.NoError(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(owner, pkgPath, privateFilesFor(pkgPath, name, "v1"))))

	got := meta(t)
	assert.Equal(t, PackageStatusLive, got.Status)
	assert.True(t, got.Pending, "the stranger's blob is still parked")
	assert.Equal(t, ReasonOwnerMismatch, got.Reason,
		"a pending submission no approver can enable must say so")

	// The owner's own parked redeploy is enablable, so it carries no reason.
	require.NoError(t, env.vmk.RejectPackage(ctx,
		MsgRejectPackage{Sender: owner, PkgPath: pkgPath}))
	params.CodeSubmissionPolicy = CodeSubmissionPolicyInert
	env.vmk.SetParams(ctx, params)
	require.NoError(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(owner, pkgPath, privateFilesFor(pkgPath, name, "v2"))))
	got = meta(t)
	assert.True(t, got.Pending)
	assert.Empty(t, got.Reason, "the owner's own pending redeploy is not blocked")
}
