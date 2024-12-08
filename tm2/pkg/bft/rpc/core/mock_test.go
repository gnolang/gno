package core

import "github.com/gnolang/gno/tm2/pkg/bft/types"

type (
	heightDelegate          func() int64
	loadBlockMetaDelegate   func(int64) *types.BlockMeta
	loadBlockDelegate       func(int64) *types.Block
	loadBlockPartDelegate   func(int64, int) *types.Part
	loadBlockCommitDelegate func(int64) *types.Commit
	loadSeenCommitDelegate  func(int64) *types.Commit

	saveBlockDelegate func(*types.Block, *types.PartSet, *types.Commit)
)

type mockBlockStore struct {
	heightFn          heightDelegate
	loadBlockMetaFn   loadBlockMetaDelegate
	loadBlockFn       loadBlockDelegate
	loadBlockPartFn   loadBlockPartDelegate
	loadBlockCommitFn loadBlockCommitDelegate
	loadSeenCommitFn  loadSeenCommitDelegate
	saveBlockFn       saveBlockDelegate
}

func (m *mockBlockStore) Height() int64 {
	if m.heightFn != nil {
		return m.heightFn()
	}

	return 0
}

func (m *mockBlockStore) LoadBlockMeta(height int64) *types.BlockMeta {
	if m.loadBlockMetaFn != nil {
		return m.loadBlockMetaFn(height)
	}

	return nil
}

func (m *mockBlockStore) LoadBlock(height int64) *types.Block {
	if m.loadBlockFn != nil {
		return m.loadBlockFn(height)
	}

	return nil
}

func (m *mockBlockStore) LoadBlockPart(height int64, index int) *types.Part {
	if m.loadBlockPartFn != nil {
		return m.loadBlockPartFn(height, index)
	}

	return nil
}

func (m *mockBlockStore) LoadBlockCommit(height int64) *types.Commit {
	if m.loadBlockCommitFn != nil {
		return m.loadBlockCommitFn(height)
	}

	return nil
}

func (m *mockBlockStore) LoadSeenCommit(height int64) *types.Commit {
	if m.loadSeenCommitFn != nil {
		return m.loadSeenCommitFn(height)
	}

	return nil
}

func (m *mockBlockStore) SaveBlock(block *types.Block, blockParts *types.PartSet, seenCommit *types.Commit) {
	if m.saveBlockFn != nil {
		m.saveBlockFn(block, blockParts, seenCommit)
	}
}
