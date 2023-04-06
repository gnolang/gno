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
	Height() int64

	LoadBlockMeta(height int64) *types.BlockMeta
	LoadBlock(height int64) *types.Block
	LoadBlockPart(height int64, index int) *types.Part

	LoadBlockCommit(height int64) *types.Commit
	LoadSeenCommit(height int64) *types.Commit
}

// BlockStore defines the BlockStore interface used by the ConsensusState.
type BlockStore interface {
	BlockStoreRPC
	SaveBlock(block *types.Block, blockParts *types.PartSet, seenCommit *types.Commit)
}
