package vm

import (
	"encoding/hex"
	"path"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// Re-deploying over a live PRIVATE realm, on both paths that allow it.
//
// AddPackage refuses to re-add over a live PUBLIC package but permits it for a
// private one, and that guard sits above every code-submission-policy branch,
// so the sequence below is reachable on the stock params an ordinary creator
// faces as well as through the "inert" policy's EnablePackage.

// redeployV1 allocates enough package-level state to push the realm's object
// clock well past what redeployV2 needs, and aliases two of those objects so
// they escape (refcount >= 2) and each get a hash in the merkleized store.
const redeployV1 = `type node struct{ N int }

var (
	first, aliasOfFirst   *node
	second, aliasOfSecond *node
	third                 *node
	fourth                *node
	fifth                 *node
)

func init() {
	a := &node{N: 1}
	first, aliasOfFirst = a, a
	b := &node{N: 2}
	second, aliasOfSecond = b, b
	third = &node{N: 3}
	fourth = &node{N: 4}
	fifth = &node{N: 5}
}

func Which() string { return "v1" }`

// redeployV2 is strictly smaller than redeployV1, so a realm clock that
// restarts lands its high-water mark below the previous deployment's.
const redeployV2 = `func Which() string { return "v2" }`

// privateRealmFiles builds a private realm at pkgPath whose package and source
// file are named after the last path segment. Files come back sorted by name,
// which AddPackage requires.
func privateRealmFiles(pkgPath, body string) []*std.MemFile {
	name := path.Base(pkgPath)
	files := []*std.MemFile{
		{Name: "gnomod.toml", Body: gno.GenGnoModLatest(pkgPath) + "\nprivate = true"},
		{Name: name + ".gno", Body: "package " + name + "\n\n" + body + "\n"},
	}
	slices.SortFunc(files, func(a, b *std.MemFile) int { return strings.Compare(a.Name, b.Name) })
	return files
}

// escapedObjectHashes reads the merkleized ("main") store's escaped-object
// hashes for pkgPath, keyed by the ObjectID's NewTime. Those keys are written
// under the bare ObjectID string, with no prefix, and they are the only realm
// state that reaches the app hash.
func escapedObjectHashes(t *testing.T, ctx sdk.Context, env testEnv, pkgPath string) map[uint64]string {
	t.Helper()
	pkgID := gno.PkgIDFromPkgPath(pkgPath)
	prefix := hex.EncodeToString(pkgID.Hashlet[:]) + ":"
	out := map[uint64]string{}
	it := ctx.Store(env.vmk.iavlKey).Iterator(nil, nil, nil)
	defer it.Close()
	for ; it.Valid(); it.Next() {
		rest, ok := strings.CutPrefix(string(it.Key()), prefix)
		if !ok {
			continue
		}
		newTime, err := strconv.ParseUint(rest, 10, 64)
		require.NoError(t, err, "unexpected key shape under the ObjectID prefix")
		out[newTime] = hex.EncodeToString(it.Value())
	}
	return out
}

// A realm's object clock is monotonic across a re-deploy on the ordinary
// AddPackage path.
//
// An ObjectID is {PkgID, NewTime} and NewTime comes off the realm record's
// counter, so a counter that restarts hands the second deployment's objects the
// ids the first deployment's objects hold. Two objects then share one ObjectID
// at different times, which breaks the stable tx-stamped NewTime a finalized
// object is guaranteed.
func TestAddPackageRedeployKeepsTheRealmObjectClock(t *testing.T) {
	const pkgPath = "gno.land/r/test/redeployclock"
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	creator := crypto.AddressFromPreimage([]byte("redeploy-clock"))
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, creator))
	require.NoError(t, env.bankk.SetCoins(ctx, creator, initialBalance))

	require.NoError(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(creator, pkgPath, privateRealmFiles(pkgPath, redeployV1))))
	env.vmk.CommitGnoTransactionStore(ctx)

	first := env.vmk.getGnoTransactionStore(ctx).GetPackageRealm(pkgPath)
	require.NotNil(t, first, "premise: the first deploy must have written a realm record")
	require.NotZero(t, first.Time, "premise: the first deploy must have minted object ids")
	firstTime := first.Time

	require.NoError(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(creator, pkgPath, privateRealmFiles(pkgPath, redeployV2))))
	env.vmk.CommitGnoTransactionStore(ctx)

	second := env.vmk.getGnoTransactionStore(ctx).GetPackageRealm(pkgPath)
	require.NotNil(t, second)
	assert.GreaterOrEqual(t, second.Time, firstTime,
		"the re-deploy restarted the realm's object-id counter at %d, below the %d ids the "+
			"first deployment already minted, so its objects took ids that are already taken",
		second.Time, firstTime)

	// Causality control: a monotonic clock is worth nothing if the re-deploy no
	// longer replaces the code. The package value lives at a reserved ObjectID
	// that the chain resolves by path, so a re-deploy that mints it anywhere
	// else leaves the previous deployment answering calls.
	which, err := env.vmk.QueryEvalString(env.ctx, pkgPath, "Which()")
	require.NoError(t, err)
	require.Equal(t, "v2", which, "the re-deployed source must be the one that answers calls")
}

// The same monotonicity on the "inert" policy's EnablePackage path, plus what
// the restart leaves in the app hash.
//
// An escaped object's value hash is written into the merkleized store under its
// bare ObjectID. Their only deleter is driven by the realm's own deleted marks,
// so a realm that restarts its counter never names the ids it minted before:
// every escaped ObjectID above the restarted high-water mark keeps a hash in
// the app hash for an id the realm will mint again.
func TestEnablePackageRedeployKeepsTheRealmObjectClock(t *testing.T) {
	const pkgPath = "gno.land/r/test/enableclock"

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("enable-clock"))
	env, ctx := inertEnv(t, approver, creator)

	require.NoError(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(creator, pkgPath, privateRealmFiles(pkgPath, redeployV1))))
	require.NoError(t, env.vmk.EnablePackage(ctx, approvalFor(t, env, ctx, approver, pkgPath)))
	env.vmk.CommitGnoTransactionStore(ctx)

	first := env.vmk.getGnoTransactionStore(ctx).GetPackageRealm(pkgPath)
	require.NotNil(t, first, "premise: the first deploy must have written a realm record")
	firstTime := first.Time
	require.NotEmpty(t, escapedObjectHashes(t, ctx, env, pkgPath),
		"premise: a realm always escapes at least its package block into the app hash")

	require.NoError(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(creator, pkgPath, privateRealmFiles(pkgPath, redeployV2))))
	require.NoError(t, env.vmk.EnablePackage(ctx, approvalFor(t, env, ctx, approver, pkgPath)))
	env.vmk.CommitGnoTransactionStore(ctx)

	second := env.vmk.getGnoTransactionStore(ctx).GetPackageRealm(pkgPath)
	require.NotNil(t, second)
	assert.GreaterOrEqual(t, second.Time, firstTime,
		"re-activation restarted the realm's object-id counter at %d, below the %d ids the "+
			"first deployment already minted", second.Time, firstTime)

	var stranded []uint64
	for newTime := range escapedObjectHashes(t, ctx, env, pkgPath) {
		if newTime > second.Time {
			stranded = append(stranded, newTime)
		}
	}
	slices.Sort(stranded)
	assert.Empty(t, stranded,
		"the app hash asserts escaped-object hashes at ObjectIDs %v, above the live realm's "+
			"high-water mark of %d, so the realm will mint those ids again", stranded, second.Time)

	which, err := env.vmk.QueryEvalString(env.ctx, pkgPath, "Which()")
	require.NoError(t, err)
	require.Equal(t, "v2", which, "the re-deployed source must be the one that answers calls")
}

// The escrow at a realm's storage-deposit address must never hold more than the
// realm's own record accounts for. Anything above the record is immobilised:
// refunds are sized from rlm.Storage and rlm.Deposit, and the escrow address is
// a truncated hash of the realm path rather than of a public key, so nothing
// can sign for the excess either.
func TestRedeployedPrivateRealmEscrowMatchesItsRecord(t *testing.T) {
	const pkgPath = "gno.land/r/test/redeployescrow"
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	creator := crypto.AddressFromPreimage([]byte("redeploy-escrow"))
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, creator))
	require.NoError(t, env.bankk.SetCoins(ctx, creator, initialBalance))

	escrow := gno.DeriveStorageDepositCryptoAddr(pkgPath)
	balance := func(addr crypto.Address) int64 {
		return env.bankk.GetCoins(ctx, addr).AmountOf(ugnot.Denom)
	}
	record := func() *gno.Realm {
		return env.vmk.getGnoTransactionStore(ctx).GetPackageRealm(pkgPath)
	}

	beforeFirst := balance(creator)
	require.NoError(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(creator, pkgPath, privateRealmFiles(pkgPath, redeployV1))))
	env.vmk.CommitGnoTransactionStore(ctx)
	afterFirst := balance(creator)

	first := record()
	require.NotNil(t, first, "premise: the first deploy must have written a realm record")
	require.NotZero(t, first.Storage, "premise: the first deploy must have accounted for storage")
	require.Equal(t, int64(first.Deposit), balance(escrow),
		"premise: after one deploy the escrow must match the realm record, so any gap "+
			"below is the re-deploy's doing")
	t.Logf("deploy 1: creator paid %d ugnot, record %d bytes / %d ugnot, escrow %d ugnot",
		beforeFirst-afterFirst, first.Storage, first.Deposit, balance(escrow))

	require.NoError(t, env.vmk.AddPackage(ctx,
		NewMsgAddPackage(creator, pkgPath, privateRealmFiles(pkgPath, redeployV2))))
	env.vmk.CommitGnoTransactionStore(ctx)
	afterSecond := balance(creator)

	second := record()
	require.NotNil(t, second)
	t.Logf("deploy 2: creator paid %d ugnot, record %d bytes / %d ugnot, escrow %d ugnot",
		afterFirst-afterSecond, second.Storage, second.Deposit, balance(escrow))

	require.Equal(t, int64(second.Deposit), balance(escrow),
		"the realm's storage-deposit address holds %d ugnot while the realm record accounts "+
			"for %d; the %d ugnot gap is what the first deploy charged, and no code path can "+
			"size it back out",
		balance(escrow), second.Deposit, balance(escrow)-int64(second.Deposit))
}
