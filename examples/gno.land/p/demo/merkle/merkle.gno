package merkle

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"errors"
)

type Hashable interface {
	Bytes() []byte
}

type nodes []Node

type Node struct {
	Hash []byte

	Position uint8
}

func (n Node) GetHexEncodedHash() string {
	return hex.EncodeToString(n.Hash[:])
}

type Tree struct {
	layers []nodes
}

// Root return the merkle root of the tree
func (t *Tree) Root() string {
	for _, l := range t.layers {
		if len(l) == 1 {
			return l[0].GetHexEncodedHash()
		}
	}
	return ""
}

// NewTree create a new Merkle Tree
func NewTree(data []Hashable) *Tree {
	tree := &Tree{}

	leaves := make([]Node, len(data))

	for i, d := range data {
		hash := sha256.Sum256(d.Bytes())
		leaves[i] = Node{Hash: hash[:]}
	}

	tree.layers = []nodes{nodes(leaves)}

	var buff bytes.Buffer
	for len(leaves) > 1 {
		level := make([]Node, 0, len(leaves)/2+1)
		for i := 0; i < len(leaves); i += 2 {
			buff.Reset()

			if i < len(leaves)-1 {
				buff.Write(leaves[i].Hash)
				buff.Write(leaves[i+1].Hash)
				hash := sha256.Sum256(buff.Bytes())
				level = append(level, Node{
					Hash: hash[:],
				})
			} else {
				level = append(level, leaves[i])
			}
		}
		leaves = level
		tree.layers = append(tree.layers, level)
	}
	return tree
}

// Proof return a MerkleProof
func (t *Tree) Proof(data Hashable) ([]Node, error) {
	targetHash := sha256.Sum256(data.Bytes())
	targetIndex := -1

	for i, layer := range t.layers[0] {
		if bytes.Equal(targetHash[:], layer.Hash) {
			targetIndex = i
			break
		}
	}

	if targetIndex == -1 {
		return nil, errors.New("target not found")
	}

	proofs := make([]Node, 0, len(t.layers))

	for _, layer := range t.layers {
		var pairIndex int

		if targetIndex%2 == 0 {
			pairIndex = targetIndex + 1
		} else {
			pairIndex = targetIndex - 1
		}
		if pairIndex < len(layer) {
			proofs = append(proofs, Node{
				Hash:     layer[pairIndex].Hash,
				Position: uint8(targetIndex) % 2,
			})
		}
		targetIndex /= 2
	}
	return proofs, nil
}

// Verify if a merkle proof is valid
func (t *Tree) Verify(leaf Hashable, proofs []Node) bool {
	return Verify(t.Root(), leaf, proofs)
}

// Verify if a merkle proof is valid
func Verify(root string, leaf Hashable, proofs []Node) bool {
	hash := sha256.Sum256(leaf.Bytes())

	for i := 0; i < len(proofs); i += 1 {
		var h []byte
		if proofs[i].Position == 0 {
			h = append(hash[:], proofs[i].Hash...)
		} else {
			h = append(proofs[i].Hash, hash[:]...)
		}
		hash = sha256.Sum256(h)
	}
	return hex.EncodeToString(hash[:]) == root
}
