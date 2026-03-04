package core

import (
	cnscfg "github.com/gnolang/gno/tm2/pkg/bft/consensus/config"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
)

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

// mockConsensus implements the Consensus interface for testing.
type mockConsensus struct {
	getConfigDeepCopyFn     func() *cnscfg.ConsensusConfig
	getStateFn              func() sm.State
	getValidatorsFn         func() (int64, []*types.Validator)
	getLastHeightFn         func() int64
	getRoundStateDeepCopyFn func() *cstypes.RoundState
	getRoundStateSimpleFn   func() cstypes.RoundStateSimple
}

func (m *mockConsensus) GetConfigDeepCopy() *cnscfg.ConsensusConfig {
	if m.getConfigDeepCopyFn != nil {
		return m.getConfigDeepCopyFn()
	}
	return nil
}

func (m *mockConsensus) GetState() sm.State {
	if m.getStateFn != nil {
		return m.getStateFn()
	}
	return sm.State{}
}

func (m *mockConsensus) GetValidators() (int64, []*types.Validator) {
	if m.getValidatorsFn != nil {
		return m.getValidatorsFn()
	}
	return 0, nil
}

func (m *mockConsensus) GetLastHeight() int64 {
	if m.getLastHeightFn != nil {
		return m.getLastHeightFn()
	}
	return 0
}

func (m *mockConsensus) GetRoundStateDeepCopy() *cstypes.RoundState {
	if m.getRoundStateDeepCopyFn != nil {
		return m.getRoundStateDeepCopyFn()
	}
	return nil
}

func (m *mockConsensus) GetRoundStateSimple() cstypes.RoundStateSimple {
	if m.getRoundStateSimpleFn != nil {
		return m.getRoundStateSimpleFn()
	}
	return cstypes.RoundStateSimple{}
}

// mockTransport implements the transport interface for testing.
type mockTransport struct {
	listenersFn   func() []string
	isListeningFn func() bool
	nodeInfoFn    func() p2pTypes.NodeInfo
}

func (m *mockTransport) Listeners() []string {
	if m.listenersFn != nil {
		return m.listenersFn()
	}
	return nil
}

func (m *mockTransport) IsListening() bool {
	if m.isListeningFn != nil {
		return m.isListeningFn()
	}
	return false
}

func (m *mockTransport) NodeInfo() p2pTypes.NodeInfo {
	if m.nodeInfoFn != nil {
		return m.nodeInfoFn()
	}
	return p2pTypes.NodeInfo{}
}
