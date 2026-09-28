package vm

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gnolang/gno/gno.land/pkg/gnoland/ugnot"
	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
)

// A redeploy replaces the package value at the path, so everything the
// previous deployment owned becomes unreachable from Gno in the same instant.
// Storage accounting has to follow: the bytes go, the deposit comes back, and
// the new deployment is charged for itself.
//
// storageSlack is the drift a redeploy leaves behind relative to a first
// deployment of the same source. The realm's object clock is monotonic across
// a redeploy, so the second deployment's object ids are larger numbers than
// the first's and serialize a few bytes wider. It is a constant handful of
// bytes, not a function of what the realm held, which is the property these
// tests are actually pinning.
const storageSlack = 256

const redeployStorageSrc = `import "strconv"

var slots []string

func Set(cur realm, s string) { slots = append(slots, s) }

func Count() string { return strconv.Itoa(len(slots)) }

func Free(cur realm) { slots = nil }`

func setupRedeployStorage(t *testing.T, seed string) (testEnv, crypto.Address) {
	t.Helper()
	env := setupTestEnv()
	addr := crypto.AddressFromPreimage([]byte(seed))
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	env.acck.SetAccount(ctx, env.acck.NewAccountWithAddress(ctx, addr))
	require.NoError(t, env.bankk.SetCoins(ctx, addr, initialBalance))
	return env, addr
}

// The headline: the state a redeploy wipes stops being charged for.
//
// Before this, the previous deployment's objects stayed in rlm.Storage with no
// code path able to free them: the package graph is cyclic (package block ->
// FuncValue -> file block -> package block), so dropping the root never
// cascades, and no Gno code in the new deployment can name them to delete
// them either.
func TestRedeployReleasesTheStateItWipes(t *testing.T) {
	const pkgPath = "gno.land/r/test/redeployrelease"
	env, addr := setupRedeployStorage(t, "redeploy-release")
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	escrow := gno.DeriveStorageDepositCryptoAddr(pkgPath)
	bal := func(a crypto.Address) int64 { return env.bankk.GetCoins(ctx, a).AmountOf(ugnot.Denom) }
	rec := func() *gno.Realm { return env.vmk.getGnoTransactionStore(ctx).GetPackageRealm(pkgPath) }
	count := func() string {
		s, err := env.vmk.QueryEvalString(env.ctx, pkgPath, "Count()")
		require.NoError(t, err)
		return s
	}
	deploy := func() {
		require.NoError(t, env.vmk.AddPackage(ctx,
			NewMsgAddPackage(addr, pkgPath, privateRealmFiles(pkgPath, redeployStorageSrc))))
		env.vmk.CommitGnoTransactionStore(ctx)
	}
	call := func(fn string, args ...string) {
		msg := NewMsgCall(addr, std.Coins{}, pkgPath, fn, args)
		msg.MaxDeposit = std.Coins{{Denom: ugnot.Denom, Amount: 1_000_000_000}}
		_, err := env.vmk.Call(ctx, msg)
		require.NoError(t, err)
	}

	deploy()
	empty := rec().Storage
	require.NotZero(t, empty, "premise: a deployment must be charged for its own objects")

	for i := 0; i < 40; i++ {
		call("Set", fmt.Sprintf("slot-%02d-%s", i, strings.Repeat("x", 64)))
	}
	env.vmk.CommitGnoTransactionStore(ctx)
	withState := rec().Storage
	require.Greater(t, withState, empty, "premise: the calls must have grown the realm")
	t.Logf("empty=%d withState=%d (state is %d bytes)", empty, withState, withState-empty)

	beforeRedeploy := bal(addr)
	deploy()
	afterRedeploy := rec().Storage
	t.Logf("afterRedeploy=%d, creator paid %d ugnot", afterRedeploy, beforeRedeploy-bal(addr))

	assert.Equal(t, "0", count(),
		"a redeploy must still reset the realm: private redeploy semantics are unchanged")
	assert.Less(t, afterRedeploy, empty+storageSlack,
		"the redeploy wiped %d bytes of state and the realm is still charged for %d, "+
			"against the %d a first deployment of the same source costs",
		withState-empty, afterRedeploy, empty)
	assert.Equal(t, int64(rec().Deposit), bal(escrow),
		"the escrow must still match the record after a release")
}

// A redeploy that shrinks the realm refunds the difference, the same way a
// call that frees state does. The control matters: the refund path itself was
// never broken, a redeploy simply never reached it.
func TestRedeployThatShrinksRefunds(t *testing.T) {
	const big = `type node struct{ N int; S string }

var all []*node

func init() {
	for i := 0; i < 200; i++ {
		all = append(all, &node{N: i, S: "padpadpadpadpadpadpadpadpadpadpadpadpadpadpadpadpadpadpadpadpad"})
	}
}`
	const tiny = `func Which() string { return "v2" }`

	const pkgPath = "gno.land/r/test/redeployshrink"
	env, addr := setupRedeployStorage(t, "redeploy-shrink")
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	escrow := gno.DeriveStorageDepositCryptoAddr(pkgPath)
	bal := func(a crypto.Address) int64 { return env.bankk.GetCoins(ctx, a).AmountOf(ugnot.Denom) }
	deploy := func(path, body string) (storage uint64, paid int64) {
		before := bal(addr)
		require.NoError(t, env.vmk.AddPackage(ctx,
			NewMsgAddPackage(addr, path, privateRealmFiles(path, body))))
		env.vmk.CommitGnoTransactionStore(ctx)
		return env.vmk.getGnoTransactionStore(ctx).GetPackageRealm(path).Storage, before - bal(addr)
	}

	bigStorage, bigPaid := deploy(pkgPath, big)
	tinyStorage, tinyPaid := deploy(pkgPath, tiny)
	freshStorage, _ := deploy("gno.land/r/test/redeployshrinkfresh", tiny)
	t.Logf("big=%d (paid %d), redeployed tiny=%d (paid %d), same tiny on a fresh path=%d",
		bigStorage, bigPaid, tinyStorage, tinyPaid, freshStorage)

	assert.Less(t, tinyStorage, bigStorage,
		"redeploying a strictly smaller package must shrink the realm")
	assert.Less(t, tinyStorage, freshStorage+storageSlack,
		"a redeploy must cost what the same source costs on an empty path, not that plus "+
			"whatever the path already held")
	assert.Negative(t, tinyPaid,
		"shrinking the realm by %d bytes must refund, not charge %d ugnot",
		bigStorage-tinyStorage, tinyPaid)
	assert.Equal(t,
		int64(env.vmk.getGnoTransactionStore(ctx).GetPackageRealm(pkgPath).Deposit),
		bal(escrow),
		"the escrow must still match the record after a refunding redeploy")
}

// Redeploying the same package repeatedly must not ratchet the realm upwards.
// This is what made the bug expensive in practice: a realm whose content is
// pushed back after every redeploy paid for every generation it ever had.
func TestRepeatedRedeployDoesNotRatchet(t *testing.T) {
	const pkgPath = "gno.land/r/test/redeployratchet"
	env, addr := setupRedeployStorage(t, "redeploy-ratchet")
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	storages := make([]uint64, 0, 5)
	for i := 0; i < 5; i++ {
		require.NoError(t, env.vmk.AddPackage(ctx,
			NewMsgAddPackage(addr, pkgPath, privateRealmFiles(pkgPath, redeployStorageSrc))))
		env.vmk.CommitGnoTransactionStore(ctx)
		storages = append(storages, env.vmk.getGnoTransactionStore(ctx).GetPackageRealm(pkgPath).Storage)
	}
	t.Logf("storage after each of 5 deployments: %v", storages)

	assert.Less(t, storages[4], storages[0]+storageSlack,
		"five deployments of one package left the realm charged for %d bytes, against %d "+
			"after the first: the previous generations are still on the books", storages[4], storages[0])
}
