package merkle

import (
	"bytes"
	"testing"

	"github.com/gnolang/gno/tm2/pkg/crypto/tmhash"
)

// computeHashFromAunts signalled "I could not compute a root" by returning nil,
// and Verify fed that straight into bytes.Equal against the caller's expected
// root. bytes.Equal(nil, []byte{}) is true, so a proof the code had already
// given up on verified against an empty root — any leaf, any garbage proof.
// Every case below must be rejected.
func TestSimpleProofVerifyRejectsUncomputableProofs(t *testing.T) {
	t.Parallel()
	leaf := []byte("arbitrary leaf")

	tests := []struct {
		name         string
		root         []byte
		index, total int
		aunts        [][]byte
	}{
		{"total 0, empty root", []byte{}, 0, 0, nil},
		{"total 0, nil root", nil, 0, 0, nil},
		{"index beyond total, empty root", []byte{}, 5, 3, nil},
		{"index equal to total, nil root", nil, 1, 1, nil},
		{"multi-leaf tree with no aunts, empty root", []byte{}, 0, 4, nil},
		{"single-leaf tree with a stray aunt, empty root", []byte{}, 0, 1, [][]byte{make([]byte, tmhash.Size)}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			sp := &SimpleProof{
				Total:    tt.total,
				Index:    tt.index,
				LeafHash: leafHash(leaf),
				Aunts:    tt.aunts,
			}
			if err := sp.Verify(tt.root, leaf); err == nil {
				t.Errorf("Verify accepted an uncomputable proof against a %d-byte root", len(tt.root))
			}
		})
	}
}

// A root that cannot be computed is reported as an error, never as an empty
// byte slice a caller might go on to compare against something.
func TestComputeRootHashErrorsInsteadOfReturningNil(t *testing.T) {
	t.Parallel()
	sp := &SimpleProof{Total: 0, Index: 0, LeafHash: leafHash([]byte("leaf"))}
	root, err := sp.ComputeRootHash()
	if err == nil {
		t.Fatalf("ComputeRootHash succeeded on a 0-total proof, returning %v", root)
	}
	if root != nil {
		t.Errorf("ComputeRootHash returned %v alongside an error; want nil", root)
	}
}

// The same collision had a second call site: SimpleValueOp.Run passed
// ComputeRootHash's result up to ProofOperators.Verify, which compares it
// against the expected root. The leaf hash below has to match what Run
// recomputes, or Run bails before ever reaching the root computation.
func TestSimpleValueOpRunRejectsUncomputableProof(t *testing.T) {
	t.Parallel()
	key, value := []byte("key"), []byte("value")

	bz := new(bytes.Buffer)
	encodeByteSlice(bz, key)
	encodeByteSlice(bz, tmhash.Sum(value))
	op := NewSimpleValueOp(key, &SimpleProof{
		Total:    0,
		Index:    0,
		LeafHash: leafHash(bz.Bytes()),
	})

	if _, err := op.Run([][]byte{value}); err == nil {
		t.Error("Run accepted a proof whose root cannot be computed")
	}
}
