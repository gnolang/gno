package consensus

import (
	cnscfg "github.com/gnolang/gno/tm2/pkg/bft/consensus/config"
	cstypes "github.com/gnolang/gno/tm2/pkg/bft/consensus/types"
	sm "github.com/gnolang/gno/tm2/pkg/bft/state"
	"github.com/gnolang/gno/tm2/pkg/bft/types"
	"github.com/gnolang/gno/tm2/pkg/p2p"
)

// Consensus exposes read-only access to consensus state for RPC handlers
type Consensus interface {
	// GetConfigDeepCopy returns a deep copy of the current consensus config
	GetConfigDeepCopy() *cnscfg.ConsensusConfig

	// GetState returns a snapshot of the current consensus state
	GetState() sm.State

	// GetValidators returns the height and validator set for that height
	GetValidators() (int64, []*types.Validator)

	// GetLastHeight returns the last block height known to consensus
	GetLastHeight() int64

	// GetRoundStateDeepCopy returns a deep copy of the full round state
	GetRoundStateDeepCopy() *cstypes.RoundState

	// GetRoundStateSimple returns a concise summary of the round state
	GetRoundStateSimple() cstypes.RoundStateSimple
}

// Peers exposes access to the current P2P peer set
type Peers interface {
	// Peers returns the current peer set
	Peers() p2p.PeerSet
}
