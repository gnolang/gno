package consensus

import (
	"context"

	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/params"
	ctypes "github.com/gnolang/gno/tm2/pkg/bft/rpc/core/types"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/core/utils"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/metadata"
	"github.com/gnolang/gno/tm2/pkg/bft/rpc/lib/server/spec"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	dbm "github.com/gnolang/gno/tm2/pkg/db"
	"github.com/gnolang/gno/tm2/pkg/telemetry/traces"
)

// Handler is the consensus RPC handler
type Handler struct {
	consensusState Consensus
	stateDB        dbm.DB
	peers          ctypes.Peers
}

// NewHandler creates a new instance of the consensus RPC handler
func NewHandler(consensusState Consensus, stateDB dbm.DB, peers ctypes.Peers) *Handler {
	return &Handler{
		consensusState: consensusState,
		stateDB:        stateDB,
		peers:          peers,
	}
}

// ValidatorsHandler returns the validator set at the given height.
// If no height is provided, it will fetch the current validator set.
// Note the validators are sorted by their address - this is the canonical
// order for the validators in the set as used in computing their Merkle root
//
//	Params:
//	- height   int64 (optional, default latest height)
func (h *Handler) ValidatorsHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	_, span := traces.Tracer().Start(context.Background(), "Validators")
	defer span.End()

	const idxHeight = 0

	heightVal, err := params.AsInt64(p, idxHeight)
	if err != nil {
		return nil, err
	}

	latest := h.consensusState.GetState().LastBlockHeight + 1

	height, normErr := utils.NormalizeHeight(latest, heightVal, 1)
	if normErr != nil {
		return nil, spec.GenerateResponseError(normErr)
	}

	validators, loadErr := sm.LoadValidators(h.stateDB, height)
	if loadErr != nil {
		return nil, spec.GenerateResponseError(loadErr)
	}

	return &ResultValidators{
		BlockHeight: height,
		Validators:  validators.Validators,
	}, nil
}

// DumpConsensusStateHandler dumps the full consensus state (UNSTABLE)
//
//	No params
func (h *Handler) DumpConsensusStateHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	if len(p) > 0 {
		return nil, spec.GenerateInvalidParamError(1)
	}

	_, span := traces.Tracer().Start(context.Background(), "DumpConsensusState")
	defer span.End()

	var (
		peers      = h.peers.Peers().List()
		peerStates = make([]PeerStateInfo, len(peers))
	)

	for i, peer := range peers {
		ps, ok := peer.Get(types.PeerStateKey).(interface {
			GetExposed() cstypes.PeerStateExposed
		})

		if !ok {
			continue
		}

		psJSON, err := ps.GetExposed().ToJSON()
		if err != nil {
			return nil, spec.GenerateResponseError(err)
		}

		peerStates[i] = PeerStateInfo{
			NodeAddress: peer.SocketAddr().String(),
			PeerState:   psJSON,
		}
	}

	var (
		config     = h.consensusState.GetConfigDeepCopy()
		roundState = h.consensusState.GetRoundStateDeepCopy()
	)

	return &ResultDumpConsensusState{
		Config:     config,
		RoundState: roundState,
		Peers:      peerStates,
	}, nil
}

// ConsensusStateHandler returns a concise summary of the consensus state (UNSTABLE)
//
//	No params
func (h *Handler) ConsensusStateHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	if len(p) > 0 {
		return nil, spec.GenerateInvalidParamError(1)
	}

	_, span := traces.Tracer().Start(context.Background(), "ConsensusState")
	defer span.End()

	return &ResultConsensusState{
		RoundState: h.consensusState.GetRoundStateSimple(),
	}, nil
}

// ConsensusParamsHandler returns consensus params at a given height.
//
//	Params:
//	- height   int64 (optional, default latest height)
func (h *Handler) ConsensusParamsHandler(_ *metadata.Metadata, p []any) (any, *spec.BaseJSONError) {
	_, span := traces.Tracer().Start(context.Background(), "ConsensusParams")
	defer span.End()

	const idxHeight = 0

	heightVal, err := params.AsInt64(p, idxHeight)
	if err != nil {
		return nil, err
	}

	latest := h.consensusState.GetState().LastBlockHeight + 1

	height, normErr := utils.NormalizeHeight(latest, heightVal, 1)
	if normErr != nil {
		return nil, spec.GenerateResponseError(normErr)
	}

	consensusParams, loadErr := sm.LoadConsensusParams(h.stateDB, height)
	if loadErr != nil {
		return nil, spec.GenerateResponseError(loadErr)
	}

	return &ResultConsensusParams{
		BlockHeight:     height,
		ConsensusParams: consensusParams,
	}, nil
}
