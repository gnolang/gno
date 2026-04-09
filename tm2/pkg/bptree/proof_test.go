package bptree

import (
	"bytes"
	"fmt"
	"testing"

	ics23 "github.com/cosmos/ics23/go"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

func newTestDB(t *testing.T) *memdb.MemDB {
	t.Helper()
	return memdb.NewMemDB()
}

func TestProof_MembershipSingleKey(t *testing.T) {
	tree := NewMutableTreeMem()
	tree.Set([]byte("hello"), []byte("world"))
	tree.SaveVersion()

	root := tree.Hash()
	proof, err := tree.GetMembershipProof([]byte("hello"))
	if err != nil {
		t.Fatalf("GetMembershipProof: %v", err)
	}
	ok := ics23.VerifyMembership(BptreeSpec, root, proof, []byte("hello"), []byte("world"))
	if !ok {
		t.Fatalf("membership proof verification failed")
	}
}

func TestProof_MembershipMultipleKeys(t *testing.T) {
	tree := NewMutableTreeMem()
	n := 100
	for i := 0; i < n; i++ {
		tree.Set(fmt.Appendf(nil, "pk%04d", i), fmt.Appendf(nil, "pv%04d", i))
	}
	tree.SaveVersion()

	root := tree.Hash()

	for _, idx := range []int{0, 1, 25, 50, 75, 99} {
		key := fmt.Appendf(nil, "pk%04d", idx)
		val := fmt.Appendf(nil, "pv%04d", idx)

		proof, err := tree.GetMembershipProof(key)
		if err != nil {
			t.Fatalf("GetMembershipProof(%s): %v", key, err)
		}
		ok := ics23.VerifyMembership(BptreeSpec, root, proof, key, val)
		if !ok {
			t.Fatalf("membership proof failed for %s", key)
		}
	}
}

func TestProof_MembershipWrongValue(t *testing.T) {
	tree := NewMutableTreeMem()
	tree.Set([]byte("k"), []byte("v"))
	tree.SaveVersion()

	root := tree.Hash()
	proof, err := tree.GetMembershipProof([]byte("k"))
	if err != nil {
		t.Fatalf("GetMembershipProof: %v", err)
	}
	ok := ics23.VerifyMembership(BptreeSpec, root, proof, []byte("k"), []byte("wrong"))
	if ok {
		t.Fatalf("proof should fail with wrong value")
	}
}

func TestProof_MembershipWrongRoot(t *testing.T) {
	tree := NewMutableTreeMem()
	tree.Set([]byte("k"), []byte("v"))
	tree.SaveVersion()

	proof, _ := tree.GetMembershipProof([]byte("k"))
	fakeRoot := []byte("00000000000000000000000000000000")
	ok := ics23.VerifyMembership(BptreeSpec, fakeRoot, proof, []byte("k"), []byte("v"))
	if ok {
		t.Fatalf("proof should fail with wrong root")
	}
}

func TestProof_MembershipMissingKey(t *testing.T) {
	tree := NewMutableTreeMem()
	tree.Set([]byte("a"), []byte("1"))
	tree.SaveVersion()

	_, err := tree.GetMembershipProof([]byte("missing"))
	if err == nil {
		t.Fatalf("GetMembershipProof should fail for missing key")
	}
}

func TestProof_MembershipLargeTree(t *testing.T) {
	tree := NewMutableTreeMem()
	n := 500
	for i := 0; i < n; i++ {
		tree.Set(fmt.Appendf(nil, "lg%05d", i), fmt.Appendf(nil, "val%05d", i))
	}
	tree.SaveVersion()

	root := tree.Hash()

	for _, idx := range []int{0, 100, 250, 499} {
		key := fmt.Appendf(nil, "lg%05d", idx)
		val := fmt.Appendf(nil, "val%05d", idx)

		proof, err := tree.GetMembershipProof(key)
		if err != nil {
			t.Fatalf("GetMembershipProof(%s): %v", key, err)
		}
		ok := ics23.VerifyMembership(BptreeSpec, root, proof, key, val)
		if !ok {
			t.Fatalf("membership proof failed for %s (height %d)", key, tree.Height())
		}
	}
}

func TestProof_NonMembershipMiddle(t *testing.T) {
	tree := NewMutableTreeMem()
	tree.Set([]byte("a"), []byte("1"))
	tree.Set([]byte("c"), []byte("3"))
	tree.SaveVersion()

	root := tree.Hash()

	proof, err := tree.GetNonMembershipProof([]byte("b"))
	if err != nil {
		t.Fatalf("GetNonMembershipProof: %v", err)
	}
	ok := ics23.VerifyNonMembership(BptreeSpec, root, proof, []byte("b"))
	if !ok {
		t.Fatalf("non-membership proof verification failed for 'b'")
	}
}

func TestProof_NonMembershipBeforeAll(t *testing.T) {
	tree := NewMutableTreeMem()
	tree.Set([]byte("b"), []byte("2"))
	tree.Set([]byte("c"), []byte("3"))
	tree.SaveVersion()

	root := tree.Hash()

	proof, err := tree.GetNonMembershipProof([]byte("a"))
	if err != nil {
		t.Fatalf("GetNonMembershipProof(before): %v", err)
	}
	ok := ics23.VerifyNonMembership(BptreeSpec, root, proof, []byte("a"))
	if !ok {
		t.Fatalf("non-membership proof failed for key before all")
	}
}

func TestProof_NonMembershipAfterAll(t *testing.T) {
	tree := NewMutableTreeMem()
	tree.Set([]byte("a"), []byte("1"))
	tree.Set([]byte("b"), []byte("2"))
	tree.SaveVersion()

	root := tree.Hash()

	proof, err := tree.GetNonMembershipProof([]byte("z"))
	if err != nil {
		t.Fatalf("GetNonMembershipProof(after): %v", err)
	}
	ok := ics23.VerifyNonMembership(BptreeSpec, root, proof, []byte("z"))
	if !ok {
		t.Fatalf("non-membership proof failed for key after all")
	}
}

func TestProof_NonMembershipExistingKey(t *testing.T) {
	tree := NewMutableTreeMem()
	tree.Set([]byte("a"), []byte("1"))
	tree.SaveVersion()

	_, err := tree.GetNonMembershipProof([]byte("a"))
	if err == nil {
		t.Fatalf("GetNonMembershipProof should fail for existing key")
	}
}

func TestProof_NonMembershipLargeTree(t *testing.T) {
	tree := NewMutableTreeMem()
	// Insert even numbers
	for i := 0; i < 200; i += 2 {
		tree.Set(fmt.Appendf(nil, "nm%04d", i), []byte("v"))
	}
	tree.SaveVersion()

	root := tree.Hash()

	for _, idx := range []int{1, 51, 99, 151, 199} {
		key := fmt.Appendf(nil, "nm%04d", idx)
		proof, err := tree.GetNonMembershipProof(key)
		if err != nil {
			t.Fatalf("GetNonMembershipProof(%s): %v", key, err)
		}
		ok := ics23.VerifyNonMembership(BptreeSpec, root, proof, key)
		if !ok {
			t.Fatalf("non-membership proof failed for %s", key)
		}
	}
}

func TestProof_NonMembershipCrossLeaf(t *testing.T) {
	tree := NewMutableTreeMem()
	// Insert even numbers — enough to create multiple leaves
	for i := 0; i < 200; i += 2 {
		tree.Set(fmt.Appendf(nil, "cl%04d", i), []byte("v"))
	}
	if tree.Height() < 1 {
		t.Fatalf("need multiple leaves (height=%d)", tree.Height())
	}
	tree.SaveVersion()

	root := tree.Hash()

	// Probe EVERY odd number — some will be cross-leaf boundaries
	for i := 1; i < 200; i += 2 {
		key := fmt.Appendf(nil, "cl%04d", i)
		proof, err := tree.GetNonMembershipProof(key)
		if err != nil {
			t.Fatalf("GetNonMembershipProof(%s): %v", key, err)
		}
		ok := ics23.VerifyNonMembership(BptreeSpec, root, proof, key)
		if !ok {
			t.Fatalf("non-membership proof failed for %s (cross-leaf test)", key)
		}
	}
}

func TestProof_MembershipMultiLevelTree(t *testing.T) {
	tree := NewMutableTreeMem()
	n := 2000
	for i := 0; i < n; i++ {
		tree.Set(fmt.Appendf(nil, "ml%06d", i), fmt.Appendf(nil, "val%06d", i))
	}
	if tree.Height() < 2 {
		t.Fatalf("need height >= 2, got %d", tree.Height())
	}
	tree.SaveVersion()

	root := tree.Hash()

	for _, idx := range []int{0, 500, 1000, 1500, 1999} {
		key := fmt.Appendf(nil, "ml%06d", idx)
		val := fmt.Appendf(nil, "val%06d", idx)

		proof, err := tree.GetMembershipProof(key)
		if err != nil {
			t.Fatalf("GetMembershipProof(%s): %v", key, err)
		}
		ok := ics23.VerifyMembership(BptreeSpec, root, proof, key, val)
		if !ok {
			t.Fatalf("membership proof failed for %s (multi-level, height=%d)", key, tree.Height())
		}
	}
}

func TestProof_MembershipDBBacked(t *testing.T) {
	db := newTestDB(t)
	tree := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	for i := 0; i < 100; i++ {
		tree.Set(fmt.Appendf(nil, "db%04d", i), fmt.Appendf(nil, "val%04d", i))
	}
	tree.SaveVersion()

	// Reload from DB
	tree2 := NewMutableTreeWithDB(db, 1000, NewNopLogger())
	tree2.Load()

	root := tree2.Hash()

	for _, idx := range []int{0, 50, 99} {
		key := fmt.Appendf(nil, "db%04d", idx)
		val := fmt.Appendf(nil, "val%04d", idx)

		proof, err := tree2.GetMembershipProof(key)
		if err != nil {
			t.Fatalf("GetMembershipProof(%s) on loaded tree: %v", key, err)
		}
		ok := ics23.VerifyMembership(BptreeSpec, root, proof, key, val)
		if !ok {
			t.Fatalf("membership proof failed for %s (DB-loaded tree)", key)
		}
	}
}

// TestProof_UsesCommittedState verifies that immutableForProof() uses
// the last committed state (lastSaved), not the working tree. Proofs
// generated after Set() but before SaveVersion() must be based on the
// committed state and verifiable against MutableTree.Hash().
func TestProof_UsesCommittedState(t *testing.T) {
	t.Run("in-memory", func(t *testing.T) {
		testProofUsesCommittedState(t, func() *MutableTree {
			return NewMutableTreeMem()
		})
	})
	t.Run("db-backed", func(t *testing.T) {
		testProofUsesCommittedState(t, func() *MutableTree {
			db := newTestDB(t)
			return NewMutableTreeWithDB(db, 1000, NewNopLogger())
		})
	})
}

func testProofUsesCommittedState(t *testing.T, newTree func() *MutableTree) {
	t.Helper()
	tree := newTree()

	// Before any SaveVersion, proof generation should fail
	tree.Set([]byte("a"), []byte("1"))
	_, err := tree.GetMembershipProof([]byte("a"))
	if err != ErrNoCommittedState {
		t.Fatalf("expected ErrNoCommittedState before SaveVersion, got: %v", err)
	}

	// Commit version 1
	committedHash, _, err := tree.SaveVersion()
	if err != nil {
		t.Fatalf("SaveVersion: %v", err)
	}

	// Insert ("b", "2") WITHOUT SaveVersion — dirty the working tree
	tree.Set([]byte("b"), []byte("2"))

	// Working hash should differ from committed hash
	workingHash := tree.WorkingHash()
	if bytes.Equal(workingHash, committedHash) {
		t.Fatalf("working hash should differ from committed hash after uncommitted Set")
	}

	// Proof for committed key "a" must verify against the committed root hash,
	// NOT against the working hash
	proofA, err := tree.GetMembershipProof([]byte("a"))
	if err != nil {
		t.Fatalf("GetMembershipProof(a): %v", err)
	}
	if ok := ics23.VerifyMembership(BptreeSpec, committedHash, proofA, []byte("a"), []byte("1")); !ok {
		t.Fatalf("proof for committed key 'a' must verify against committed root hash")
	}

	// Proof for uncommitted key "b" should fail — it doesn't exist in lastSaved
	_, err = tree.GetMembershipProof([]byte("b"))
	if err == nil {
		t.Fatalf("GetMembershipProof(b) should fail for uncommitted key")
	}

	// After committing version 2, both keys should produce valid proofs
	tree.SaveVersion()
	hash2 := tree.Hash()

	proofA2, err := tree.GetMembershipProof([]byte("a"))
	if err != nil {
		t.Fatalf("GetMembershipProof(a) after second save: %v", err)
	}
	if ok := ics23.VerifyMembership(BptreeSpec, hash2, proofA2, []byte("a"), []byte("1")); !ok {
		t.Fatalf("proof for 'a' should verify after second save")
	}

	proofB2, err := tree.GetMembershipProof([]byte("b"))
	if err != nil {
		t.Fatalf("GetMembershipProof(b) after second save: %v", err)
	}
	if ok := ics23.VerifyMembership(BptreeSpec, hash2, proofB2, []byte("b"), []byte("2")); !ok {
		t.Fatalf("proof for 'b' should verify after second save")
	}
}

func TestProof_VerifyMethods(t *testing.T) {
	tree := NewMutableTreeMem()
	tree.Set([]byte("a"), []byte("1"))
	tree.Set([]byte("b"), []byte("2"))
	tree.SaveVersion()

	imm, err := tree.immutableForProof()
	if err != nil {
		t.Fatalf("immutableForProof: %v", err)
	}

	// VerifyMembership
	proof, _ := imm.GetMembershipProof([]byte("a"))
	ok, err := imm.VerifyMembership(proof, []byte("a"))
	if err != nil {
		t.Fatalf("VerifyMembership error: %v", err)
	}
	if !ok {
		t.Fatalf("VerifyMembership failed")
	}

	// VerifyNonMembership
	proof, _ = imm.GetNonMembershipProof([]byte("ab"))
	ok, err = imm.VerifyNonMembership(proof, []byte("ab"))
	if err != nil {
		t.Fatalf("VerifyNonMembership error: %v", err)
	}
	if !ok {
		t.Fatalf("VerifyNonMembership failed")
	}
}
