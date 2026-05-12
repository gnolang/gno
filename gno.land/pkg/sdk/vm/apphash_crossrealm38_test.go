package vm

// This test is a forward guard: it pins the current (fixed) committed
// multistore hash so any future change that silently shifts the save set
// trips the assertion. It does not demonstrate the old behavior.
//
// It pins the committed multistore hash (apphash equivalent) after running
// the scenario from gnovm/tests/files/zrealm_crossrealm38.gno at the SDK
// layer. It is the direct consensus-level evidence complementing the
// filetest's save-set golden.
//
// Why an apphash test is needed:
//   The zrealm_crossrealm38.gno filetest only exercises the opslog (which
//   objects enter the save set). The save set drives writes to the iavlStore
//   for every escaped object, and the iavlStore Merkle root is what surfaces
//   as the app hash. So the filetest is an indirect proxy for the commitment.
//   This test closes the loop by pinning the commitment itself.
//
// What this test proves:
//   - Running the crossrealm38 scenario deterministically produces the pinned
//     multistore hash. Any change to the save set shifts the hash
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

// Expected multistore commit hash after running the crossrealm38 scenario.
// Capture recipe:
//
//	go test ./gno.land/pkg/sdk/vm/ -run TestAppHashCrossrealm38 -v
//
// then paste the "observed hash" from the failure message here.
//
// If this test fails after an intentional change to realm/ownership logic,
// verify the change is actually consensus-breaking before updating this
// constant — re-run the zrealm_crossrealm38.gno filetest and inspect the
// save-set diff first.
const expectedCrossrealm38Hash = "71ceb2a2fdb652e741a4f85cc4d045a5dc6139aeb0ddae523bc5f2ca868708fc"

func TestAppHashCrossrealm38(t *testing.T) {
	env := setupTestEnv()
	ctx := env.vmk.MakeGnoTransactionStore(env.ctx)

	// Fund deployer.
	addr := crypto.AddressFromPreimage([]byte("crossrealm38-deployer"))
	acc := env.acck.NewAccountWithAddress(ctx, addr)
	env.acck.SetAccount(ctx, acc)
	env.bankk.SetCoins(ctx, addr, initialBalance)

	// Tx1: deploy crossrealm_f (the collection realm with the growing slice).
	const crossrealmFPkg = "gno.land/r/tests/vm/crossrealm_f"
	require.NoError(t, deployExamplePackage(env, ctx, addr, crossrealmFPkg))
	env.vmk.CommitGnoTransactionStore(ctx)

	// Tx2: deploy an impl realm whose init() does one append into crossrealm_f.
	// After this tx the persisted state has a HeapItemValue whose OwnerID
	// points at the cap-1 backing array. That array will be deleted in Tx3
	// (when AddC's append grows the backing), turning the OwnerID into a
	// stale cross-tx pointer — the exact condition the bug requires.
	const implPkg = "gno.land/r/test/crossrealm38impl"
	implFiles := []*std.MemFile{
		{Name: "gnomod.toml", Body: gno.GenGnoModLatest(implPkg)},
		{Name: "impl.gno", Body: `
package crossrealm38impl

import "gno.land/r/tests/vm/crossrealm_f"

func init() {
	crossrealm_f.Add(cross, &crossrealm_f.Entry{Key: "a", Value: 1})
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
	// bug would silently truncate; the fix rehydrates the owner via
	// store.GetObjectSafe and the walk reaches the escaped PackageValue,
	// whose updated hash lands in iavlStore and changes the commit hash.
	ctx = env.vmk.MakeGnoTransactionStore(env.ctx)
	_, err := env.vmk.Call(ctx, NewMsgCall(addr, nil, implPkg, "AddC", []string{}))
	require.NoError(t, err)
	env.vmk.CommitGnoTransactionStore(ctx)

	got := commitMultiStoreHash(t, env)
	if expectedCrossrealm38Hash == "" {
		t.Fatalf("expected hash not pinned; observed hash = %s\n"+
			"update expectedCrossrealm38Hash to this value if the scenario "+
			"is the one you intend to pin.", got)
	}
	require.Equal(t, expectedCrossrealm38Hash, got,
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
