package grc6150

import (
	"std"
	"testing"

	"gno.land/r/demo/users"
)

var (
	parentTokenId = uint256(1)
	childTokenId  = uint256(4)
	treeNFT       = map[uint256]uint256{
		1: 1,
		2: 1,
		3: 1,
		4: 1,
	}
)

func TestTreeNFTCreation(t *testing.T) {
	dummy := NewTreeNFT()
	if dummy == nil {
		t.Errorf("should not be nil")
	}
}
