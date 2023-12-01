package iavl

import (
	"fmt"
	"strings"
)

// pathWithLeaf is a path to a leaf node and the leaf node itself.
type pathWithLeaf struct {
	Path PathToLeaf    `json:"path"`
	Leaf proofLeafNode `json:"leaf"`
}

func (pwl pathWithLeaf) String() string {
	return pwl.StringIndented("")
}

func (pwl pathWithLeaf) StringIndented(indent string) string {
	return fmt.Sprintf(`pathWithLeaf{
%s  Path: %v
%s  Leaf: %v
%s}`,
		indent, pwl.Path.stringIndented(indent+"  "),
		indent, pwl.Leaf.stringIndented(indent+"  "),
		indent)
}

// `computeRootHash` computes the root hash with leaf node.
// Does not verify the root hash.
func (pwl pathWithLeaf) computeRootHash() []byte {
	leafHash := pwl.Leaf.Hash()
	return pwl.Path.computeRootHash(leafHash)
}

// ----------------------------------------

// PathToLeaf represents an inner path to a leaf node.
// Note that the nodes are ordered such that the last one is closest
// to the root of the tree.
type PathToLeaf []proofInnerNode

func (pl PathToLeaf) String() string {
	return pl.stringIndented("")
}

func (pl PathToLeaf) stringIndented(indent string) string {
	if len(pl) == 0 {
		return "empty-PathToLeaf"
	}
	strs := make([]string, len(pl))
	for i, pin := range pl {
		if i == 20 {
			strs[i] = fmt.Sprintf("... (%v total)", len(pl))
			break
		}
		strs[i] = fmt.Sprintf("%v:%v", i, pin.stringIndented(indent+"  "))
	}
	return fmt.Sprintf(`PathToLeaf{
%s  %v
%s}`,
		indent, strings.Join(strs, "\n"+indent+"  "),
		indent)
}

// `computeRootHash` computes the root hash assuming some leaf hash.
// Does not verify the root hash.
func (pl PathToLeaf) computeRootHash(leafHash []byte) []byte {
	hash := leafHash
	for i := len(pl) - 1; i >= 0; i-- {
		pin := pl[i]
		hash = pin.Hash(hash)
	}
	return hash
}

func (pl PathToLeaf) isLeftmost() bool {
	for _, node := range pl {
		if len(node.Left) > 0 {
			return false
		}
	}
	return true
}

func (pl PathToLeaf) isRightmost() bool {
	for _, node := range pl {
		if len(node.Right) > 0 {
			return false
		}
	}
	return true
}

// returns -1 if invalid.
func (pl PathToLeaf) Index() (idx int64) {
	for i, node := range pl {
		if node.Left == nil {
			continue
		} else if node.Right == nil {
			if i < len(pl)-1 {
				idx += node.Size - pl[i+1].Size
			} else {
				idx += node.Size - 1
			}
		} else {
			return -1
		}
	}
	return idx
}
