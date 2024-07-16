// Package valopers is designed around the permissionless lifecycle of valoper profiles.
// It also includes parts designed for govdao to propose valset changes based on registered valopers.
package valopers

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	pVals "gno.land/p/sys/validators"
	govdao "gno.land/r/gov/dao/v2"
	validators "gno.land/r/sys/validators/v2"
)

const (
	errValoperExists        = "valoper already exists"
	errValoperMissing       = "valoper does not exist"
	errInvalidAddressUpdate = "valoper updated address exists"
	errValoperNotCaller     = "valoper is not the caller"
)

// valopers keeps track of all the active validator operators
var valopers *avl.Tree // Address -> Valoper

// Valoper represents a validator operator profile
type Valoper struct {
	Name        string // the display name of the valoper
	Moniker     string // the moniker of the valoper
	Description string // the description of the valoper

	Address      std.Address // The bech32 gno address of the validator
	PubKey       string      // the bech32 public key of the validator
	P2PAddresses []string    // the publicly reachable P2P addresses of the validator
	Active       bool        // flag indicating if the valoper is active
}

// Register registers a new valoper
func Register(v Valoper) {
	// Check if the valoper is already registered
	if isValoper(v.Address) {
		panic(errValoperExists)
	}

	// TODO add address derivation from public key
	// (when the laws of gno make it possible)

	// Save the valoper to the set
	valopers.Set(v.Address.String(), v)
}

// Update updates an existing valoper
func Update(address std.Address, v Valoper) {
	// Check if the valoper is present
	if !isValoper(address) {
		panic(errValoperMissing)
	}

	// Check that the valoper wouldn't be
	// overwriting an existing one
	isAddressUpdate := address != v.Address
	if isAddressUpdate && isValoper(v.Address) {
		panic(errInvalidAddressUpdate)
	}

	// Remove the old valoper info
	// in case the address changed
	if address != v.Address {
		valopers.Remove(address.String())
	}

	// Save the new valoper info
	valopers.Set(v.Address.String(), v)
}

// GetByAddr fetches the valoper using the address, if present
func GetByAddr(address std.Address) Valoper {
	valoperRaw, exists := valopers.Get(address.String())
	if !exists {
		panic(errValoperMissing)
	}

	return valoperRaw.(Valoper)
}

// Render renders the current valoper set
func Render(_ string) string {
	if valopers.Size() == 0 {
		return "No valopers to display."
	}

	output := "Valset changes to apply:\n"
	valopers.Iterate("", "", func(_ string, value interface{}) bool {
		valoper := value.(Valoper)

		output += valoper.Render()

		return false
	})

	return output
}

// Render renders a single valoper with their information
func (v Valoper) Render() string {
	output := ufmt.Sprintf("## %s (%s)\n", v.Name, v.Moniker)
	output += ufmt.Sprintf("%s\n\n", v.Description)
	output += ufmt.Sprintf("- Address: %s\n", v.Address.String())
	output += ufmt.Sprintf("- PubKey: %s\n", v.PubKey)
	output += "- P2P Addresses: [\n"

	if len(v.P2PAddresses) == 0 {
		output += "]\n"

		return output
	}

	for index, addr := range v.P2PAddresses {
		output += addr

		if index == len(v.P2PAddresses)-1 {
			output += "]\n"

			continue
		}

		output += ",\n"
	}

	return output
}

// isValoper checks if the valoper exists
func isValoper(address std.Address) bool {
	_, exists := valopers.Get(address.String())

	return exists
}

// GovDAOProposal creates a proposal to the GovDAO
// for adding the given valoper to the validator set.
// This function is meant to serve as a helper
// for generating the govdao proposal
func GovDAOProposal(address std.Address) {
	var (
		valoper     = GetByAddr(address)
		votingPower = uint64(1)
	)

	// Make sure the valoper is the caller
	if std.GetOrigCaller() != address {
		panic(errValoperNotCaller)
	}

	// Determine the voting power
	if !valoper.Active {
		votingPower = uint64(0)
	}

	changesFn := func() []pVals.Validator {
		return []pVals.Validator{
			{
				Address:     valoper.Address,
				PubKey:      valoper.PubKey,
				VotingPower: votingPower,
			},
		}
	}

	// Create the executor
	executor := validators.NewPropExecutor(changesFn)

	// Craft the proposal comment
	comment := ufmt.Sprintf(
		"Proposal to add valoper %s (Address: %s; PubKey: %s) to the valset",
		valoper.Name,
		valoper.Address.String(),
		valoper.PubKey,
	)

	// Create the govdao proposal
	govdao.Propose(comment, executor)
}
