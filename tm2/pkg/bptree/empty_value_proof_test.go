package bptree

import (
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"

	ics23 "github.com/cosmos/ics23/go"
)

// Empty values are storable (the store layer rejects only nil, and IAVL
// accepts them) but unprovable: ics23's LeafOp.Apply rejects len(value)==0,
// so no existence proof over an empty value can verify — in this tree or in
// IAVL. This is the value-side twin of the M24 empty-key constraint. This
// test pins the current behavior: GetMembershipProof succeeds but the proof
// does not verify. If GetMembershipProof is later changed to return a defined
// error for empty values, update this test to assert that instead.
func TestProof_EmptyValueUnprovable(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 0, NewNopLogger())
	if _, err := tree.Set([]byte("a"), []byte{}); err != nil {
		t.Fatalf("Set empty value: %v", err)
	}
	if _, err := tree.Set([]byte("b"), []byte("x")); err != nil {
		t.Fatal(err)
	}
	hash, _, err := tree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}
	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatal(err)
	}
	defer imm.Close()

	// Empty-valued key: proof generation succeeds, verification cannot.
	proof, err := imm.GetMembershipProof([]byte("a"))
	if err != nil {
		t.Fatalf("GetMembershipProof: %v", err)
	}
	if ics23.VerifyMembership(BptreeSpec, hash, proof, []byte("a"), []byte{}) {
		t.Fatal("empty-value membership proof verified — ics23 lifted its LeafOp empty-value rejection? update the harness/proof docs")
	}

	// The non-empty sibling in the same tree proves fine.
	proof, err = imm.GetMembershipProof([]byte("b"))
	if err != nil {
		t.Fatalf("GetMembershipProof(b): %v", err)
	}
	if !ics23.VerifyMembership(BptreeSpec, hash, proof, []byte("b"), []byte("x")) {
		t.Fatal("membership proof for non-empty sibling does not verify")
	}
}
