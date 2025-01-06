package users

import (
	"regexp"
	"std"
	"strings"

	"gno.land/p/demo/avl"
	"gno.land/p/demo/pausable"
)

var (
	callerWhitelist = avl.NewTree() // address > struct{}{}

	nameStore    = avl.NewTree() // name/aliases > *UserData
	addressStore = avl.NewTree() // address > *UserData - always points to the latest Alias

	reAddressLookalike = regexp.MustCompile(`^g1[a-z0-9]{20,38}$`)
)

const (
	RegisterUserEvent  = "UserRegistered"
	RegisterAliasEvent = "AliasRegistered"
	DeleteUserEvent    = "UserDeleted"

	maxUsernameLen = 64
)

type UserData struct {
	Addr     std.Address
	Username string
	Alias    *UserData
	Deleted  bool
}

func (u UserData) IsDeleted() bool {
	// No need to resolve, any UserData this is called upon will already be resolved
	return u.Deleted
}

// RegisterUser adds a new user to the system.
func RegisterUser(name string, address std.Address) error {
	// Validate inputs
	if err := validateInputs(name, address); err != nil {
		return err
	}

	if err := validateCall(); err != nil {
		return err
	}

	// Check if name is taken
	if nameStore.Has(name) {
		return ErrNameTaken
	}

	raw, ok := addressStore.Get(address.String())
	if ok {
		// Cannot re-register after deletion
		if raw.(*UserData).IsDeleted() {
			return ErrDeletedUser
		}

		// For a second name, user RegisterAlias
		return ErrAlreadyHasName
	}

	// Create UserData
	data := &UserData{
		Addr:     address,
		Username: name,
		Alias:    nil,
		Deleted:  false,
	}

	// Set corresponding stores
	nameStore.Set(name, data)
	addressStore.Set(address.String(), data)

	std.Emit(RegisterUserEvent, "name", name, "address", address.String())
	return nil
}

// RegisterAlias adds an Alias pointing to an existing user.
func RegisterAlias(alias string, address std.Address) error {
	// Validate inputs
	if err := validateInputs(alias, address); err != nil {
		return err
	}

	if err := validateCall(); err != nil {
		return err
	}

	// Check if requested Alias is already taken
	if nameStore.Has(alias) {
		return ErrNameTaken
	}

	// Check if user has a name before an Alias
	raw, ok := addressStore.Get(address.String())
	if !ok {
		return ErrAliasBeforeName
	}

	if raw.(*UserData).IsDeleted() {
		return ErrDeletedUser
	}

	// Create new Alias
	newData := &UserData{
		Addr:     address,
		Username: alias,
		Alias:    nil,
		Deleted:  false,
	}

	oldData := raw.(*UserData)
	// Point to new Alias
	oldData.Alias = newData
	oldData.Addr = ""

	// Set latest Alias to stores
	addressStore.Set(address.String(), newData)
	nameStore.Set(alias, newData)

	std.Emit(RegisterAliasEvent, "Alias", alias, "address", address.String())
	return nil
}

// Delete marks a user and all their aliases as Deleted.
func Delete(addr std.Address) error {
	if err := validateCall(); err != nil {
		return err
	}

	data, deleted := ResolveAddress(addr)
	if data == nil {
		return ErrUserNotExist
	}

	if deleted {
		return ErrAlreadyDeleted
	}

	data.Deleted = true
	// data.Addr = "" // I don't think we should do this, it will likely cause user errors
	// data.Username = "ghost" // could cause issues with users named "ghost"

	std.Emit(DeleteUserEvent, "address", data.Addr.String())
	return nil
}

func validateCall() error {
	if !IsOnWhitelist(std.PrevRealm().Addr()) {
		return ErrNotWhitelisted
	}

	if Pausable.IsPaused() {
		return pausable.ErrPaused
	}

	return nil
}

// Validate validates Username and address passed in
func validateInputs(username string, address std.Address) error {
	if !address.IsValid() {
		return ErrInvalidAddress
	}

	if strings.TrimSpace(username) == "" {
		return ErrEmptyUsername
	}

	if len(username) > maxUsernameLen {
		return ErrNameTooLong
	}

	// Check if the Username can be decoded or looks like a valid address
	if _, _, ok := std.DecodeBech32(std.Address(username)); ok || reAddressLookalike.MatchString(username) {
		return ErrNameLikeAddress
	}

	return nil
}
