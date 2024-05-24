package vals

import (
	"std"

	"gno.land/p/demo/json"
	"gno.land/p/demo/ufmt"
	"gno.land/r/sys/vals/poa"
	"gno.land/r/sys/vals/types"
)

// Protocol defines the validator set protocol (PoA / PoS)
type Protocol interface {
	// AddValidator adds a new validator to the validator set.
	// If the validator is already present, the method should error out
	//
	// TODO: This API is not ideal -- the address should be derived from
	// the public key, and not be passed in as such, but currently Gno
	// does not support crypto address derivation
	AddValidator(address std.Address, pubKey string)

	// RemoveValidator removes the given validator from the set.
	// If the validator is not present in the set, the method should error out
	RemoveValidator(address std.Address)

	// IsValidator returns a flag indicating if the given
	// bech32 address is part of the validator set
	IsValidator(address std.Address) bool

	// GetValidators returns the JSON-encoded validator set.
	// The reason for returning JSON, and not a `[]Validator`,
	// is because Gno does not yet support an ABI-like functionality
	// for clients interacting with the Realm / Package methods
	GetValidators() []*types.Validator
}

// p is the underlying validator set protocol (PoA / PoS)
var p Protocol = poa.NewPoA()

// AddValidator adds a new validator to the validator set.
// If the validator is already present, the method errors out
func AddValidator(address, pubKey string) {
	p.AddValidator(std.Address(address), pubKey)
}

// RemoveValidator removes the given validator from the set.
// If the validator is not present in the set, the method errors out
func RemoveValidator(address string) {
	p.RemoveValidator(std.Address(address))
}

// IsValidator returns a flag indicating if the given bech32 address
// is part of the validator set
func IsValidator(address string) bool {
	return p.IsValidator(std.Address(address))
}

// GetValidators returns the JSON-encoded validator set
func GetValidators() string {
	encodedSet, err := json.Marshal(prepareForJSON(p.GetValidators()))
	if err != nil {
		panic(ufmt.Sprintf("unable to marshal set, %s", err))
	}

	return string(encodedSet)
}

// prepareForJSON prepares the validator set for JSON encoding
func prepareForJSON(vals []*types.Validator) *json.Node {
	nodes := make([]*json.Node, 0, len(vals))

	for _, v := range vals {
		node := json.ObjectNode("", map[string]*json.Node{
			"address":      json.StringNode("address", v.Address.String()),
			"pub_key":      json.StringNode("pub_key", v.PubKey),
			"voting_power": json.NumberNode("voting_power", float64(v.VotingPower)),
		})

		nodes = append(nodes, node)
	}

	return json.ArrayNode("", nodes)
}
