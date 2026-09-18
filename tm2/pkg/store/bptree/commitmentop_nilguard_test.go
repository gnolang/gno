package bptree

import (
	"testing"

	ics23 "github.com/cosmos/ics23/go"

	bp "github.com/gnolang/gno/tm2/pkg/bptree"
	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// CommitmentOp.Run calls op.Proof.Calculate() and ics23.Verify*, both of which
// nil-deref on a hand-built nil-inner proof. The guard must turn that into an
// error (never a panic) for either args length, while leaving genuine proofs
// untouched. Not reachable from a wire-decoded proof; defense-in-depth.
func TestBptreeCommitmentOpRun_RejectsNilInnerProof(t *testing.T) {
	malformed := map[string]*ics23.CommitmentProof{
		"nil-inner Exist":    {Proof: &ics23.CommitmentProof_Exist{Exist: nil}},
		"nil-inner Nonexist": {Proof: &ics23.CommitmentProof_Nonexist{Nonexist: nil}},
		"nil proof":          nil,
		"empty proof":        {},
	}
	for name, proof := range malformed {
		op := NewBptreeCommitmentOp([]byte("k"), proof)
		// args len 0 (absence) and len 1 (existence) both hit Calculate() first.
		for _, args := range [][][]byte{nil, {[]byte("v")}} {
			res, err := op.Run(args)
			if err == nil || res != nil {
				t.Fatalf("%s args=%d: res=%v err=%v, want (nil, error)", name, len(args), res, err)
			}
		}
	}

	// A genuine proof still runs to a root through the same path.
	tree := bp.NewMutableTreeWithDB(memdb.NewMemDB(), 0, bp.NewNopLogger())
	tree.Set([]byte("b"), []byte("vb"))
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatal(err)
	}
	defer imm.Close()

	exist, err := imm.GetMembershipProof([]byte("b"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewBptreeCommitmentOp([]byte("b"), exist).Run([][]byte{[]byte("vb")}); err != nil {
		t.Fatalf("genuine membership proof must run: %v", err)
	}

	nonexist, err := imm.GetNonMembershipProof([]byte("c"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewBptreeCommitmentOp([]byte("c"), nonexist).Run(nil); err != nil {
		t.Fatalf("genuine non-membership proof must run: %v", err)
	}
}
