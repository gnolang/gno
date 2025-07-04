package tendermint

import (
	"time"

	"gno.land/p/aib/ibc/types"
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
	FrozenHeight types.Height
	// Latest height the client was updated to
	LatestHeight types.Height

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

// ConsensusState defines the consensus state from Tendermint.
type ConsensusState struct {
	// timestamp that corresponds to the block height in which the ConsensusState
	// was stored.
	Timestamp time.Time
	// commitment root (i.e app hash)
	Root               []byte
	NextValidatorsHash []byte
}
