package iavl

import (
	"testing"

	ics23 "github.com/cosmos/ics23/go"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"
)

// VerifyMembership / VerifyNonMembership must reject a nil / nil-inner /
// wrong-type proof with an error, never nil-deref inside ics23. Mirrors the
// bptree wrapper guard (TestVerifyNonMembership_RejectsNilInnerProof). A
// nil-inner proof is not reachable from a wire-decoded proof, but a hand-built
// one would panic without the guard.
func TestIavlVerify_RejectsNilInnerProof(t *testing.T) {
	tree := NewMutableTree(memdb.NewMemDB(), 0, false, NewNopLogger())
	if _, err := tree.Set([]byte("b"), []byte("vb")); err != nil {
		t.Fatal(err)
	}
	if _, err := tree.Set([]byte("d"), []byte("vd")); err != nil {
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

	// nil-inner and nil-proof both nil-deref-panic in ics23 pre-guard; empty and
	// wrong-type don't panic but must still be classified as the wrong proof kind.
	malformed := map[string]*ics23.CommitmentProof{
		"nil-inner Exist":    {Proof: &ics23.CommitmentProof_Exist{Exist: nil}},
		"nil-inner Nonexist": {Proof: &ics23.CommitmentProof_Nonexist{Nonexist: nil}},
		"nil proof":          nil,
		"empty proof":        {},
	}
	for name, proof := range malformed {
		t.Run("membership/"+name, func(t *testing.T) {
			ok, verr := imm.VerifyMembership(proof, []byte("b"))
			if ok || verr == nil {
				t.Fatalf("malformed proof: ok=%v err=%v, want (false, error)", ok, verr)
			}
		})
		t.Run("nonmembership/"+name, func(t *testing.T) {
			ok, verr := imm.VerifyNonMembership(proof, []byte("c"))
			if ok || verr == nil {
				t.Fatalf("malformed proof: ok=%v err=%v, want (false, error)", ok, verr)
			}
		})
	}

	// A genuine membership proof for an existing key still verifies.
	good, err := imm.GetMembershipProof([]byte("b"))
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := imm.VerifyMembership(good, []byte("b")); err != nil || !ok {
		t.Fatalf("genuine membership proof must verify: ok=%v err=%v", ok, err)
	}

	// A genuine non-existence proof for an absent key still verifies.
	goodNon, err := imm.GetNonMembershipProof([]byte("c"))
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := imm.VerifyNonMembership(goodNon, []byte("c")); err != nil || !ok {
		t.Fatalf("genuine non-membership proof must verify: ok=%v err=%v", ok, err)
	}
}
