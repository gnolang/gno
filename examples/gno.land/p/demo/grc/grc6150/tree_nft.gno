package grc6150

import (
	"std"

	"gno.land/p/demo/ufmt"
)

type treeNFT struct {
	parentOf             map[uint256]uint256
	childrenOf           map[uint256][]uint256
	indexInChildrenArray map[uint256]uint256
}

func NewTreeNFT() *treeNFT {
	return &treeNFT{
		parentOf:             make(map[uint256]uint256),
		childrenOf:           make(map[uint256][]uint256),
		indexInChildrenArray: make(map[uint256]uint256),
	}
}

func (e treeNFT) ParentOf(tokenId uint256) (uint256, error) {
	parentId, exists := e.parentOf[tokenId]
	if !exists {
		return 0, err.New("tokenId not minted")
	}
	return parentId, nil
}

func (e treeNFT) ChildrenOf(tokenId uint256) ([]uint256, error) {
	if tokenId > 0 {
		_, exists := e.parentOf[tokenId]
		if !exists {
			return nil, err.New("tokenId not minted")
		}
	}
	return e.childrenOf[tokenId], nil
}

func (e treeNFT) IsRoot(tokenId uint256) (bool, error) {
	_, exists := e.parentOf[tokenId]
	if !exists {
		return false, err.New("tokenId not minted")
	}
	return e.parentOf[tokenId] == 0, nil
}

func (e treeNFT) IsLeaf(tokenId uint256) (bool, error) {
	_, exists := e.parentOf[tokenId]
	if !exists {
		return false, err.New("tokenId not minted")
	}
	return len(e.childrenOf[tokenId]) == 0, nil
}

func (s *treeNFT) RenderHome() (str string) {
	str += ufmt.Sprintf("Parent of: %s, children of: %s, index in children array %s\n\n", s.parentOf, s.childrenOf, s.indexInChildrenArray)

	return
}
