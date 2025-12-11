package consensus

import (
	"net"

	cnscfg "github.com/gnolang/gno/tm2/pkg/bft/consensus/config"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/p2p"
	p2pTypes "github.com/gnolang/gno/tm2/pkg/p2p/types"
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

type (
	peersDelegate func() p2p.PeerSet
)

type mockPeers struct {
	peersFn peersDelegate
}

func (m *mockPeers) Peers() p2p.PeerSet {
	if m.peersFn != nil {
		return m.peersFn()
	}

	return nil
}

type (
	addDelegate         func(p2p.PeerConn)
	removeDelegate      func(p2pTypes.ID) bool
	hasDelegate         func(p2pTypes.ID) bool
	hasIPDelegate       func(net.IP) bool
	getPeerDelegate     func(p2pTypes.ID) p2p.PeerConn
	listDelegate        func() []p2p.PeerConn
	numInboundDelegate  func() uint64
	numOutboundDelegate func() uint64
)

type mockPeerSet struct {
	addFn         addDelegate
	removeFn      removeDelegate
	hasFn         hasDelegate
	hasIPFn       hasIPDelegate
	getFn         getPeerDelegate
	listFn        listDelegate
	numInboundFn  numInboundDelegate
	numOutboundFn numOutboundDelegate
}

func (m *mockPeerSet) Add(peer p2p.PeerConn) {
	if m.addFn != nil {
		m.addFn(peer)
	}
}

func (m *mockPeerSet) Remove(key p2pTypes.ID) bool {
	if m.removeFn != nil {
		m.removeFn(key)
	}

	return false
}

func (m *mockPeerSet) Has(key p2pTypes.ID) bool {
	if m.hasFn != nil {
		return m.hasFn(key)
	}

	return false
}

func (m *mockPeerSet) Get(key p2pTypes.ID) p2p.PeerConn {
	if m.getFn != nil {
		return m.getFn(key)
	}

	return nil
}

func (m *mockPeerSet) List() []p2p.PeerConn {
	if m.listFn != nil {
		return m.listFn()
	}

	return nil
}

func (m *mockPeerSet) NumInbound() uint64 {
	if m.numInboundFn != nil {
		return m.numInboundFn()
	}

	return 0
}

func (m *mockPeerSet) NumOutbound() uint64 {
	if m.numOutboundFn != nil {
		return m.numOutboundFn()
	}

	return 0
}
