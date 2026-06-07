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
	hash []byte

	position uint8
}

func NewNode(hash []byte, position uint8) Node {
	return Node{
		hash:     hash,
		position: position,
	}
}

func (n Node) Position() uint8 {
	return n.position
}

func (n Node) Hash() string {
	return hex.EncodeToString(n.hash[:])
}

type Tree struct {
	layers []nodes
}

// Root return the merkle root of the tree
func (t *Tree) Root() string {
	for _, l := range t.layers {
		if len(l) == 1 {
			return l[0].Hash()
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
		leaves[i] = Node{hash: hash[:]}
	}

	tree.layers = []nodes{nodes(leaves)}

	var buff bytes.Buffer
	for len(leaves) > 1 {
		level := make([]Node, 0, len(leaves)/2+1)
		for i := 0; i < len(leaves); i += 2 {
			buff.Reset()

			if i < len(leaves)-1 {
				buff.Write(leaves[i].hash)
				buff.Write(leaves[i+1].hash)
				hash := sha256.Sum256(buff.Bytes())
				level = append(level, Node{
					hash: hash[:],
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
		if bytes.Equal(targetHash[:], layer.hash) {
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
				hash:     layer[pairIndex].hash,
				position: uint8(targetIndex) % 2,
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
		if proofs[i].position == 0 {
			h = append(hash[:], proofs[i].hash...)
		} else {
			h = append(proofs[i].hash, hash[:]...)
		}
		hash = sha256.Sum256(h)
	}
	return hex.EncodeToString(hash[:]) == root
}
