// Package valopers is designed around the permissionless lifecycle of valoper profiles.
// It also includes parts designed for govdao to propose valset changes based on registered valopers.
package valopers

import (
	"crypto/bech32"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/profile"
)

const (
	ErrValoperExists       = "valoper already exists"
	ErrValoperMissing      = "valoper does not exist"
	ErrInvalidAddress      = "valoper updated address exists"
	ErrValoperNotCaller    = "valoper is not the caller"
	ErrInvalidMoniker      = "moniker cannot be empty"
	ErrInvalidDescription  = "description cannot be empty"
	ErrInvalidPubKey       = "invalid public key"
	ErrInvalidP2PAddresses = "invalid P2P addresses"
)

// valopers keeps track of all the active validator operators
var valopers *avl.Tree // Address -> Valoper

// Valoper represents a validator operator profile
type Valoper struct {
	Moniker     string // Human-readable name
	Description string // Description and details about the valoper

	Address std.Address // The bech32 gno address of the validator
	PubKey  string      // the bech32 public key of the validator
	Active  bool        // flag indicating if the valoper is active

	Caller std.Address // The address of the caller
}

// Register registers a new valoper
func Register(moniker string, description string, address std.Address, pubKey string, active bool) {
>>>>>>> 8a4a916ec (feat: use Description to reply questions)
	// Check if the valoper is already registered
	if isValoper(address) {
		panic(ErrValoperExists)
	}

	v := Valoper{
		Moniker:     moniker,
		Description: description,
		Address:     address,
		PubKey:      pubKey,
		Active:      active,
		Caller:      std.PreviousRealm().Address(),
	}

	if err := v.Validate(); err != nil {
		panic(err)
	}

	// TODO add address derivation from public key
	// (when the laws of gno make it possible)

	// Save the valoper to the set
	valopers.Set(v.Address.String(), v)
}

// Update updates an existing valoper
func Update(address std.Address, v Valoper) {
	old := GetByAddr(address)

	// Check that the caller is the valoper
	if old.Caller != std.PreviousRealm().Address() {
		panic(ErrValoperNotCaller)
	}

	// Check that the valoper wouldn't be
	// overwriting an existing one
	isAddressUpdate := address != v.Address
	if isAddressUpdate && isValoper(v.Address) {
		panic(ErrInvalidAddress)
	}

	if err := v.Validate(); err != nil {
		panic(err)
	}

	// Remove the old valoper info
	// in case the address changed
	if address != v.Address {
		valopers.Remove(address.String())
	}

	// Save the new valoper info
	valopers.Set(v.Address.String(), v)
}

// UpdateMoniker updates an existing valoper's moniker
func UpdateMoniker(address std.Address, moniker string) {
	v := GetByAddr(address)

	// Check that the caller is the valoper
	if v.Caller != std.PreviousRealm().Address() {
		panic(ErrValoperNotCaller)
	}

	// Check that the moniker is not empty
	if moniker == "" {
		panic(ErrInvalidMoniker)
	}

	// Update the moniker
	v.Moniker = moniker

	// Save the valoper info
	valopers.Set(address.String(), v)
}

// UpdateDescription updates an existing valoper's description
func UpdateDescription(address std.Address, description string) {
	v := GetByAddr(address)

	// Check that the caller is the valoper
	if v.Caller != std.PreviousRealm().Address() {
		panic(ErrValoperNotCaller)
	}

	// Check that the description is not empty
	if description == "" {
		panic(ErrInvalidDescription)
	}

	// Update the description
	v.Description = description

	// Save the valoper info
	valopers.Set(address.String(), v)
}

// UpdateAddress updates an existing valoper's address
func UpdateAddress(address std.Address, newAddress std.Address) {
	// Check that the new address is different
	if address == newAddress {
		panic(ErrInvalidAddress)
	}

	v := GetByAddr(address)

	// Check that the caller is the valoper
	if v.Caller != std.PreviousRealm().Address() {
		panic(ErrValoperNotCaller)
	}

	// Check that the address is valid
	if err := validateAddress(newAddress); err != nil {
		panic(ErrInvalidAddress)
	}

	// Check that the new address wouldn't be
	// overwriting an existing one
	if isValoper(newAddress) {
		panic(ErrInvalidAddress)
	}

	// Update the address
	v.Address = newAddress

	// Remove the old valoper info
	valopers.Remove(address.String())

	// Save the new valoper info
	valopers.Set(newAddress.String(), v)
}

// UpdatePubKey updates an existing valoper's public key
func UpdatePubKey(address std.Address, pubKey string) {
	v := GetByAddr(address)

	// Check that the caller is the valoper
	if v.Caller != std.PreviousRealm().Address() {
		panic(ErrValoperNotCaller)
	}

	// Check that the pubKey is valid
	if err := validatePubKey(pubKey); err != nil {
		panic(ErrInvalidPubKey)
	}

	// Update the description
	v.PubKey = pubKey

	// Save the valoper info
	valopers.Set(address.String(), v)
}

// UpdateActive updates an existing valoper's active status
func UpdateActive(address std.Address, active bool) {
	v := GetByAddr(address)

	// Check that the caller is the valoper
	if v.Caller != std.PreviousRealm().Address() {
		panic(ErrValoperNotCaller)
	}

	// Update active status
	v.Active = active

	// Save the valoper info
	valopers.Set(address.String(), v)
}

// UpdateP2PAddresses updates an existing valoper's P2P addresses
// p2pAddresses is a comma-separated list of P2P addresses
func UpdateP2PAddresses(address std.Address, p2pAddresses string) {
	v := GetByAddr(address)

	// Check that the caller is the valoper
	if v.Caller != std.PreviousRealm().Address() {
		panic(ErrValoperNotCaller)
	}

	// Split the P2P addresses
	addresses := strings.Split(p2pAddresses, ",")
	if len(addresses) == 0 {
		panic(ErrInvalidP2PAddresses)
	}

	// Update the P2P addresses
	v.P2PAddresses = p2pAddresses
	// Save the valoper info
	valopers.Set(address.String(), v)
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

// Validate checks if the fields of the Valoper are valid
func (v *Valoper) Validate() error {
	if _, _, err := bech32.Decode(v.Address.String()); err != nil {
		return ufmt.Errorf("Valoper Address must be bech32 \"%s\": %s", v.Address.String(), err)
	}
	if _, _, err := bech32.DecodeNoLimit(v.PubKey); err != nil {
		return ufmt.Errorf("Valoper PubKey must be bech32: \"%s\": %s", v.PubKey, err)
	}

	return nil
}

// Render renders a single valoper with their information
func (v Valoper) Render() string {
	output := ufmt.Sprintf("## %s\n", v.Name)

	if v.Description != "" {
		output += ufmt.Sprintf("%s\n\n", v.Description)
	}

	output += ufmt.Sprintf("- Address: %s\n", v.Address.String())
	output += ufmt.Sprintf("- PubKey: %s\n", v.PubKey)
	output += "- P2P Addresses: [\n"

	if len(v.P2PAddresses) == 0 {
		output += "]\n"
	}

	for index, addr := range v.P2PAddresses {
		output += addr

		if index == len(v.P2PAddresses)-1 {
			output += "]\n"

			continue
		}

		output += ",\n"
	}

	if v.ProfileAddress != "" {
		profile := profile.Render(ufmt.Sprintf("u/%s", v.ProfileAddress))
		output += ufmt.Sprintf("%s\n", profile)
	}

	if v.OtherNetworks != "" {
		output += ufmt.Sprintf("- What networks are you currently validating? %s\n", v.OtherNetworks)
	}

	if v.DigitalPresence != "" {
		output += ufmt.Sprintf("- Digital presence? %s\n", v.DigitalPresence)
	}

	if v.AssetsManagement != "" {
		output += ufmt.Sprintf("- Assets under management? %s\n", v.AssetsManagement)
	}

	if v.ValidatingInterest != "" {
		output += ufmt.Sprintf("- Why are you interested in validating on gno.land? %s\n", v.ValidatingInterest)
	}

	if v.Contributions != "" {
		output += ufmt.Sprintf("- What contributions have you made to gno.land? Or what contributions are you willing to make to gno.land? %s\n", v.Contributions)
	}

	return output
}

// isValoper checks if the valoper exists
func isValoper(address std.Address) bool {
	_, exists := valopers.Get(address.String())

	return exists
}
