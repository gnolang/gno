package users

import (
	"chain"
	"chain/runtime"
	"regexp"

	"gno.land/p/nt/avl/v0"
	"gno.land/p/nt/ufmt/v0"
)

var (
	nameStore    = avl.NewTree() // name/aliases > *UserData
	addressStore = avl.NewTree() // address > *UserData

	reAddressLookalike = regexp.MustCompile(`^g1[a-z0-9]{20,38}$`)
	reAlphanum         = regexp.MustCompile(`^[a-zA-Z0-9_]{1,64}$`)
)

const (
	RegisterUserEvent = "Registered"
	UpdateNameEvent   = "Updated"
	DeleteUserEvent   = "Deleted"
)

type UserData struct {
	addr     address
	username string // contains the latest name of a user
	deleted  bool
}

func (u UserData) Name() string {
	return u.username
}

func (u UserData) Addr() address {
	return u.addr
}

func (u UserData) IsDeleted() bool {
	return u.deleted
}

// RenderLink provides a render link to the user page on gnoweb
// `linkText` is optional
func (u UserData) RenderLink(linkText string) string {
	if linkText == "" {
		return ufmt.Sprintf("[@%s](/u/%s)", u.username, u.username)
	}

	return ufmt.Sprintf("[%s](/u/%s)", linkText, u.username)
}

// RegisterUser adds a new user to the system.
func RegisterUser(cur realm, name string, address_XXX address) error {
	// Validate caller
	if !controllers.Has(runtime.PreviousRealm().Address()) {
		return NewErrNotWhitelisted()
	}

	// Validate name
	if err := validateName(name); err != nil {
		return err
	}

	// Validate address
	if !address_XXX.IsValid() {
		return ErrInvalidAddress
	}

	// Check if name is taken
	if nameStore.Has(name) {
		return ErrNameTaken
	}

	raw, ok := addressStore.Get(address_XXX.String())
	if ok {
		// Cannot re-register after deletion
		if raw.(*UserData).IsDeleted() {
			return ErrDeletedUser
		}

		// For a second name, use UpdateName
		return ErrAlreadyHasName
	}

	// Create UserData
	data := &UserData{
		addr:     address_XXX,
		username: name,
		deleted:  false,
	}

	// Set corresponding stores
	nameStore.Set(name, data)
	addressStore.Set(address_XXX.String(), data)

	chain.Emit(RegisterUserEvent,
		"name", name,
		"address", address_XXX.String(),
	)
	return nil
}

// UpdateName adds a name that is associated with a specific address
// All previous names are preserved and resolvable.
// The new name is the default value returned for address lookups.
func (u *UserData) UpdateName(newName string) error {
	if u == nil { // either doesnt exists or was deleted
		return ErrUserNotExistOrDeleted
	}

	// Validate caller
	if !controllers.Has(runtime.CurrentRealm().Address()) {
		return NewErrNotWhitelisted()
	}

	// Validate name
	if err := validateName(newName); err != nil {
		return err
	}

	// Check if the requested Alias is already taken
	if nameStore.Has(newName) {
		return ErrNameTaken
	}

	u.username = newName
	nameStore.Set(newName, u)

	chain.Emit(UpdateNameEvent,
		"alias", newName,
		"address", u.addr.String(),
	)
	return nil
}

// Delete marks a user and all their aliases as deleted.
func (u *UserData) Delete() error {
	if u == nil {
		return ErrUserNotExistOrDeleted
	}

	// Validate caller
	if !controllers.Has(runtime.CurrentRealm().Address()) {
		return NewErrNotWhitelisted()
	}

	u.deleted = true

	chain.Emit(DeleteUserEvent, "address", u.addr.String())
	return nil
}

// Validate validates username and address passed in
// Most of the validation is done in the controllers
// This provides more flexibility down the line
func validateName(username string) error {
	if username == "" {
		return ErrEmptyUsername
	}

	if !reAlphanum.MatchString(username) {
		return ErrInvalidUsername
	}

	// Check if the username can be decoded or looks like a valid address
	if address(username).IsValid() || reAddressLookalike.MatchString(username) {
		return ErrNameLikeAddress
	}

	return nil
}
