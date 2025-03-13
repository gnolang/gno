// Package valopers is designed around the permissionless lifecycle of valoper profiles.
package valopers

import (
	"crypto/bech32"
	"errors"
	"regexp"
	"std"
	"strconv"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/avl/pager"
	"gno.land/p/demo/combinederr"
	"gno.land/p/demo/ownable/exts/authorizable"
	"gno.land/p/demo/ufmt"
	"gno.land/p/moul/realmpath"
)

const (
	MonikerMaxLength     = 32
	DescriptionMaxLength = 2048
)

var (
	ErrValoperExists       = errors.New("valoper already exists")
	ErrValoperMissing      = errors.New("valoper does not exist")
	ErrAddressExists       = errors.New("valoper updated address exists")
	ErrSameAddress         = errors.New("valoper updated address is the same")
	ErrInvalidAddress      = errors.New("invalid address")
	ErrOriginCaller        = errors.New("origin call is not the same")
	ErrInvalidMoniker      = errors.New("moniker is not valid")
	ErrInvalidDescription  = errors.New("description is not valid")
	ErrInvalidPubKey       = errors.New("invalid public key")
)

var (
	valopers     *avl.Tree                  // valopers keeps track of all the valoper profiles. Address -> Valoper
	instructions string                     // markdown instructions for valoper's registration
	minFee       int64     = 20 * 1_000_000 // minimum gnot must be paid to register.
	minCoin                = std.NewCoin("ugnot", minFee)

	monikerMaxLengthMiddle     = ufmt.Sprintf("%d", MonikerMaxLength-2)
	validateMonikerRe          = regexp.MustCompile(`^[a-zA-Z0-9][\w -]{0,` + monikerMaxLengthMiddle + `}[a-zA-Z0-9]$`) // 32 characters, including spaces, hyphens or underscores in the middle
	validateListeningAddressRe = regexp.MustCompile(`^(?:\w+://)?[\w\.-]+:\d+$`)                                        // <protocol>://<hostname>:<port> where <protocol> is optional
)

// Valoper represents a validator operator profile
type Valoper struct {
	Moniker     string // A human-readable name
	Description string // A description and details about the valoper

	Address      std.Address // The bech32 gno address of the validator
	PubKey       string      // The bech32 public key of the validator
	KeepRunning  bool        // Flag indicating if the owner wants to keep the validator running

	auth *authorizable.Authorizable // The authorizer system for the valoper
}

func (v Valoper) Auth() *authorizable.Authorizable {
	return v.auth
}

func AddToAuthList(address std.Address, member std.Address) {
	v := GetByAddr(address)

	if err := v.Auth().AddToAuthList(member); err != nil {
		panic(err)
	}
}

func DeleteFromAuthList(address std.Address, member std.Address) {
	v := GetByAddr(address)

	if err := v.Auth().DeleteFromAuthList(member); err != nil {
		panic(err)
	}
}

// Register registers a new valoper
func Register(moniker string, description string, address std.Address, pubKey string) {
	if err := validateSentCoins(); err != nil {
		panic(err)
	}

	// Check if the valoper is already registered
	if isValoper(address) {
		panic(ErrValoperExists)
	}

	v := Valoper{
		Moniker:      moniker,
		Description:  description,
		Address:      address,
		PubKey:       pubKey,
		KeepRunning:  true,
		auth:         authorizable.NewAuthorizable(),
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
func UpdateMoniker(address std.Address, moniker string) {
	// Check that the moniker is not empty
	if err := validateMoniker(moniker); err != nil {
		panic(err)
	}

	v := GetByAddr(address)

	// Check that the caller has permissions
	v.Auth().AssertOnAuthList()

	// Update the moniker
	v.Moniker = moniker

	// Save the valoper info
	valopers.Set(address.String(), v)
}

// UpdateDescription updates an existing valoper's description
func UpdateDescription(address std.Address, description string) {
	// Check that the description is not empty
	if err := validateDescription(description); err != nil {
		panic(err)
	}

	v := GetByAddr(address)

	// Check that the caller has permissions
	v.Auth().AssertOnAuthList()

	// Update the description
	v.Description = description

	// Save the valoper info
	valopers.Set(address.String(), v)
}

// UpdateKeepRunning updates an existing valoper's active status
func UpdateKeepRunning(address std.Address, keepRunning bool) {
	v := GetByAddr(address)

	// Check that the caller has permissions
	v.Auth().AssertOnAuthList()

	// Update status
	v.KeepRunning = keepRunning

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
		output += ufmt.Sprintf(" * [%s](/r/gnoland/valopers:%s) - [profile](/r/demo/profile:u/%s)\n",
			v.Moniker, v.Address, v.Auth().Owner())
	}

	output += "\n"
	output += page.Picker()
	return output
}

// Validate checks if the fields of the Valoper are valid
func (v *Valoper) Validate() error {
	errs := &combinederr.CombinedError{}

	errs.Add(validateMoniker(v.Moniker))
	errs.Add(validateDescription(v.Description))
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
	output += ufmt.Sprintf("- PubKey: %s\n\n", v.PubKey)
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

	if len(sentCoins) == 1 && sentCoins[0].IsGTE(minCoin) {
		return nil
	}
	return errors.New("payment must not be less than " + strconv.Itoa(int(minFee)))
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
func validateBech32(address std.Address) error {
	if !std.Address.IsValid(address) {
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
