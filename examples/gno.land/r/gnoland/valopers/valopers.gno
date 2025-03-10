// Package valopers is designed around the permissionless lifecycle of valoper profiles.
package valopers

import (
	"crypto/bech32"
	"errors"
	"regexp"
	"std"
	"strconv"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/combinederr"
	"gno.land/p/demo/ufmt"
)

var (
	ErrValoperExists       = errors.New("valoper already exists")
	ErrValoperMissing      = errors.New("valoper does not exist")
	ErrAddressExists       = errors.New("valoper updated address exists")
	ErrSameAddress         = errors.New("valoper updated address is the same")
	ErrInvalidAddress      = errors.New("invalid address")
	ErrOriginCaller        = errors.New("origin call is not the same")
	ErrInvalidMoniker      = errors.New("moniker cannot be empty")
	ErrInvalidDescription  = errors.New("description cannot be empty")
	ErrInvalidPubKey       = errors.New("invalid public key")
	ErrInvalidP2PAddress   = errors.New("invalid P2P address")
	ErrP2PAddressesMissing = errors.New("P2P addresses cannot be empty")

	minFee     int64 = 20 * 1_000_000 // minimum gnot must be paid to register.
	maxFeeMult int64 = 10             // maximum multiples of minFee accepted.
)

// valopers keeps track of all the active validator operators
var valopers *avl.Tree  // Address -> Valoper
var instructions string // markdown instructions for valoper's registration

// Valoper represents a validator operator profile
type Valoper struct {
	Moniker     string // A human-readable name
	Description string // A description and details about the valoper

	Address      std.Address // The bech32 gno address of the validator
	PubKey       string      // The bech32 public key of the validator
	Active       bool        // Flag indicating if the valoper is active
	P2PAddresses []string    // The publicly reachable P2P addresses of the validator

	Caller std.Address // The address of the caller
}

// Register registers a new valoper
func Register(moniker string, description string, address std.Address, pubKey string, p2pAddresses string) {
	// Check if the valoper is already registered
	if isValoper(address) {
		panic(ErrValoperExists)
	}

	if err := validateSentCoins(); err != nil {
		panic(err)
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
func Update(address std.Address, moniker string, description string, newAddress std.Address, pubKey string, p2pAddresses string, active bool) {
	old := GetByAddr(address)

	// Check that the caller is the valoper
	if old.Caller != std.PreviousRealm().Address() {
		panic(ErrOriginCaller)
	}

	// Check that the valoper wouldn't be
	// overwriting an existing one
	isAddressUpdate := address != newAddress
	if isAddressUpdate && isValoper(newAddress) {
		panic(ErrAddressExists)
	}

	v := Valoper{
		Moniker:      moniker,
		Description:  description,
		Address:      newAddress,
		PubKey:       pubKey,
		Active:       active,
		P2PAddresses: splitP2PAddresses(p2pAddresses),
		Caller:       std.PreviousRealm().Address(),
	}

	if err := v.Validate(); err != nil {
		panic(err)
	}

	// Remove the old valoper info
	// in case the address changed
	if address != newAddress {
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
		panic(ErrOriginCaller)
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
		panic(ErrOriginCaller)
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
		panic(ErrSameAddress)
	}

	v := GetByAddr(address)

	// Check that the caller is the valoper
	if v.Caller != std.PreviousRealm().Address() {
		panic(ErrOriginCaller)
	}

	// Check that the address is valid
	if err := validateBech32(newAddress.String()); err != nil {
		panic(ErrInvalidAddress)
	}

	// Check that the new address wouldn't be
	// overwriting an existing one
	if isValoper(newAddress) {
		panic(ErrAddressExists)
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
		panic(ErrOriginCaller)
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
		panic(ErrOriginCaller)
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
		panic(ErrOriginCaller)
	}

	// Split the P2P addresses
	addresses := splitP2PAddresses(p2pAddresses)

	if len(addresses) == 0 {
		panic(ErrP2PAddressesMissing)
	}

	// Check addresses
	for _, addr := range addresses {
		if err := validateP2PAddress(addr); err != nil {
			panic(err)
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

// Render renders the current valoper set.
// "/r/gnoland/valopers" lists all valopers, paginated.
// "/r/gnoland/valopers:addr" shows the detail for the valoper with the addr.
func Render(fullPath string) string {
	path, _ := splitPathAndQuery(fullPath)
	if path == "" {
		return renderHome(fullPath)
	} else {
		if len(path) < 2 || path[:2] != "g1" {
			return "invalid address " + path
		}
		valoperRaw, exists := valopers.Get(path)
		if !exists {
			return "unknown address " + path
		}
		v := valoperRaw.(Valoper)
		return "Valoper's details:\n" + v.Render()
	}
}

func renderHome(path string) string {
	// if there are no valopers, display instructions
	if valopers.Size() == 0 {
		return ufmt.Sprintf("%s\n\nNo valopers to display.", instructions)
	}

	page := pager.NewPager(valopers, 50, false).MustGetPageByPath(path)

	output := ""

	// if we are on the first page, display instructions
	if page.PageNumber == 1 {
		output += ufmt.Sprintf("%s\n\n", instructions)
	}

	for _, item := range page.Items {
		v := item.Value.(Valoper)
		output += ufmt.Sprintf(" * [%s](/r/gnoland/valopers:%s) - [profile](/r/demo/profile:u/%s)\n",
			v.Moniker, v.Address, v.Caller)
	}

	output += "\n"
	output += page.Picker()
	return output
}

func splitPathAndQuery(fullPath string) (string, string) {
	parts := strings.SplitN(fullPath, "?", 2)
	path := parts[0]
	queryString := ""
	if len(parts) > 1 {
		queryString = "?" + parts[1]
	}
	return path, queryString
}

// Validate checks if the fields of the Valoper are valid
func (v *Valoper) Validate() error {
	errs := &combinederr.CombinedError{}

	errs.Add(validateMoniker(v.Moniker))
	errs.Add(validateDescription(v.Description))
	errs.Add(validateBech32(v.Address.String()))
	errs.Add(validatePubKey(v.PubKey))

	for _, addr := range v.P2PAddresses {
		errs.Add(validateP2PAddress(addr))
	}

	if errs.Size() == 0 {
		return nil
	}

	return errs
}

// Render renders a single valoper with their information
func (v Valoper) Render() string {
	output := ufmt.Sprintf("## %s\n", v.Moniker)

	if v.Description != "" {
		output += ufmt.Sprintf("%s\n\n", v.Description)
	}

	output += ufmt.Sprintf("- Address: %s\n", v.Address.String())
	output += ufmt.Sprintf("- PubKey: %s\n", v.PubKey)
	output += "- P2P Addresses: [\n\n"

	if len(v.P2PAddresses) == 0 {
		output += "]\n\n"
	}

	for index, addr := range v.P2PAddresses {
		output += "\t- " + addr

		if index == len(v.P2PAddresses)-1 {
			output += "\n\n]\n\n"

			continue
		}

		output += ",\n"
	}

	output += ufmt.Sprintf("[Profile link](/r/demo/profile:u/%s)\n", v.Address)

	return output
}

// isValoper checks if the valoper exists
func isValoper(address std.Address) bool {
	_, exists := valopers.Get(address.String())

	return exists
}

func validateSentCoins() error {
	sentCoins := std.OriginSend()
	minCoin := std.NewCoin("ugnot", minFee)

	if len(sentCoins) == 1 && sentCoins[0].IsGTE(minCoin) {
		if sentCoins[0].Amount > minFee*maxFeeMult {
			return errors.New("payment must not be greater than " + strconv.Itoa(int(minFee*maxFeeMult)))
		} else {
			return nil
		}
	} else {
		return errors.New("payment must not be less than " + strconv.Itoa(int(minFee)))
	}
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
		return err
	}

	return nil
}

func validateListeningAddress(addr string) bool {
	re := regexp.MustCompile(`^(?:\w+://)?[\w\.-]+:\d+$`)
	return re.MatchString(addr)
}

// validateP2PAddress checks if the P2P address is valid
// address is of the form: <ID>@<HOSTNAME> or <ID>@<IP>
func validateP2PAddress(p2pAddress string) error {
	parts := strings.Split(p2pAddress, "@")

	if len(parts) != 2 {
		return ErrInvalidP2PAddress
	}

	// test part 1 which is the ID
	if err := validateBech32(parts[0]); err != nil {
		return ErrInvalidP2PAddress
	}

	if !validateListeningAddress(parts[1]) {
		return ErrInvalidP2PAddress
	}

	return nil
}
