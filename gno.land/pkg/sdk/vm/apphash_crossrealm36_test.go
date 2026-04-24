package vm

// This test pins the committed multistore hash (apphash equivalent) after
// running the scenario from gnovm/tests/files/zrealm_crossrealm36.gno at the
// SDK layer. It is the direct consensus-level evidence complementing the
// filetest's save-set golden.
//
// Why an apphash test is needed:
//   The zrealm_crossrealm36.gno filetest only exercises the opslog (which
//   objects enter the save set). The save set drives writes to the iavlStore
//   for every escaped object, and the iavlStore Merkle root is what surfaces
//   as the app hash. So the filetest is an indirect proxy for the commitment.
//   This test closes the loop by pinning the commitment itself.
//
// What this test proves:
//   - Running the crossrealm36 scenario deterministically produces the pinned
//     multistore hash. Any change to the save set — including silently-missed
//     dirty-ancestor marks (the class of bug the #5291 regression introduced
//     when getOwner stopped rehydrating a deleted owner) — shifts the hash
//     and fails this test.
//
// What this test does NOT prove:
//   - That two different code versions (buggy vs fixed) produce DIFFERENT
//     apphashes for the same input. Proving that requires a version-gated
//     runtime switch on getOwner, which doesn't exist in this tree yet.
//     See the ADR note in the PR description; that work belongs with the
//     chain-upgrade gating effort, not here.

import (
	"fmt"
	"testing"

	gno "github.com/gnolang/gno/gnovm/pkg/gnolang"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/std"
	"github.com/gnolang/gno/tm2/pkg/store/types"
	"github.com/stretchr/testify/require"
)

// Expected multistore commit hash after running the crossrealm36 scenario.
// Capture recipe:
//
//	go test ./gno.land/pkg/sdk/vm/ -run TestAppHashCrossrealm36 -v
//
// then paste the "observed hash" from the failure message here.
//
// If this test fails after an intentional change to realm/ownership logic,
// verify the change is actually consensus-breaking before updating this
// constant — re-run the zrealm_crossrealm36.gno filetest and inspect the
// save-set diff first.
const expectedCrossrealm36Hash = "9b6ce282244408d891ba4c59b62d87761e2bbfd7883345d0abbd8b6a30de1470"

func TestAppHashCrossrealm36(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Fund deployer.
	addr := crypto.AddressFromPreimage([]byte("crossrealm36-deployer"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)

	// Tx1: deploy crossrealm_f (the collection realm with the growing slice).
	const crossrealmFPkg = "gno.land/r/tests/vm/crossrealm_f"
	require.NoError(t, deployExamplePackage(env, ctx, addr, crossrealmFPkg))
	env.vmk.CommitGnoTransactionStore(ctx)

	// Tx2: deploy an impl realm whose init() does two appends into crossrealm_f.
	// This finalization is what leaves a HeapItemValue with a stale OwnerID
	// pointing at the replaced (and deleted) backing array.
	const implPkg = "gno.land/r/test/crossrealm36impl"
	implFiles := []*std.MemFile{
		{Name: "gnomod.toml", Body: gno.GenGnoModLatest(implPkg)},
		{Name: "impl.gno", Body: `
package crossrealm36impl

import "gno.land/r/tests/vm/crossrealm_f"

func init() {
	crossrealm_f.Add(cross, &crossrealm_f.Entry{Key: "a", Value: 1})
	crossrealm_f.Add(cross, &crossrealm_f.Entry{Key: "b", Value: 2})
}

func AddC(cur realm) {
	crossrealm_f.Add(cross, &crossrealm_f.Entry{Key: "c", Value: 3})
}
`},
	}
	ctx = env.vmk.MakeGnoTransactionStore(env.ctx)
	require.NoError(t,
		env.vmk.AddPackage(ctx, NewMsgAddPackage(addr, implPkg, implFiles)),
	)
	env.vmk.CommitGnoTransactionStore(ctx)

	// Tx3: call AddC — the subsequent mutation that forces markDirtyAncestors
	// to walk through the stale-owner HeapItemValue. This is the path the
	// #5291 bug would silently truncate; the fix rehydrates the owner via
	// store.GetObjectSafe and the walk reaches the escaped PackageValue,
	// whose updated hash lands in iavlStore and changes the commit hash.
	ctx = env.vmk.MakeGnoTransactionStore(env.ctx)
	_, err := env.vmk.Call(ctx, NewMsgCall(addr, nil, implPkg, "AddC", []string{}))
	require.NoError(t, err)
	env.vmk.CommitGnoTransactionStore(ctx)

	got := commitMultiStoreHash(t, env)
	if expectedCrossrealm36Hash == "" {
		t.Fatalf("expected hash not pinned; observed hash = %s\n"+
			"update expectedCrossrealm36Hash to this value if the scenario "+
			"is the one you intend to pin.", got)
	}
	require.Equal(t, expectedCrossrealm36Hash, got,
		"multistore commit hash changed — the save set (and therefore the "+
			"iavlStore Merkle root) shifted. Verify this is an intentional "+
			"consensus-breaking change before updating the pinned value.")
}

// commitMultiStoreHash commits the test multistore and returns the resulting
// root hash as a lowercase hex string. The test environment's MultiStore is a
// CommitMultiStore (see setupTestEnv); we type-assert through the Committer
// interface to call Commit() and pull the CommitID.Hash.
func commitMultiStoreHash(t *testing.T, env testEnv) string {
	t.Helper()
	committer, ok := env.ctx.MultiStore().(types.Committer)
	require.Truef(t, ok, "MultiStore does not implement types.Committer")
	cid := committer.Commit()
	return fmt.Sprintf("%x", cid.Hash)
}
