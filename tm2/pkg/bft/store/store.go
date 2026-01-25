package store

import (
	"fmt"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/amino"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/errors"
)

/*
BlockStore is a simple low level store for blocks.

There are three types of information stored:
  - BlockMeta:   Meta information about each block
  - Block part:  Parts of each block, aggregated w/ PartSet
  - Commit:      The commit part of each block, for gossiping precommit votes

Currently the precommit signatures are duplicated in the Block parts as
well as the Commit.  In the future this may change, perhaps by moving
the Commit data outside the Block. (TODO)

// NOTE: BlockStore methods will panic if they encounter errors
// deserializing loaded data, indicating probable corruption on disk.
*/
type BlockStore struct {
	db dbm.DB

	mtx    sync.RWMutex
	height int64
}

// NewBlockStore returns a new BlockStore with the given DB,
// initialized to the last height that was committed to the DB.
func NewBlockStore(db dbm.DB) *BlockStore {
	bsjson := LoadBlockStoreStateJSON(db)
	return &BlockStore{
		height: bsjson.Height,
		db:     db,
	}
}

// Height returns the last known contiguous block height.
func (bs *BlockStore) Height() int64 {
	bs.mtx.RLock()
	defer bs.mtx.RUnlock()
	return bs.height
}

// NewBatch returns a new database batch for grouping block store writes.
func (bs *BlockStore) NewBatch() dbm.Batch {
	return bs.db.NewBatch()
}

// LoadBlock returns the block with the given height.
// If no block is found for that height, it returns nil.
func (bs *BlockStore) LoadBlock(height int64) *types.Block {
	blockMeta := bs.LoadBlockMeta(height)
	if blockMeta == nil {
		return nil
	}

	block := new(types.Block)
	buf := []byte{}
	for i := range blockMeta.BlockID.PartsHeader.Total {
		part := bs.LoadBlockPart(height, i)
		buf = append(buf, part.Bytes...)
	}
	err := amino.UnmarshalSized(buf, block)
	if err != nil {
		// NOTE: The existence of meta should imply the existence of the
		// block. So, make sure meta is only saved after blocks are saved.
		panic(errors.Wrap(err, "Error reading block"))
	}
	return block
}

// LoadBlockPart returns the Part at the given index
// from the block at the given height.
// If no part is found for the given height and index, it returns nil.
func (bs *BlockStore) LoadBlockPart(height int64, index int) *types.Part {
	part := new(types.Part)
	bz, err := bs.db.Get(calcBlockPartKey(height, index))
	if err != nil {
		panic(err)
	}
	if len(bz) == 0 {
		return nil
	}
	err = amino.Unmarshal(bz, part)
	if err != nil {
		panic(errors.Wrap(err, "Error reading block part"))
	}
	return part
}

// LoadBlockMeta returns the BlockMeta for the given height.
// If no block is found for the given height, it returns nil.
func (bs *BlockStore) LoadBlockMeta(height int64) *types.BlockMeta {
	blockMeta := new(types.BlockMeta)
	bz, err := bs.db.Get(calcBlockMetaKey(height))
	if err != nil {
		panic(err)
	}
	if len(bz) == 0 {
		return nil
	}
	err = amino.Unmarshal(bz, blockMeta)
	if err != nil {
		panic(errors.Wrap(err, "Error reading block meta"))
	}
	return blockMeta
}

// LoadBlockCommit returns the Commit for the given height.
// This commit consists of the +2/3 and other Precommit-votes for block at `height`,
// and it comes from the block.LastCommit for `height+1`.
// If no commit is found for the given height, it returns nil.
func (bs *BlockStore) LoadBlockCommit(height int64) *types.Commit {
	commit := new(types.Commit)
	bz, err := bs.db.Get(calcBlockCommitKey(height))
	if err != nil {
		panic(err)
	}
	if len(bz) == 0 {
		return nil
	}
	err = amino.Unmarshal(bz, commit)
	if err != nil {
		panic(errors.Wrap(err, "Error reading block commit"))
	}
	return commit
}

// LoadSeenCommit returns the locally seen Commit for the given height.
// This is useful when we've seen a commit, but there has not yet been
// a new block at `height + 1` that includes this commit in its block.LastCommit.
func (bs *BlockStore) LoadSeenCommit(height int64) *types.Commit {
	commit := new(types.Commit)
	bz, err := bs.db.Get(calcSeenCommitKey(height))
	if err != nil {
		panic(err)
	}
	if len(bz) == 0 {
		return nil
	}
	err = amino.Unmarshal(bz, commit)
	if err != nil {
		panic(errors.Wrap(err, "Error reading block seen commit"))
	}
	return commit
}

// SaveBlock persists the given block, blockParts, and seenCommit to the underlying db.
// blockParts: Must be parts of the block
// seenCommit: The +2/3 precommits that were seen which committed at height.
//
//	If all the nodes restart after committing a block,
//	we need this to reload the precommits to catch-up nodes to the
//	most recent height.  Otherwise they'd stall at H-1.
func (bs *BlockStore) SaveBlock(block *types.Block, blockParts *types.PartSet, seenCommit *types.Commit) {
	batch := bs.NewBatch()
	bs.SaveBlockWithBatch(batch, block, blockParts, seenCommit)
	err := batch.WriteSync()
	if err != nil {
		panic(err)
	}
	batch.Close()
}

func (bs *BlockStore) SaveBlockWithBatch(batch dbm.Batch, block *types.Block, blockParts *types.PartSet, seenCommit *types.Commit) {
	if block == nil {
		panic("BlockStore can only save a non-nil block")
	}
	height := block.Height
	if g, w := height, bs.Height()+1; g != w {
		panic(fmt.Sprintf("BlockStore can only save contiguous blocks. Wanted %v, got %v", w, g))
	}
	if !blockParts.IsComplete() {
		panic("BlockStore can only save complete block part sets")
	}

	// Save block meta
	blockMeta := types.NewBlockMeta(block, blockParts)
	metaBytes := amino.MustMarshal(blockMeta)
	err := batch.Set(calcBlockMetaKey(height), metaBytes)
	if err != nil {
		panic(err)
	}

	// Save block parts
	for i := range blockParts.Total() {
		part := blockParts.GetPart(i)
		bs.saveBlockPart(batch, height, i, part)
	}

	// Save block commit (duplicate and separate from the Block)
	blockCommitBytes := amino.MustMarshal(block.LastCommit)
	err = batch.Set(calcBlockCommitKey(height-1), blockCommitBytes)
	if err != nil {
		panic(err)
	}

	// Save seen commit (seen +2/3 precommits for block)
	// NOTE: we can delete this at a later height
	seenCommitBytes := amino.MustMarshal(seenCommit)
	err = batch.Set(calcSeenCommitKey(height), seenCommitBytes)
	if err != nil {
		panic(err)
	}

	// Save new BlockStoreStateJSON descriptor
	BlockStoreStateJSON{Height: height}.Save(batch)

	// Done!
	bs.mtx.Lock()
	bs.height = height
	bs.mtx.Unlock()
}

func (bs *BlockStore) saveBlockPart(batch dbm.Batch, height int64, index int, part *types.Part) {
	if height != bs.Height()+1 {
		panic(fmt.Sprintf("BlockStore can only save contiguous blocks. Wanted %v, got %v", bs.Height()+1, height))
	}
	partBytes := amino.MustMarshal(part)
	err := batch.Set(calcBlockPartKey(height, index), partBytes)
	if err != nil {
		panic(err)
	}
}

//-----------------------------------------------------------------------------

func calcBlockMetaKey(height int64) []byte {
	return fmt.Appendf(nil, "H:%v", height)
}

func calcBlockPartKey(height int64, partIndex int) []byte {
	return fmt.Appendf(nil, "P:%v:%v", height, partIndex)
}

func calcBlockCommitKey(height int64) []byte {
	return fmt.Appendf(nil, "C:%v", height)
}

func calcSeenCommitKey(height int64) []byte {
	return fmt.Appendf(nil, "SC:%v", height)
}

//-----------------------------------------------------------------------------

var blockStoreKey = []byte("blockStore")

// BlockStoreStateJSON is the block store state JSON structure.
type BlockStoreStateJSON struct {
	Height int64 `json:"height"`
}

// Save persists the blockStore state to the database as JSON.
func (bsj BlockStoreStateJSON) Save(batch dbm.Batch) {
	bytes, err := amino.MarshalJSON(bsj)
	if err != nil {
		panic(fmt.Sprintf("Could not marshal state bytes: %v", err))
	}
	err = batch.Set(blockStoreKey, bytes)
	if err != nil {
		panic(fmt.Sprintf("Could not save blockStore state: %v", err))
	}
}

// LoadBlockStoreStateJSON returns the BlockStoreStateJSON as loaded from disk.
// If no BlockStoreStateJSON was previously persisted, it returns the zero value.
func LoadBlockStoreStateJSON(db dbm.DB) BlockStoreStateJSON {
	bytes, err := db.Get(blockStoreKey)
	if err != nil {
		panic(err)
	}
	if len(bytes) == 0 {
		return BlockStoreStateJSON{
			Height: 0,
		}
	}
	bsj := BlockStoreStateJSON{}
	err = amino.UnmarshalJSON(bytes, &bsj)
	if err != nil {
		panic(fmt.Sprintf("Could not unmarshal bytes: %X", bytes))
	}
	return bsj
}
