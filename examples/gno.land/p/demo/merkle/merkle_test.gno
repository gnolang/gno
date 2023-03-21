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
	var leaves []Hashable
	for i := 0; i < 10; i++ {
		leaves = append(leaves, testData{fmt.Sprintf("node_%d", i)})
	}

	tree := NewTree(leaves)

	if tree == nil {
		t.Error("Merkle tree creation failed")
	}

	root := tree.Root()

	expected := "cd8a40502b0b92bf58e7432a5abb2d8b60121cf2b7966d6ebaf103f907a1bc21"
	if root != expected {
		t.Fatalf("merkle.Tree.Root(), expected: %s; got: %s", expected, root)
	}

	proofs, err := tree.Proof(leaves[3])
	if err != nil {
		t.Fatal("err: ", err)
	}

	ok := Verify(root, leaves[3], proofs)
	if !ok {
		t.Fatal("failed to verify merkle tree")
	}
}
