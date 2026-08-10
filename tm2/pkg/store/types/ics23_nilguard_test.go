package types

import (
	"testing"

	ics23 "github.com/cosmos/ics23/go"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
	"github.com/gnolang/gno/tm2/pkg/iavl"
)

// CommitmentOp.Run (the shared iavl / simple-merkle op) calls
// op.Proof.Calculate() and ics23.Verify*, both of which nil-deref on a
// hand-built nil-inner proof. The guard must turn that into an error (never a
// panic) for either args length. Not reachable from a wire-decoded proof
// (the decoder's Unmarshal always allocates the inner); defense-in-depth.
func TestCommitmentOpRun_RejectsNilInnerProof(t *testing.T) {
	malformed := map[string]*ics23.CommitmentProof{
		"nil-inner Exist":    {Proof: &ics23.CommitmentProof_Exist{Exist: nil}},
		"nil-inner Nonexist": {Proof: &ics23.CommitmentProof_Nonexist{Nonexist: nil}},
		"nil proof":          nil,
		"empty proof":        {},
	}
	for name, proof := range malformed {
		op := NewIavlCommitmentOp([]byte("k"), proof)
		// args len 0 (absence) and len 1 (existence) both hit Calculate() first.
		for _, args := range [][][]byte{nil, {[]byte("v")}} {
			res, err := op.Run(args)
			if err == nil || res != nil {
				t.Fatalf("%s args=%d: res=%v err=%v, want (nil, error)", name, len(args), res, err)
			}
		}
	}

	// A genuine IavlSpec proof still runs to a root through the same path, so the
	// guard is not over-rejecting. Uses iavl (cycle-free: iavl does not import
	// store/types) to mint a real proof the shared op can verify.
	tree := iavl.NewMutableTree(memdb.NewMemDB(), 0, false, iavl.NewNopLogger())
	if _, err := tree.Set([]byte("b"), []byte("vb")); err != nil {
		t.Fatal(err)
	}
	_, version, err := tree.SaveVersion()
	if err != nil {
		t.Fatal(err)
	}
	imm, err := tree.GetImmutable(version)
	if err != nil {
		t.Fatal(err)
	}

	exist, err := imm.GetMembershipProof([]byte("b"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewIavlCommitmentOp([]byte("b"), exist).Run([][]byte{[]byte("vb")}); err != nil {
		t.Fatalf("genuine membership proof must run: %v", err)
	}

	nonexist, err := imm.GetNonMembershipProof([]byte("c"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := NewIavlCommitmentOp([]byte("c"), nonexist).Run(nil); err != nil {
		t.Fatalf("genuine non-membership proof must run: %v", err)
	}
}
