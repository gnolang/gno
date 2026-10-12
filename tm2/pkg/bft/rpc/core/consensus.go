package core

import (
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	rpctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/types"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

// Validators returns the validator set at the given block height. If no
// height is provided, it fetches the current validator set. Note the
// validators are sorted by their address — this is the canonical order for
// the validators in the set as used in computing their Merkle root.
func (env *Environment) Validators(ctx *rpctypes.Context, heightPtr *int64) (*ctypes.ResultValidators, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "Validators")
	defer span.End()
	// The latest validator that we know is the NextValidator of the last block.
	height := env.Consensus.GetState().LastBlockHeight + 1
	height, err := getHeight(height, heightPtr)
	if err != nil {
		return nil, err
	}

	validators, err := sm.LoadValidators(env.StateDB, height)
	if err != nil {
		return nil, err
	}
	return &ctypes.ResultValidators{
		BlockHeight: height,
		Validators:  validators.Validators,
	}, nil
}

// DumpConsensusState dumps consensus state. UNSTABLE.
func (env *Environment) DumpConsensusState(ctx *rpctypes.Context) (*ctypes.ResultDumpConsensusState, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "DumpConsensusState")
	defer span.End()
	// Get peer consensus states.
	peers := env.P2PPeers.Peers().List()
	peerStates := make([]ctypes.PeerStateInfo, len(peers))
	for i, peer := range peers {
		peerState, ok := peer.Get(types.PeerStateKey).(interface {
			GetExposed() cstypes.PeerStateExposed
		})
		if !ok { // peer does not have a state yet
			continue
		}
		peerStateJSON, err := peerState.GetExposed().ToJSON()
		if err != nil {
			return nil, err
		}
		peerStates[i] = ctypes.PeerStateInfo{
			// Peer basic info.
			NodeAddress: peer.SocketAddr().String(),
			// Peer consensus state.
			PeerState: peerStateJSON,
		}
	}
	// Get self round state.
	config := env.Consensus.GetConfigDeepCopy()
	roundState := env.Consensus.GetRoundStateDeepCopy()
	return &ctypes.ResultDumpConsensusState{
		Config:     config,
		RoundState: roundState,
		Peers:      peerStates,
	}, nil
}

// ConsensusState returns a concise summary of the consensus state. UNSTABLE.
func (env *Environment) ConsensusState(ctx *rpctypes.Context) (*ctypes.ResultConsensusState, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "ConsensusState")
	defer span.End()
	// Get self round state.
	rs := env.Consensus.GetRoundStateSimple()
	return &ctypes.ResultConsensusState{RoundState: rs}, nil
}

// ConsensusParams returns the consensus parameters at the given block height.
// If no height is provided, it fetches the current consensus params.
func (env *Environment) ConsensusParams(ctx *rpctypes.Context, heightPtr *int64) (*ctypes.ResultConsensusParams, error) {
	_, span := traces.Tracer().Start(ctx.Context(), "ConsensusParams")
	defer span.End()
	height := env.Consensus.GetState().LastBlockHeight + 1
	height, err := getHeight(height, heightPtr)
	if err != nil {
		return nil, err
	}

	consensusparams, err := sm.LoadConsensusParams(env.StateDB, height)
	if err != nil {
		return nil, err
	}
	return &ctypes.ResultConsensusParams{
		BlockHeight:     height,
		ConsensusParams: consensusparams,
	}, nil
}
