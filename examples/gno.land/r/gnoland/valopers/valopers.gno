// Package valopers is designed around the permissionless lifecycle of valoper profiles.
// It also includes parts designed for govdao to propose valset changes based on registered valopers.
package valopers

import (
	"std"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/profile"
)

const (
	ErrValoperExists        = "valoper already exists"
	ErrValoperMissing       = "valoper does not exist"
	ErrInvalidAddressUpdate = "valoper updated address exists"
	ErrValoperNotCaller     = "valoper is not the caller"
)

// valopers keeps track of all the active validator operators
var valopers *avl.Tree // Address -> Valoper

// Valoper represents a validator operator profile
type Valoper struct {
	Name               string // the display name of the valoper
	Description        string // the description of the valoper
	ProfileAddress     string // The address of the r/demo/profile you want to link to
	OtherNetworks      string // What networks are you currently validating?
	DigitalPresence    string // Website, social links, contact details
	AssetsManagement   string // How do you manage your assets?
	ValidatingInterest string // Why are you interested in validating on gno.land?
	Contributions      string // What contributions have you made to gno.land? Or what contributions are you willing to make to gno.land?

	Address      std.Address // The bech32 gno address of the validator
	PubKey       string      // the bech32 public key of the validator
	P2PAddresses []string    // the publicly reachable P2P addresses of the validator
	Active       bool        // flag indicating if the valoper is active
}

// Register registers a new valoper
func Register(v Valoper) {
	// Check if the valoper is already registered
	if isValoper(v.Address) {
		panic(ErrValoperExists)
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
		panic(ErrValoperMissing)
	}

	// Check that the valoper wouldn't be
	// overwriting an existing one
	isAddressUpdate := address != v.Address
	if isAddressUpdate && isValoper(v.Address) {
		panic(ErrInvalidAddressUpdate)
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
		panic(ErrValoperMissing)
	}

	return valoperRaw.(Valoper)
}

// Render renders the current valoper set
func Render(_ string) string {
	if valopers.Size() == 0 {
		return "No valopers to display."
	}

	output := "Valset changes to apply:\n"
	valopers.Iterate("", "", func(_ string, value any) bool {
		valoper := value.(Valoper)

		output += valoper.Render()

		return false
	})

	return output
}

// Render renders a single valoper with their information
func (v Valoper) Render() string {
	output := ufmt.Sprintf("## %s\n", v.Name)

	if v.Description != "" {
		output += ufmt.Sprintf("%s\n\n", v.Description)
	}

	if v.ProfileAddress != "" {
		profile := profile.Render(ufmt.Sprintf("u/%s", v.ProfileAddress))
		output += ufmt.Sprintf("%s\n\n", profile)
	}

	if v.OtherNetworks != "" {
		output += ufmt.Sprintf("What networks are you currently validating?\n%s\n", v.OtherNetworks)
	}

	if v.DigitalPresence != "" {
		output += ufmt.Sprintf("Digital presence?\n%s\n", v.DigitalPresence)
	}

	if v.AssetsManagement != "" {
		output += ufmt.Sprintf("Assets under management?\n%s\n", v.AssetsManagement)
	}

	if v.ValidatingInterest != "" {
		output += ufmt.Sprintf("Why are you interested in validating on gno.land?\n%s\n", v.ValidatingInterest)
	}

	if v.Contributions != "" {
		output += ufmt.Sprintf("What contributions have you made to gno.land? Or what contributions are you willing to make to gno.land?\n%s\n", v.Contributions)
	}
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
