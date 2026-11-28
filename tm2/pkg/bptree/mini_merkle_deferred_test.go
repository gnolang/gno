package bptree

import (
	"bytes"
	"fmt"
	"math/rand"
	"testing"

	ics23 "github.com/cosmos/ics23/go"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// poisonUnbuiltMiniMerkles overwrites the mini merkle of every cached LEAF
// that has not built one, leaving miniBuilt false. Any code that reads an
// unbuilt leaf's miniTree instead of going through Hash /
// builtMiniMerkleInto / Clone then produces a visibly wrong hash. (Inner mini merkles stay eager,
// so there is nothing deferred to poison there.)
//
// This is the enforcement for the deferred mini merkle: the rule ("a decoded
// node is immutable, and its mini merkle is derived on demand into a local")
// is not expressible in the type system, so it is pinned by making a
// violation loud.
func poisonUnbuiltMiniMerkles(ndb *nodeDB) int {
	var poison Hash
	for i := range poison {
		poison[i] = 0xA5
	}
	n := 0
	for _, k := range ndb.nodeCache.Keys() {
		node, ok := ndb.nodeCache.Peek(k)
		if !ok {
			continue
		}
		leaf, ok := node.(*LeafNode)
		if !ok || leaf.miniBuilt {
			continue
		}
		for i := range leaf.miniTree.tree {
			leaf.miniTree.tree[i] = poison
		}
		n++
	}
	return n
}

// TestDeferredMiniMerkle_PoisonedCache builds a tree, reloads it so every node
// comes back from the decoder with a deferred mini merkle, poisons those
// caches, and then exercises every surface that depends on node hashes.
func TestDeferredMiniMerkle_PoisonedCache(t *testing.T) {
	db := memdb.NewMemDB()
	tr := NewMutableTreeWithDB(db, 4096, NewNopLogger())
	r := rand.New(rand.NewSource(31))
	model := map[string][]byte{}
	keys := make([][]byte, 0, 2000)
	for i := range 2000 {
		k := fmt.Appendf(nil, "key-%06d-%04d", i, r.Intn(10000))
		v := fmt.Appendf(nil, "value-%d", i)
		mustSet(t, tr, k, v)
		model[string(k)] = v
		keys = append(keys, k)
	}
	ver := mustSave(t, tr)
	wantHash := append([]byte(nil), tr.Hash()...)
	tr.Close()

	// Fresh tree: every node is decoded, so every mini merkle is deferred.
	tr2 := NewMutableTreeWithDB(db, 4096, NewNopLogger())
	defer tr2.Close()
	if _, err := tr2.LoadVersion(ver); err != nil {
		t.Fatalf("LoadVersion: %v", err)
	}
	// Touch every key so the whole tree is resident in the node cache.
	for _, k := range keys {
		if _, err := tr2.Get(k); err != nil {
			t.Fatal(err)
		}
	}
	poisoned := poisonUnbuiltMiniMerkles(tr2.ndb)
	if poisoned < 10 {
		t.Fatalf("only %d cached nodes had a deferred mini merkle; "+
			"the test is not exercising the deferred path", poisoned)
	}
	t.Logf("poisoned %d deferred mini merkles", poisoned)

	if !bytes.Equal(tr2.Hash(), wantHash) {
		t.Fatalf("Hash() = %x, want %x", tr2.Hash(), wantHash)
	}
	if !bytes.Equal(tr2.WorkingHash(), wantHash) {
		t.Fatalf("WorkingHash() = %x, want %x", tr2.WorkingHash(), wantHash)
	}

	// Proofs must verify against the committed root.
	imm, err := tr2.GetImmutable(ver)
	if err != nil {
		t.Fatalf("GetImmutable: %v", err)
	}
	defer imm.Close()
	if !bytes.Equal(imm.Hash(), wantHash) {
		t.Fatalf("imm.Hash() = %x, want %x", imm.Hash(), wantHash)
	}
	for _, k := range []([]byte){keys[0], keys[len(keys)/2], keys[len(keys)-1]} {
		proof, err := tr2.GetMembershipProof(k)
		if err != nil {
			t.Fatalf("GetMembershipProof(%s): %v", k, err)
		}
		if !ics23.VerifyMembership(BptreeSpec, wantHash, proof, k, model[string(k)]) {
			t.Fatalf("VerifyMembership(%s) failed", k)
		}
	}
	for _, k := range []([]byte){[]byte("absent-aaa"), []byte("zzzzzz")} {
		proof, err := tr2.GetNonMembershipProof(k)
		if err != nil {
			t.Fatalf("GetNonMembershipProof(%s): %v", k, err)
		}
		if !ics23.VerifyNonMembership(BptreeSpec, wantHash, proof, k) {
			t.Fatalf("VerifyNonMembership(%s) failed", k)
		}
	}

	// Writing on top of poisoned nodes: each mutated node is cloned, and the
	// clone must build its mini merkle from the node's own inputs rather than
	// inherit the poison. The resulting root must match a tree that was
	// never poisoned.
	ref := NewMutableTreeWithDB(memdb.NewMemDB(), 4096, NewNopLogger())
	defer ref.Close()
	for _, k := range keys {
		mustSet(t, ref, k, model[string(k)])
	}
	mustSave(t, ref)

	for i := range 300 {
		k := keys[r.Intn(len(keys))]
		v := fmt.Appendf(nil, "rewritten-%d", i)
		mustSet(t, tr2, k, v)
		mustSet(t, ref, k, v)
		model[string(k)] = v
	}
	// Also delete some keys so the underflow paths run over poisoned nodes.
	for range 200 {
		idx := r.Intn(len(keys))
		k := keys[idx]
		if _, ok := model[string(k)]; !ok {
			continue
		}
		if _, _, err := tr2.Remove(k); err != nil {
			t.Fatal(err)
		}
		if _, _, err := ref.Remove(k); err != nil {
			t.Fatal(err)
		}
		delete(model, string(k))
	}
	if !bytes.Equal(tr2.WorkingHash(), ref.WorkingHash()) {
		t.Fatalf("poisoned tree working hash %x != clean tree %x", tr2.WorkingHash(), ref.WorkingHash())
	}
	h2, v2, err := tr2.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}
	hRef, _, err := ref.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(h2, hRef) {
		t.Fatalf("poisoned tree saved hash %x != clean tree %x", h2, hRef)
	}

	// And it must still reload to the same hash from disk.
	tr3 := NewMutableTreeWithDB(db, 0, NewNopLogger())
	defer tr3.Close()
	if _, err := tr3.LoadVersion(v2); err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(tr3.Hash(), h2) {
		t.Fatalf("reloaded hash %x != %x", tr3.Hash(), h2)
	}
	for k, want := range model {
		got, err := tr3.Get([]byte(k))
		if err != nil || !bytes.Equal(got, want) {
			t.Fatalf("Get(%s) = %q, %v; want %q", k, got, err, want)
		}
	}
}

// TestDeferredMiniMerkle_DecodedNodeIsUnbuilt pins the decode-side behaviour
// the whole optimisation rests on, and that Hash() still answers correctly
// for such a node.
func TestDeferredMiniMerkle_DecodedNodeIsUnbuilt(t *testing.T) {
	leaf := makeLeaf([]string{"a", "b", "c"})
	want := leaf.Hash()
	rec, err := leaf.appendSerialized(nil)
	if err != nil {
		t.Fatal(err)
	}
	node, err := ReadNode(&NodeKey{Version: 1, Nonce: 1}, rec)
	if err != nil {
		t.Fatal(err)
	}
	decoded := node.(*LeafNode)
	if decoded.miniBuilt {
		t.Fatal("decoded leaf has a built mini merkle; the deferral is gone")
	}
	if got := decoded.Hash(); got != want {
		t.Fatalf("deferred Hash() = %x, want %x", got, want)
	}
	if decoded.miniBuilt {
		t.Fatal("Hash() built the mini merkle in place; it must fold into a local")
	}
	clone := decoded.Clone()
	if !clone.miniBuilt {
		t.Fatal("Clone did not build the mini merkle")
	}
	if got := clone.Hash(); got != want {
		t.Fatalf("clone Hash() = %x, want %x", got, want)
	}
	if decoded.miniBuilt {
		t.Fatal("Clone built the mini merkle on the ORIGINAL; the shared node must stay immutable")
	}
	// builtMiniMerkleInto must agree with the built form at every level.
	var built MiniMerkle
	decoded.builtMiniMerkleInto(&built)
	ref := decoded.Clone()
	if built.tree != ref.miniTree.tree {
		t.Fatal("builtMiniMerkleInto differs from the built mini merkle")
	}
}
