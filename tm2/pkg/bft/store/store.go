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

// DB returns the database the store reads and writes.
func (bs *BlockStore) DB() dbm.DB { return bs.db }

// Height returns the last known contiguous block height.
func (bs *BlockStore) Height() int64 {
	bs.mtx.RLock()
	defer bs.mtx.RUnlock()
	return bs.height
}

// LowerHeight lowers the store's height to height, on disk and in memory. The
// records above it stay in place, unread, until SaveBlock saves those heights
// again over them. The handshake calls it after a crash in a block-sync flush
// point (see FlushPoint).
func (bs *BlockStore) LowerHeight(height int64) {
	if height < 0 || height >= bs.Height() {
		panic(fmt.Sprintf("BlockStore can only lower its height: %d to %d", bs.Height(), height))
	}
	BlockStoreStateJSON{Height: height}.Save(bs.db)
	bs.mtx.Lock()
	bs.height = height
	bs.mtx.Unlock()
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
		// If the part is missing (e.g. since it has been deleted after we
		// loaded the block meta) we consider the whole block to be missing.
		if part == nil {
			return nil
		}
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
	if block == nil {
		panic("BlockStore can only save a non-nil block")
	}
	height := block.Height
	// When the store is empty (Height() == 0) the chain may start at InitialHeight > 1,
	// so any height is valid for the first block.  Once the store has blocks, saves
	// must be strictly contiguous.
	if bs.Height() != 0 {
		if g, w := height, bs.Height()+1; g != w {
			panic(fmt.Sprintf("BlockStore can only save contiguous blocks. Wanted %v, got %v", w, g))
		}
	}
	if !blockParts.IsComplete() {
		panic("BlockStore can only save complete block part sets")
	}

	// Save block meta
	blockMeta := types.NewBlockMeta(block, blockParts)
	metaBytes := amino.MustMarshal(blockMeta)
	bs.db.Set(calcBlockMetaKey(height), metaBytes)

	// Save block parts
	for i := range blockParts.Total() {
		part := blockParts.GetPart(i)
		bs.saveBlockPart(height, i, part)
	}

	// Save block commit (duplicate and separate from the Block)
	blockCommitBytes := amino.MustMarshal(block.LastCommit)
	bs.db.Set(calcBlockCommitKey(height-1), blockCommitBytes)

	// Save seen commit (seen +2/3 precommits for block)
	// NOTE: we can delete this at a later height
	seenCommitBytes := amino.MustMarshal(seenCommit)
	bs.db.Set(calcSeenCommitKey(height), seenCommitBytes)

	// Save new BlockStoreStateJSON descriptor. This also flushes the database.
	BlockStoreStateJSON{Height: height}.Save(bs.db)

	// Done!
	bs.mtx.Lock()
	bs.height = height
	bs.mtx.Unlock()
}

func (bs *BlockStore) saveBlockPart(height int64, index int, part *types.Part) {
	// Allow the genesis block at any height when the store is empty (InitialHeight > 1).
	if bs.Height() != 0 && height != bs.Height()+1 {
		panic(fmt.Sprintf("BlockStore can only save contiguous blocks. Wanted %v, got %v", bs.Height()+1, height))
	}
	partBytes := amino.MustMarshal(part)
	bs.db.Set(calcBlockPartKey(height, index), partBytes)
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
func (bsj BlockStoreStateJSON) Save(db dbm.DB) {
	bytes, err := amino.MarshalJSON(bsj)
	if err != nil {
		panic(fmt.Sprintf("Could not marshal state bytes: %v", err))
	}
	if err := db.SetSync(blockStoreKey, bytes); err != nil {
		panic(fmt.Sprintf("Could not save the block store state: %v", err))
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

//-----------------------------------------------------------------------------

var flushPointKey = []byte("blockStoreFlushPoint")

// FlushPoint is the record a block-sync flush point writes with the block
// store's batch, its first write: the window it writes, from the last flush
// point's height to its own. A crash before its second write, the state's,
// leaves the block store ahead of the state by that window; the record lets
// the handshake tell that shape from any other, such as a partial restore,
// which it refuses as before.
type FlushPoint struct {
	From int64 `json:"from"`
	To   int64 `json:"to"`
}

// Save writes the record to db, without a sync: it rides in the batch of the
// block store's held writes.
func (fp FlushPoint) Save(db dbm.DB) {
	bytes, err := amino.MarshalJSON(fp)
	if err != nil {
		panic(fmt.Sprintf("Could not marshal the flush point: %v", err))
	}
	if err := db.Set(flushPointKey, bytes); err != nil {
		panic(fmt.Sprintf("Could not save the flush point: %v", err))
	}
}

// FlushPoint returns the last block-sync flush point's record, and false if
// there is none.
func (bs *BlockStore) FlushPoint() (FlushPoint, bool) {
	bytes, err := bs.db.Get(flushPointKey)
	if err != nil {
		panic(err)
	}
	if len(bytes) == 0 {
		return FlushPoint{}, false
	}
	var fp FlushPoint
	if err := amino.UnmarshalJSON(bytes, &fp); err != nil {
		panic(fmt.Sprintf("Could not unmarshal the flush point: %X", bytes))
	}
	return fp, true
}
