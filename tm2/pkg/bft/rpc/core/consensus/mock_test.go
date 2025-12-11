package consensus

import (
	cnscfg "github.com/gnolang/gno/tm2/pkg/bft/consensus/config"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

type (
	getConfigDeepCopyDelegate     func() *cnscfg.ConsensusConfig
	getStateDelegate              func() sm.State
	getValidatorsDelegate         func() (int64, []*types.Validator)
	getLastHeightDelegate         func() int64
	getRoundStateDeepCopyDelegate func() *cstypes.RoundState
	getRoundStateSimpleDelegate   func() cstypes.RoundStateSimple
)

type mockConsensus struct {
	getConfigDeepCopyFn     getConfigDeepCopyDelegate
	getStateFn              getStateDelegate
	getValidatorsFn         getValidatorsDelegate
	getLastHeightFn         getLastHeightDelegate
	getRoundStateDeepCopyFn getRoundStateDeepCopyDelegate
	getRoundStateSimpleFn   getRoundStateSimpleDelegate
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
