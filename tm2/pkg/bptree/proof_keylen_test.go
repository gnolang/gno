package bptree

import (
	"bytes"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/db/memdb"

	ics23 "github.com/cosmos/ics23/go"
)

// Long keys: the leaf hash length-prefixes the key with a uvarint
// (HashLeafSlotFromValueHash), which turns multi-byte above 127 bytes, and
// ics23 VAR_PROTO recomputes it independently at verification. No other proof
// test exceeds 20-byte keys, so the multi-byte agreement is pinned here.
func TestProof_LongKeys(t *testing.T) {
	tree := NewMutableTreeWithDB(memdb.NewMemDB(), 0, NewNopLogger())
	keys := [][]byte{
		bytes.Repeat([]byte{'a'}, 127), // longest 1-byte uvarint
		bytes.Repeat([]byte{'b'}, 128), // first 2-byte uvarint
		bytes.Repeat([]byte{'c'}, 300),
	}
	for _, k := range keys {
		if _, err := tree.Set(k, []byte("x")); err != nil {
			t.Fatal(err)
		}
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

	for _, k := range keys {
		proof, err := imm.GetMembershipProof(k)
		if err != nil {
			t.Fatalf("membership proof (len %d): %v", len(k), err)
		}
		if !ics23.VerifyMembership(BptreeSpec, hash, proof, k, []byte("x")) {
			t.Fatalf("membership proof for %d-byte key does not verify", len(k))
		}
	}

	// Interior gap whose both neighbors are long keys.
	absent := append(bytes.Repeat([]byte{'b'}, 128), '!')
	proof, err := imm.GetNonMembershipProof(absent)
	if err != nil {
		t.Fatalf("non-membership proof: %v", err)
	}
	if !ics23.VerifyNonMembership(BptreeSpec, hash, proof, absent) {
		t.Fatal("non-membership proof between long keys does not verify")
	}
}
