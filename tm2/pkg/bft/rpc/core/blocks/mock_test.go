package blocks

import "github.com/gnolang/gno/tm2/pkg/bft/types"

type (
	heightDelegate          func() int64
	loadBlockMetaDelegate   func(int64) *types.BlockMeta
	loadBlockDelegate       func(int64) *types.Block
	loadSeenCommitDelegate  func(int64) *types.Commit
	loadBlockCommitDelegate func(int64) *types.Commit
	loadBlockByHashDelegate func([]byte) *types.Block
	loadBlockPartDelegate   func(int64, int) *types.Part
	saveBlockDelegate       func(*types.Block, *types.PartSet, *types.Commit)
)

type mockBlockStore struct {
	heightFn          heightDelegate
	loadBlockMetaFn   loadBlockMetaDelegate
	loadBlockFn       loadBlockDelegate
	loadSeenCommitFn  loadSeenCommitDelegate
	loadBlockCommitFn loadBlockCommitDelegate
	loadBlockByHashFn loadBlockByHashDelegate
	loadBlockPartFn   loadBlockPartDelegate
	saveBlockFn       saveBlockDelegate
}

func (m *mockBlockStore) Height() int64 {
	if m.heightFn != nil {
		return m.heightFn()
	}

	return 0
}

func (m *mockBlockStore) LoadBlockMeta(h int64) *types.BlockMeta {
	if m.loadBlockMetaFn != nil {
		return m.loadBlockMetaFn(h)
	}

	return nil
}

func (m *mockBlockStore) LoadBlock(h int64) *types.Block {
	if m.loadBlockFn != nil {
		return m.loadBlockFn(h)
	}

	return nil
}

func (m *mockBlockStore) LoadSeenCommit(h int64) *types.Commit {
	if m.loadSeenCommitFn != nil {
		return m.loadSeenCommitFn(h)
	}

	return nil
}

func (m *mockBlockStore) LoadBlockCommit(h int64) *types.Commit {
	if m.loadBlockCommitFn != nil {
		return m.loadBlockCommitFn(h)
	}

	return nil
}

func (m *mockBlockStore) LoadBlockByHash(hash []byte) *types.Block {
	if m.loadBlockByHashFn != nil {
		return m.loadBlockByHashFn(hash)
	}

	return nil
}
func (m *mockBlockStore) LoadBlockPart(height int64, index int) *types.Part {
	if m.loadBlockPartFn != nil {
		return m.loadBlockPartFn(height, index)
	}

	return nil
}
func (m *mockBlockStore) SaveBlock(block *types.Block, set *types.PartSet, commit *types.Commit) {
	if m.saveBlockFn != nil {
		m.saveBlockFn(block, set, commit)
	}
}
