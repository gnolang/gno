package mock

import (
	cnscfg "github.com/gnolang/gno/tm2/pkg/bft/consensus/config"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
)

type (
	GetConfigDeepCopyDelegate     func() *cnscfg.ConsensusConfig
	GetStateDelegate              func() sm.State
	GetRoundStateDeepCopyDelegate func() *cstypes.RoundState
	GetRoundStateSimpleDelegate   func() cstypes.RoundStateSimple
)

type Consensus struct {
	GetConfigDeepCopyFn     GetConfigDeepCopyDelegate
	GetStateFn              GetStateDelegate
	GetRoundStateDeepCopyFn GetRoundStateDeepCopyDelegate
	GetRoundStateSimpleFn   GetRoundStateSimpleDelegate
}

func (m *Consensus) GetConfigDeepCopy() *cnscfg.ConsensusConfig {
	if m.GetConfigDeepCopyFn != nil {
		return m.GetConfigDeepCopyFn()
	}

	return nil
}

func (m *Consensus) GetState() sm.State {
	if m.GetStateFn != nil {
		return m.GetStateFn()
	}

	return sm.State{}
}

func (m *Consensus) GetRoundStateDeepCopy() *cstypes.RoundState {
	if m.GetRoundStateDeepCopyFn != nil {
		return m.GetRoundStateDeepCopyFn()
	}

	return nil
}

func (m *Consensus) GetRoundStateSimple() cstypes.RoundStateSimple {
	if m.GetRoundStateSimpleFn != nil {
		return m.GetRoundStateSimpleFn()
	}

	return cstypes.RoundStateSimple{}
}
