// Package valopers is designed around the permissionless lifecycle of valoper profiles.
package valopers

import (
	"chain"
	"chain/banker"
	"crypto/bech32"
	"errors"
	"regexp"

	"gno.land/p/moul/realmpath"
	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/avl/v0/pager"
	"gno.land/p/nt/combinederr/v0"
	"gno.land/p/nt/ownable/v0/exts/authorizable"
	"gno.land/p/nt/ufmt/v0"
)

const (
	MonikerMaxLength     = 32
	DescriptionMaxLength = 2048

	// Valid server types
	ServerTypeCloud      = "cloud"
	ServerTypeOnPrem     = "on-prem"
	ServerTypeDataCenter = "data-center"
)

var (
	ErrValoperExists      = errors.New("valoper already exists")
	ErrValoperMissing     = errors.New("valoper does not exist")
	ErrInvalidAddress     = errors.New("invalid address")
	ErrInvalidMoniker     = errors.New("moniker is not valid")
	ErrInvalidDescription = errors.New("description is not valid")
	ErrInvalidServerType  = errors.New("server type is not valid")
)

var (
	valopers     *avl.Tree                              // valopers keeps track of all the valoper profiles. Address -> Valoper
	instructions string                                 // markdown instructions for valoper's registration
	minFee       = chain.NewCoin("ugnot", 20*1_000_000) // minimum gnot must be paid to register.

	monikerMaxLengthMiddle = ufmt.Sprintf("%d", MonikerMaxLength-2)
	validateMonikerRe      = regexp.MustCompile(`^[a-zA-Z0-9][\w -]{0,` + monikerMaxLengthMiddle + `}[a-zA-Z0-9]$`) // 32 characters, including spaces, hyphens or underscores in the middle
)

// Valoper represents a validator operator profile
type Valoper struct {
	Moniker     string // A human-readable name
	Description string // A description and details about the valoper
	ServerType  string // The type of server (cloud/on-prem/data-center)

	Address     address // The bech32 gno address of the validator
	PubKey      string  // The bech32 public key of the validator
	KeepRunning bool    // Flag indicating if the owner wants to keep the validator running

	auth *authorizable.Authorizable // The authorizer system for the valoper
}

func (v Valoper) Auth() *authorizable.Authorizable {
	return v.auth
}

func AddToAuthList(cur realm, address_XXX address, member address) {
	v := GetByAddr(address_XXX)
	if err := v.Auth().AddToAuthListByPrevious(member); err != nil {
		panic(err)
	}
}

func DeleteFromAuthList(cur realm, address_XXX address, member address) {
	v := GetByAddr(address_XXX)
	if err := v.Auth().DeleteFromAuthListByPrevious(member); err != nil {
		panic(err)
	}
}

// Register registers a new valoper
func Register(cur realm, moniker string, description string, serverType string, address_XXX address, pubKey string) {
	// Check if a fee is enforced
	if !minFee.IsZero() {
		sentCoins := banker.OriginSend()

		// Coins must be sent and cover the min fee
		if len(sentCoins) != 1 || sentCoins[0].IsLT(minFee) {
			panic(ufmt.Sprintf("payment must not be less than %d%s", minFee.Amount, minFee.Denom))
		}
	}

	// Check if the valoper is already registered
	if isValoper(address_XXX) {
		panic(ErrValoperExists)
	}

	v := Valoper{
		Moniker:     moniker,
		Description: description,
		ServerType:  serverType,
		Address:     address_XXX,
		PubKey:      pubKey,
		KeepRunning: true,
		auth:        authorizable.NewAuthorizableWithOrigin(),
	}

	if err := v.Validate(); err != nil {
		panic(err)
	}

	// TODO add address derivation from public key
	// (when the laws of gno make it possible)

	// Save the valoper to the set
	valopers.Set(v.Address.String(), v)
}

// UpdateMoniker updates an existing valoper's moniker
func UpdateMoniker(cur realm, address_XXX address, moniker string) {
	// Check that the moniker is not empty
	if err := validateMoniker(moniker); err != nil {
		panic(err)
	}

	v := GetByAddr(address_XXX)

	// Check that the caller has permissions
	v.Auth().AssertPreviousOnAuthList()

	// Update the moniker
	v.Moniker = moniker

	// Save the valoper info
	valopers.Set(address_XXX.String(), v)
}

// UpdateDescription updates an existing valoper's description
func UpdateDescription(cur realm, address_XXX address, description string) {
	// Check that the description is not empty
	if err := validateDescription(description); err != nil {
		panic(err)
	}

	v := GetByAddr(address_XXX)

	// Check that the caller has permissions
	v.Auth().AssertPreviousOnAuthList()

	// Update the description
	v.Description = description

	// Save the valoper info
	valopers.Set(address_XXX.String(), v)
}

// UpdateKeepRunning updates an existing valoper's active status
func UpdateKeepRunning(cur realm, address_XXX address, keepRunning bool) {
	v := GetByAddr(address_XXX)

	// Check that the caller has permissions
	v.Auth().AssertPreviousOnAuthList()

	// Update status
	v.KeepRunning = keepRunning

	// Save the valoper info
	valopers.Set(address_XXX.String(), v)
}

// UpdateServerType updates an existing valoper's server type
func UpdateServerType(cur realm, address_XXX address, serverType string) {
	// Check that the server type is valid
	if err := validateServerType(serverType); err != nil {
		panic(err)
	}

	v := GetByAddr(address_XXX)

	// Check that the caller has permissions
	v.Auth().AssertPreviousOnAuthList()

	// Update server type
	v.ServerType = serverType

	// Save the valoper info
	valopers.Set(address_XXX.String(), v)
}

// GetByAddr fetches the valoper using the address, if present
func GetByAddr(address_XXX address) Valoper {
	valoperRaw, exists := valopers.Get(address_XXX.String())
	if !exists {
		panic(ErrValoperMissing)
	}

	return valoperRaw.(Valoper)
}

// Render renders the current valoper set.
// "/r/gnops/valopers" lists all valopers, paginated.
// "/r/gnops/valopers:addr" shows the detail for the valoper with the addr.
func Render(fullPath string) string {
	req := realmpath.Parse(fullPath)
	if req.Path == "" {
		return renderHome(fullPath)
	} else {
		addr := req.Path
		if len(addr) < 2 || addr[:2] != "g1" {
			return "invalid address " + addr
		}
		valoperRaw, exists := valopers.Get(addr)
		if !exists {
			return "unknown address " + addr
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
		output += ufmt.Sprintf(" * [%s](/r/gnops/valopers:%s) - [profile](/r/demo/profile:u/%s)\n",
			v.Moniker, v.Address, v.Address)
	}

	output += "\n"
	output += page.Picker(path)
	return output
}

// Validate checks if the fields of the Valoper are valid
func (v *Valoper) Validate() error {
	errs := &combinederr.CombinedError{}

	errs.Add(validateMoniker(v.Moniker))
	errs.Add(validateDescription(v.Description))
	errs.Add(validateServerType(v.ServerType))
	errs.Add(validateBech32(v.Address))
	errs.Add(validatePubKey(v.PubKey))

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
	output += ufmt.Sprintf("- Server Type: %s\n\n", v.ServerType)
	output += ufmt.Sprintf("[Profile link](/r/demo/profile:u/%s)\n", v.Address)

	return output
}

// isValoper checks if the valoper exists
func isValoper(address_XXX address) bool {
	_, exists := valopers.Get(address_XXX.String())

	return exists
}

// validateMoniker checks if the moniker is valid
func validateMoniker(moniker string) error {
	if moniker == "" {
		return ErrInvalidMoniker
	}

	if len(moniker) > MonikerMaxLength {
		return ErrInvalidMoniker
	}

	if !validateMonikerRe.MatchString(moniker) {
		return ErrInvalidMoniker
	}

	return nil
}

// validateDescription checks if the description is valid
func validateDescription(description string) error {
	if description == "" {
		return ErrInvalidDescription
	}

	if len(description) > DescriptionMaxLength {
		return ErrInvalidDescription
	}

	return nil
}

// validateBech32 checks if the value is a valid bech32 address
func validateBech32(address_XXX address) error {
	if !address_XXX.IsValid() {
		return ErrInvalidAddress
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

// validateServerType checks if the server type is valid
func validateServerType(serverType string) error {
	if serverType != ServerTypeCloud &&
		serverType != ServerTypeOnPrem &&
		serverType != ServerTypeDataCenter {
		return ErrInvalidServerType
	}

	return nil
}
