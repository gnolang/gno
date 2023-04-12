package store

import (
	"errors"
	"fmt"
	"sync"

	"github.com/gnolang/gno/tm2/pkg/amino"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
)

var _ sm.BlockStore = &BlockStore{}

/*
BlockStore is a simple low level store for blocks.

There are three types of information stored:
  - BlockMeta:   Meta information about each block
  - Block part:  Parts of each block, aggregated w/ PartSet
  - Commit:      The commit part of each block, for gossiping precommit votes

Currently the precommit signatures are duplicated in the Block parts as
well as the Commit.  In the future this may change, perhaps by moving
the Commit data outside the Block. (TODO)
*/
type BlockStore struct {
	db dbm.DB

	mtx    sync.RWMutex
	height int64
}

// NewBlockStore returns a new BlockStore with the given DB,
// initialized to the last height that was committed to the DB.
func NewBlockStore(db dbm.DB) (*BlockStore, error) {
	bsjson, err := LoadBlockStoreStateJSON(db)
	if err != nil {
		return nil, err
	}

	return &BlockStore{
		height: bsjson.Height,
		db:     db,
	}, nil
}

// Height returns the last known contiguous block height.
func (bs *BlockStore) Height() (int64, error) {
	bs.mtx.RLock()
	defer bs.mtx.RUnlock()
	return bs.height, nil
}

// LoadBlock returns the block with the given height.
// If no block is found for that height, it returns nil.
func (bs *BlockStore) LoadBlock(height int64) (*types.Block, error) {
	blockMeta, err := bs.LoadBlockMeta(height)
	if err != nil {
		return nil, err
	}

	if blockMeta == nil {
		return nil, nil
	}

	block := new(types.Block)
	buf := []byte{}
	for i := 0; i < blockMeta.BlockID.PartsHeader.Total; i++ {
		part, err := bs.LoadBlockPart(height, i)
		if err != nil {
			return nil, err
		}
		buf = append(buf, part.Bytes...)
	}

	if err := amino.UnmarshalSized(buf, block); err != nil {
		// NOTE: The existence of meta should imply the existence of the
		// block. So, make sure meta is only saved after blocks are saved.
		return nil, fmt.Errorf("error reading block: %w", err)
	}

	return block, nil
}

// LoadBlockPart returns the Part at the given index
// from the block at the given height.
// If no part is found for the given height and index, it returns nil.
func (bs *BlockStore) LoadBlockPart(height int64, index int) (*types.Part, error) {
	part := new(types.Part)
	bz := bs.db.Get(calcBlockPartKey(height, index))
	if len(bz) == 0 {
		return nil, nil
	}

	if err := amino.Unmarshal(bz, part); err != nil {
		return nil, fmt.Errorf("error reading block part: %w", err)
	}
	return part, nil
}

// LoadBlockMeta returns the BlockMeta for the given height.
// If no block is found for the given height, it returns nil.
func (bs *BlockStore) LoadBlockMeta(height int64) (*types.BlockMeta, error) {
	blockMeta := new(types.BlockMeta)
	bz := bs.db.Get(calcBlockMetaKey(height))
	if len(bz) == 0 {
		return nil, nil
	}

	if err := amino.Unmarshal(bz, blockMeta); err != nil {
		return nil, fmt.Errorf("error reading block meta: %w", err)
	}

	return blockMeta, nil
}

// LoadBlockCommit returns the Commit for the given height.
// This commit consists of the +2/3 and other Precommit-votes for block at `height`,
// and it comes from the block.LastCommit for `height+1`.
// If no commit is found for the given height, it returns nil.
func (bs *BlockStore) LoadBlockCommit(height int64) (*types.Commit, error) {
	commit := new(types.Commit)
	bz := bs.db.Get(calcBlockCommitKey(height))
	if len(bz) == 0 {
		return nil, nil
	}

	if err := amino.Unmarshal(bz, commit); err != nil {
		return nil, fmt.Errorf("error reading block commit: %w", err)
	}

	return commit, nil
}

// LoadSeenCommit returns the locally seen Commit for the given height.
// This is useful when we've seen a commit, but there has not yet been
// a new block at `height + 1` that includes this commit in its block.LastCommit.
func (bs *BlockStore) LoadSeenCommit(height int64) (*types.Commit, error) {
	commit := new(types.Commit)
	bz := bs.db.Get(calcSeenCommitKey(height))
	if len(bz) == 0 {
		return nil, nil
	}

	if err := amino.Unmarshal(bz, commit); err != nil {
		return nil, fmt.Errorf("error reading block seen commit: %w", err)
	}

	return commit, nil
}

// SaveBlock persists the given block, blockParts, and seenCommit to the underlying db.
// blockParts: Must be parts of the block
// seenCommit: The +2/3 precommits that were seen which committed at height.
//
//	If all the nodes restart after committing a block,
//	we need this to reload the precommits to catch-up nodes to the
//	most recent height.  Otherwise they'd stall at H-1.
func (bs *BlockStore) SaveBlock(block *types.Block, blockParts *types.PartSet, seenCommit *types.Commit) error {
	if block == nil {
		return errors.New("blockStore can only save a non-nil block")
	}
	height := block.Height
	bsh, err := bs.Height()
	if err != nil {
		return err
	}
	if g, w := height, bsh+1; g != w {
		return fmt.Errorf("blockStore can only save contiguous blocks. Wanted %v, got %v", w, g)
	}
	if !blockParts.IsComplete() {
		return errors.New("blockStore can only save complete block part sets")
	}

	// Save block meta
	blockMeta := types.NewBlockMeta(block, blockParts)
	metaBytes := amino.MustMarshal(blockMeta)
	bs.db.Set(calcBlockMetaKey(height), metaBytes)

	// Save block parts
	for i := 0; i < blockParts.Total(); i++ {
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

	// Save new BlockStoreStateJSON descriptor
	BlockStoreStateJSON{Height: height}.Save(bs.db)

	// Done!
	bs.mtx.Lock()
	bs.height = height
	bs.mtx.Unlock()

	// Flush
	bs.db.SetSync(nil, nil)

	return nil
}

func (bs *BlockStore) saveBlockPart(height int64, index int, part *types.Part) error {
	h, err := bs.Height()
	if err != nil {
		return err
	}
	if height != h+1 {
		return fmt.Errorf("blockStore can only save contiguous blocks. Wanted %v, got %v", h+1, height)
	}
	partBytes := amino.MustMarshal(part)
	bs.db.Set(calcBlockPartKey(height, index), partBytes)

	return nil
}

//-----------------------------------------------------------------------------

func calcBlockMetaKey(height int64) []byte {
	return []byte(fmt.Sprintf("H:%v", height))
}

func calcBlockPartKey(height int64, partIndex int) []byte {
	return []byte(fmt.Sprintf("P:%v:%v", height, partIndex))
}

func calcBlockCommitKey(height int64) []byte {
	return []byte(fmt.Sprintf("C:%v", height))
}

func calcSeenCommitKey(height int64) []byte {
	return []byte(fmt.Sprintf("SC:%v", height))
}

//-----------------------------------------------------------------------------

var blockStoreKey = []byte("blockStore")

// BlockStoreStateJSON is the block store state JSON structure.
type BlockStoreStateJSON struct {
	Height int64 `json:"height"`
}

// Save persists the blockStore state to the database as JSON.
func (bsj BlockStoreStateJSON) Save(db dbm.DB) error {
	bytes, err := amino.MarshalJSON(bsj)
	if err != nil {
		return fmt.Errorf("error marshalling state bytes: %w", err)
	}

	db.SetSync(blockStoreKey, bytes)

	return nil
}

// LoadBlockStoreStateJSON returns the BlockStoreStateJSON as loaded from disk.
// If no BlockStoreStateJSON was previously persisted, it returns the zero value.
func LoadBlockStoreStateJSON(db dbm.DB) (BlockStoreStateJSON, error) {
	bytes := db.Get(blockStoreKey)
	if len(bytes) == 0 {
		return BlockStoreStateJSON{
			Height: 0,
		}, nil
	}
	bsj := BlockStoreStateJSON{}
	err := amino.UnmarshalJSON(bytes, &bsj)
	if err != nil {
		return bsj, fmt.Errorf("could not unmarshall bytes %X: %w", bytes, err)
	}
	return bsj, nil
}
