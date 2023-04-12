package state

import (
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

//------------------------------------------------------
// blockchain services types
// NOTE: Interfaces used by RPC must be thread safe!
//------------------------------------------------------

//------------------------------------------------------
// blockstore

// BlockStoreRPC is the block store interface used by the RPC.
type BlockStoreRPC interface {
	Height() (int64, error)

	LoadBlockMeta(height int64) (*types.BlockMeta, error)
	LoadBlock(height int64) (*types.Block, error)
	LoadBlockPart(height int64, index int) (*types.Part, error)

	LoadBlockCommit(height int64) (*types.Commit, error)
	LoadSeenCommit(height int64) (*types.Commit, error)
}

// BlockStore defines the BlockStore interface used by the ConsensusState.
type BlockStore interface {
	BlockStoreRPC
	SaveBlock(block *types.Block, blockParts *types.PartSet, seenCommit *types.Commit) error
}
