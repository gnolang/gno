package tendermint

import (
	"time"

	"gno.land/p/aib/ibc/types"
)

// ClientState from Tendermint tracks the current validator set, latest height,
// and a possible frozen height.
type ClientState struct {
	ChainId    string   `json:"chain_id,omitempty"`
	TrustLevel Fraction `json:"trust_level"`
	// duration of the period since the LatestTimestamp during which the
	// submitted headers are valid for upgrade
	TrustingPeriod time.Duration `json:"trusting_period"`
	// duration of the staking unbonding period
	UnbondingPeriod time.Duration `json:"unbonding_period"`
	// defines how much new (untrusted) header's Time can drift into the future.
	MaxClockDrift time.Duration `json:"max_clock_drift"`
	// Block height when the client was frozen due to a misbehaviour
	FrozenHeight types.Height `json:"frozen_height"`
	// Latest height the client was updated to
	LatestHeight types.Height `json:"latest_height"`
	// Proof specifications used in verifying counterparty state
	// TODO write gno type for proofspec
	//ProofSpecs []*_go.ProofSpec `json:"proof_specs,omitempty"`

	// Path at which next upgraded client will be committed.
	// Each element corresponds to the key for a single CommitmentProof in the
	// chained proof. NOTE: ClientState must stored under
	// `{upgradePath}/{upgradeHeight}/clientState` ConsensusState must be stored
	// under `{upgradepath}/{upgradeHeight}/consensusState` For SDK chains using
	// the default upgrade module, upgrade_path should be []string{"upgrade",
	// "upgradedIBCState"}`
	UpgradePath []string `json:"upgrade_path,omitempty"`
	// allow_update_after_expiry is deprecated
	AllowUpdateAfterExpiry bool `json:"allow_update_after_expiry,omitempty"` // Deprecated: Do not use.
	// allow_update_after_misbehaviour is deprecated
	AllowUpdateAfterMisbehaviour bool `json:"allow_update_after_misbehaviour,omitempty"` // Deprecated: Do not use.
}

type Fraction struct {
	Numerator   uint64 `json:"numerator,omitempty"`
	Denominator uint64 `json:"denominator,omitempty"`
}

// ConsensusState defines the consensus state from Tendermint.
type ConsensusState struct {
	// timestamp that corresponds to the block height in which the ConsensusState
	// was stored.
	Timestamp time.Time `json:"timestamp"`
	// commitment root (i.e app hash)
	Root               MerkleRoot `json:"root"`
	NextValidatorsHash []byte     `json:"next_validator_hash,omitempty"`
}

type MerkleRoot struct {
	Hash []byte `json:"hash,omitempty"`
}
