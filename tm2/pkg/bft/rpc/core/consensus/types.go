package consensus

import (
	"encoding/json"

	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	cnscfg "github.com/gnolang/gno/tm2/pkg/bft/consensus/config"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
)

// Consensus exposes read-only access to consensus state for RPC handlers
type Consensus interface {
	// GetConfigDeepCopy returns a deep copy of the current consensus config
	GetConfigDeepCopy() *cnscfg.ConsensusConfig

	// GetState returns a snapshot of the current consensus state
	GetState() sm.State

	// GetRoundStateDeepCopy returns a deep copy of the full round state
	GetRoundStateDeepCopy() *cstypes.RoundState

	// GetRoundStateSimple returns a concise summary of the round state
	GetRoundStateSimple() cstypes.RoundStateSimple
}

type ResultValidators struct {
	BlockHeight int64              `json:"block_height"`
	Validators  []*types.Validator `json:"validators"`
}

type ResultConsensusParams struct {
	BlockHeight     int64                `json:"block_height"`
	ConsensusParams abci.ConsensusParams `json:"consensus_params"`
}

type ResultDumpConsensusState struct {
	Config     *cnscfg.ConsensusConfig `json:"config"`
	RoundState *cstypes.RoundState     `json:"round_state"`
	Peers      []PeerStateInfo         `json:"peers"`
}

type PeerStateInfo struct {
	NodeAddress string          `json:"node_address"`
	PeerState   json.RawMessage `json:"peer_state"`
}

type ResultConsensusState struct {
	RoundState cstypes.RoundStateSimple `json:"round_state"`
}
