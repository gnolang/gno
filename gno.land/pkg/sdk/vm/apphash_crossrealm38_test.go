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
//     runtime switch on getOwner, which belongs with the chain-upgrade
//     gating effort, not here.

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
// Hash bumped 2026-05-26: adding crypto/{bn254,cometbls,cometblszk,keccak256,merkle,modexp}
// to the genesis stdlib set shifts the iavlStore Merkle root. New stdlibs always do — this
// PR is the test13 chain-upgrade vehicle, so the shift is intentional.
// Hash bumped 2026-06-01: this branch's foreign-markdown work changes the genesis
// package set (notably the chain/markdown stdlib), which shifts the iavlStore Merkle
// root — same class of change as the crypto-stdlib bump above. Verified this is NOT
// the merged nil-realm write-gate fix (#5758): crossrealm38 still produces e37075fb
// on a clean origin/master. Behavior is unchanged (the zrealm_crossrealm38.gno
// filetest passes); only the genesis encoding shifted.
// Hash bumped 2026-06-07: adding the errors stdlib (Unwrap/Is/Join) to the genesis
// stdlib set shifts the iavlStore Merkle root. Behavior is unchanged (the
// zrealm_crossrealm38.gno filetest still passes); only the genesis encoding shifted.
//
// Hash bumped again by the Example-test PR: editing
// gnovm/stdlibs/math/rand/example_test.gno changes the math/rand stdlib
// MemPackage that is committed into genesis state (stdlib MemPackages include
// their *_test.gno source bytes), which shifts the iavlStore Merkle root. This
// is the only consensus-relevant change in that PR; verified by bisection that
// no other change in the PR moves this hash. The shift is therefore expected.
//
// Hash bumped 2026-07-10 (bptree mount PR), two coinciding causes: (1) the
// test env's main store switched from IAVL to the B+32 bptree store
// (different commitment structure — every multistore hash moves); (2) the
// depth gas pins committed into "vm:p" changed (Fixed = Min: 300/200/440 →
// 100/200/540). Behavior is unchanged (the zrealm_crossrealm38.gno filetest
// still passes).
//
// Hash bumped by the realm.Sub PR (#5890): the realm interface gained
// Sub/Subpath (shifting its TypeID) and the chain/banker + chain/address
// stdlib source changed (NewBanker IsCurrent guard, sub-realm helpers) —
// stdlib MemPackage source bytes are committed into genesis state, so the
// committed multistore root shifts. The crossrealm38 scenario itself does
// not use sub-realms; the move is purely the interface/stdlib change and is
// an intended consensus break for that PR.
//
// Hash bumped by the mempackage prod/test storage split (#5891): MP*All
// packages now store production files under pkg:<path> (typed MP*Prod) and
// test/filetest files under a pkg:<path>#allbutprod sibling, changing stored
// package bytes and the committed multistore root. Behavior is unchanged;
// only the storage encoding shifted.
//
// Hash bumped by the preprocess-gas PR (#5892): the new PreprocessGasPerByte
// vm param (default 1250) has a non-zero default serialized into the genesis
// vm params state, shifting the committed multistore root. Behavior is
// unchanged; only the genesis params encoding shifted. (Value re-derived
// after merging master, so it reflects the bptree store + #5890 + #5891 +
// this param together.)
//
// Hash bumped by the apd -> math/big.Rat PR (#5867): BigdecValue (untyped
// float constant representation) now amino-serializes in rational form
// ("1/3") instead of the old decimal string ("0.3333333333"), shifting the
// committed multistore root for any realm state containing bigdec constants.
// Behavior is unchanged for all typed values; only the constant-folding
// arithmetic is corrected (fixes #5862). Re-derived after merging master, so
// it reflects the bptree store + #5890 + #5891 + #5892 + this change together.
//
// Bumped again by a doc-comment-only edit to chain/banker's package comment
// (the NewBanker capability-persistence warning). Stdlib .gno sources are
// stored in chain state, so their bytes — comments included — are covered by
// the multistore root: editing a comment in a stdlib package is a
// consensus-breaking change, even though no behavior changes. Confirmed by
// reverting that comment alone, which restores the previous hash. No
// executable code was touched.
// Hash bumped by adding banker.GetCoin: the chain/banker stdlib gained an
// interface method, a native declaration and a method body, and stdlib .gno
// source bytes are committed into genesis state, so the multistore root moves.
// The crossrealm38 scenario itself does not call GetCoin; the shift is purely the
// stdlib source change and is intended. Note this is the *only* reason this
// branch moves the pin — the balance split alone does not, because that scenario
// holds no non-gas denom.
//
// Bumped again by the OriginSend banker lifetime fix. banker.gno gains the
// realm-handle field that pins such a banker to its message, plus the
// accompanying doc. Same reason as the comment-only bump noted above:
// stdlib .gno source bytes are genesis state, so any edit to them moves the
// root. This one does change behavior, and is intentionally
// consensus-breaking.
//
// Hash bumped by the phase-2 inert-packages PR (#5888): new vm params
// (code_submission_policy, code_submitters, pkg_approvers) plus inert-storage
// keying serialize additional defaults into the genesis vm params state,
// shifting the committed multistore root. Behavior is unchanged (policy
// defaults to permissionless). Re-derived after merging master.
//
// Re-derived once more for the merge of master into #5888: both sides moved
// the root, so neither side's value survives the combination.
//
// Bumped by the run_submitters vm param (MsgRun allowlist). A new Params field
// moves this root even though its default is an empty list: params.SetStruct
// goes through encodeStructFields (tm2/pkg/sdk/params/amino_helper.go), which
// writes one store key per field unconditionally and does not skip zero values.
// So the genesis vm params state gains a `vm:p:run_submitters` key holding
// `null`. Behavior at this hash is unchanged — the scenario sends no MsgRun.
//
// Bumped again by lowering the DefaultDeposit param from 600000000ugnot to
// 100000000ugnot. Unlike the bumps above this is a value change, not a new
// key: `vm:p:default_deposit` is genesis state, so its contents are in the
// root. It does change behavior at the margin — the param is the fallback
// CEILING on a storage deposit when a message declares no MaxDeposit, so a
// single message may now add at most 1 MB of realm state rather than 6 MB
// before it is refused. Measured against all 321 genesis packages the largest
// deploy is r/gnoland/boards2/v1 at 276,098 bytes (27,609,800ugnot), so the
// new ceiling clears the worst real case by 3.6x.
// Bumped again by the two inert-charge vm params, for the same reason as
// run_submitters above: two more keys, written unconditionally. Behavior at
// this hash is unchanged — inert_submission_charge defaults to empty, which
// means off, and the scenario submits nothing under the "inert" policy anyway.
//
// Hash bumped by the native-input-bounds PR: crypto/bn254's G1Add/G1Mul got
// their length checks moved into the .gno wrapper (ahead of the native call, so
// an oversized input is not copied into Go memory for a flat fee), plus a test
// for that. Both bn254.gno and bn254_test.gno are stdlib source bytes committed
// into genesis state, so the root moves. Attributed by bisection against the
// OriginSend value above: base .gno files give b43e5fd5, bn254_test.gno alone
// gives a25dc7a4, both give the value below; the innerHash gas-table change moves
// nothing (gas is not committed state). Behavior is unchanged — the crossrealm38
// filetest passes and the bn254 EIP-196/197 vectors are untouched.
//
// Bumped again by the entity-reference hardening of PercentEncodeURL. The
// change to chain/markdown is comment-only on the .gno side — the encoding
// rule itself lives in the injected Go implementation — but stdlib .gno
// source bytes are genesis state, so documenting the new rule moves the
// root just as the GetCoin bump above did. Re-derived after merging develop,
// so the value below covers the bn254 wrapper bounds above and this change
// together.
//
// Bumped once more within this branch by extending that same PercentEncodeURL
// doc comment (the `&amp;` round-trip note from review). Still comment-only,
// still consensus-breaking for the reason above; the merge commit pins
// 0e8e8714 without it.
//
// Re-derived for the combination: master's params/deposit bumps and this
// branch's stdlib source bumps both move the root, so neither side's value
// survives the merge.
//
// Bumped again by the chain/params reader API: params.gno gains six GetXxx
// declarations and the doc describing what a wrong-type read does. Stdlib .gno
// source bytes are genesis state, so both the declarations and the comment move
// the root. The crossrealm38 scenario calls none of them; the shift is the
// stdlib source change alone. Re-derived after merging master, whose own
// encode/decode work moved the root too, so neither side's value survives.
//
// Bumped 2026-08-26 by the chain.NewCoins copy fix: it edits coins.gno and
// coins_test.gno, and stdlib MemPackages carry both files' source bytes into
// genesis state. The scenario never calls NewCoins, and the
// zrealm_crossrealm38.gno filetest still passes, so behavior is unchanged.
// Checked that the source bytes are the whole cause and not the copy the fix
// adds: starting from the pre-fix file, one comment line moves the root on its
// own.
//
// Bumped 2026-08-27 by a doc fix on Coin.Add and Coin.Sub in coins.gno. Both
// claimed an invalid result panics; neither checks the sign, and 5ugnot.Sub
// (10ugnot) returns -5ugnot. Comments only — no code changed, the scenario
// calls neither method, and the zrealm_crossrealm38.gno filetest still passes.
//
// Hash bumped by the crypto/secp256k1 PR: adding the crypto/secp256k1 stdlib
// (with native Verify) introduces a new stdlib MemPackage into the genesis
// stdlib set, shifting the iavlStore Merkle root. This is the intended
// consensus-breaking change of adding a new stdlib; the zrealm_crossrealm38.gno
// filetest still passes.
const expectedCrossrealm38Hash = "9b3516aa56fb12e03f97c7579107ff7266a61356a924d89b20f4d44ad97fd381"

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

func init(cur realm) {
	crossrealm_f.Add(cross(cur), crossrealm_f.NewEntry("a", 1))
}

func AddC(cur realm) {
	crossrealm_f.Add(cross(cur), crossrealm_f.NewEntry("c", 3))
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
