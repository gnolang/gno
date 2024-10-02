package merkle

import (
	"fmt"
	"testing"
)

type testData struct {
	content string
}

func (d testData) Bytes() []byte {
	return []byte(d.content)
}

func TestMerkleTree(t *testing.T) {
	tests := []struct {
		size     int
		expected string
	}{
		{
			size:     1,
			expected: "cf9f824bce7f5bc63d557b23591f58577f53fe29f974a615bdddbd0140f912f4",
		},
		{
			size:     3,
			expected: "1a4a5f0fa267244bf9f74a63fdf2a87eed5e97e4bd104a9e94728c8fb5442177",
		},
		{
			size:     10,
			expected: "cd8a40502b0b92bf58e7432a5abb2d8b60121cf2b7966d6ebaf103f907a1bc21",
		},
		{
			size:     1000,
			expected: "fa533d2efdf12be26bc410dfa42936ac63361324e35e9b1ff54d422a1dd2388b",
		},
	}

	for _, test := range tests {
		var leaves []Hashable
		for i := 0; i < test.size; i++ {
			leaves = append(leaves, testData{fmt.Sprintf("node_%d", i)})
		}

		tree := NewTree(leaves)

		if tree == nil {
			t.Error("Merkle tree creation failed")
		}

		root := tree.Root()

		if root != test.expected {
			t.Fatalf("merkle.Tree.Root(), expected: %s; got: %s", test.expected, root)
		}

		for _, leaf := range leaves {
			proofs, err := tree.Proof(leaf)
			if err != nil {
				t.Fatal("failed to proof leaf: %v, on tree: %v", leaf, test)
			}

			ok := Verify(root, leaf, proofs)
			if !ok {
				t.Fatal("failed to verify leaf: %v, on tree: %v", leaf, tree)
			}
		}
	}
}
