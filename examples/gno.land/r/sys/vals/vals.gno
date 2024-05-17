package vals

import (
	"gno.land/p/demo/json"
	"gno.land/p/demo/ufmt"
	"gno.land/r/sys/vals/pos"
	"gno.land/r/sys/vals/types"
)

// Protocol defines the validator set protocol (PoA / PoS)
type Protocol interface {
	// AddValidator adds a new validator to the validator set.
	// If the validator is already present, the method should error out
	AddValidator(address, pubKey string, votingPower uint64)

	// RemoveValidator removes the given validator from the set.
	// If the validator is not present in the set, the method should error out
	RemoveValidator(address string)

	// IsValidator returns a flag indicating if the given
	// bech32 address is part of the validator set
	IsValidator(address string) bool

	// GetValidators returns the JSON-encoded validator set.
	// The reason for returning JSON, and not a `[]Validator`,
	// is because Gno does not yet support an ABI-like functionality
	// for clients interacting with the Realm / Package methods
	GetValidators() []types.Validator
}

// p is the underlying validator set protocol (PoA / PoS)
var p Protocol = pos.NewPoS()

// AddValidator adds a new validator to the validator set.
// If the validator is already present, the method errors out
func AddValidator(address, pubKey string, votingPower uint64) {
	p.AddValidator(address, pubKey, votingPower)
}

// RemoveValidator removes the given validator from the set.
// If the validator is not present in the set, the method errors out
func RemoveValidator(address string) {
	p.RemoveValidator(address)
}

// IsValidator returns a flag indicating if the given bech32 address
// is part of the validator set
func IsValidator(address string) bool {
	return p.IsValidator(address)
}

// GetValidators returns the JSON-encoded validator set
func GetValidators() string {
	encodedSet, err := json.Marshal(p.GetValidators())
	if err != nil {
		panic(ufmt.Sprintf("unable to marshal set, %s", err))
	}

	return string(encodedSet)
}
