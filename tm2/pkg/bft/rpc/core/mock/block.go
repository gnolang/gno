package mock

import "github.com/gnolang/gno/tm2/pkg/bft/types"

type (
	HeightDelegate          func() int64
	LoadBlockMetaDelegate   func(int64) *types.BlockMeta
	LoadBlockDelegate       func(int64) *types.Block
	LoadSeenCommitDelegate  func(int64) *types.Commit
	LoadBlockCommitDelegate func(int64) *types.Commit
	LoadBlockByHashDelegate func([]byte) *types.Block
	LoadBlockPartDelegate   func(int64, int) *types.Part
	SaveBlockDelegate       func(*types.Block, *types.PartSet, *types.Commit)
)

type BlockStore struct {
	HeightFn          HeightDelegate
	LoadBlockMetaFn   LoadBlockMetaDelegate
	LoadBlockFn       LoadBlockDelegate
	LoadSeenCommitFn  LoadSeenCommitDelegate
	LoadBlockCommitFn LoadBlockCommitDelegate
	LoadBlockByHashFn LoadBlockByHashDelegate
	LoadBlockPartFn   LoadBlockPartDelegate
	SaveBlockFn       SaveBlockDelegate
}

func (m *BlockStore) Height() int64 {
	if m.HeightFn != nil {
		return m.HeightFn()
	}

	return 0
}

func (m *BlockStore) LoadBlockMeta(h int64) *types.BlockMeta {
	if m.LoadBlockMetaFn != nil {
		return m.LoadBlockMetaFn(h)
	}

	return nil
}

func (m *BlockStore) LoadBlock(h int64) *types.Block {
	if m.LoadBlockFn != nil {
		return m.LoadBlockFn(h)
	}

	return nil
}

func (m *BlockStore) LoadSeenCommit(h int64) *types.Commit {
	if m.LoadSeenCommitFn != nil {
		return m.LoadSeenCommitFn(h)
	}

	return nil
}

func (m *BlockStore) LoadBlockCommit(h int64) *types.Commit {
	if m.LoadBlockCommitFn != nil {
		return m.LoadBlockCommitFn(h)
	}

	return nil
}

func (m *BlockStore) LoadBlockByHash(hash []byte) *types.Block {
	if m.LoadBlockByHashFn != nil {
		return m.LoadBlockByHashFn(hash)
	}

	return nil
}
func (m *BlockStore) LoadBlockPart(height int64, index int) *types.Part {
	if m.LoadBlockPartFn != nil {
		return m.LoadBlockPartFn(height, index)
	}

	return nil
}
func (m *BlockStore) SaveBlock(block *types.Block, set *types.PartSet, commit *types.Commit) {
	if m.SaveBlockFn != nil {
		m.SaveBlockFn(block, set, commit)
	}
}
