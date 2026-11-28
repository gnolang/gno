package vm

import (
	"slices"
	"testing"

	"github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/log"
	"github.com/gnolang/gno/tm2/pkg/sdk"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/stretchr/testify/require"
)

const (
	redeployRealmPath = "gno.land/r/test/redeploy"
	redeployDepPath   = "gno.land/p/test/redeploydep"
)

// bootIndexOrder drains the package index in index order, which is the order
// PreprocessAllFilesAndSaveBlockNodes preprocesses packages in at every node
// restart.
func bootIndexOrder(env testEnv, ctx sdk.Context) []string {
	var order []string
	for mpkg := range env.vmk.getGnoTransactionStore(ctx).IterMemPackage() {
		order = append(order, mpkg.Path)
	}
	return order
}

// bootIndexPositions returns every position path occupies in a boot-index
// order, i.e. how many times the boot pass preprocesses it and where.
func bootIndexPositions(order []string, path string) []int {
	var at []int
	for i, p := range order {
		if p == path {
			at = append(at, i)
		}
	}
	return at
}

// restartBootPreprocess drops the keeper's store and re-runs VMKeeper.Initialize
// the way a node restart does, returning the value the boot pass panicked with,
// or nil if it completed. gno.land/pkg/gnoland/app.go calls vmk.Initialize bare,
// with no recover on the path, so a panic here kills the process.
func restartBootPreprocess(env testEnv, ctx sdk.Context) (recovered any) {
	defer func() { recovered = recover() }()

	env.vmk.CommitGnoTransactionStore(ctx)
	env.vmk.gnoStore = nil
	mcw := env.ctx.MultiStore().MultiCacheWrap()
	env.vmk.Initialize(log.NewNoopLogger(), mcw)
	mcw.MultiWrite()
	return nil
}

// Built fresh per submission: stampGnomod rewrites gnomod.toml in place, so a
// shared slice would carry one submission's stamp into the next.
func redeployRealmFiles(body string) []*std.MemFile {
	return []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(redeployRealmPath) + "\nprivate = true"},
		{Name: "redeploy.gno", Body: body},
	}
}

func redeployDepFiles() []*std.MemFile {
	return []*std.MemFile{
		{Name: "gnomod.toml", Body: gnolang.GenGnoModLatest(redeployDepPath)},
		{Name: "redeploydep.gno", Body: `package redeploydep

func Tag() string { return "dep-v1" }`},
	}
}

const (
	redeployRealmV1 = `package redeploy

func Which(cur realm) string { return "v1" }`

	redeployRealmV2 = `package redeploy

import "` + redeployDepPath + `"

func Which(cur realm) string { return redeploydep.Tag() }`
)

// requireRedeployPreprocessedOnceAboveDep asserts what the boot index has to
// hand a restarting node after a redeploy: the path once, above the dependency
// its current content imports, and a pass over that order that runs through to
// the end.
//
// The count and the position are separate properties and both are load-bearing.
// A de-duplication that kept a path's FIRST entry would satisfy the count and
// still preprocess the redeployed content below the dependency it imports,
// which is the boot that dies.
//
// The restart is asserted before them because they are fatal: behind them, the
// failure that actually kills a node would never be reached. The order has to
// be captured first, since the restart commits the transaction store and drops
// the keeper's.
func requireRedeployPreprocessedOnceAboveDep(t *testing.T, env testEnv, ctx sdk.Context) {
	t.Helper()

	order := bootIndexOrder(env, ctx)

	bootPanic := restartBootPreprocess(env, ctx)
	require.Nil(t, bootPanic, "the boot preprocess pass panicked: %v", bootPanic)

	at := bootIndexPositions(order, redeployRealmPath)
	depAt := slices.Index(order, redeployDepPath)
	require.NotEqual(t, -1, depAt, "the dependency is indexed")
	require.Len(t, at, 1, "the redeployed path is preprocessed once (order: %v)", order)
	require.Greater(t, at[0], depAt,
		"it is preprocessed above the dependency its current content imports (order: %v)", order)
}

// redeployPrivateRealmViaEnable plays out ordinary maintenance of a private
// realm under the inert policy: the realm is parked and enabled, a library is
// parked and enabled, then the realm is parked and enabled again, now importing
// that library. depFirst enables the library BEFORE the realm's first
// deployment instead of after.
//
// Every AddPackage and every EnablePackage here is a real message from a real
// creator and a real approver in PkgApprovers, and every one of them succeeds.
func redeployPrivateRealmViaEnable(t *testing.T, depFirst bool) (testEnv, sdk.Context) {
	t.Helper()

	approver := crypto.AddressFromPreimage([]byte("oracle"))
	creator := crypto.AddressFromPreimage([]byte("redeploy-creator"))
	env, ctx := inertEnv(t, approver, creator)

	parkAndEnable := func(path string, files []*std.MemFile) {
		require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, path, files)))
		require.NoError(t, env.vmk.EnablePackage(ctx, approvalFor(t, env, ctx, approver, path)))
	}
	enableDep := func() { parkAndEnable(redeployDepPath, redeployDepFiles()) }

	if depFirst {
		enableDep()
	}
	parkAndEnable(redeployRealmPath, redeployRealmFiles(redeployRealmV1))
	if !depFirst {
		enableDep()
	}
	parkAndEnable(redeployRealmPath, redeployRealmFiles(redeployRealmV2))

	return env, ctx
}

// redeployPrivateRealmViaAddPackage plays out the same maintenance on the
// ordinary permissionless path, where AddPackage both stores and runs the
// package and no approver is involved.
func redeployPrivateRealmViaAddPackage(t *testing.T, depFirst bool) (testEnv, sdk.Context) {
	t.Helper()

	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)
	creator := crypto.AddressFromPreimage([]byte("redeploy-creator"))
	acc := env.acck.NewAccountWithAddress(ctx, creator)
	env.acck.SetAccount(ctx, acc)
	require.NoError(t, env.bankk.SetCoins(ctx, creator, initialBalance))

	deploy := func(path string, files []*std.MemFile) {
		require.NoError(t, env.vmk.AddPackage(ctx, NewMsgAddPackage(creator, path, files)))
	}
	deployDep := func() { deploy(redeployDepPath, redeployDepFiles()) }

	if depFirst {
		deployDep()
	}
	deploy(redeployRealmPath, redeployRealmFiles(redeployRealmV1))
	if !depFirst {
		deployDep()
	}
	deploy(redeployRealmPath, redeployRealmFiles(redeployRealmV2))

	return env, ctx
}

// TestRedeployOverLivePrivateRealmPreprocessesThePathOnce pins the boot pass a
// redeploy leaves behind, on both paths that can perform one.
//
// A redeploy deletes the path's mempackage blobs and re-adds them, which
// writes a second index entry and leaves the first. The boot pass must reach
// the path at its new position only: it walks index order, so the entry at the
// original position would preprocess the redeployed content before a
// dependency deployed in between, and GetBlockNode's panic on the unsaved
// dependency takes the node's boot with it.
//
// The "dependency deployed before" case is the causality control. It carries
// the same redeploy with no position inversion available, so it boots clean
// whatever the index holds, and only the count assertion can fail there.
func TestRedeployOverLivePrivateRealmPreprocessesThePathOnce(t *testing.T) {
	flows := map[string]func(*testing.T, bool) (testEnv, sdk.Context){
		"enabled by an approver": redeployPrivateRealmViaEnable,
		"deployed directly":      redeployPrivateRealmViaAddPackage,
	}
	order := map[string]bool{
		"dependency deployed after the original deploy":  false,
		"dependency deployed before the original deploy": true,
	}
	for flow, redeploy := range flows {
		t.Run(flow, func(t *testing.T) {
			for name, depFirst := range order {
				t.Run(name, func(t *testing.T) {
					env, ctx := redeploy(t, depFirst)
					requireRedeployPreprocessedOnceAboveDep(t, env, ctx)
				})
			}
		})
	}
}
