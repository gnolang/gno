// Package valopers is designed around the permissionless lifecycle of valoper profiles.
// It also includes parts designed for govdao to propose valset changes based on registered valopers.
package valopers

import (
	"crypto/bech32"
	"errors"
	"regexp"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/combinederr"
	"gno.land/p/demo/ufmt"
	"gno.land/r/demo/profile"
)

var (
	ErrValoperExists       = errors.New("valoper already exists")
	ErrValoperMissing      = errors.New("valoper does not exist")
	ErrInvalidAddress      = errors.New("valoper updated address exists")
	ErrValoperNotCaller    = errors.New("valoper is not the caller")
	ErrInvalidMoniker      = errors.New("moniker cannot be empty")
	ErrInvalidDescription  = errors.New("description cannot be empty")
	ErrInvalidPubKey       = errors.New("invalid public key")
	ErrInvalidP2PAddress   = errors.New("invalid P2P address")
	ErrP2PAddressesMissing = errors.New("P2P addresses cannot be empty")
)

// valopers keeps track of all the active validator operators
var valopers *avl.Tree // Address -> Valoper

// Valoper represents a validator operator profile
type Valoper struct {
	Moniker     string // Human-readable name
	Description string // Description and details about the valoper

	Address      std.Address // The bech32 gno address of the validator
	PubKey       string      // the bech32 public key of the validator
	Active       bool        // flag indicating if the valoper is active
	P2PAddresses []string    // the publicly reachable P2P addresses of the validator

	Caller std.Address // The address of the caller
}

// Register registers a new valoper
func Register(moniker string, description string, address std.Address, pubKey string, p2pAddresses string) {
	// Check if the valoper is already registered
	if isValoper(address) {
		panic(ErrValoperExists)
	}

	v := Valoper{
		Moniker:      moniker,
		Description:  description,
		Address:      address,
		PubKey:       pubKey,
		Active:       true,
		P2PAddresses: splitP2PAddresses(p2pAddresses),
		Caller:       std.PreviousRealm().Address(),
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
	if err := validateMoniker(moniker); err != nil {
		panic(err)
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
	if err := validateDescription(description); err != nil {
		panic(err)
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
	if err := validateBech32(newAddress); err != nil {
		panic(ufmt.Errorf("Valoper Address must be bech32 \"%s\": %s", newAddress.String(), err))
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
		panic(err)
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
	addresses := splitP2PAddresses(p2pAddresses)

	if len(addresses) == 0 {
		panic(ErrP2PAddressesMissing)
	}

	// Check addresses
	for _, addr := range addresses {
		if err := validateP2PAddress(addr); err != nil {
			panic(ErrInvalidP2PAddress)
		}
	}

	// Update the P2P addresses
	v.P2PAddresses = addresses

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
	errs := &combinederr.CombinedError{}

	errs = errs.Add(validateMoniker(v.Moniker))
	errs = errs.Add(validateDescription(v.Description))
	errs = errs.Add(validateBech32(v.Address))
	errs = errs.Add(validatePubKey(v.PubKey))

	for _, addr := range v.P2PAddresses {
		errs = errs.Add(validateP2PAddress(addr))
	}

	if errs.Size() == 0 {
		return nil
	}

	return errs
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

func splitP2PAddresses(p2pAddresses string) []string {
	return strings.Split(p2pAddresses, ",")
}

// validateMoniker checks if the moniker is valid
func validateMoniker(moniker string) error {
	if moniker == "" {
		return ErrInvalidMoniker
	}

	return nil
}

// validateDescription checks if the description is valid
func validateDescription(description string) error {
	if description == "" {
		return ErrInvalidDescription
	}

	return nil
}

// validateBech32 checks if the value is a valid bech32 address
func validateBech32(address string) error {
	if _, _, err := bech32.Decode(address); err != nil {
		return err
	}

	return nil
}

// validatePubKey checks if the public key is valid
func validatePubKey(pubKey string) error {
	if _, _, err := bech32.DecodeNoLimit(pubKey); err != nil {
		return ufmt.Errorf("Valoper PubKey must be bech32: \"%s\": %s", pubKey, err)
	}

	return nil
}

func isValidIPv4(ip string) bool {
	ipRegex := regexp.MustCompile("^([0-9]{1,3}\\.){3}[0-9]{1,3}(\\/([0-9]|[1-2][0-9]|3[0-2]))?$")
	return ipRegex.MatchString(ip)
}

func isValidIPv6(ip string) bool {
	ipRegex := regexp.MustCompile(`^s*((([0-9A-Fa-f]{1,4}:){7}([0-9A-Fa-f]{1,4}|:))|(([0-9A-Fa-f]{1,4}:){6}(:[0-9A-Fa-f]{1,4}|((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d|1dd|[1-9]?d)){3})|:))|(([0-9A-Fa-f]{1,4}:){5}(((:[0-9A-Fa-f]{1,4}){1,2})|:((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d|1dd|[1-9]?d)){3})|:))|(([0-9A-Fa-f]{1,4}:){4}(((:[0-9A-Fa-f]{1,4}){1,3})|((:[0-9A-Fa-f]{1,4})?:((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d|1dd|[1-9]?d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){3}(((:[0-9A-Fa-f]{1,4}){1,4})|((:[0-9A-Fa-f]{1,4}){0,2}:((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d|1dd|[1-9]?d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){2}(((:[0-9A-Fa-f]{1,4}){1,5})|((:[0-9A-Fa-f]{1,4}){0,3}:((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d|1dd|[1-9]?d)){3}))|:))|(([0-9A-Fa-f]{1,4}:){1}(((:[0-9A-Fa-f]{1,4}){1,6})|((:[0-9A-Fa-f]{1,4}){0,4}:((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d|1dd|[1-9]?d)){3}))|:))|(:(((:[0-9A-Fa-f]{1,4}){1,7})|((:[0-9A-Fa-f]{1,4}){0,5}:((25[0-5]|2[0-4]d|1dd|[1-9]?d)(.(25[0-5]|2[0-4]d|1dd|[1-9]?d)){3}))|:)))(%.+)?s*(\/([0-9]|[1-9][0-9]|1[0-1][0-9]|12[0-8]))?$`)
	return ipRegex.MatchString(ip)
}

// isValidIP checks both IPv4 and IPv6
func isValidIP(ip string) bool {
	return isValidIPv4(ip) || isValidIPv6(ip)
}

func isValidHostname(hostname string) bool {
	// Validation rules for a hostname:
	// 1. Total length max of 255 characters
	// 2. Each label (part between dots):
	//    - Max 63 characters
	//    - Starts and ends with an alphanumeric character
	//    - Contains only alphanumeric characters and hyphens
	if len(hostname) > 255 {
		return false
	}

	// Regular expression to validate hostname labels
	hostnameRegex := regexp.MustCompile(`^([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])(\.([a-zA-Z0-9]|[a-zA-Z0-9][a-zA-Z0-9\-]{0,61}[a-zA-Z0-9]))*$`)

	return hostnameRegex.MatchString(hostname)
}

// validateP2PAddress checks if the P2P address is valid
// address is of the form: <ID>@<HOSTNAME> or <ID>@<IP>
func validateP2PAddress(p2pAddress string) error {
	parts := strings.Split(p2pAddress, "@")

	if len(parts) != 2 {
		return ufmt.Errorf("P2P Address must be of the form <ID>@<HOSTNAME> or <ID>@<IP>: %s", p2pAddress)
	}

	// test part 1 which is the ID
	if err := validateBech32(parts[0]); err != nil {
		return ufmt.Errorf("ID is not a valid bech32 address: %s", p2pAddress)
	}

	// test part 2 which is the hostname or IP
	if !isValidIP(parts[1]) && !isValidHostname(parts[1]) {
		return ufmt.Errorf("Invalid IP/hostname: %s", p2pAddress)
	}

	return nil
}
