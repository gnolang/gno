package tendermint

import (
	"time"

	"gno.land/p/demo/ufmt"
)

// ClientState from Tendermint tracks the current validator set, latest height,
// and a possible frozen height.
type ClientState struct {
	ChainID    string
	TrustLevel Fraction
	// duration of the period since the LatestTimestamp during which the
	// submitted headers are valid for upgrade
	TrustingPeriod time.Duration
	// duration of the staking unbonding period
	UnbondingPeriod time.Duration
	// defines how much new (untrusted) header's Time can drift into the future.
	MaxClockDrift time.Duration
	// Block height when the client was frozen due to a misbehaviour
	FrozenHeight Height
	// Latest height the client was updated to
	LatestHeight Height

	// Proof specifications used in verifying counterparty state
	// TODO write gno type for proofspec
	// ProofSpecs []*_go.ProofSpec

	// Path at which next upgraded client will be committed.
	// Each element corresponds to the key for a single CommitmentProof in the
	// chained proof. NOTE: ClientState must stored under
	// `{upgradePath}/{upgradeHeight}/clientState` ConsensusState must be stored
	// under `{upgradepath}/{upgradeHeight}/consensusState` For SDK chains using
	// the default upgrade module, upgrade_path should be []string{"upgrade",
	// "upgradedIBCState"}`
	UpgradePath []string
}

type Fraction struct {
	Numerator   uint64
	Denominator uint64
}

// Height is a monotonically increasing data type
// that can be compared against another Height for the purposes of updating and
// freezing clients
//
// Normally the RevisionHeight is incremented at each height while keeping
// RevisionNumber the same. However some consensus algorithms may choose to
// reset the height in certain conditions e.g. hard forks, state-machine
// breaking changes In these cases, the RevisionNumber is incremented so that
// height continues to be monitonically increasing even as the RevisionHeight
// gets reset
//
// Please note that json tags for generated Go code are overridden to explicitly exclude the omitempty jsontag.
// This enforces the Go json marshaller to always emit zero values for both revision_number and revision_height.
type Height struct {
	// the revision that the client is currently on
	RevisionNumber uint64
	// the height within the given revision
	RevisionHeight uint64
}

func (h Height) String() string {
	return ufmt.Sprintf("%d/%d", h.RevisionNumber, h.RevisionHeight)
}

// ConsensusState defines the consensus state from Tendermint.
type ConsensusState struct {
	// timestamp that corresponds to the block height in which the ConsensusState
	// was stored.
	Timestamp time.Time
	// commitment root (i.e app hash)
	Root               []byte
	NextValidatorsHash []byte
}
