package grc6150

import "std"

type IGRC6150 interface {
	ParentOf(tid TokenID) error
	ChildrenOf(tid TokenID) error
	IsRoot(tid TokenID) error
	IsLeaf(tid TokenID) error
}

type IGRC6150 struct {
	parentOf             map[uint256]uint256
	childrenOf           map[uint256][]uint256
	indexInChildrenArray map[uint256]uint256
}
