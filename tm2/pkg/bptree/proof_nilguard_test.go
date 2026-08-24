package bptree

import (
	"testing"

	ics23 "github.com/cosmos/ics23/go"
)

// R2c-1: VerifyNonMembership must reject a nil-inner / nil / wrong-type proof
// with an error, not nil-deref panic inside ics23. Mirrors VerifyMembership's
// GetExist guard.
func TestVerifyNonMembership_RejectsNilInnerProof(t *testing.T) {
	tree := newMemTree()
	tree.Set([]byte("b"), []byte("vb"))
	tree.Set([]byte("d"), []byte("vd"))
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatal(err)
	}
	defer imm.Close()

	cases := []struct {
		name  string
		proof *ics23.CommitmentProof
	}{
		// nil-inner and nil-proof both nil-deref-panic in ics23 pre-guard.
		{"nil-inner Nonexist", &ics23.CommitmentProof{Proof: &ics23.CommitmentProof_Nonexist{Nonexist: nil}}},
		{"nil proof", nil},
		// empty / wrong-type don't panic pre-guard (ics23 returns false), but
		// the guard should still classify them as not-a-non-existence-proof.
		{"empty proof", &ics23.CommitmentProof{}},
		{"wrong-type (Exist) proof", &ics23.CommitmentProof{Proof: &ics23.CommitmentProof_Exist{Exist: &ics23.ExistenceProof{}}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Must return (false, error) — never panic.
			ok, verr := imm.VerifyNonMembership(tc.proof, []byte("c"))
			if ok || verr == nil {
				t.Fatalf("malformed proof: ok=%v err=%v, want (false, error)", ok, verr)
			}
		})
	}

	// A genuine non-existence proof for an absent key still verifies.
	good, err := imm.GetNonMembershipProof([]byte("c"))
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := imm.VerifyNonMembership(good, []byte("c")); err != nil || !ok {
		t.Fatalf("genuine non-membership proof must verify: ok=%v err=%v", ok, err)
	}
}

// Membership-side twin: VerifyMembership must reject a nil-inner / nil /
// wrong-type proof with an error (the GetExist guard), and still verify a
// genuine existence proof. Mirrors the iavl wrapper's both-direction coverage.
func TestVerifyMembership_RejectsNilInnerProof(t *testing.T) {
	tree := newMemTree()
	tree.Set([]byte("b"), []byte("vb"))
	if _, _, err := tree.SaveVersion(); err != nil {
		t.Fatal(err)
	}
	imm, err := tree.GetImmutable(1)
	if err != nil {
		t.Fatal(err)
	}
	defer imm.Close()

	cases := []struct {
		name  string
		proof *ics23.CommitmentProof
	}{
		{"nil-inner Exist", &ics23.CommitmentProof{Proof: &ics23.CommitmentProof_Exist{Exist: nil}}},
		{"nil proof", nil},
		{"empty proof", &ics23.CommitmentProof{}},
		{"wrong-type (Nonexist) proof", &ics23.CommitmentProof{Proof: &ics23.CommitmentProof_Nonexist{Nonexist: &ics23.NonExistenceProof{}}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			ok, verr := imm.VerifyMembership(tc.proof, []byte("b"))
			if ok || verr == nil {
				t.Fatalf("malformed proof: ok=%v err=%v, want (false, error)", ok, verr)
			}
		})
	}

	// A genuine existence proof for a present key still verifies.
	good, err := imm.GetMembershipProof([]byte("b"))
	if err != nil {
		t.Fatal(err)
	}
	if ok, err := imm.VerifyMembership(good, []byte("b")); err != nil || !ok {
		t.Fatalf("genuine membership proof must verify: ok=%v err=%v", ok, err)
	}
}
