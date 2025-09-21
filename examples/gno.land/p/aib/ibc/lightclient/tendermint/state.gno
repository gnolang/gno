package tendermint

import (
	"crypto/sha256"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"time"

	"gno.land/p/aib/ibc/lightclient"
	"gno.land/p/aib/ibc/types"
	"gno.land/p/nt/ufmt"
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
	ProofSpecs []*types.ProofSpec `json:"proof_specs,omitempty"`

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

func NewClientState(chainID string, trustLevel Fraction,
	trustingPeriod, ubdPeriod, maxClockDrift time.Duration,
	latestHeight types.Height, specs []*types.ProofSpec, upgradePath []string) *ClientState {
	return &ClientState{
		ChainId:         chainID,
		TrustLevel:      trustLevel,
		TrustingPeriod:  trustingPeriod,
		UnbondingPeriod: ubdPeriod,
		MaxClockDrift:   maxClockDrift,
		LatestHeight:    latestHeight,
		FrozenHeight:    types.ZeroHeight(),
		ProofSpecs:      specs,
		UpgradePath:     upgradePath,
	}
}

type Fraction struct {
	Numerator   uint64 `json:"numerator,omitempty"`
	Denominator uint64 `json:"denominator,omitempty"`
}

func NewFraction(numerator, denominator uint64) Fraction {
	return Fraction{
		Numerator:   numerator,
		Denominator: denominator,
	}
}

// DefaultTrustLevel is the tendermint light client default trust level
var DefaultTrustLevel = NewFraction(1, 3)

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

func NewMerkleRoot(hash []byte) MerkleRoot {
	return MerkleRoot{Hash: hash}
}

// Empty returns true if the root is empty
func (mr MerkleRoot) Empty() bool {
	return len(mr.Hash) == 0
}

const (
	// MaxChainIDLen is a maximum length of the chain ID.
	MaxChainIDLen = 50
	// SentinelRoot is used as a stand-in root value for the consensus state set at the upgrade height
	SentinelRoot = "sentinel_root"
)

// Implements lightclient.ClientState
func (ClientState) ClientType() string {
	return lightclient.Tendermint
}

// Implements lightclient.ClientState
func (cs ClientState) Validate() error {
	if strings.TrimSpace(cs.ChainId) == "" {
		return ufmt.Errorf("chainID cannot be empty string")
	}

	if len(cs.ChainId) > MaxChainIDLen {
		return ufmt.Errorf("chainID is too long; got: %d, max: %d", len(cs.ChainId), MaxChainIDLen)
	}

	if err := ValidateTrustLevel(cs.TrustLevel); err != nil {
		return err
	}
	if cs.TrustingPeriod <= 0 {
		return errors.New("trusting period must be greater than zero")
	}
	if cs.UnbondingPeriod <= 0 {
		return errors.New("unbonding period must be greater than zero")
	}
	if cs.MaxClockDrift <= 0 {
		return errors.New("max clock drift must be greater than zero")
	}

	// the latest height revision number must match the chain id revision number
	if cs.LatestHeight.RevisionNumber != ParseChainID(cs.ChainId) {
		return ufmt.Errorf("latest height revision number must match chain id revision number (%d != %d)", cs.LatestHeight.RevisionNumber, ParseChainID(cs.ChainId))
	}
	if cs.LatestHeight.RevisionHeight == 0 {
		return ufmt.Errorf("tendermint client's latest height revision height cannot be zero")
	}
	if cs.TrustingPeriod >= cs.UnbondingPeriod {
		return ufmt.Errorf("trusting period (%s) should be < unbonding period (%s)", cs.TrustingPeriod, cs.UnbondingPeriod)
	}

	if cs.ProofSpecs == nil {
		return ufmt.Errorf("proof specs cannot be nil for tm client")
	}
	for i, spec := range cs.ProofSpecs {
		if spec == nil {
			return ufmt.Errorf("proof spec cannot be nil at index: %d", i)
		}
	}
	// UpgradePath may be empty, but if it isn't, each key must be non-empty
	for i, k := range cs.UpgradePath {
		if strings.TrimSpace(k) == "" {
			return ufmt.Errorf("key in upgrade path at index %d cannot be empty", i)
		}
	}

	return nil
}

// ValidateTrustLevel checks that trustLevel is within the allowed range [1/3,
// 1]. If not, it returns an error. 1/3 is the minimum amount of trust needed
// which does not break the security model.
func ValidateTrustLevel(lvl Fraction) error {
	if lvl.Numerator*3 < lvl.Denominator || // < 1/3
		lvl.Numerator > lvl.Denominator || // > 1
		lvl.Denominator == 0 {
		return ufmt.Errorf("trustLevel must be within [1/3, 1], given %v", lvl)
	}
	return nil
}

// IsRevisionFormat checks if a chainID is in the format required for parsing revisions
// The chainID must be in the form: `{chainID}-{revision}`.
// 24-host may enforce stricter checks on chainID
var IsRevisionFormat = regexp.MustCompile(`^.*[^\n-]-{1}[1-9][0-9]*$`).MatchString

// ParseChainID is a utility function that returns a revision number from the given ChainID.
// ParseChainID attempts to parse a chain id in the format: `{chainID}-{revision}`
// and return the revisionnumber as a uint64.
// If the chainID is not in the expected format, a default revision value of 0 is returned.
func ParseChainID(chainID string) uint64 {
	if !IsRevisionFormat(chainID) {
		// chainID is not in revision format, return 0 as default
		return 0
	}
	splitStr := strings.Split(chainID, "-")
	revision, err := strconv.ParseUint(splitStr[len(splitStr)-1], 10, 64)
	// sanity check: error should always be nil since regex only allows numbers in last element
	if err != nil {
		panic(ufmt.Errorf("regex allowed non-number value as last split element for chainID: %s", chainID))
	}
	return revision
}

// Implements lightclient.ConsensusState
func (ConsensusState) ClientType() string {
	return lightclient.Tendermint
}

// Implements lightclient.ConsensusState
func (cs ConsensusState) Validate() error {
	if cs.Root.Empty() {
		return errors.New("root cannot be empty")
	}
	if err := validateHash(cs.NextValidatorsHash); err != nil {
		return ufmt.Errorf("next validators hash is invalid: %v", err)
	}
	if cs.Timestamp.Unix() <= 0 {
		return errors.New("timestamp must be a positive Unix time")
	}
	return nil
}

// validateHash returns an error if the hash is not empty, but its
// size != tmhash.Size.
func validateHash(h []byte) error {
	if len(h) > 0 && len(h) != sha256.Size {
		return ufmt.Errorf("expected size to be %d bytes, got %d bytes",
			sha256.Size,
			len(h),
		)
	}
	return nil
}
