package types

import (
	"fmt"
	"os"
	"time"

	"github.com/gnolang/gno/tm2/pkg/amino"
	abci "github.com/gnolang/gno/tm2/pkg/bft/abci/types"
	tmtime "github.com/gnolang/gno/tm2/pkg/bft/types/time"
	"github.com/gnolang/gno/tm2/pkg/crypto"
	"github.com/gnolang/gno/tm2/pkg/errors"
	osm "github.com/gnolang/gno/tm2/pkg/os"
)

const (
	// MaxChainIDLen is a maximum length of the chain ID.
	MaxChainIDLen = 50
)

var (
	ErrEmptyChainID                = errors.New("chain ID is empty")
	ErrLongChainID                 = fmt.Errorf("chain ID cannot be longer than %d chars", MaxChainIDLen)
	ErrInvalidGenesisTime          = errors.New("invalid genesis time")
	ErrNoValidators                = errors.New("no validators in set")
	ErrInvalidValidatorVotingPower = errors.New("validator has no voting power")
	ErrInvalidValidatorAddress     = errors.New("invalid validator address")
	ErrValidatorPubKeyMismatch     = errors.New("validator public key and address mismatch")
)

// ------------------------------------------------------------
// core types for a genesis definition
// NOTE: any changes to the genesis definition should
// be reflected in the documentation:
// docs/tendermint-core/using-tendermint.md

// GenesisValidator is an initial validator.
type GenesisValidator struct {
	Address Address       `json:"address"`
	PubKey  crypto.PubKey `json:"pub_key"`
	Power   int64         `json:"power"`
	Name    string        `json:"name"`
}

// GenesisDoc defines the initial conditions for a tendermint blockchain, in particular its validator set.
type GenesisDoc struct {
	GenesisTime     time.Time            `json:"genesis_time"`
	ChainID         string               `json:"chain_id"`
	ConsensusParams abci.ConsensusParams `json:"consensus_params,omitempty"`
	Validators      []GenesisValidator   `json:"validators,omitempty"`
	AppHash         []byte               `json:"app_hash"`
	AppState        any                  `json:"app_state,omitempty"`
}

// SaveAs is a utility method for saving GenensisDoc as a JSON file.
func (genDoc *GenesisDoc) SaveAs(file string) error {
	genDocBytes, err := amino.MarshalJSONIndent(genDoc, "", "  ")
	if err != nil {
		return err
	}
	return osm.WriteFile(file, genDocBytes, 0o644)
}

// ValidatorHash returns the hash of the validator set contained in the GenesisDoc
func (genDoc *GenesisDoc) ValidatorHash() []byte {
	vals := make([]*Validator, len(genDoc.Validators))
	for i, v := range genDoc.Validators {
		vals[i] = NewValidator(v.PubKey, v.Power)
	}
	vset := NewValidatorSet(vals)
	return vset.Hash()
}

// Validate validates the genesis doc
func (genDoc *GenesisDoc) Validate() error {
	// Make sure the chain ID is not empty
	if genDoc.ChainID == "" {
		return ErrEmptyChainID
	}

	// Make sure the chain ID is < max chain ID length
	if len(genDoc.ChainID) > MaxChainIDLen {
		return ErrLongChainID
	}

	// Make sure the genesis time is valid
	if genDoc.GenesisTime.IsZero() {
		return ErrInvalidGenesisTime
	}

	// Validate the consensus params
	if consensusParamsErr := ValidateConsensusParams(genDoc.ConsensusParams); consensusParamsErr != nil {
		return consensusParamsErr
	}

	// Make sure there are validators in the set
	if len(genDoc.Validators) == 0 {
		return ErrNoValidators
	}

	// Make sure the validators are valid
	for _, v := range genDoc.Validators {
		// Check the voting power
		if v.Power == 0 {
			return fmt.Errorf("%w, %s", ErrInvalidValidatorVotingPower, v.Name)
		}

		// Check the address
		if v.Address.IsZero() {
			return fmt.Errorf("%w, %s", ErrInvalidValidatorAddress, v.Name)
		}

		// Check the pub key -> address matching
		if v.PubKey.Address() != v.Address {
			return fmt.Errorf("%w, %s", ErrValidatorPubKeyMismatch, v.Name)
		}
	}

	return nil
}

// ValidateAndComplete checks that all necessary fields are present
// and fills in defaults for optional fields left empty
func (genDoc *GenesisDoc) ValidateAndComplete() error {
	if genDoc.ChainID == "" {
		return errors.New("Genesis doc must include non-empty chain_id")
	}
	if len(genDoc.ChainID) > MaxChainIDLen {
		return errors.New("chain_id in genesis doc is too long (max: %d)", MaxChainIDLen)
	}

	// Start from defaults and fill in consensus params from GenesisDoc.
	genDoc.ConsensusParams = DefaultConsensusParams().Update(genDoc.ConsensusParams)
	if err := ValidateConsensusParams(genDoc.ConsensusParams); err != nil {
		return err
	}

	for i, v := range genDoc.Validators {
		if v.Power == 0 {
			return errors.New("The genesis file cannot contain validators with no voting power: %v", v)
		}
		if v.Address.IsZero() {
			genDoc.Validators[i].Address = v.PubKey.Address()
		} else if v.PubKey.Address() != v.Address {
			return errors.New("Incorrect address for validator %v in the genesis file, should be %v", v, v.PubKey.Address())
		}
	}

	if genDoc.GenesisTime.IsZero() {
		genDoc.GenesisTime = tmtime.Now()
	}

	return nil
}

// ------------------------------------------------------------
// Make genesis state from file

// GenesisDocFromJSON unmarshalls JSON data into a GenesisDoc.
func GenesisDocFromJSON(jsonBlob []byte) (*GenesisDoc, error) {
	genDoc := GenesisDoc{}
	err := amino.UnmarshalJSON(jsonBlob, &genDoc)
	if err != nil {
		return nil, err
	}

	if err := genDoc.ValidateAndComplete(); err != nil {
		return nil, err
	}

	return &genDoc, err
}

// GenesisDocFromFile reads JSON data from a file and unmarshalls it into a GenesisDoc.
func GenesisDocFromFile(genDocFile string) (*GenesisDoc, error) {
	jsonBlob, err := os.ReadFile(genDocFile)
	if err != nil {
		return nil, errors.Wrap(err, "Couldn't read GenesisDoc file")
	}
	genDoc, err := GenesisDocFromJSON(jsonBlob)
	if err != nil {
		return nil, errors.Wrapf(err, "Error reading GenesisDoc at %v", genDocFile)
	}
	return genDoc, nil
}

// ----------------------------------------
// Mock AppState (for testing)

type MockAppState struct {
	AccountOwner string `json:"account_owner"`
}
